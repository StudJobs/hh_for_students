package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

// SearchSkills возвращает список тегов компетенций по подстроке
// @Summary Поиск компетенций
// @Description Поиск тегов компетенций по подстроке slug или name (для автокомплита)
// @Tags Skills
// @Produce json
// @Security BearerAuth
// @Param q query string false "Подстрока для поиска"
// @Param category query int false "Фильтр по категории" minimum(0) maximum(8)
// @Param limit query int false "Количество элементов" default(20) minimum(1) maximum(100)
// @Success 200 {array} models.Skill
// @Failure 500 {object} models.Error
// @Router /skills/search [get]
func (h *Handler) SearchSkills(c *fiber.Ctx) error {
	query := c.Query("q", "")
	category, _ := strconv.Atoi(c.Query("category", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	skills, err := h.apiService.Skills.Search(c.Context(), query, int32(category), int32(limit))
	if err != nil {
		log.Printf("SearchSkills: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to search skills",
		})
	}
	return c.JSON(skills)
}

// PopularSkills возвращает топ популярных тегов
// @Summary Популярные компетенции
// @Description Возвращает топ популярных тегов (для подсказок при пустом вводе)
// @Tags Skills
// @Produce json
// @Security BearerAuth
// @Param category query int false "Фильтр по категории" minimum(0) maximum(8)
// @Param limit query int false "Количество элементов" default(20) minimum(1) maximum(100)
// @Success 200 {array} models.Skill
// @Failure 500 {object} models.Error
// @Router /skills/popular [get]
func (h *Handler) PopularSkills(c *fiber.Ctx) error {
	category, _ := strconv.Atoi(c.Query("category", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	skills, err := h.apiService.Skills.Popular(c.Context(), int32(category), int32(limit))
	if err != nil {
		log.Printf("PopularSkills: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to fetch popular skills",
		})
	}
	return c.JSON(skills)
}

// BulkSkills возвращает теги по списку slug-ов
// @Summary Получить теги по slug-ам
// @Description Возвращает информацию о тегах по списку slug-ов (для денормализации в Profile/Vacancy)
// @Tags Skills
// @Produce json
// @Security BearerAuth
// @Param slugs query string true "Список slug-ов через запятую"
// @Success 200 {array} models.Skill
// @Failure 400 {object} models.Error
// @Failure 500 {object} models.Error
// @Router /skills/bulk [get]
func (h *Handler) BulkSkills(c *fiber.Ctx) error {
	raw := c.Query("slugs", "")
	if raw == "" {
		return c.JSON([]*models.Skill{})
	}
	parts := strings.Split(raw, ",")
	slugs := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			slugs = append(slugs, t)
		}
	}

	skills, err := h.apiService.Skills.Bulk(c.Context(), slugs)
	if err != nil {
		log.Printf("BulkSkills: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to fetch skills",
		})
	}
	return c.JSON(skills)
}
