package reindexer

import (
	"context"
	"fmt"
	"log"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"

	"github.com/studjobs/hh_for_students/search/internal/clients"
	"github.com/studjobs/hh_for_students/search/internal/esclient"
	"github.com/studjobs/hh_for_students/search/internal/indexer"
)

const reindexBatchSize = 100

type Reindexer struct {
	es      *esclient.Client
	idx     *indexer.Indexer
	clients *clients.Clients
}

func New(es *esclient.Client, idx *indexer.Indexer, c *clients.Clients) *Reindexer {
	return &Reindexer{es: es, idx: idx, clients: c}
}

// EnsureIndices гарантирует существование индексов. recreate=true пересоздаёт их.
func (r *Reindexer) EnsureIndices(ctx context.Context, recreate bool) error {
	if err := r.es.EnsureIndex(ctx, esclient.IndexProfiles, esclient.ProfilesMapping, recreate); err != nil {
		return fmt.Errorf("reindexer: profiles: %w", err)
	}
	if err := r.es.EnsureIndex(ctx, esclient.IndexVacancies, esclient.VacanciesMapping, recreate); err != nil {
		return fmt.Errorf("reindexer: vacancies: %w", err)
	}
	if err := r.es.EnsureIndex(ctx, esclient.IndexMicroTasks, esclient.MicroTasksMapping, recreate); err != nil {
		return fmt.Errorf("reindexer: microtasks: %w", err)
	}
	return nil
}

// Run перечитывает все профили, вакансии и микрозадачи из соответствующих сервисов и индексирует их.
func (r *Reindexer) Run(ctx context.Context, recreate bool) (profiles, vacancies, microtasks int32, err error) {
	if err := r.EnsureIndices(ctx, recreate); err != nil {
		return 0, 0, 0, err
	}

	profiles, err = r.reindexProfiles(ctx)
	if err != nil {
		return profiles, 0, 0, fmt.Errorf("reindex profiles: %w", err)
	}

	vacancies, err = r.reindexVacancies(ctx)
	if err != nil {
		return profiles, vacancies, 0, fmt.Errorf("reindex vacancies: %w", err)
	}

	microtasks, err = r.reindexMicroTasks(ctx)
	if err != nil {
		return profiles, vacancies, microtasks, fmt.Errorf("reindex microtasks: %w", err)
	}

	log.Printf("reindex done: profiles=%d vacancies=%d microtasks=%d", profiles, vacancies, microtasks)
	return profiles, vacancies, microtasks, nil
}

func (r *Reindexer) reindexProfiles(ctx context.Context) (int32, error) {
	var total int32
	page := int32(1)
	for {
		resp, err := r.clients.Users.GetAllProfiles(ctx, &usersv1.GetAllProfilesRequest{
			Pagination: &commonv1.Pagination{Page: page, Limit: reindexBatchSize},
		})
		if err != nil {
			return total, fmt.Errorf("users.GetAllProfiles page=%d: %w", page, err)
		}
		profiles := resp.GetProfiles()
		if len(profiles) == 0 {
			return total, nil
		}
		for _, p := range profiles {
			if err := r.idx.IndexProfile(ctx, p); err != nil {
				log.Printf("reindexer: skip profile %s: %v", p.GetId(), err)
				continue
			}
			total++
		}
		if resp.GetPagination() != nil && page >= resp.GetPagination().GetPages() {
			return total, nil
		}
		page++
	}
}

func (r *Reindexer) reindexMicroTasks(ctx context.Context) (int32, error) {
	if r.clients.MicroTasks == nil {
		log.Printf("reindexer: microtasks client not configured, skipping")
		return 0, nil
	}
	var total int32
	page := int32(1)
	for {
		resp, err := r.clients.MicroTasks.List(ctx, &microtaskv1.ListMicroTasksRequest{
			Pagination: &commonv1.Pagination{Page: page, Limit: reindexBatchSize},
		})
		if err != nil {
			return total, fmt.Errorf("microtasks.List page=%d: %w", page, err)
		}
		tasks := resp.GetTasks()
		if len(tasks) == 0 {
			return total, nil
		}
		for _, t := range tasks {
			if err := r.idx.IndexMicroTask(ctx, t); err != nil {
				log.Printf("reindexer: skip microtask %s: %v", t.GetId(), err)
				continue
			}
			total++
		}
		if resp.GetPagination() != nil && page >= resp.GetPagination().GetPages() {
			return total, nil
		}
		page++
	}
}

func (r *Reindexer) reindexVacancies(ctx context.Context) (int32, error) {
	var total int32
	page := int32(1)
	for {
		resp, err := r.clients.Vacancy.GetAllVacancies(ctx, &vacancyv1.GetAllVacanciesRequest{
			Pagination: &commonv1.Pagination{Page: page, Limit: reindexBatchSize},
		})
		if err != nil {
			return total, fmt.Errorf("vacancy.GetAllVacancies page=%d: %w", page, err)
		}
		vacs := resp.GetVacancies()
		if len(vacs) == 0 {
			return total, nil
		}
		for _, v := range vacs {
			if err := r.idx.IndexVacancy(ctx, v); err != nil {
				log.Printf("reindexer: skip vacancy %s: %v", v.GetId(), err)
				continue
			}
			total++
		}
		if resp.GetPagination() != nil && page >= resp.GetPagination().GetPages() {
			return total, nil
		}
		page++
	}
}
