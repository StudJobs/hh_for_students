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
// Возвращает (ok, errMessage). На gRPC-уровне используется один объединённый
// thread_id `<kind>:<rid>` — это контракт сервиса чата, никак не влияет на URL.
func (h *Handler) canAccessThread(ctx context.Context, userID, kind, rid string) (bool, string) {
	if rid == "" {
		return false, "invalid resource id"
	}
	switch kind {
	case "application":
		// Участники треда отклика: студент-автор + HR-assignee + owner-компании-вакансии.
		app, err := h.apiService.Application.Get(ctx, rid)
		if err != nil {
			return false, "application not found"
		}
		// Студент-автор отклика — пускаем.
		if app.StudentID == userID {
			return true, ""
		}
		// HR-assignee (если задан) — пускаем.
		if app.HRAssigneeID != "" && app.HRAssigneeID == userID {
			return true, ""
		}
		// Owner компании — пускаем (он видит всё). Достаём company_id из вакансии.
		v, vErr := h.apiService.Vacancy.GetVacancy(ctx, app.VacancyID)
		if vErr == nil && v != nil && v.CompanyID == userID {
			return true, ""
		}
		// HR ещё не assignee, но является сотрудником компании этой вакансии — auto-assign + пускаем.
		if v != nil {
			ms, mErr := h.apiService.Company.GetMembershipByUser(ctx, userID)
			if mErr == nil && ms != nil && ms.CompanyID == v.CompanyID {
				_, _ = h.apiService.Application.AssignHR(ctx, rid, userID)
				return true, ""
			}
		}
		return false, "not a participant"
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

// GetChatThreads — inbox юзера. Обогащает тред метаданными собеседника и контекста
// (название вакансии / задачи / навыка квеста). Делаем дополнительные RPC-вызовы на чтение,
// что для inbox-страницы приемлемо (N тредов, обычно <50).
func (h *Handler) GetChatThreads(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	threads, err := h.apiService.Chat.ListUserThreads(c.Context(), userID, 100)
	if err != nil {
		log.Printf("GetChatThreads user=%s failed: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load threads"})
	}

	// Обогащаем: для каждого треда определяем собеседника и контекст.
	for _, t := range threads {
		parts := strings.SplitN(t.ThreadID, ":", 2)
		if len(parts) != 2 {
			continue
		}
		t.Kind = parts[0]
		t.ResourceID = parts[1]

		switch t.Kind {
		case "task", "quest":
			task, err := h.apiService.MicroTasks.Get(c.Context(), t.ResourceID)
			if err != nil {
				continue
			}
			t.ContextTitle = task.Title
			// Собеседник = тот участник, кто НЕ я.
			peerID := task.AssignedTo
			if peerID == userID {
				peerID = task.CompanyID
			}
			t.PeerID = peerID
		case "application":
			// applications.Get RPC отсутствует — собеседника берём по первому сообщению (other_party).
			// ContextTitle подгружается через первый last_message preview.
			t.ContextTitle = "Отклик"
			// Возьмём сообщения треда, найдём first author != me.
			msgs, _ := h.apiService.Chat.ListMessages(c.Context(), t.ThreadID, 1, 100)
			if msgs != nil {
				for _, m := range msgs.Messages {
					if m.FromUserID != userID {
						t.PeerID = m.FromUserID
						break
					}
				}
			}
		}

		// Имя/роль/аватар собеседника подгружаем из Users.
		if t.PeerID != "" {
			p, err := h.apiService.User.GetUser(c.Context(), t.PeerID)
			if err == nil && p != nil {
				name := strings.TrimSpace(p.FirstName + " " + p.LastName)
				if name == "" {
					name = p.Email
				}
				t.PeerName = name
				t.PeerRole = roleHumanLabel(p.Role)
			}
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
	if ok, why := h.canAccessThread(c.Context(), userID, kind, rid); !ok {
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

func (h *Handler) SendChatMessage(c *fiber.Ctx) error {
	kind, rid, threadID := threadIDFromParams(c)
	userID := getUserIDFromContext(c)
	if threadID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if ok, why := h.canAccessThread(c.Context(), userID, kind, rid); !ok {
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
