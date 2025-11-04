package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

// GetCompanies возвращает список компаний с пагинацией и фильтрами
// @Summary Получить список компаний
// @Description Возвращает список компаний с поддержкой пагинации и фильтрации по городу и типу
// @Tags Companies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(10) minimum(1) maximum(100)
// @Param city query string false "Фильтр по городу"
// @Param type query string false "Фильтр по типу компании"
// @Success 200 {object} models.CompanyList "Список компаний"
// @Failure 400 {object} models.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /company [get]
func (h *Handler) GetCompanies(c *fiber.Ctx) error {
	log.Printf("GetCompanies: Getting companies list")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	city := c.Query("city", "")
	companyType := c.Query("type", "")

	pagination := &models.Pagination{
		Page:  int32(page),
		Limit: int32(limit),
	}

	companies, err := h.apiService.Company.GetAllCompanies(c.Context(), pagination, city, companyType)
	if err != nil {
		log.Printf("GetCompanies: Failed to get companies: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get companies",
		})
	}

	h.enrichCompanyListWithFiles(c.Context(), companies.Companies)

	log.Printf("GetCompanies: Successfully retrieved %d companies", len(companies.Companies))
	return c.JSON(companies)
}

// GetCompanyByID возвращает компанию по ID
// @Summary Получить компанию по ID
// @Description Возвращает информацию о компании по указанному идентификатору
// @Tags Companies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID компании" format(uuid)
// @Success 200 {object} models.Company "Информация о компании"
// @Failure 400 {object} models.ErrorResponse "Неверный ID компании"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Компания не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /company/{id} [get]
func (h *Handler) GetCompanyByID(c *fiber.Ctx) error {
	companyID := c.Params("id")
	log.Printf("GetCompanyByID: Getting company with ID: %s", companyID)

	company, err := h.apiService.Company.GetCompany(c.Context(), companyID)
	if err != nil {
		log.Printf("GetCompanyByID: Failed to get company %s: %v", companyID, err)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "COMPANY_NOT_FOUND",
			Message: "Company not found",
		})
	}

	h.enrichCompanyWithFiles(c.Context(), company)

	log.Printf("GetCompanyByID: Successfully retrieved company: %s", companyID)
	return c.JSON(company)
}

// GetCompanyMe возвращает текущую компанию пользователя
// @Summary Получить текущую компанию
// @Description Возвращает информацию о компании текущего аутентифицированного пользователя
// @Tags Companies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Company "Информация о компании"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Компания не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /company/me [get]
func (h *Handler) GetCompanyMe(c *fiber.Ctx) error {
	companyID := getUserIDFromContext(c)
	log.Printf("GetCompanyMe: Getting company with ID: %s", companyID)

	company, err := h.apiService.Company.GetCompany(c.Context(), companyID)
	if err != nil {
		log.Printf("GetCompanyMe: Failed to get company %s: %v", companyID, err)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "COMPANY_NOT_FOUND",
			Message: "Company not found",
		})
	}

	h.enrichCompanyWithFiles(c.Context(), company)

	log.Printf("GetCompanyMe: Successfully retrieved company: %s", companyID)
	return c.JSON(company)
}

// CreateCompany создает новую компанию
// @Summary Создать компанию
// @Description Создает новую компанию для текущего пользователя
// @Tags Companies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Company true "Данные компании"
// @Success 201 {object} models.Company "Созданная компания"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /company [post]
func (h *Handler) CreateCompany(c *fiber.Ctx) error {
	log.Printf("CreateCompany: Creating new company")

	var req models.Company
	if err := c.BodyParser(&req); err != nil {
		log.Printf("CreateCompany: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	if req.Name == "" {
		log.Printf("CreateCompany: Missing required field 'name'")
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "MISSING_FIELD",
			Message: "Company name is required",
		})
	}

	company, err := h.apiService.Company.CreateCompany(c.Context(), &req)
	if err != nil {
		log.Printf("CreateCompany: Failed to create company: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "CREATE_FAILED",
			Message: "Failed to create company",
		})
	}

	log.Printf("CreateCompany: Successfully created company: %s", company.ID)
	return c.Status(fiber.StatusCreated).JSON(company)
}

// UpdateCompany обновляет компанию
// @Summary Обновить компанию
// @Description Обновляет данные компании текущего пользователя
// @Tags Companies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Company true "Данные для обновления"
// @Success 200 {object} models.Company "Обновленная компания"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} models.ErrorResponse "Компания не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /company [patch]
func (h *Handler) UpdateCompany(c *fiber.Ctx) error {
	companyID := getUserIDFromContext(c)
	log.Printf("UpdateCompany: Updating company %s", companyID)

	var req models.Company
	if err := c.BodyParser(&req); err != nil {
		log.Printf("UpdateCompany: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	company, err := h.apiService.Company.UpdateCompany(c.Context(), companyID, &req)
	if err != nil {
		log.Printf("UpdateCompany: Failed to update company %s: %v", companyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "UPDATE_FAILED",
			Message: "Failed to update company",
		})
	}

	h.enrichCompanyWithFiles(c.Context(), company)

	log.Printf("UpdateCompany: Successfully updated company: %s", companyID)
	return c.JSON(company)
}

// DeleteCompany удаляет компанию
// @Summary Удалить компанию
// @Description Удаляет компанию текущего пользователя
// @Tags Companies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Сообщение об успешном удалении"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} models.ErrorResponse "Компания не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /company [delete]
func (h *Handler) DeleteCompany(c *fiber.Ctx) error {
	companyID := getUserIDFromContext(c)
	log.Printf("DeleteCompany: Deleting company %s", companyID)

	company, err := h.apiService.Company.GetCompany(c.Context(), companyID)
	if err == nil && company.LogoID != nil && *company.LogoID != "" {
		if err := h.fileHandler.DeleteFile(c.Context(), companyID, *company.LogoID); err != nil {
			log.Printf("DeleteCompany: Failed to delete logo for company %s: %v", companyID, err)
		}
	}

	if err := h.apiService.Company.DeleteCompany(c.Context(), companyID); err != nil {
		log.Printf("DeleteCompany: Failed to delete company %s: %v", companyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "DELETE_FAILED",
			Message: "Failed to delete company",
		})
	}

	if err := h.apiService.Auth.DeleteUser(c.Context(), companyID); err != nil {
		log.Printf("DeleteCompany: Failed to delete company %s: %v", companyID, err)
	}

	log.Printf("DeleteCompany: Successfully deleted company: %s", companyID)
	return c.JSON(fiber.Map{
		"message": "Company deleted successfully",
	})
}
