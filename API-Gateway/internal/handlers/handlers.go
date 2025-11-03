package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
	"log"
	"time"
)

type Handler struct {
	app        *fiber.App
	apiService *services.ApiGateway
}

// NewHandler создает новый экземпляр Handler
func NewHandler(apiService *services.ApiGateway) *Handler {
	log.Printf("Creating new Handler")
	return &Handler{
		apiService: apiService,
	}
}
func (h *Handler) Init() *fiber.App {

	h.app = fiber.New(fiber.Config{
		AppName:       "API Gateway",
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   30 * time.Second,
		Prefork:       false, // TODO: read
		CaseSensitive: true,
		StrictRouting: false,
	})
	h.app.Use(AuthMiddleware(h.apiService))

	h.initRoutes()

	return h.app
}

func (h *Handler) initRoutes() {
	api := h.app.Group("/api/v1")

	// === Auth routes ===
	auth := api.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)
	auth.Post("/logout", h.Logout)

	// === User routes ===
	users := api.Group("/users")
	users.Get("/", h.GetUsers, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	users.Get("/me", h.GetMe, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT))
	users.Get("/:id", h.GetUser, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	users.Patch("/edit", h.UpdateUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	users.Delete("/", h.DeleteUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))

	// === User Achievement routes ===
	userAchievement := users.Group("/achievement")
	userAchievement.Get("/", h.GetUserAchievements, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	userAchievement.Get("/:id", h.CreateUserAchievement, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	userAchievement.Post("/", h.CreateUserAchievement, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	userAchievement.Delete("/:id", h.DeleteAchievement, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))

	// === HR routes ===
	profileHR := api.Group("/hr")
	profileHR.Get("/me", h.GetMe, RoleMiddleware(ROLE_DEVELOPER, ROLE_HR))
	profileHR.Patch("/edit", h.UpdateUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))
	profileHR.Delete("/", h.DeleteUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))

	// === HR vacancy routes ===
	HRVacancy := profileHR.Group("/vacancy")
	HRVacancy.Get("/", h.GetHRVacancies, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))
	HRVacancy.Get("/:id", h.GetVacancy, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))
	HRVacancy.Post("/", h.CreateHRVacancy, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))
	HRVacancy.Patch("/:id", h.UpdateVacancy, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))
	HRVacancy.Delete("/:id", h.DeleteVacancy, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))

	//  === Vacancy routes ===
	vacancy := api.Group("/vacancy")
	vacancy.Get("/", h.GetVacancies, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	vacancy.Get("/:id", h.GetVacancy, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))

}
