package handlers

import (
	"context"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// GetVacancies возвращает список вакансий с пагинацией и фильтрами
// @Summary Получить список вакансий
// @Description Возвращает список вакансий. Если задан skill_slugs или search_title — поиск через Elasticsearch.
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
// @Param search_title query string false "Поиск по названию вакансии (через Elasticsearch)"
// @Param skill_slugs query string false "Slug-и навыков через запятую (через Elasticsearch)"
// @Success 200 {object} models.VacancyList "Список вакансий"
// @Failure 400 {object} models.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /vacancy [get]
func (h *Handler) GetVacancies(c *fiber.Ctx) error {
	log.Printf("GetVacancies: Getting vacancies list")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	page, limit = normalizePagination(page, limit)
	companyID := c.Query("company_id", "")
	positionStatus := c.Query("position_status", "")
	workFormat := c.Query("work_format", "")
	schedule := c.Query("schedule", "")
	minSalary, _ := strconv.Atoi(c.Query("min_salary", "0"))
	maxSalary, _ := strconv.Atoi(c.Query("max_salary", "0"))
	minExperience, _ := strconv.Atoi(c.Query("min_experience", "0"))
	maxExperience, _ := strconv.Atoi(c.Query("max_experience", "0"))
	searchTitle := c.Query("search_title", "")
	skillSlugs := splitCSV(c.Query("skill_slugs", ""))

	var vacancies *models.VacancyList
	var err error

	// Через Search идём ТОЛЬКО когда нужен фильтр по skill_slugs (его в SQL нет).
	// Текстовый поиск (search_title), формат, график и зарплата — через SQL ILIKE + индексы:
	// предсказуемо отдают partial match и работают вместе с остальными фильтрами.
	if h.apiService.Search.Available() && len(skillSlugs) > 0 {
		log.Printf("GetVacancies: routing through Search (skill_slugs=%v search_title=%q)", skillSlugs, searchTitle)
		vacancies, err = h.apiService.Search.SearchVacanciesAsModel(c.Context(), searchTitle, skillSlugs,
			int32(minSalary), int32(maxExperience), companyID, int32(page), int32(limit))
		// Пост-фильтрация в Gateway: ES в текущем mapping-е не моделирует
		// work_format/schedule/position_status, а salary/experience моделирует
		// только наполовину (передаётся min-salary и max-experience). Применяем
		// все оставшиеся ограничения к выдаче Search.
		if err == nil && vacancies != nil {
			vacancies.Vacancies = filterVacanciesInMemory(
				vacancies.Vacancies,
				workFormat, schedule, positionStatus,
				int32(minSalary), int32(maxSalary),
				int32(minExperience), int32(maxExperience),
			)
		}
	} else {
		pagination := &models.Pagination{
			Page:  int32(page),
			Limit: int32(limit),
		}
		vacancies, err = h.apiService.Vacancy.GetAllVacancies(c.Context(), pagination,
			companyID, positionStatus, workFormat, schedule,
			int32(minSalary), int32(maxSalary), int32(minExperience), int32(maxExperience),
			searchTitle)
	}
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
	page, limit = normalizePagination(page, limit)
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
	userRole := getRoleFromContext(c)
	log.Printf("CreateHRVacancy: user=%s role=%s", userID, userRole)

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

	// Закрытие B3+B4: владелец компании публикует под СВОЙ company_id напрямую (PUBLISHED).
	// HR-сотрудник публикует под company_id компании, в которой состоит (APPROVED-membership),
	// но в статусе PENDING — owner затем модерирует.
	switch userRole {
	case ROLE_COMPANY:
		req.CompanyID = userID
		req.ModerationStatus = 2 // PUBLISHED
		req.AuthorID = userID
	case ROLE_HR:
		ms, err := h.apiService.Company.GetMembershipByUser(c.Context(), userID)
		if err != nil || ms == nil || ms.Status != 2 { // 2 = APPROVED
			return c.Status(fiber.StatusForbidden).JSON(models.Error{
				Code:    "MEMBERSHIP_REQUIRED",
				Message: "HR должен быть подтверждённым сотрудником компании. Подайте заявку через /hr-profile.",
			})
		}
		req.CompanyID = ms.CompanyID
		req.ModerationStatus = 1 // PENDING — owner смодерирует
		req.AuthorID = userID
	case ROLE_DEVELOPER:
		// devloper-режим — оставляем как есть, требуем явный company_id
		if req.CompanyID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(models.Error{
				Code:    "MISSING_FIELD",
				Message: "Company ID is required",
			})
		}
	default:
		return c.Status(fiber.StatusForbidden).JSON(models.Error{
			Code:    "FORBIDDEN",
			Message: "Insufficient permissions",
		})
	}

	vacancy, err := h.apiService.Vacancy.CreateVacancy(c.Context(), &req)
	if err != nil {
		log.Printf("CreateHRVacancy: Failed to create vacancy: %v", err)
		// Маппим gRPC-коды Vacancy-сервиса в HTTP: InvalidArgument → 400 с
		// человекочитаемым сообщением (его потом покажет фронт), иначе 500.
		if st, ok := grpcstatus.FromError(err); ok {
			switch st.Code() {
			case grpccodes.InvalidArgument:
				return c.Status(fiber.StatusBadRequest).JSON(models.Error{
					Code:    "INVALID_VACANCY_DATA",
					Message: st.Message(),
				})
			case grpccodes.PermissionDenied:
				return c.Status(fiber.StatusForbidden).JSON(models.Error{
					Code:    "FORBIDDEN",
					Message: st.Message(),
				})
			}
		}
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
	userRole := getRoleFromContext(c)
	log.Printf("UpdateVacancy: Updating vacancy %s by user %s (role %s)", vacancyID, userID, userRole)

	if err := h.assertVacancyOwnership(c, vacancyID, userID, userRole); err != nil {
		return err
	}

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
	userRole := getRoleFromContext(c)
	log.Printf("DeleteVacancy: Deleting vacancy %s by user %s (role %s)", vacancyID, userID, userRole)

	if err := h.assertVacancyOwnership(c, vacancyID, userID, userRole); err != nil {
		return err
	}

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

// filterVacanciesInMemory применяет фильтры, не моделируемые ES-mapping-ом
// (workFormat / schedule / positionStatus) или моделируемые частично
// (salary-min/max — Search принимает только min; experience-min/max — Search
// принимает только max). Прогоняется поверх Search-выдачи, безопасно для MVP.
func filterVacanciesInMemory(
	list []*models.Vacancy,
	workFormat, schedule, positionStatus string,
	minSalary, maxSalary, minExperience, maxExperience int32,
) []*models.Vacancy {
	if workFormat == "" && schedule == "" && positionStatus == "" &&
		minSalary == 0 && maxSalary == 0 && minExperience == 0 && maxExperience == 0 {
		return list
	}
	out := make([]*models.Vacancy, 0, len(list))
	for _, v := range list {
		if v == nil {
			continue
		}
		if workFormat != "" && v.WorkFormat != workFormat {
			continue
		}
		if schedule != "" && v.Schedule != schedule {
			continue
		}
		if positionStatus != "" && v.PositionStatus != positionStatus {
			continue
		}
		if minSalary > 0 && v.Salary < minSalary {
			continue
		}
		if maxSalary > 0 && v.Salary > maxSalary {
			continue
		}
		if minExperience > 0 && v.Experience < minExperience {
			continue
		}
		if maxExperience > 0 && v.Experience > maxExperience {
			continue
		}
		out = append(out, v)
	}
	return out
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

// assertVacancyOwnership: только владелец компании, под которой опубликована вакансия,
// или ROLE_DEVELOPER может её править/удалять. ROLE_HR без membership пока заблокирован
// (см. CreateHRVacancy и B4 в BUGS_E2E.md).
func (h *Handler) assertVacancyOwnership(c *fiber.Ctx, vacancyID, userID string, userRole Role) error {
	if userRole == ROLE_DEVELOPER {
		return nil
	}
	if userRole == ROLE_HR {
		return c.Status(fiber.StatusForbidden).JSON(models.Error{
			Code:    "MEMBERSHIP_REQUIRED",
			Message: "HR может управлять вакансиями только в составе компании; HR-membership пока не реализован",
		})
	}
	if userRole != ROLE_COMPANY {
		return c.Status(fiber.StatusForbidden).JSON(models.Error{
			Code:    "FORBIDDEN",
			Message: "Insufficient permissions",
		})
	}
	vacancy, err := h.apiService.Vacancy.GetVacancy(c.Context(), vacancyID)
	if err != nil {
		log.Printf("assertVacancyOwnership: vacancy %s not found: %v", vacancyID, err)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "VACANCY_NOT_FOUND",
			Message: "Vacancy not found",
		})
	}
	if vacancy.CompanyID != userID {
		log.Printf("assertVacancyOwnership: user %s tried to mutate vacancy of company %s", userID, vacancy.CompanyID)
		return c.Status(fiber.StatusForbidden).JSON(models.Error{
			Code:    "FORBIDDEN",
			Message: "Vacancy belongs to another company",
		})
	}
	return nil
}
