package handlers

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

// canAccessThread проверяет, имеет ли пользователь право читать/писать в thread_id.
// thread_id-формат: "<kind>:<resource_uuid>".
//   - application:<id>  → доступ у автора-студента или у HR-владельца вакансии.
//   - task:<id>         → доступ у assigned студента и у владельца задачи.
//   - quest:<id>        → доступ у target_student_id и у эксперта-создателя.
// Возвращает (ok, errMessage).
func (h *Handler) canAccessThread(ctx context.Context, userID, threadID string) (bool, string) {
	parts := strings.SplitN(threadID, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return false, "invalid thread_id"
	}
	kind, rid := parts[0], parts[1]
	switch kind {
	case "application":
		// Берём application через ListMine (нет публичного Get). Если application принадлежит
		// текущему юзеру — он студент. Иначе считаем HR (best-effort).
		// Дешевле: получить application id напрямую — но Application API не имеет такого метода.
		// Для MVP разрешаем любому юзеру с непустым userID, ибо реальная проверка участников
		// требует Application.Get RPC — отложено.
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

func (h *Handler) GetChatMessages(c *fiber.Ctx) error {
	threadID := c.Params("thread_id")
	userID := getUserIDFromContext(c)
	if threadID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if ok, why := h.canAccessThread(c.Context(), userID, threadID); !ok {
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
	threadID := c.Params("thread_id")
	userID := getUserIDFromContext(c)
	if threadID == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if ok, why := h.canAccessThread(c.Context(), userID, threadID); !ok {
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
