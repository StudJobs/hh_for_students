// Package cleaner — фоновый воркер автоматического удаления старых
// closed-вакансий и completed-микрозадач, согласно настройкам компании
// (Company.CleanupVacanciesAfterDays / Company.CleanupTasksAfterDays).
//
// Запускается в Gateway, потому что здесь уже есть gRPC-клиенты ко всем трём
// сервисам (Company, Vacancy, MicroTasks). Альтернатива — добавить клиент
// Company в Vacancy/MicroTasks и крутить локальный цикл, но дублирование
// dial-кода в двух сервисах хуже, чем централизованный цикл.
//
// Каждая итерация делает не более одного round-trip GetAllCompanies + по
// одному List+Delete на компанию с активной policy. Нагрузка на бэк — раз
// в N часов; для MVP допустимо. При росте — переехать на dedicated job.
package cleaner

import (
	"context"
	"log"
	"time"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
)

const (
	// position_status вакансий, которые подходят под чистку.
	statusClosed = "closed"
	// MicroTask status: 3 = COMPLETED.
	microtaskCompleted int32 = 3
	// Лимиты на странице при сборе списков — для MVP хватает.
	listLimit int32 = 200
)

type Cleaner struct {
	svc      *services.ApiGateway
	interval time.Duration
}

func New(svc *services.ApiGateway, interval time.Duration) *Cleaner {
	return &Cleaner{svc: svc, interval: interval}
}

// Run запускает цикл (блокирующий — вызывать в goroutine). interval <= 0 — отключено.
func (c *Cleaner) Run(ctx context.Context) {
	if c.interval <= 0 {
		log.Printf("cleaner: disabled (interval=%v)", c.interval)
		return
	}
	log.Printf("cleaner: starting loop, interval=%v", c.interval)

	// Первый прогон с задержкой, чтобы не сталкиваться с остальной инициализацией.
	first := time.NewTimer(2 * time.Minute)
	select {
	case <-ctx.Done():
		return
	case <-first.C:
	}
	c.runOnce(ctx)

	t := time.NewTicker(c.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("cleaner: stopping")
			return
		case <-t.C:
			c.runOnce(ctx)
		}
	}
}

func (c *Cleaner) runOnce(ctx context.Context) {
	pag := &models.Pagination{Page: 1, Limit: 1000}
	companies, err := c.svc.Company.GetAllCompanies(ctx, pag, "", "", "")
	if err != nil || companies == nil {
		log.Printf("cleaner: GetAllCompanies failed: %v", err)
		return
	}
	for _, comp := range companies.Companies {
		if comp == nil {
			continue
		}
		c.cleanupVacancies(ctx, comp)
		c.cleanupMicrotasks(ctx, comp)
	}
}

func (c *Cleaner) cleanupVacancies(ctx context.Context, comp *models.Company) {
	days := comp.CleanupVacanciesAfterDays
	if days <= 0 {
		return
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	pag := &models.Pagination{Page: 1, Limit: listLimit}
	list, err := c.svc.Vacancy.GetAllVacancies(ctx, pag, comp.ID, statusClosed, "", "", 0, 0, 0, 0, "")
	if err != nil || list == nil {
		log.Printf("cleaner: vacancy list failed for company=%s: %v", comp.ID, err)
		return
	}
	deleted := 0
	for _, v := range list.Vacancies {
		if v == nil {
			continue
		}
		// CreateAt — RFC3339 string из proto. Если parsed older than cutoff — soft-delete.
		t, ok := parseTime(v.CreateAt)
		if !ok {
			continue
		}
		if t.After(cutoff) {
			continue
		}
		if err := c.svc.Vacancy.DeleteVacancy(ctx, v.ID); err != nil {
			log.Printf("cleaner: delete vacancy %s failed: %v", v.ID, err)
			continue
		}
		deleted++
	}
	if deleted > 0 {
		log.Printf("cleaner: company=%s — soft-deleted %d closed vacancies older than %d days", comp.ID, deleted, days)
	}
}

func (c *Cleaner) cleanupMicrotasks(ctx context.Context, comp *models.Company) {
	days := comp.CleanupTasksAfterDays
	if days <= 0 {
		return
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	list, err := c.svc.MicroTasks.ListByCompany(ctx, comp.ID, 1, listLimit)
	if err != nil || list == nil {
		log.Printf("cleaner: microtasks list failed for company=%s: %v", comp.ID, err)
		return
	}
	deleted := 0
	for _, t := range list.Tasks {
		if t == nil {
			continue
		}
		if t.Status != microtaskCompleted {
			continue
		}
		tCreated, ok := parseTime(t.CreatedAt)
		if !ok {
			continue
		}
		if tCreated.After(cutoff) {
			continue
		}
		if err := c.svc.MicroTasks.Delete(ctx, t.ID); err != nil {
			log.Printf("cleaner: delete microtask %s failed: %v", t.ID, err)
			continue
		}
		deleted++
	}
	if deleted > 0 {
		log.Printf("cleaner: company=%s — soft-deleted %d completed microtasks older than %d days", comp.ID, deleted, days)
	}
}

// parseTime пытается распарсить RFC3339 / common date formats из бэка.
func parseTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
