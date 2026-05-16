package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ApplyMembership — HR подаёт заявку быть сотрудником компании company_id.
func (h *Handler) ApplyMembership(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	// route: /company/:id/membership/apply — параметр "id" по имени Fiber.
	companyID := c.Params("id")
	if userID == "" || companyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&body)
	m, err := h.apiService.Company.ApplyMembership(c.Context(), companyID, userID, body.Note)
	if err != nil {
		log.Printf("ApplyMembership user=%s company=%s failed: %v", userID, companyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(m)
}

// MyMembership — текущему HR возвращает его membership (или 404).
func (h *Handler) MyMembership(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	m, err := h.apiService.Company.GetMembershipByUser(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no membership"})
	}
	return c.JSON(m)
}

// ListMyCompanyMembers — owner видит сотрудников своей компании (с фильтром по статусу).
func (h *Handler) ListMyCompanyMembers(c *fiber.Ctx) error {
	ownerID := getUserIDFromContext(c)
	if ownerID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	// owner.userID == owner.companyID по соглашению Company-сервиса.
	statusStr := c.Query("status", "0")
	st, _ := strconv.Atoi(statusStr)
	list, err := h.apiService.Company.ListMembers(c.Context(), ownerID, int32(st))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"members": list})
}

// ReviewMembership — owner approve/reject membership.
func (h *Handler) ReviewMembership(c *fiber.Ctx) error {
	ownerID := getUserIDFromContext(c)
	membershipID := c.Params("membership_id")
	if ownerID == "" || membershipID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	var body struct {
		Status int32 `json:"status"` // 2 = APPROVED, 3 = REJECTED
	}
	if err := c.BodyParser(&body); err != nil || (body.Status != 2 && body.Status != 3) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be 2 or 3"})
	}
	m, err := h.apiService.Company.ReviewMembership(c.Context(), membershipID, body.Status)
	if err != nil {
		log.Printf("ReviewMembership failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(m)
}

// ModerateVacancy — owner approve/reject вакансии своего HR.
func (h *Handler) ModerateVacancy(c *fiber.Ctx) error {
	ownerID := getUserIDFromContext(c)
	id := c.Params("id")
	if ownerID == "" || id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	// Проверим что вакансия в компании owner-а.
	v, err := h.apiService.Vacancy.GetVacancy(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vacancy not found"})
	}
	if v.CompanyID != ownerID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not your company"})
	}
	var body struct {
		Status  int32  `json:"status"` // 2=PUBLISHED, 3=REJECTED
		Comment string `json:"comment,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil || (body.Status != 2 && body.Status != 3) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be 2 or 3"})
	}
	out, err := h.apiService.Vacancy.ModerateVacancy(c.Context(), id, body.Status, body.Comment)
	if err != nil {
		log.Printf("ModerateVacancy failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}
