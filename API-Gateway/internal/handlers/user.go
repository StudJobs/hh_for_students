package handlers

import (
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
	"strconv"
)

// GetUsers возвращает список пользователей с пагинацией
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
	var profileList models.ProfileList
	profileList.Profiles = make([]models.User, len(profiles.Profiles))

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

	profileList.Pagination = models.PaginationResponse{
		Total:       profiles.Pagination.Total,
		Pages:       profiles.Pagination.Pages,
		CurrentPage: profiles.Pagination.CurrentPage,
	}

	log.Printf("GetUsers: Successfully retrieved %d users", len(profileList.Profiles))
	return c.JSON(profileList)
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	log.Printf("GetUsers: Getting me")
	userID := getUserIDFromContext(c)

	// Валидация UUID
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("GetUser: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Вызываем users service
	profile, err := h.apiService.User.GetUser(c.Context(), userID)
	if err != nil {
		log.Printf("GetUser: Failed to get user %s: %v", userID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Конвертируем в HTTP модель
	user := models.User{
		ID:                 uuid.MustParse(profile.Id),
		FirstName:          profile.FirstName,
		LastName:           profile.LastName,
		Age:                profile.Age,
		Tg:                 profile.Tg,
		Email:              profile.Email,
		Description:        profile.Description,
		ProfessionCategory: profile.ProfessionCategory,
	}

	// Обрабатываем resume_id
	if profile.ResumeId != "" {
		user.ResumeID = uuid.MustParse(profile.ResumeId)
	}

	log.Printf("GetUser: Successfully retrieved user: %s", userID)
	return c.JSON(user)

}

// GetUser возвращает профиль пользователя по ID
func (h *Handler) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	log.Printf("GetUser: Getting user with ID: %s", userID)

	// Валидация UUID
	if _, err := uuid.Parse(userID); err != nil {
		log.Printf("GetUser: Invalid UUID format: %s", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Вызываем users service
	profile, err := h.apiService.User.GetUser(c.Context(), userID)
	if err != nil {
		log.Printf("GetUser: Failed to get user %s: %v", userID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Конвертируем в HTTP модель
	user := models.User{
		ID:                 uuid.MustParse(profile.Id),
		FirstName:          profile.FirstName,
		LastName:           profile.LastName,
		Age:                profile.Age,
		Tg:                 profile.Tg,
		Email:              profile.Email,
		Description:        profile.Description,
		ProfessionCategory: profile.ProfessionCategory,
	}

	// Обрабатываем resume_id
	if profile.ResumeId != "" {
		user.ResumeID = uuid.MustParse(profile.ResumeId)
	}

	log.Printf("GetUser: Successfully retrieved user: %s", userID)
	return c.JSON(user)
}

// UpdateUser обновляет профиль пользователя (PATCH)
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
	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		log.Printf("UpdateUser: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Строим профиль для обновления (только указанные поля)
	profile := &usersv1.Profile{}

	if firstName, ok := updateData["first_name"].(string); ok && firstName != "" {
		profile.FirstName = firstName
	}
	if lastName, ok := updateData["last_name"].(string); ok && lastName != "" {
		profile.LastName = lastName
	}
	if age, ok := updateData["age"].(float64); ok && age > 0 {
		profile.Age = int32(age)
	}
	if tg, ok := updateData["tg"].(string); ok && tg != "" {
		profile.Tg = tg
	}
	if email, ok := updateData["email"].(string); ok && email != "" {
		profile.Email = email
	}
	if description, ok := updateData["description"].(string); ok && description != "" {
		profile.Description = description
	}
	if professionCategory, ok := updateData["profession_category"].(string); ok && professionCategory != "" {
		profile.ProfessionCategory = professionCategory
	}
	if resumeID, ok := updateData["resume_id"].(string); ok && resumeID != "" {
		profile.ResumeId = resumeID
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
	user := models.User{
		ID:                 uuid.MustParse(updatedProfile.Id),
		FirstName:          updatedProfile.FirstName,
		LastName:           updatedProfile.LastName,
		Age:                updatedProfile.Age,
		Tg:                 updatedProfile.Tg,
		Email:              updatedProfile.Email,
		Description:        updatedProfile.Description,
		ProfessionCategory: updatedProfile.ProfessionCategory,
	}

	// Обрабатываем resume_id
	if updatedProfile.ResumeId != "" {
		user.ResumeID = uuid.MustParse(updatedProfile.ResumeId)
	}

	log.Printf("UpdateUser: Successfully updated user: %s", userID)
	return c.JSON(user)
}

// DeleteUser удаляет профиль пользователя
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

	// удаляем пользователя при успешном удалении аккаунта
	if err := h.apiService.Auth.DeleteUser(c.Context(), userID); err != nil {
		log.Printf("DeleteUser: Failed to delete user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	log.Printf("DeleteUser: Successfully deleted user: %s", userID)
	return c.JSON(fiber.Map{
		"message": "Profile deleted successfully",
	})
}
