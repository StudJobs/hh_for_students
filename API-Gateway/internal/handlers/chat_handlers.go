package handlers

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

// canAccessThread проверяет, имеет ли пользователь право читать/писать в треде.
//   - application/<id>  → доступ у автора-студента и у HR (упрощённо, см. ниже).
//   - task/<id>         → доступ у assigned студента и у владельца задачи.
//   - quest/<id>        → доступ у target_student_id и у эксперта-создателя.
//   - direct/<idA>_<idB> → 1:1 чат; rid — это пара UUID, отсортированных по
//     возрастанию (фронт строит детерминированно). Тред создаётся «на лету» при
//     первом сообщении; ListUserThreads потом вернёт его обеим сторонам.
// Возвращает (ok, errMessage). На gRPC-уровне используется один объединённый
// thread_id `<kind>:<rid>` — это контракт сервиса чата, никак не влияет на URL.
func (h *Handler) canAccessThread(ctx context.Context, userID string, userRole Role, kind, rid string) (bool, string) {
	if rid == "" {
		return false, "invalid resource id"
	}
	switch kind {
	case "direct":
		// rid: "<uuidA>_<uuidB>". Доступ — если userID совпадает с любым из них.
		parts := strings.Split(rid, "_")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return false, "invalid direct id"
		}
		if parts[0] == userID || parts[1] == userID {
			return true, ""
		}
		return false, "not a participant"
	case "application":
		// Участники треда отклика:
		//   - студент-автор: только он сам.
		//   - HR/COMPANY/EXPERT/DEVELOPER: разрешаем — в MVP не дёргаем строгую
		//     проверку membership (UX-блокеры были болезненны). Если HR APPROVED
		//     и ещё не assignee — auto-assign'имся.
		app, err := h.apiService.Application.Get(ctx, rid)
		if err != nil {
			return false, "application not found"
		}
		// Студент: только автор отклика.
		if userRole == ROLE_STUDENT {
			if app.StudentID == userID {
				return true, ""
			}
			return false, "not a participant"
		}
		// HR/COMPANY/EXPERT/DEVELOPER → пускаем. Дополнительно auto-assign HR в
		// approved-компании этой вакансии, чтобы тред был корректно атрибутирован.
		if userRole == ROLE_HR {
			if v, vErr := h.apiService.Vacancy.GetVacancy(ctx, app.VacancyID); vErr == nil && v != nil {
				if ms, mErr := h.apiService.Company.GetMembershipByUser(ctx, userID); mErr == nil && ms != nil && ms.CompanyID == v.CompanyID && ms.Status == 2 {
					_, _ = h.apiService.Application.AssignHR(ctx, rid, userID)
				}
			}
		}
		return true, ""
	case "task":
		t, err := h.apiService.MicroTasks.Get(ctx, rid)
		if err != nil {
			return false, "task not found"
		}
		if t.AssignedTo == userID || t.CompanyID == userID {
			return true, ""
		}
		return false, "not a participant"
	case "quest":
		t, err := h.apiService.MicroTasks.Get(ctx, rid)
		if err != nil {
			return false, "quest not found"
		}
		if !t.IsSkillQuest {
			return false, "not a quest"
		}
		if t.TargetStudentID == userID || t.CompanyID == userID {
			return true, ""
		}
		return false, "not a participant"
	default:
		return false, "unknown thread kind"
	}
}

func threadIDFromParams(c *fiber.Ctx) (kind, rid, joined string) {
	kind = c.Params("kind")
	rid = c.Params("rid")
	if kind == "" || rid == "" {
		return "", "", ""
	}
	return kind, rid, kind + ":" + rid
}

// GetChatThreads — inbox юзера. Собирает треды из бизнес-источников
// (application / task / quest), а не из chat_messages — это даёт получателю
// видеть тред даже если он ещё ни разу не отвечал.
//
// По ролям:
//   - STUDENT: свои applications, свои task/quest (assigned_to=me).
//   - HR (EMPLOYER): applications и task'и компании (через membership).
//   - COMPANY_OWNER: applications своих вакансий + task своей компании.
//   - EXPERT: квесты, которые он создал (company_id=expert_id).
func (h *Handler) GetChatThreads(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	role := getRoleFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	ctx := c.Context()

	// Скрытые юзером треды (анти-зачистка): после сбора будем фильтровать.
	hidden := make(map[string]bool)
	if hids, err := h.apiService.Chat.ListHiddenThreads(ctx, userID); err == nil {
		for _, tid := range hids {
			hidden[tid] = true
		}
	}

	threads := make([]*models.ChatThread, 0, 32)
	added := make(map[string]bool)
	add := func(t *models.ChatThread) {
		if t == nil || t.ThreadID == "" || added[t.ThreadID] {
			return
		}
		if hidden[t.ThreadID] {
			return
		}
		added[t.ThreadID] = true
		threads = append(threads, t)
	}

	// Определяем company-id для HR/OWNER. У OWNER company_id == userID.
	companyID := ""
	if role == ROLE_COMPANY {
		companyID = userID
	} else if role == ROLE_HR {
		if ms, err := h.apiService.Company.GetMembershipByUser(ctx, userID); err == nil && ms != nil && ms.Status == 2 {
			companyID = ms.CompanyID
		}
	}

	// 1) STUDENT (и DEVELOPER) — свои отклики, свои микрозадачи/квесты.
	if role == ROLE_STUDENT || role == ROLE_DEVELOPER {
		if list, err := h.apiService.Application.ListMine(ctx, userID, 0, 1, 100); err == nil && list != nil {
			for _, a := range list.Applications {
				add(&models.ChatThread{
					ThreadID:     "application:" + a.ID,
					Kind:         "application",
					ResourceID:   a.ID,
					PeerID:       a.HRAssigneeID,
					ContextTitle: "Отклик",
				})
			}
		}
		if mine, err := h.apiService.MicroTasks.ListByStudent(ctx, userID, 0, 1, 100); err == nil && mine != nil {
			for _, t := range mine.Tasks {
				kind := "task"
				if t.IsSkillQuest {
					kind = "quest"
				}
				add(&models.ChatThread{
					ThreadID:     kind + ":" + t.ID,
					Kind:         kind,
					ResourceID:   t.ID,
					PeerID:       t.CompanyID,
					ContextTitle: t.Title,
				})
			}
		}
	}

	// 2) HR / OWNER — отклики на вакансии компании + микрозадачи компании.
	if companyID != "" {
		if vacList, err := h.apiService.Vacancy.GetHRVacancies(ctx, &models.Pagination{Page: 1, Limit: 100}, companyID, "", "", "", 0, 0, 0, 0, ""); err == nil && vacList != nil {
			for _, v := range vacList.Vacancies {
				if alist, err := h.apiService.Application.ListForVacancy(ctx, v.ID, 0, 1, 100); err == nil && alist != nil {
					for _, a := range alist.Applications {
						add(&models.ChatThread{
							ThreadID:     "application:" + a.ID,
							Kind:         "application",
							ResourceID:   a.ID,
							PeerID:       a.StudentID,
							ContextTitle: "Отклик на «" + v.Title + "»",
						})
					}
				}
			}
		}
		if tasks, err := h.apiService.MicroTasks.ListByCompany(ctx, companyID, 1, 100); err == nil && tasks != nil {
			for _, t := range tasks.Tasks {
				if t.AssignedTo == "" {
					continue
				}
				kind := "task"
				if t.IsSkillQuest {
					kind = "quest"
				}
				add(&models.ChatThread{
					ThreadID:     kind + ":" + t.ID,
					Kind:         kind,
					ResourceID:   t.ID,
					PeerID:       t.AssignedTo,
					ContextTitle: t.Title,
				})
			}
		}
	}

	// 3) EXPERT — его квесты (company_id == expert_id).
	if role == ROLE_EXPERT {
		if tasks, err := h.apiService.MicroTasks.ListByCompany(ctx, userID, 1, 100); err == nil && tasks != nil {
			for _, t := range tasks.Tasks {
				if !t.IsSkillQuest || t.TargetStudentID == "" {
					continue
				}
				add(&models.ChatThread{
					ThreadID:     "quest:" + t.ID,
					Kind:         "quest",
					ResourceID:   t.ID,
					PeerID:       t.TargetStudentID,
					ContextTitle: t.Title,
				})
			}
		}
	}

	// 4) Fallback: треды, где юзер уже писал хотя бы раз. Это страхует случаи,
	//    которые не покрылись бизнес-источниками (HR без membership, expert как
	//    co-reviewer и т.д.). Если запись уже есть в added — пропускается.
	if extra, err := h.apiService.Chat.ListUserThreads(ctx, userID, 100); err == nil {
		for _, t := range extra {
			if t.ThreadID == "" || added[t.ThreadID] {
				continue
			}
			parts := strings.SplitN(t.ThreadID, ":", 2)
			if len(parts) != 2 {
				continue
			}
			kind, rid := parts[0], parts[1]
			thread := &models.ChatThread{
				ThreadID:    t.ThreadID,
				Kind:        kind,
				ResourceID:  rid,
				LastMessage: t.LastMessage,
				LastAt:      t.LastAt,
			}
			switch kind {
			case "task", "quest":
				if task, err := h.apiService.MicroTasks.Get(ctx, rid); err == nil && task != nil {
					thread.ContextTitle = task.Title
					if task.AssignedTo != "" && task.AssignedTo != userID {
						thread.PeerID = task.AssignedTo
					} else {
						thread.PeerID = task.CompanyID
					}
				}
			case "application":
				if a, err := h.apiService.Application.Get(ctx, rid); err == nil && a != nil {
					thread.ContextTitle = "Отклик"
					if a.StudentID != userID {
						thread.PeerID = a.StudentID
					} else if a.HRAssigneeID != "" {
						thread.PeerID = a.HRAssigneeID
					}
				}
			case "direct":
				// rid = "<uuidA>_<uuidB>" — peer == та сторона, что не я.
				parts := strings.SplitN(rid, "_", 2)
				if len(parts) == 2 {
					if parts[0] == userID {
						thread.PeerID = parts[1]
					} else if parts[1] == userID {
						thread.PeerID = parts[0]
					}
				}
				thread.ContextTitle = "Личный чат"
			}
			add(thread)
		}
	}

	// Подтягиваем last_message + автора по каждому треду (N+1 — для inbox ≤ 100 ок).
	// Запрашиваем всегда, не только для пустых — нужно знать from_user_id, чтобы
	// фронт мог отличать «своё» сообщение от «чужого» (для уведомлений).
	//
	// ListByThread в репо отдаёт ORDER BY created_at ASC. Берём последний из
	// первой страницы как «свежайший». 50 — потолок, для активных тредов хватает;
	// уведомления приходят не позже чем после следующего тика.
	for _, t := range threads {
		if msgs, err := h.apiService.Chat.ListMessages(ctx, t.ThreadID, 1, 50); err == nil && msgs != nil && len(msgs.Messages) > 0 {
			last := msgs.Messages[len(msgs.Messages)-1]
			t.LastMessage = last.Body
			t.LastAt = last.CreatedAt
			t.LastFromUserID = last.FromUserID
		}
	}

	// Имя/роль собеседника.
	for _, t := range threads {
		if t.PeerID == "" {
			continue
		}
		p, err := h.apiService.User.GetUser(ctx, t.PeerID)
		if err == nil && p != nil {
			name := strings.TrimSpace(p.FirstName + " " + p.LastName)
			if name == "" {
				name = p.Email
			}
			t.PeerName = name
			t.PeerRole = roleHumanLabel(p.Role)
		}
	}

	return c.JSON(fiber.Map{"threads": threads})
}

func roleHumanLabel(r string) string {
	switch r {
	case "ROLE_STUDENT":
		return "Студент"
	case "ROLE_EMPLOYER":
		return "HR"
	case "ROLE_COMPANY_OWNER":
		return "Владелец компании"
	case "ROLE_EXPERT":
		return "Эксперт"
	}
	return r
}

func (h *Handler) GetChatMessages(c *fiber.Ctx) error {
	kind, rid, threadID := threadIDFromParams(c)
	userID := getUserIDFromContext(c)
	if threadID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if ok, why := h.canAccessThread(c.Context(), userID, getRoleFromContext(c), kind, rid); !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": why})
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	page, limit = normalizePagination(page, limit)
	list, err := h.apiService.Chat.ListMessages(c.Context(), threadID, int32(page), int32(limit))
	if err != nil {
		log.Printf("GetChatMessages: thread=%s failed: %v", threadID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load messages"})
	}
	return c.JSON(list)
}

// EditChatMessage — редактирование своего сообщения.
func (h *Handler) EditChatMessage(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	msgID := c.Params("msg_id")
	if userID == "" || msgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	var req models.ChatSendRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body is required"})
	}
	m, err := h.apiService.Chat.EditMessage(c.Context(), msgID, userID, strings.TrimSpace(req.Body))
	if err != nil {
		log.Printf("EditChatMessage failed user=%s msg=%s: %v", userID, msgID, err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "не ваше сообщение или не существует"})
	}
	return c.JSON(m)
}

// HideChatThread — скрыть тред со своей стороны (мягкое удаление, не трогает БД).
func (h *Handler) HideChatThread(c *fiber.Ctx) error {
	_, _, threadID := threadIDFromParams(c)
	userID := getUserIDFromContext(c)
	if threadID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := h.apiService.Chat.HideThread(c.Context(), userID, threadID); err != nil {
		log.Printf("HideChatThread failed user=%s thread=%s: %v", userID, threadID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hide"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) SendChatMessage(c *fiber.Ctx) error {
	kind, rid, threadID := threadIDFromParams(c)
	userID := getUserIDFromContext(c)
	if threadID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if ok, why := h.canAccessThread(c.Context(), userID, getRoleFromContext(c), kind, rid); !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": why})
	}
	var req models.ChatSendRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body is required"})
	}
	m, err := h.apiService.Chat.SendMessage(c.Context(), threadID, userID, strings.TrimSpace(req.Body))
	if err != nil {
		log.Printf("SendChatMessage: thread=%s user=%s failed: %v", threadID, userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to send"})
	}
	return c.Status(fiber.StatusCreated).JSON(m)
}
