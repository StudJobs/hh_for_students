package handlers

import (
	apigatewayV1 "github.com/StudJobs/proto_srtucture/gen/go/proto/apigateway/v1"
	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
	"time"
)

type Handler struct {
	app           *fiber.App
	gatewayClient apigatewayV1.ApiGatewayServiceClient
	apiService    *services.ServiceAPIGateway
}

func NewHandler(gatewayClient apigatewayV1.ApiGatewayServiceClient, apiService *services.ServiceAPIGateway) *Handler {
	return &Handler{
		gatewayClient: gatewayClient,
		apiService:    apiService,
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
	h.app.Use(AuthMiddleware(h.gatewayClient))

	h.initRoutes()

	return h.app
}

func (h *Handler) initRoutes() {
	api := h.app.Group("/api/v1")

	// === Auth routes ===
	auth := api.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)

	// === User routes ===
	users := api.Group("/users")
	users.Get("/", h.GetUsers, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))

	profileUser := users.Group("/:id")
	profileUser.Get("/", h.GetUser, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	profileUser.Patch("/edit", h.UpdateUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	profileUser.Delete("/", h.DeleteUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))

	// === User Achievement routes ===
	userAchievement := profileUser.Group("/achievement")
	userAchievement.Get("/", h.GetUserAchievements, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	userAchievement.Get("/id", h.CreateUserAchievement, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	userAchievement.Post("/", h.CreateUserAchievement, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	userAchievement.Delete("/id", h.DeleteAchievement, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))

	// === HR routes ===
	profileHR := api.Group("/hr/:id")
	profileHR.Delete("/", h.DeleteUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))

	// === HR vacancy routes ===
	HRVacancy := api.Group("/vacancy")
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
