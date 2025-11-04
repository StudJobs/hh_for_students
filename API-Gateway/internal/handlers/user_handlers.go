package handlers

import (
	"context"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
	"strconv"
)

// GetUsers возвращает список пользователей с пагинацией
// @Summary Получить список пользователей
// @Description Возвращает список пользователей с поддержкой пагинации и фильтрации по категории
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Количество элементов на странице" default(10) minimum(1) maximum(100)
// @Param category query string false "Фильтр по категории профессии"
// @Success 200 {object} models.ProfileList "Список пользователей"
// @Failure 400 {object} models.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users [get]
func (h *Handler) GetUsers(c *fiber.Ctx) error {
	log.Printf("GetUsers: Getting users list")

	// Получаем параметры пагинации
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	category := c.Query("category", "")

	// Строим запрос к users service
	req := &usersv1.GetAllProfilesRequest{
		Pagination: &commonv1.Pagination{
			Page:  int32(page),
			Limit: int32(limit),
		},
	}

	if category != "" {
		req.ProfessionCategory = category
	}

	// Вызываем users service
	profiles, err := h.apiService.User.GetUsers(c.Context(), req)
	if err != nil {
		log.Printf("GetUsers: Failed to get users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get users",
		})
	}

	// Конвертируем в HTTP модель
	profileList := models.ProfileList{
		Profiles: make([]models.User, len(profiles.Profiles)),
		Pagination: models.PaginationResponse{
			Total:       profiles.Pagination.Total,
			Pages:       profiles.Pagination.Pages,
			CurrentPage: profiles.Pagination.CurrentPage,
		},
	}

	// Заполняем базовую информацию
	for i, profile := range profiles.Profiles {
		profileList.Profiles[i] = models.User{
			ID:                 uuid.MustParse(profile.Id),
			FirstName:          profile.FirstName,
			LastName:           profile.LastName,
			Age:                profile.Age,
			Tg:                 profile.Tg,
			Email:              profile.Email,
			Description:        profile.Description,
			ProfessionCategory: profile.ProfessionCategory,
		}
	}

	// Обогащаем информацией о файлах
	h.enrichUserListWithFiles(c.Context(), profileList.Profiles, profiles.Profiles)

	log.Printf("GetUsers: Successfully retrieved %d users", len(profileList.Profiles))
	return c.JSON(profileList)
}

// GetMe возвращает текущего пользователя
// @Summary Получить текущего пользователя
// @Description Возвращает профиль текущего аутентифицированного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User "Профиль пользователя"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/me [get]
func (h *Handler) GetMe(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("GetMe: Getting current user: %s", userID)

	user, err := h.getUserWithFiles(c, userID)
	if err != nil {
		return err
	}

	log.Printf("GetMe: Successfully retrieved current user: %s", userID)
	return c.JSON(user)
}

// GetUser возвращает профиль пользователя по ID
// @Summary Получить пользователя по ID
// @Description Возвращает профиль пользователя по указанному идентификатору
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID пользователя" format(uuid)
// @Success 200 {object} models.User "Профиль пользователя"
// @Failure 400 {object} models.ErrorResponse "Неверный ID пользователя"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/{id} [get]
func (h *Handler) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	log.Printf("GetUser: Getting user with ID: %s", userID)

	user, err := h.getUserWithFiles(c, userID)
	if err != nil {
		return err
	}

	log.Printf("GetUser: Successfully retrieved user: %s", userID)
	return c.JSON(user)
}

// getUserWithFiles вспомогательная функция для получения пользователя с файлами
func (h *Handler) getUserWithFiles(c *fiber.Ctx, userID string) (*models.User, error) {
	// Валидация UUID
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("getUserWithFiles: Invalid UUID format: %s", userID)
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Вызываем users service
	profile, err := h.apiService.User.GetUser(c.Context(), userID)
	if err != nil {
		log.Printf("getUserWithFiles: Failed to get user %s: %v", userID, err)
		return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Конвертируем в HTTP модель
	user := &models.User{
		ID:                 uuid.MustParse(profile.Id),
		FirstName:          profile.FirstName,
		LastName:           profile.LastName,
		Age:                profile.Age,
		Tg:                 profile.Tg,
		Email:              profile.Email,
		Description:        profile.Description,
		ProfessionCategory: profile.ProfessionCategory,
	}

	// Обогащаем информацией о файлах
	h.enrichUserWithFiles(c.Context(), user, profile)

	return user, nil
}

// enrichUserWithFiles обогащает пользователя информацией о файлах
func (h *Handler) enrichUserWithFiles(ctx context.Context, user *models.User, profile *usersv1.Profile) {
	// Обрабатываем аватар
	if profile.AvatarId != "" {
		avatarUUID := uuid.MustParse(profile.AvatarId)
		user.AvatarID = &avatarUUID

		fileInfo, err := h.fileHandler.GetFileInfo(ctx, profile.Id, profile.AvatarId, "avatar")
		if err == nil && fileInfo != nil {
			// Для аватара отдаем прямую ссылку (если есть) или presigned URL
			if fileInfo.DirectURL != nil {
				user.AvatarURL = fileInfo.DirectURL
			} else {
				user.AvatarURL = fileInfo.URL
			}
		}
	}

	// Обрабатываем резюме
	if profile.ResumeId != "" {
		resumeUUID := uuid.MustParse(profile.ResumeId)
		user.ResumeID = &resumeUUID

		fileInfo, err := h.fileHandler.GetFileInfo(ctx, profile.Id, profile.ResumeId, "resume")
		if err == nil && fileInfo != nil {
			// Для резюме всегда отдаем presigned URL
			user.ResumeURL = fileInfo.URL
		}
	}
}

// enrichUserListWithFiles обогащает список пользователей информацией о файлах
func (h *Handler) enrichUserListWithFiles(ctx context.Context, users []models.User, profiles []*usersv1.Profile) {
	for i, profile := range profiles {
		// Обрабатываем аватар для каждого пользователя
		if profile.AvatarId != "" {
			avatarUUID := uuid.MustParse(profile.AvatarId)
			users[i].AvatarID = &avatarUUID

			fileInfo, err := h.fileHandler.GetFileInfo(ctx, profile.Id, profile.AvatarId, "avatar")
			if err == nil && fileInfo != nil {
				if fileInfo.DirectURL != nil {
					users[i].AvatarURL = fileInfo.DirectURL
				} else {
					users[i].AvatarURL = fileInfo.URL
				}
			}
		}

		// Обрабатываем резюме для каждого пользователя
		if profile.ResumeId != "" {
			resumeUUID := uuid.MustParse(profile.ResumeId)
			users[i].ResumeID = &resumeUUID

			fileInfo, err := h.fileHandler.GetFileInfo(ctx, profile.Id, profile.ResumeId, "resume")
			if err == nil && fileInfo != nil {
				users[i].ResumeURL = fileInfo.URL
			}
		}
	}
}

// UpdateUser обновляет профиль пользователя (PATCH)
// @Summary Обновить профиль пользователя
// @Description Обновляет данные профиля текущего пользователя. Поддерживает частичное обновление.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UserUpdateRequest true "Данные для обновления"
// @Success 200 {object} models.User "Обновленный профиль пользователя"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/edit [patch]
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("UpdateUser: Updating user %s", userID)

	// Валидация UUID
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("UpdateUser: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Парсим тело запроса
	var updateData models.UserUpdateRequest
	if err := c.BodyParser(&updateData); err != nil {
		log.Printf("UpdateUser: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Строим профиль для обновления
	profile := &usersv1.Profile{}

	if updateData.FirstName != nil {
		profile.FirstName = *updateData.FirstName
	}
	if updateData.LastName != nil {
		profile.LastName = *updateData.LastName
	}
	if updateData.Age != nil {
		profile.Age = *updateData.Age
	}
	if updateData.Tg != nil {
		profile.Tg = *updateData.Tg
	}
	if updateData.Email != nil {
		profile.Email = *updateData.Email
	}
	if updateData.Description != nil {
		profile.Description = *updateData.Description
	}
	if updateData.ProfessionCategory != nil {
		profile.ProfessionCategory = *updateData.ProfessionCategory
	}
	if updateData.ResumeID != nil {
		profile.ResumeId = *updateData.ResumeID
	}
	if updateData.AvatarID != nil {
		profile.AvatarId = *updateData.AvatarID
	}

	// Вызываем users service
	updatedProfile, err := h.apiService.User.UpdateUser(c.Context(), &usersv1.UpdateProfileRequest{
		Id:      userID,
		Profile: profile,
	})
	if err != nil {
		log.Printf("UpdateUser: Failed to update user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	// Конвертируем в HTTP модель
	user := &models.User{
		ID:                 uuid.MustParse(updatedProfile.Id),
		FirstName:          updatedProfile.FirstName,
		LastName:           updatedProfile.LastName,
		Age:                updatedProfile.Age,
		Tg:                 updatedProfile.Tg,
		Email:              updatedProfile.Email,
		Description:        updatedProfile.Description,
		ProfessionCategory: updatedProfile.ProfessionCategory,
	}

	// Обогащаем информацией о файлах
	h.enrichUserWithFiles(c.Context(), user, updatedProfile)

	log.Printf("UpdateUser: Successfully updated user: %s", userID)
	return c.JSON(user)
}

// DeleteUser удаляет профиль пользователя
// @Summary Удалить пользователя
// @Description Удаляет профиль текущего пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Сообщение об успешном удалении"
// @Failure 400 {object} models.ErrorResponse "Неверный ID пользователя"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 403 {object} models.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users [delete]
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	currentUserRole := getRoleFromContext(c)
	log.Printf("DeleteUser: Deleting user %s (role: %s)", userID, currentUserRole)

	// Валидация UUID
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("DeleteUser: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Вызываем users service
	if err := h.apiService.User.DeleteUser(c.Context(), userID); err != nil {
		log.Printf("DeleteUser: Failed to delete user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete profile",
		})
	}

	// Удаляем пользователя из auth service
	if err := h.apiService.Auth.DeleteUser(c.Context(), userID); err != nil {
		log.Printf("DeleteUser: Failed to delete user from auth %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	log.Printf("DeleteUser: Successfully deleted user: %s", userID)
	return c.JSON(fiber.Map{
		"message": "Profile deleted successfully",
	})
}
