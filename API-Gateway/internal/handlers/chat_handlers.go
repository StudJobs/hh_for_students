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
		// Для MVP разрешаем любому залогиненному пользователю; полная проверка
		// участников отклика требует Application.Get RPC (отложено).
		_ = rid
		return userID != "", ""
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
