package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

// RespondToVacancy — студент создаёт отклик на вакансию.
// POST /api/v1/vacancy/:id/respond
//
// Идемпотентность гарантируется на уровне БД (UNIQUE на (vacancy_id, student_id)
// для активных записей): повторный отклик вернёт уже существующий, не создаст дубликат.
func (h *Handler) RespondToVacancy(c *fiber.Ctx) error {
	if !h.apiService.Application.Available() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Error{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Applications service is not configured",
		})
	}

	vacancyID := c.Params("id")
	studentID := getUserIDFromContext(c)
	if studentID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Error{
			Code:    "UNAUTHORIZED",
			Message: "Cannot determine current user",
		})
	}

	var req models.ApplyRequest
	if len(c.Body()) > 0 {
		// Тело опционально (cover_letter может отсутствовать).
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Code:    "INVALID_BODY",
				Message: "Invalid JSON body",
			})
		}
	}

	app, err := h.apiService.Application.Apply(c.Context(), vacancyID, studentID, req.CoverLetter)
	if err != nil {
		log.Printf("RespondToVacancy: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create application",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(app)
}

// ListMyApplications — список откликов текущего студента.
// GET /api/v1/user/applications?page=&limit=&status=
func (h *Handler) ListMyApplications(c *fiber.Ctx) error {
	if !h.apiService.Application.Available() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Error{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Applications service is not configured",
		})
	}
	studentID := getUserIDFromContext(c)
	if studentID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Error{
			Code:    "UNAUTHORIZED",
			Message: "Cannot determine current user",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	page, limit = normalizePagination(page, limit)
	status, _ := strconv.Atoi(c.Query("status", "0"))

	list, err := h.apiService.Application.ListMine(c.Context(), studentID, int32(status), int32(page), int32(limit))
	if err != nil {
		log.Printf("ListMyApplications: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list applications",
		})
	}
	return c.JSON(list)
}

// WithdrawApplication — студент отзывает свой отклик.
// DELETE /api/v1/user/applications/:id
func (h *Handler) WithdrawApplication(c *fiber.Ctx) error {
	if !h.apiService.Application.Available() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Error{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Applications service is not configured",
		})
	}
	studentID := getUserIDFromContext(c)
	if studentID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(models.Error{
			Code:    "UNAUTHORIZED",
			Message: "Cannot determine current user",
		})
	}

	id := c.Params("id")
	if err := h.apiService.Application.Withdraw(c.Context(), id, studentID); err != nil {
		log.Printf("WithdrawApplication: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to withdraw application",
		})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListVacancyApplications — HR видит отклики на конкретную вакансию.
// GET /api/v1/hr/vacancy/:id/applications?page=&limit=&status=
func (h *Handler) ListVacancyApplications(c *fiber.Ctx) error {
	if !h.apiService.Application.Available() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Error{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Applications service is not configured",
		})
	}
	vacancyID := c.Params("id")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	page, limit = normalizePagination(page, limit)
	status, _ := strconv.Atoi(c.Query("status", "0"))

	list, err := h.apiService.Application.ListForVacancy(c.Context(), vacancyID, int32(status), int32(page), int32(limit))
	if err != nil {
		log.Printf("ListVacancyApplications: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list applications",
		})
	}
	return c.JSON(list)
}

// ReviewApplication — HR принимает решение по отклику.
// PATCH /api/v1/hr/applications/:id  body: { decision: 2|3, comment?: string }
func (h *Handler) ReviewApplication(c *fiber.Ctx) error {
	if !h.apiService.Application.Available() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Error{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Applications service is not configured",
		})
	}
	id := c.Params("id")

	var req models.ApplicationReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_BODY",
			Message: "Invalid JSON body",
		})
	}
	if req.Decision != 2 && req.Decision != 3 {
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_DECISION",
			Message: "decision must be 2 (ACCEPT) or 3 (REJECT)",
		})
	}

	app, err := h.apiService.Application.UpdateStatus(c.Context(), id, req.Decision, req.Comment)
	if err != nil {
		log.Printf("ReviewApplication: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to review application",
		})
	}
	return c.JSON(app)
}

