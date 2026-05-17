package handlers

import (
	"log"
	"strconv"
	"strings"

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
//
// Авторизация: HR может рассматривать только отклики на вакансии своей
// (APPROVED-membership) компании; COMPANY_OWNER — только на вакансии своей
// компании. Без этой проверки любой HR мог принимать/отклонять чужие отклики.
//
// После успешного review комментарий HR (если непустой) дублируется в чат
// треда `application:<id>` — кандидат увидит решение в «Сообщениях», иначе
// hr_comment висел только в карточке отклика и легко терялся.
func (h *Handler) ReviewApplication(c *fiber.Ctx) error {
	if !h.apiService.Application.Available() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Error{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Applications service is not configured",
		})
	}
	id := c.Params("id")
	userID := getUserIDFromContext(c)
	role := getRoleFromContext(c)

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

	// Авторизация: тянем отклик → его вакансию → company_id; сверяем с moей
	// membership (status=APPROVED, 2). DEVELOPER пропускаем — для smoke-сценариев.
	if role != ROLE_DEVELOPER {
		existing, err := h.apiService.Application.Get(c.Context(), id)
		if err != nil || existing == nil {
			return c.Status(fiber.StatusNotFound).JSON(models.Error{
				Code:    "NOT_FOUND",
				Message: "Application not found",
			})
		}
		vac, err := h.apiService.Vacancy.GetVacancy(c.Context(), existing.VacancyID)
		if err != nil || vac == nil {
			return c.Status(fiber.StatusNotFound).JSON(models.Error{
				Code:    "NOT_FOUND",
				Message: "Vacancy not found",
			})
		}
		allowed := false
		switch role {
		case ROLE_COMPANY:
			// owner.userID == owner.companyID по соглашению Company-сервиса.
			allowed = vac.CompanyID == userID
		case ROLE_HR:
			if ms, mErr := h.apiService.Company.GetMembershipByUser(c.Context(), userID); mErr == nil && ms != nil {
				allowed = ms.Status == 2 && ms.CompanyID == vac.CompanyID
			}
		}
		if !allowed {
			log.Printf("ReviewApplication: forbidden — user=%s role=%s app=%s vacCompany=%s", userID, role, id, vac.CompanyID)
			return c.Status(fiber.StatusForbidden).JSON(models.Error{
				Code:    "FORBIDDEN",
				Message: "Эта вакансия не относится к вашей компании",
			})
		}
	}

	app, err := h.apiService.Application.UpdateStatus(c.Context(), id, req.Decision, req.Comment)
	if err != nil {
		log.Printf("ReviewApplication: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to review application",
		})
	}

	// Комментарий HR — в чат треда. Без блокировки на ошибку: статус уже обновлён,
	// чат — best-effort UX (если упадёт — отклик всё равно закрыт корректно).
	if comment := strings.TrimSpace(req.Comment); comment != "" {
		threadID := "application:" + id
		body := comment
		if req.Decision == 2 {
			body = "✓ Отклик принят. " + comment
		} else if req.Decision == 3 {
			body = "✗ Отклик отклонён. " + comment
		}
		if _, sErr := h.apiService.Chat.SendMessage(c.Context(), threadID, userID, body); sErr != nil {
			log.Printf("ReviewApplication: chat message failed (non-fatal): %v", sErr)
		}
	}
	return c.JSON(app)
}

