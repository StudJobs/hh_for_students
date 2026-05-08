package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

// GetTasks — публичный листинг микрозадач для студентов.
// При наличии skill_slugs или q маршрут идёт через Search (Elasticsearch),
// иначе через прямой вызов MicroTasks.List.
func (h *Handler) GetTasks(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	page, limit = normalizePagination(page, limit)

	skillSlugs := splitCSV(c.Query("skill_slugs", ""))
	query := c.Query("q", "")
	rewardMin, _ := strconv.Atoi(c.Query("reward_min", "0"))
	statusInt, _ := strconv.Atoi(c.Query("status", "0"))

	var (
		list *models.MicroTaskList
		err  error
	)

	if h.apiService.Search.Available() && (len(skillSlugs) > 0 || query != "" || rewardMin > 0) {
		log.Printf("GetTasks: routing through Search (skill_slugs=%v q=%q reward_min=%d)", skillSlugs, query, rewardMin)
		list, err = h.apiService.Search.SearchMicroTasksAsModel(c.Context(), query, skillSlugs, int32(rewardMin), int32(statusInt), "", int32(page), int32(limit))
	} else {
		list, err = h.apiService.MicroTasks.List(c.Context(), int32(statusInt), skillSlugs, int32(page), int32(limit))
	}
	if err != nil {
		log.Printf("GetTasks: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get tasks"})
	}
	return c.JSON(list)
}

// GetTask — детали одной задачи.
func (h *Handler) GetTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task id"})
	}
	t, err := h.apiService.MicroTasks.Get(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
	}
	return c.JSON(t)
}

// ApplyToTask — студент берёт задачу. assigned_to = текущий user.
func (h *Handler) ApplyToTask(c *fiber.Ctx) error {
	id := c.Params("id")
	studentID := getUserIDFromContext(c)
	if id == "" || studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	t, err := h.apiService.MicroTasks.Apply(c.Context(), id, studentID)
	if err != nil {
		log.Printf("ApplyToTask: failed task=%s student=%s: %v", id, studentID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(t)
}

// SubmitTask — студент отправляет решение.
func (h *Handler) SubmitTask(c *fiber.Ctx) error {
	id := c.Params("id")
	studentID := getUserIDFromContext(c)
	if id == "" || studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	var req models.SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.SolutionURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "solution_url is required"})
	}
	s, err := h.apiService.MicroTasks.Submit(c.Context(), id, studentID, req.SolutionURL, req.Comment)
	if err != nil {
		log.Printf("SubmitTask: failed task=%s student=%s: %v", id, studentID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(s)
}

// ListMySubmissions — студент видит свои отправленные решения.
func (h *Handler) ListMySubmissions(c *fiber.Ctx) error {
	studentID := getUserIDFromContext(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	page, limit = normalizePagination(page, limit)
	list, err := h.apiService.MicroTasks.ListSubmissions(c.Context(), "", studentID, int32(page), int32(limit))
	if err != nil {
		log.Printf("ListMySubmissions: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load submissions"})
	}
	return c.JSON(list)
}

// === HR-операции ===

// GetHRTasks — список задач компании текущего HR. company_id = user_id (одна компания на HR).
func (h *Handler) GetHRTasks(c *fiber.Ctx) error {
	companyID := getUserIDFromContext(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	page, limit = normalizePagination(page, limit)
	list, err := h.apiService.MicroTasks.ListByCompany(c.Context(), companyID, int32(page), int32(limit))
	if err != nil {
		log.Printf("GetHRTasks: failed company=%s: %v", companyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load tasks"})
	}
	return c.JSON(list)
}

// CreateHRTask — HR создаёт новую задачу. company_id берётся из контекста.
func (h *Handler) CreateHRTask(c *fiber.Ctx) error {
	companyID := getUserIDFromContext(c)
	var req models.MicroTaskCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
	}
	t := &models.MicroTask{
		CompanyID:   companyID,
		Title:       req.Title,
		Description: req.Description,
		Reward:      req.Reward,
		Deadline:    req.Deadline,
		SkillSlugs:  req.SkillSlugs,
	}
	created, err := h.apiService.MicroTasks.Create(c.Context(), t)
	if err != nil {
		log.Printf("CreateHRTask: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create task"})
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// UpdateHRTask — HR редактирует свою задачу.
func (h *Handler) UpdateHRTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task id"})
	}
	var req models.MicroTaskUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	t := &models.MicroTask{}
	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Reward != nil {
		t.Reward = *req.Reward
	}
	if req.Deadline != nil {
		t.Deadline = *req.Deadline
	}
	if req.SkillSlugs != nil {
		t.SkillSlugs = req.SkillSlugs
	}
	updated, err := h.apiService.MicroTasks.Update(c.Context(), id, t)
	if err != nil {
		log.Printf("UpdateHRTask: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update task"})
	}
	return c.JSON(updated)
}

// DeleteHRTask — HR удаляет (soft) свою задачу.
func (h *Handler) DeleteHRTask(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task id"})
	}
	if err := h.apiService.MicroTasks.Delete(c.Context(), id); err != nil {
		log.Printf("DeleteHRTask: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete task"})
	}
	return c.JSON(fiber.Map{"message": "Task deleted"})
}

// ListTaskSubmissions — HR видит submission'ы по конкретной задаче (для ревью).
func (h *Handler) ListTaskSubmissions(c *fiber.Ctx) error {
	id := c.Params("id")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	page, limit = normalizePagination(page, limit)
	list, err := h.apiService.MicroTasks.ListSubmissions(c.Context(), id, "", int32(page), int32(limit))
	if err != nil {
		log.Printf("ListTaskSubmissions: failed task=%s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load submissions"})
	}
	return c.JSON(list)
}

// ReviewSubmission — HR апрувит/реджектит submission. При APPROVE задача переходит в COMPLETED.
func (h *Handler) ReviewSubmission(c *fiber.Ctx) error {
	subID := c.Params("submission_id")
	if subID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid submission id"})
	}
	var req models.ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Status != 2 && req.Status != 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be 2 (APPROVED) or 3 (REJECTED)"})
	}
	s, err := h.apiService.MicroTasks.Review(c.Context(), subID, req.Status, req.ReviewComment)
	if err != nil {
		log.Printf("ReviewSubmission: failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to review submission"})
	}
	return c.JSON(s)
}
