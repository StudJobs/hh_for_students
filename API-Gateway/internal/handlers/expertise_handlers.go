package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// GetExpertiseTest — отдаёт набор вопросов для теста по навыку.
func (h *Handler) GetExpertiseTest(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "slug is required"})
	}
	t, err := h.apiService.User.GetExpertiseTest(c.Context(), slug)
	if err != nil {
		log.Printf("GetExpertiseTest slug=%s failed: %v", slug, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	// Конвертация proto → JSON.
	type questionDTO struct {
		ID      int32    `json:"id"`
		Text    string   `json:"text"`
		Options []string `json:"options"`
	}
	qs := make([]questionDTO, 0, len(t.GetQuestions()))
	for _, q := range t.GetQuestions() {
		qs = append(qs, questionDTO{ID: q.GetId(), Text: q.GetText(), Options: q.GetOptions()})
	}
	return c.JSON(fiber.Map{
		"skill_slug":         t.GetSkillSlug(),
		"available":          t.GetAvailable(),
		"reason":             t.GetReason(),
		"questions":          qs,
		"pass_threshold_pct": t.GetPassThresholdPct(),
	})
}

// SubmitExpertiseTest — принимает ответы, возвращает результат.
type submitTestBody struct {
	AnswerIndices []int32 `json:"answer_indices"`
}

func (h *Handler) SubmitExpertiseTest(c *fiber.Ctx) error {
	slug := c.Params("slug")
	userID := getUserIDFromContext(c)
	if slug == "" || userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	var body submitTestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	resp, err := h.apiService.User.SubmitExpertiseTest(c.Context(), userID, slug, body.AnswerIndices)
	if err != nil {
		log.Printf("SubmitExpertiseTest user=%s slug=%s failed: %v", userID, slug, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"passed":    resp.GetPassed(),
		"correct":   resp.GetCorrect(),
		"total":     resp.GetTotal(),
		"score_pct": resp.GetScorePct(),
		"message":   resp.GetMessage(),
	})
}
