package handlers

import (
	"context"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
)

// ServeFileDirect отдает файл напрямую через бекенд
// @Summary Прямая загрузка файла
// @Description Перенаправляет на presigned URL для скачивания файла. Используется для отображения изображений и файлов напрямую в браузере.
// @Tags Files
// @Accept json
// @Produce json
// @Param entity_id path string true "ID сущности (пользователь, компания, вакансия)" example("user-123")
// @Param file_name path string true "Имя файла" example("avatar.jpg")
// @Success 302 {string} string "Перенаправление на URL файла"
// @Failure 404 {object} models.Error "Файл не найден"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /files/{entity_id}/{file_name} [get]
func (h *Handler) ServeFileDirect(c *fiber.Ctx) error {
	entityID := c.Params("entity_id")
	fileName := c.Params("file_name")

	log.Printf("ServeFileDirect: Serving file %s for entity %s", fileName, entityID)

	downloadURL, err := h.apiService.Achievement.GetAchievementDownloadUrl(c.Context(), entityID, fileName)
	if err != nil {
		log.Printf("ServeFileDirect: Failed to get download URL for file %s: %v", fileName, err)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "FILE_NOT_FOUND",
			Message: "File not found",
		})
	}

	return c.Redirect(downloadURL.URL, fiber.StatusFound)
}

// UploadUserAvatar загружает аватар пользователя
// @Summary Загрузить аватар пользователя
// @Description Загружает аватар для текущего пользователя. Поддерживаемые форматы: JPG, PNG, GIF. Максимальный размер: 5MB.
// @Tags Users/Files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Файл аватара (макс. 5MB)"
// @Success 200 {object} models.FileUploadResponse "Информация о загруженном файле"
// @Failure 400 {object} models.Error "Неверный запрос или файл слишком большой"
// @Failure 401 {object} models.Error "Неавторизованный доступ"
// @Failure 413 {object} models.Error "Файл превышает максимальный размер"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /users/files/avatar [post]
func (h *Handler) UploadUserAvatar(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("UploadUserAvatar: Uploading avatar for user: %s", userID)

	file, err := c.FormFile("avatar")
	if err != nil {
		log.Printf("UploadUserAvatar: No avatar file provided")
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_FILE",
			Message: "No avatar file provided",
		})
	}

	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "FILE_TOO_LARGE",
			Message: "File size too large. Maximum size is 5MB",
		})
	}

	fileInfo, err := h.fileHandler.UploadFileDirect(
		c.Context(),
		userID,
		"user",
		"avatar",
		file,
	)
	if err != nil {
		log.Printf("UploadUserAvatar: Failed to upload avatar for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "UPLOAD_FAILED",
			Message: "Failed to upload avatar",
		})
	}

	avatarID := fileInfo.Name
	_, err = h.apiService.User.UpdateUser(c.Context(), &usersv1.UpdateProfileRequest{
		Id: userID,
		Profile: &usersv1.Profile{
			AvatarId: avatarID,
		},
	})
	if err != nil {
		log.Printf("UploadUserAvatar: Failed to update user avatar in profile: %v", err)
	}

	log.Printf("UploadUserAvatar: Successfully uploaded avatar for user: %s", userID)
	return c.JSON(models.FileUploadResponse{
		FileInfo: fileInfo,
		Message:  "Avatar uploaded successfully",
	})
}

// UploadUserResume загружает резюме пользователя
// @Summary Загрузить резюме пользователя
// @Description Загружает резюме для текущего пользователя. Поддерживаемые форматы: PDF, DOC, DOCX. Максимальный размер: 10MB.
// @Tags Users/Files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param resume formData file true "Файл резюме (макс. 10MB)"
// @Success 200 {object} models.FileUploadResponse "Информация о загруженном файле"
// @Failure 400 {object} models.Error "Неверный запрос или файл слишком большой"
// @Failure 401 {object} models.Error "Неавторизованный доступ"
// @Failure 413 {object} models.Error "Файл превышает максимальный размер"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /users/files/resume [post]
func (h *Handler) UploadUserResume(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("UploadUserResume: Uploading resume for user: %s", userID)

	file, err := c.FormFile("resume")
	if err != nil {
		log.Printf("UploadUserResume: No resume file provided")
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_FILE",
			Message: "No resume file provided",
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
		userID,
		"user",
		"resume",
		file,
	)
	if err != nil {
		log.Printf("UploadUserResume: Failed to upload resume for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "UPLOAD_FAILED",
			Message: "Failed to upload resume",
		})
	}

	resumeID := fileInfo.Name
	_, err = h.apiService.User.UpdateUser(c.Context(), &usersv1.UpdateProfileRequest{
		Id: userID,
		Profile: &usersv1.Profile{
			ResumeId: resumeID,
		},
	})
	if err != nil {
		log.Printf("UploadUserResume: Failed to update user resume in profile: %v", err)
	}

	log.Printf("UploadUserResume: Successfully uploaded resume for user: %s", userID)
	return c.JSON(models.FileUploadResponse{
		FileInfo: fileInfo,
		Message:  "Resume uploaded successfully",
	})
}

// DeleteUserAvatar удаляет аватар пользователя
// @Summary Удалить аватар пользователя
// @Description Удаляет аватар текущего пользователя. После удаления будет установлен аватар по умолчанию.
// @Tags Users/Files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse "Сообщение об успешном удалении"
// @Failure 400 {object} models.Error "У пользователя нет аватара"
// @Failure 401 {object} models.Error "Неавторизованный доступ"
// @Failure 404 {object} models.Error "Пользователь не найден"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /users/files/avatar [delete]
func (h *Handler) DeleteUserAvatar(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("DeleteUserAvatar: Deleting avatar for user: %s", userID)

	profile, err := h.apiService.User.GetUser(c.Context(), userID)
	if err != nil {
		log.Printf("DeleteUserAvatar: User not found: %s", userID)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}

	if profile.AvatarId == "" {
		log.Printf("DeleteUserAvatar: User %s doesn't have an avatar", userID)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_AVATAR",
			Message: "User doesn't have an avatar",
		})
	}

	err = h.fileHandler.DeleteFile(c.Context(), userID, profile.AvatarId)
	if err != nil {
		log.Printf("DeleteUserAvatar: Failed to delete avatar file for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "DELETE_FAILED",
			Message: "Failed to delete avatar",
		})
	}

	_, err = h.apiService.User.UpdateUser(c.Context(), &usersv1.UpdateProfileRequest{
		Id: userID,
		Profile: &usersv1.Profile{
			AvatarId: "",
		},
	})
	if err != nil {
		log.Printf("DeleteUserAvatar: Failed to update user profile: %v", err)
	}

	log.Printf("DeleteUserAvatar: Successfully deleted avatar for user: %s", userID)
	return c.JSON(models.SuccessResponse{
		Message: "Avatar deleted successfully",
	})
}

// DeleteUserResume удаляет резюме пользователя
// @Summary Удалить резюме пользователя
// @Description Удаляет резюме текущего пользователя. После удаления поле resume_id будет очищено.
// @Tags Users/Files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse "Сообщение об успешном удалении"
// @Failure 400 {object} models.Error "У пользователя нет резюме"
// @Failure 401 {object} models.Error "Неавторизованный доступ"
// @Failure 404 {object} models.Error "Пользователь не найден"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /users/files/resume [delete]
func (h *Handler) DeleteUserResume(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("DeleteUserResume: Deleting resume for user: %s", userID)

	profile, err := h.apiService.User.GetUser(c.Context(), userID)
	if err != nil {
		log.Printf("DeleteUserResume: User not found: %s", userID)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}

	if profile.ResumeId == "" {
		log.Printf("DeleteUserResume: User %s doesn't have a resume", userID)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_RESUME",
			Message: "User doesn't have a resume",
		})
	}

	err = h.fileHandler.DeleteFile(c.Context(), userID, profile.ResumeId)
	if err != nil {
		log.Printf("DeleteUserResume: Failed to delete resume file for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "DELETE_FAILED",
			Message: "Failed to delete resume",
		})
	}

	_, err = h.apiService.User.UpdateUser(c.Context(), &usersv1.UpdateProfileRequest{
		Id: userID,
		Profile: &usersv1.Profile{
			ResumeId: "",
		},
	})
	if err != nil {
		log.Printf("DeleteUserResume: Failed to update user profile: %v", err)
	}

	log.Printf("DeleteUserResume: Successfully deleted resume for user: %s", userID)
	return c.JSON(models.SuccessResponse{
		Message: "Resume deleted successfully",
	})
}

// UploadCompanyLogo загружает логотип компании
// @Summary Загрузить логотип компании
// @Description Загружает логотип для компании текущего пользователя. Поддерживаемые форматы: JPG, PNG, SVG. Максимальный размер: 5MB.
// @Tags Companies/Files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID компании" example("comp-123")
// @Param logo formData file true "Файл логотипа (макс. 5MB)"
// @Success 200 {object} models.FileUploadResponse "Информация о загруженном файле"
// @Failure 400 {object} models.Error "Неверный запрос или файл слишком большой"
// @Failure 401 {object} models.Error "Неавторизованный доступ"
// @Failure 403 {object} models.Error "Доступ запрещен"
// @Failure 404 {object} models.Error "Компания не найдена"
// @Failure 413 {object} models.Error "Файл превышает максимальный размер"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /company/{id}/files/logo [post]
func (h *Handler) UploadCompanyLogo(c *fiber.Ctx) error {
	companyID := c.Params("id")
	log.Printf("UploadCompanyLogo: Uploading logo for company: %s", companyID)

	file, err := c.FormFile("logo")
	if err != nil {
		log.Printf("UploadCompanyLogo: No logo file provided for company %s", companyID)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_FILE",
			Message: "No logo file provided",
		})
	}

	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "FILE_TOO_LARGE",
			Message: "File size too large. Maximum size is 5MB",
		})
	}

	fileInfo, err := h.fileHandler.UploadFileDirect(
		c.Context(),
		companyID,
		"company",
		"logo",
		file,
	)
	if err != nil {
		log.Printf("UploadCompanyLogo: Failed to upload logo for company %s: %v", companyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "UPLOAD_FAILED",
			Message: "Failed to upload logo",
		})
	}

	logoID := fileInfo.Name
	_, err = h.apiService.Company.UpdateCompany(c.Context(), companyID, &models.Company{
		LogoID: &logoID,
	})
	if err != nil {
		log.Printf("UploadCompanyLogo: Failed to update company logo: %v", err)
	}

	log.Printf("UploadCompanyLogo: Successfully uploaded logo for company: %s", companyID)
	return c.JSON(models.FileUploadResponse{
		FileInfo: fileInfo,
		Message:  "Company logo uploaded successfully",
	})
}

// UploadCompanyDocument загружает документ компании
// @Summary Загрузить документ компании
// @Description Загружает документ для компании текущего пользователя. Поддерживаемые форматы: PDF, DOC, DOCX, XLS, XLSX. Максимальный размер: 20MB.
// @Tags Companies/Files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID компании" example("comp-123")
// @Param document formData file true "Файл документа (макс. 20MB)"
// @Success 200 {object} models.FileUploadResponse "Информация о загруженном файле"
// @Failure 400 {object} models.Error "Неверный запрос или файл слишком большой"
// @Failure 401 {object} models.Error "Неавторизованный доступ"
// @Failure 403 {object} models.Error "Доступ запрещен"
// @Failure 404 {object} models.Error "Компания не найдена"
// @Failure 413 {object} models.Error "Файл превышает максимальный размер"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /company/{id}/files/documents [post]
func (h *Handler) UploadCompanyDocument(c *fiber.Ctx) error {
	companyID := c.Params("id")
	log.Printf("UploadCompanyDocument: Uploading document for company: %s", companyID)

	file, err := c.FormFile("document")
	if err != nil {
		log.Printf("UploadCompanyDocument: No document file provided for company %s", companyID)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_FILE",
			Message: "No document file provided",
		})
	}

	if file.Size > 20*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "FILE_TOO_LARGE",
			Message: "File size too large. Maximum size is 20MB",
		})
	}

	fileInfo, err := h.fileHandler.UploadFileDirect(
		c.Context(),
		companyID,
		"company",
		"document",
		file,
	)
	if err != nil {
		log.Printf("UploadCompanyDocument: Failed to upload document for company %s: %v", companyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "UPLOAD_FAILED",
			Message: "Failed to upload document",
		})
	}

	log.Printf("UploadCompanyDocument: Successfully uploaded document for company: %s", companyID)
	return c.JSON(models.FileUploadResponse{
		FileInfo: fileInfo,
		Message:  "Company document uploaded successfully",
	})
}

// DeleteCompanyLogo удаляет логотип компании
// @Summary Удалить логотип компании
// @Description Удаляет логотип компании текущего пользователя. После удаления будет установлен логотип по умолчанию.
// @Tags Companies/Files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID компании" example("comp-123")
// @Success 200 {object} models.SuccessResponse "Сообщение об успешном удалении"
// @Failure 400 {object} models.Error "У компании нет логотипа"
// @Failure 401 {object} models.Error "Неавторизованный доступ"
// @Failure 403 {object} models.Error "Доступ запрещен"
// @Failure 404 {object} models.Error "Компания не найдена"
// @Failure 500 {object} models.Error "Внутренняя ошибка сервера"
// @Router /company/{id}/files/logo [delete]
func (h *Handler) DeleteCompanyLogo(c *fiber.Ctx) error {
	companyID := c.Params("id")
	log.Printf("DeleteCompanyLogo: Deleting logo for company: %s", companyID)

	company, err := h.apiService.Company.GetCompany(c.Context(), companyID)
	if err != nil {
		log.Printf("DeleteCompanyLogo: Company not found: %s", companyID)
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "COMPANY_NOT_FOUND",
			Message: "Company not found",
		})
	}

	if company.LogoID == nil || *company.LogoID == "" {
		log.Printf("DeleteCompanyLogo: Company %s doesn't have a logo", companyID)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "NO_LOGO",
			Message: "Company doesn't have a logo",
		})
	}

	err = h.fileHandler.DeleteFile(c.Context(), companyID, *company.LogoID)
	if err != nil {
		log.Printf("DeleteCompanyLogo: Failed to delete logo file for company %s: %v", companyID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "DELETE_FAILED",
			Message: "Failed to delete logo",
		})
	}

	emptyLogoID := ""
	_, err = h.apiService.Company.UpdateCompany(c.Context(), companyID, &models.Company{
		LogoID: &emptyLogoID,
	})
	if err != nil {
		log.Printf("DeleteCompanyLogo: Failed to update company: %v", err)
	}

	log.Printf("DeleteCompanyLogo: Successfully deleted logo for company: %s", companyID)
	return c.JSON(models.SuccessResponse{
		Message: "Company logo deleted successfully",
	})
}

// enrichCompanyWithFiles обогащает компанию информацией о файлах
func (h *Handler) enrichCompanyWithFiles(ctx context.Context, company *models.Company) {
	if company.LogoID != nil && *company.LogoID != "" {
		fileInfo, err := h.fileHandler.GetFileInfo(ctx, company.ID, *company.LogoID, "logo")
		if err == nil && fileInfo != nil {
			if fileInfo.DirectURL != nil {
				company.LogoURL = fileInfo.DirectURL
			} else {
				company.LogoURL = fileInfo.URL
			}
		} else {
			log.Printf("enrichCompanyWithFiles: Failed to get logo info for company %s: %v", company.ID, err)
		}
	}
}

// enrichCompanyListWithFiles обогащает список компаний информацией о файлах
func (h *Handler) enrichCompanyListWithFiles(ctx context.Context, companies []*models.Company) {
	for _, company := range companies {
		h.enrichCompanyWithFiles(ctx, company)
	}
}
