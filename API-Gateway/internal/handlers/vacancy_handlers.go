package handlers

import (
	"context"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

// GetVacancies возвращает список вакансий с пагинацией и фильтрами
// @Summary Получить список вакансий
// @Description Возвращает список вакансий с поддержкой пагинации и фильтрации
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(10) minimum(1) maximum(100)
// @Param company_id query string false "Фильтр по ID компании"
// @Param position_status query string false "Фильтр по статусу позиции"
// @Param work_format query string false "Фильтр по формату работы"
// @Param schedule query string false "Фильтр по графику работы"
// @Param min_salary query int false "Минимальная зарплата"
// @Param max_salary query int false "Максимальная зарплата"
// @Param min_experience query int false "Минимальный опыт работы (лет)"
// @Param max_experience query int false "Максимальный опыт работы (лет)"
// @Param search_title query string false "Поиск по названию вакансии"
// @Success 200 {object} models.VacancyList "Список вакансий"
// @Failure 400 {object} models.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /vacancy [get]
func (h *Handler) GetVacancies(c *fiber.Ctx) error {
	log.Printf("GetVacancies: Getting vacancies list")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	companyID := c.Query("company_id", "")
	positionStatus := c.Query("position_status", "")
	workFormat := c.Query("work_format", "")
	schedule := c.Query("schedule", "")
	minSalary, _ := strconv.Atoi(c.Query("min_salary", "0"))
	maxSalary, _ := strconv.Atoi(c.Query("max_salary", "0"))
	minExperience, _ := strconv.Atoi(c.Query("min_experience", "0"))
	maxExperience, _ := strconv.Atoi(c.Query("max_experience", "0"))
	searchTitle := c.Query("search_title", "")

	pagination := &models.Pagination{
		Page:  int32(page),
		Limit: int32(limit),
	}

	vacancies, err := h.apiService.Vacancy.GetAllVacancies(c.Context(), pagination,
		companyID, positionStatus, workFormat, schedule,
		int32(minSalary), int32(maxSalary), int32(minExperience), int32(maxExperience),
		searchTitle)
	if err != nil {
		log.Printf("GetVacancies: Failed to get vacancies: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get vacancies",
		})
	}

	h.enrichVacancyListWithFiles(c.Context(), vacancies.Vacancies)

	log.Printf("GetVacancies: Successfully retrieved %d vacancies with filters", len(vacancies.Vacancies))
	return c.JSON(vacancies)
}

// GetVacancy возвращает вакансию по ID
// @Summary Получить вакансию по ID
// @Description Возвращает информацию о вакансии по указанному идентификатору
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID вакансии" format(uuid)
// @Success 200 {object} models.Vacancy "Информация о вакансии"
// @Failure 400 {object} models.ErrorResponse "Неверный ID вакансии"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Вакансия не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /vacancy/{id} [get]
func (h *Handler) GetVacancy(c *fiber.Ctx) error {
	vacancyID := c.Params("id")
	log.Printf("GetVacancy: Getting vacancy with ID: %s", vacancyID)

	vacancy, err := h.apiService.Vacancy.GetVacancy(c.Context(), vacancyID)
	if err != nil {
		log.Printf("GetVacancy: Failed to get vacancy %s: %v", vacancyID, err)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "VACANCY_NOT_FOUND",
			Message: "Vacancy not found",
		})
	}

	h.enrichVacancyWithFiles(c.Context(), vacancy)

	log.Printf("GetVacancy: Successfully retrieved vacancy: %s", vacancyID)
	return c.JSON(vacancy)
}

// GetHRVacancies возвращает вакансии с фильтрами для HR
// @Summary Получить вакансии HR
// @Description Возвращает список вакансий для текущего HR с поддержкой фильтрации
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(10) minimum(1) maximum(100)
// @Param company_id query string false "Фильтр по ID компании"
// @Param position_status query string false "Фильтр по статусу позиции"
// @Param work_format query string false "Фильтр по формату работы"
// @Param schedule query string false "Фильтр по графику работы"
// @Param min_salary query int false "Минимальная зарплата"
// @Param max_salary query int false "Максимальная зарплата"
// @Param min_experience query int false "Минимальный опыт работы (лет)"
// @Param max_experience query int false "Максимальный опыт работы (лет)"
// @Param search_title query string false "Поиск по названию вакансии"
// @Success 200 {object} models.VacancyList "Список вакансий HR"
// @Failure 400 {object} models.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /hr/vacancy [get]
func (h *Handler) GetHRVacancies(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("GetHRVacancies: Getting vacancies for HR: %s", userID)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	companyID := c.Query("company_id", "")
	positionStatus := c.Query("position_status", "")
	workFormat := c.Query("work_format", "")
	schedule := c.Query("schedule", "")
	minSalary, _ := strconv.Atoi(c.Query("min_salary", "0"))
	maxSalary, _ := strconv.Atoi(c.Query("max_salary", "0"))
	minExperience, _ := strconv.Atoi(c.Query("min_experience", "0"))
	maxExperience, _ := strconv.Atoi(c.Query("max_experience", "0"))
	searchTitle := c.Query("search_title", "")

	pagination := &models.Pagination{
		Page:  int32(page),
		Limit: int32(limit),
	}

	vacancies, err := h.apiService.Vacancy.GetHRVacancies(c.Context(), pagination,
		companyID, positionStatus, workFormat, schedule,
		int32(minSalary), int32(maxSalary), int32(minExperience), int32(maxExperience),
		searchTitle)
	if err != nil {
		log.Printf("GetHRVacancies: Failed to get vacancies for HR %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get vacancies",
		})
	}

	h.enrichVacancyListWithFiles(c.Context(), vacancies.Vacancies)

	log.Printf("GetHRVacancies: Successfully retrieved %d vacancies for HR %s",
		len(vacancies.Vacancies), userID)
	return c.JSON(vacancies)
}

// CreateHRVacancy создает новую вакансию для HR
// @Summary Создать вакансию
// @Description Создает новую вакансию для текущего HR
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Vacancy true "Данные вакансии"
// @Success 201 {object} models.Vacancy "Созданная вакансия"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /hr/vacancy [post]
func (h *Handler) CreateHRVacancy(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("CreateHRVacancy: Creating new vacancy for HR: %s", userID)

	var req models.Vacancy
	if err := c.BodyParser(&req); err != nil {
		log.Printf("CreateHRVacancy: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	if req.Title == "" {
		log.Printf("CreateHRVacancy: Missing required field 'title'")
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "MISSING_FIELD",
			Message: "Vacancy title is required",
		})
	}

	if req.CompanyID == "" {
		log.Printf("CreateHRVacancy: Missing required field 'company_id'")
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "MISSING_FIELD",
			Message: "Company ID is required",
		})
	}

	vacancy, err := h.apiService.Vacancy.CreateVacancy(c.Context(), &req)
	if err != nil {
		log.Printf("CreateHRVacancy: Failed to create vacancy: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "CREATE_FAILED",
			Message: "Failed to create vacancy",
		})
	}

	h.enrichVacancyWithFiles(c.Context(), vacancy)

	log.Printf("CreateHRVacancy: Successfully created vacancy: %s for company: %s by HR: %s",
		vacancy.ID, vacancy.CompanyID, userID)
	return c.Status(fiber.StatusCreated).JSON(vacancy)
}

// UpdateVacancy обновляет вакансию
// @Summary Обновить вакансию
// @Description Обновляет данные вакансии текущего HR
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID вакансии"
// @Param request body models.Vacancy true "Данные для обновления"
// @Success 200 {object} models.Vacancy "Обновленная вакансия"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} models.ErrorResponse "Вакансия не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /hr/vacancy/{id} [patch]
func (h *Handler) UpdateVacancy(c *fiber.Ctx) error {
	vacancyID := c.Params("id")
	userID := getUserIDFromContext(c)
	log.Printf("UpdateVacancy: Updating vacancy %s by user %s", vacancyID, userID)

	var req models.Vacancy
	if err := c.BodyParser(&req); err != nil {
		log.Printf("UpdateVacancy: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	vacancy, err := h.apiService.Vacancy.UpdateVacancy(c.Context(), vacancyID, &req)
	if err != nil {
		log.Printf("UpdateVacancy: Failed to update vacancy %s: %v", vacancyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "UPDATE_FAILED",
			Message: "Failed to update vacancy",
		})
	}

	h.enrichVacancyWithFiles(c.Context(), vacancy)

	log.Printf("UpdateVacancy: Successfully updated vacancy: %s by user: %s", vacancyID, userID)
	return c.JSON(vacancy)
}

// DeleteVacancy удаляет вакансию
// @Summary Удалить вакансию
// @Description Удаляет вакансию текущего HR
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID вакансии"
// @Success 200 {object} map[string]string "Сообщение об успешном удалении"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} models.ErrorResponse "Вакансия не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /hr/vacancy/{id} [delete]
func (h *Handler) DeleteVacancy(c *fiber.Ctx) error {
	vacancyID := c.Params("id")
	userID := getUserIDFromContext(c)
	log.Printf("DeleteVacancy: Deleting vacancy %s by user %s", vacancyID, userID)

	vacancy, err := h.apiService.Vacancy.GetVacancy(c.Context(), vacancyID)
	if err != nil {
		log.Printf("DeleteVacancy: Failed to get vacancy %s: %v", vacancyID, err)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "VACANCY_NOT_FOUND",
			Message: "Vacancy not found",
		})
	}

	if vacancy.AttachmentID != nil && *vacancy.AttachmentID != "" {
		if err := h.fileHandler.DeleteFile(c.Context(), vacancyID, *vacancy.AttachmentID); err != nil {
			log.Printf("DeleteVacancy: Failed to delete attachment for vacancy %s: %v", vacancyID, err)
		}
	}

	if err := h.apiService.Vacancy.DeleteVacancy(c.Context(), vacancyID); err != nil {
		log.Printf("DeleteVacancy: Failed to delete vacancy %s: %v", vacancyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "DELETE_FAILED",
			Message: "Failed to delete vacancy",
		})
	}

	log.Printf("DeleteVacancy: Successfully deleted vacancy: %s by user: %s", vacancyID, userID)
	return c.JSON(fiber.Map{
		"message": "Vacancy deleted successfully",
	})
}

// GetPositions возвращает список всех доступных позиций
// @Summary Получить список позиций
// @Description Возвращает список всех доступных должностей/позиций
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Список позиций"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /positions [get]
func (h *Handler) GetPositions(c *fiber.Ctx) error {
	log.Printf("GetPositions: Getting all positions")

	positions, err := h.apiService.Vacancy.GetAllPositions(c.Context())
	if err != nil {
		log.Printf("GetPositions: Failed to get positions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get positions",
		})
	}

	log.Printf("GetPositions: Successfully retrieved %d positions", len(positions))
	return c.JSON(fiber.Map{
		"positions": positions,
	})
}

// UploadVacancyAttachment загружает прикрепленный файл для вакансии
// @Summary Загрузить вложение вакансии
// @Description Загружает прикрепленный файл для вакансии
// @Tags Vacancies
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID вакансии"
// @Param attachment formData file true "Файл вложения (макс. 10MB)"
// @Success 200 {object} models.FileUploadResponse "Информация о загруженном файле"
// @Failure 400 {object} models.ErrorResponse "Неверный запрос или файл слишком большой"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /vacancy/{id}/files/attachment [post]
func (h *Handler) UploadVacancyAttachment(c *fiber.Ctx) error {
	vacancyID := c.Params("id")
	log.Printf("UploadVacancyAttachment: Uploading attachment for vacancy: %s", vacancyID)

	file, err := c.FormFile("attachment")
	if err != nil {
		log.Printf("UploadVacancyAttachment: No attachment file provided for vacancy %s", vacancyID)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_FILE",
			Message: "No attachment file provided",
		})
	}

	if file.Size > 10*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "FILE_TOO_LARGE",
			Message: "File size too large. Maximum size is 10MB",
		})
	}

	fileInfo, err := h.fileHandler.UploadFileDirect(
		c.Context(),
		vacancyID,
		"vacancy",
		"attachment",
		file,
	)
	if err != nil {
		log.Printf("UploadVacancyAttachment: Failed to upload attachment for vacancy %s: %v", vacancyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "UPLOAD_FAILED",
			Message: "Failed to upload attachment",
		})
	}

	attachmentID := fileInfo.Name
	_, err = h.apiService.Vacancy.UpdateVacancy(c.Context(), vacancyID, &models.Vacancy{
		AttachmentID: &attachmentID,
	})
	if err != nil {
		log.Printf("UploadVacancyAttachment: Failed to update vacancy with attachment: %v", err)
	}

	log.Printf("UploadVacancyAttachment: Successfully uploaded attachment for vacancy: %s", vacancyID)
	return c.JSON(models.FileUploadResponse{
		FileInfo: fileInfo,
		Message:  "Vacancy attachment uploaded successfully",
	})
}

// DeleteVacancyAttachment удаляет прикрепленный файл вакансии
// @Summary Удалить вложение вакансии
// @Description Удаляет прикрепленный файл вакансии
// @Tags Vacancies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID вакансии"
// @Success 200 {object} map[string]string "Сообщение об успешном удалении"
// @Failure 400 {object} models.ErrorResponse "У вакансии нет вложения"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} models.ErrorResponse "Вакансия не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /vacancy/{id}/files/attachment [delete]
func (h *Handler) DeleteVacancyAttachment(c *fiber.Ctx) error {
	vacancyID := c.Params("id")
	log.Printf("DeleteVacancyAttachment: Deleting attachment for vacancy: %s", vacancyID)

	vacancy, err := h.apiService.Vacancy.GetVacancy(c.Context(), vacancyID)
	if err != nil {
		log.Printf("DeleteVacancyAttachment: Vacancy not found: %s", vacancyID)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "VACANCY_NOT_FOUND",
			Message: "Vacancy not found",
		})
	}

	if vacancy.AttachmentID == nil || *vacancy.AttachmentID == "" {
		log.Printf("DeleteVacancyAttachment: Vacancy %s doesn't have an attachment", vacancyID)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_ATTACHMENT",
			Message: "Vacancy doesn't have an attachment",
		})
	}

	err = h.fileHandler.DeleteFile(c.Context(), vacancyID, *vacancy.AttachmentID)
	if err != nil {
		log.Printf("DeleteVacancyAttachment: Failed to delete attachment file for vacancy %s: %v", vacancyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "DELETE_FAILED",
			Message: "Failed to delete attachment",
		})
	}

	emptyAttachmentID := ""
	_, err = h.apiService.Vacancy.UpdateVacancy(c.Context(), vacancyID, &models.Vacancy{
		AttachmentID: &emptyAttachmentID,
	})
	if err != nil {
		log.Printf("DeleteVacancyAttachment: Failed to update vacancy: %v", err)
	}

	log.Printf("DeleteVacancyAttachment: Successfully deleted attachment for vacancy: %s", vacancyID)
	return c.JSON(fiber.Map{
		"message": "Vacancy attachment deleted successfully",
	})
}

// enrichVacancyWithFiles обогащает вакансию информацией о файлах
func (h *Handler) enrichVacancyWithFiles(ctx context.Context, vacancy *models.Vacancy) {
	if vacancy.AttachmentID != nil && *vacancy.AttachmentID != "" {
		fileInfo, err := h.fileHandler.GetFileInfo(ctx, vacancy.ID, *vacancy.AttachmentID, "attachment")
		if err == nil && fileInfo != nil {
			vacancy.AttachmentURL = fileInfo.URL
		} else {
			log.Printf("enrichVacancyWithFiles: Failed to get attachment info for vacancy %s: %v", vacancy.ID, err)
		}
	}
}

// enrichVacancyListWithFiles обогащает список вакансий информацией о файлах
func (h *Handler) enrichVacancyListWithFiles(ctx context.Context, vacancies []*models.Vacancy) {
	for _, vacancy := range vacancies {
		h.enrichVacancyWithFiles(ctx, vacancy)
	}
}
