package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
	"time"
)

// GetUserAchievements возвращает все достижения пользователя
// @Summary Получить достижения пользователя
// @Description Возвращает список всех достижений текущего пользователя
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.AchievementList "Список достижений"
// @Failure 400 {object} models.ErrorResponse "Неверный ID пользователя"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/achievements [get]
func (h *Handler) GetUserAchievements(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("GetUserAchievements: Getting achievements for user: %s", userID)

	// Валидация UUID
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("GetUserAchievements: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Вызываем achievement service
	achievements, err := h.apiService.Achievement.GetAllAchievements(c.Context(), userID)
	if err != nil {
		log.Printf("GetUserAchievements: Failed to get achievements for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get achievements",
		})
	}

	log.Printf("GetUserAchievements: Successfully retrieved %d achievements for user: %s",
		len(achievements.Achievements), userID)
	return c.JSON(achievements)
}

// CreateUserAchievement создает новое достижение (подготавливает загрузку)
// @Summary Создать достижение
// @Description Создает новое достижение и возвращает URL для загрузки файла
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.AchievementUploadRequest true "Данные достижения"
// @Success 200 {object} models.AchievementCreateResponse "Данные для загрузки файла"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/achievements [post]
func (h *Handler) CreateUserAchievement(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	log.Printf("CreateUserAchievement: Creating achievement for user: %s", userID)

	// Валидация UUID
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("CreateUserAchievement: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Парсим тело запроса
	var uploadReq models.AchievementUploadRequest
	if err := c.BodyParser(&uploadReq); err != nil {
		log.Printf("CreateUserAchievement: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Валидация обязательных полей
	if uploadReq.Name == "" || uploadReq.FileName == "" || uploadReq.FileType == "" {
		log.Printf("CreateUserAchievement: Missing required fields")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, file_name and file_type are required",
		})
	}

	if uploadReq.FileSize <= 0 {
		log.Printf("CreateUserAchievement: Invalid file size: %d", uploadReq.FileSize)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File size must be positive",
		})
	}

	// Получаем URL для загрузки и S3 ключ
	uploadResponse, err := h.apiService.Achievement.GetAchievementUploadUrl(
		c.Context(),
		userID,
		uploadReq.Name,
		uploadReq.FileName,
		uploadReq.FileType,
		uploadReq.FileSize,
	)
	if err != nil {
		log.Printf("CreateUserAchievement: Failed to get upload URL for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get upload URL",
		})
	}

	// Создаем метаданные достижения
	achievementMeta := models.AchievementMeta{
		Name:      uploadReq.Name,
		UserUUID:  userID,
		FileName:  uploadReq.FileName,
		FileType:  uploadReq.FileType,
		FileSize:  uploadReq.FileSize,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Формируем полный ответ
	response := models.AchievementCreateResponse{
		Meta: achievementMeta,
		UploadURL: models.UploadUrlResponse{
			UploadURL: uploadResponse.UploadURL,
			S3Key:     uploadResponse.S3Key,
			ExpiresAt: uploadResponse.ExpiresAt,
		},
	}

	log.Printf("CreateUserAchievement: Successfully prepared upload for user: %s, S3 key: %s",
		userID, uploadResponse.S3Key)
	return c.JSON(response)
}

// ConfirmAchievementUpload подтверждает успешную загрузку файла
// @Summary Подтвердить загрузку достижения
// @Description Подтверждает успешную загрузку файла достижения и сохраняет метаданные
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Название достижения"
// @Param request body map[string]string true "S3 ключ файла" example({"s3_key": "achievements/user123/file.pdf"})
// @Success 200 {object} map[string]string "Сообщение об успешном подтверждении"
// @Failure 400 {object} models.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Достижение не найдено"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/achievements/{id}/confirm [post]
func (h *Handler) ConfirmAchievementUpload(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	achievementName := c.Params("id")
	log.Printf("ConfirmAchievementUpload: Confirming upload for achievement %s, user: %s",
		achievementName, userID)

	// Валидация
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("ConfirmAchievementUpload: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	if achievementName == "" {
		log.Printf("ConfirmAchievementUpload: Achievement name is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Achievement name is required",
		})
	}

	// Парсим тело запроса с S3 ключом
	var confirmReq struct {
		S3Key string `json:"s3_key"`
	}
	if err := c.BodyParser(&confirmReq); err != nil {
		log.Printf("ConfirmAchievementUpload: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if confirmReq.S3Key == "" {
		log.Printf("ConfirmAchievementUpload: S3 key is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "S3 key is required",
		})
	}

	// Получаем информацию о достижении (для метаданных)
	// В реальном приложении эти данные должны сохраняться после CreateUserAchievement
	achievements, err := h.apiService.Achievement.GetAllAchievements(c.Context(), userID)
	if err != nil {
		log.Printf("ConfirmAchievementUpload: Failed to get achievements for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get achievement info",
		})
	}

	// Находим нужное достижение по имени
	var achievementMeta *models.AchievementMeta
	for _, achievement := range achievements.Achievements {
		if achievement.Name == achievementName {
			achievementMeta = &achievement
			break
		}
	}

	if achievementMeta == nil {
		log.Printf("ConfirmAchievementUpload: Achievement not found: %s", achievementName)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Achievement not found",
		})
	}

	// Добавляем метаданные с S3 ключом
	err = h.apiService.Achievement.AddAchievementMeta(c.Context(), achievementMeta, confirmReq.S3Key)
	if err != nil {
		log.Printf("ConfirmAchievementUpload: Failed to confirm upload for achievement %s, user %s: %v",
			achievementName, userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to confirm upload",
		})
	}

	log.Printf("ConfirmAchievementUpload: Successfully confirmed upload for achievement %s, user: %s",
		achievementName, userID)
	return c.JSON(fiber.Map{
		"message": "Achievement upload confirmed successfully",
	})
}

// GetAchievementDownloadUrl возвращает URL для скачивания достижения
// @Summary Получить URL для скачивания достижения
// @Description Возвращает presigned URL для скачивания файла достижения
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Название достижения"
// @Success 200 {object} models.AchievementUrl "URL для скачивания"
// @Failure 400 {object} models.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Достижение не найдено"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/achievements/{id}/download [get]
func (h *Handler) GetAchievementDownloadUrl(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	achievementName := c.Params("id")
	log.Printf("GetAchievementDownloadUrl: Getting download URL for achievement %s, user: %s",
		achievementName, userID)

	// Валидация
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("GetAchievementDownloadUrl: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	if achievementName == "" {
		log.Printf("GetAchievementDownloadUrl: Achievement name is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Achievement name is required",
		})
	}

	// Вызываем achievement service
	downloadUrl, err := h.apiService.Achievement.GetAchievementDownloadUrl(
		c.Context(),
		userID,
		achievementName,
	)
	if err != nil {
		log.Printf("GetAchievementDownloadUrl: Failed to get download URL for achievement %s, user %s: %v",
			achievementName, userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get download URL",
		})
	}

	log.Printf("GetAchievementDownloadUrl: Successfully retrieved download URL for achievement %s, user: %s",
		achievementName, userID)
	return c.JSON(downloadUrl)
}

// DeleteAchievement удаляет достижение
// @Summary Удалить достижение
// @Description Удаляет достижение пользователя по названию
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Название достижения"
// @Success 200 {object} map[string]string "Сообщение об успешном удалении"
// @Failure 400 {object} models.ErrorResponse "Неверные параметры запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Достижение не найдено"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /user/achievements/{id} [delete]
func (h *Handler) DeleteAchievement(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	achievementName := c.Params("id")
	log.Printf("DeleteAchievement: Deleting achievement %s for user: %s", achievementName, userID)

	// Валидация
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("DeleteAchievement: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	if achievementName == "" {
		log.Printf("DeleteAchievement: Achievement name is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Achievement name is required",
		})
	}

	// Вызываем achievement service
	err := h.apiService.Achievement.DeleteAchievement(c.Context(), userID, achievementName)
	if err != nil {
		log.Printf("DeleteAchievement: Failed to delete achievement %s for user %s: %v",
			achievementName, userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete achievement",
		})
	}

	log.Printf("DeleteAchievement: Successfully deleted achievement %s for user: %s",
		achievementName, userID)
	return c.JSON(fiber.Map{
		"message": "Achievement deleted successfully",
	})
}
