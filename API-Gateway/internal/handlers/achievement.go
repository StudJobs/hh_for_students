package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
)

// GetUserAchievements возвращает все достижения пользователя
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

// CreateUserAchievement создает новое достижение
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
	var achievementMeta models.AchievementMeta
	if err := c.BodyParser(&achievementMeta); err != nil {
		log.Printf("CreateUserAchievement: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Валидация обязательных полей
	if achievementMeta.Name == "" || achievementMeta.FileName == "" || achievementMeta.FileType == "" {
		log.Printf("CreateUserAchievement: Missing required fields")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, file_name and file_type are required",
		})
	}

	// Устанавливаем userID из контекста
	achievementMeta.UserUUID = userID

	// Получаем URL для загрузки
	uploadUrl, err := h.apiService.Achievement.GetAchievementUploadUrl(
		c.Context(),
		userID,
		achievementMeta.Name,
		achievementMeta.FileName,
		achievementMeta.FileType,
	)
	if err != nil {
		log.Printf("CreateUserAchievement: Failed to get upload URL for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get upload URL",
		})
	}

	// Добавляем метаданные
	if err := h.apiService.Achievement.AddAchievementMeta(c.Context(), &achievementMeta); err != nil {
		log.Printf("CreateUserAchievement: Failed to add achievement meta for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create achievement",
		})
	}

	// Получаем URL для скачивания
	downloadUrl, err := h.apiService.Achievement.GetAchievementDownloadUrl(
		c.Context(),
		userID,
		achievementMeta.Name,
	)
	if err != nil {
		log.Printf("CreateUserAchievement: Failed to get download URL for user %s: %v", userID, err)
		// Не прерываем выполнение, т.к. основная операция выполнена
	}

	// Формируем полный ответ
	achievementData := models.AchievementData{
		Meta:        achievementMeta,
		UploadURL:   *uploadUrl,
		DownloadURL: *downloadUrl,
	}

	log.Printf("CreateUserAchievement: Successfully created achievement for user: %s", userID)
	return c.JSON(achievementData)
}

// GetAchievementDownloadUrl возвращает URL для скачивания достижения
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
