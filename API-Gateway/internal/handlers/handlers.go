package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
	"github.com/studjobs/hh_for_students/api-gateway/internal/utils"
	"log"
	"time"

	swagger "github.com/arsmn/fiber-swagger/v2"
)

type Handler struct {
	app         *fiber.App
	apiService  *services.ApiGateway
	fileHandler *utils.FileHandler
}

// NewHandler создает новый экземпляр Handler
func NewHandler(apiService *services.ApiGateway) *Handler {
	log.Printf("Creating new Handler")
	return &Handler{
		apiService:  apiService,
		fileHandler: utils.NewFileHandler(apiService),
	}
}

func (h *Handler) Init() *fiber.App {
	h.app = fiber.New(fiber.Config{
		AppName:       "API Gateway",
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   30 * time.Second,
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: false,
	})
	h.app.Use(AuthMiddleware(h.apiService))

	// Swagger документация
	h.app.Get("/swagger/*", swagger.HandlerDefault)

	h.initRoutes()

	return h.app
}

func (h *Handler) initRoutes() {
	api := h.app.Group("/api/v1")

	// === Auth routes ===
	auth := api.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)

	// === File routes ===
	files := api.Group("/files")
	files.Get("/:entity_id/:file_name", h.ServeFileDirect)

	// === User routes ===
	users := api.Group("/users")
	users.Get("/", h.GetUsers, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY))
	users.Get("/me", h.GetMe, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT))
	users.Get("/:id", h.GetUser, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	users.Patch("/edit", h.UpdateUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	users.Delete("/", h.DeleteUser, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))

	// === User File routes ===
	userFiles := users.Group("/files")
	userFiles.Post("/avatar", h.UploadUserAvatar, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY))
	userFiles.Post("/resume", h.UploadUserResume, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	userFiles.Delete("/avatar", h.DeleteUserAvatar, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY))
	userFiles.Delete("/resume", h.DeleteUserResume, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))

	// === User Achievement routes ===
	userAchievement := api.Group("/user/achievements")
	userAchievement.Get("/", h.GetUserAchievements, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	userAchievement.Post("/", h.CreateUserAchievement, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	userAchievement.Post("/:id/confirm", h.ConfirmAchievementUpload, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT))
	userAchievement.Get("/:id/download", h.GetAchievementDownloadUrl, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
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

	api.Get("/positions", h.GetPositions, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))

	// === Vacancy routes ===
	vacancy := api.Group("/vacancy")
	vacancy.Get("/", h.GetVacancies, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))
	vacancy.Get("/:id", h.GetVacancy, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR))

	// === Vacancy File routes ===
	vacancyFiles := vacancy.Group("/:id/files")
	vacancyFiles.Post("/attachment", h.UploadVacancyAttachment, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))
	vacancyFiles.Delete("/attachment", h.DeleteVacancyAttachment, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR))

	// === Company ===
	company := api.Group("/company")
	company.Get("/", h.GetCompanies, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY))
	company.Get("/me", h.GetCompanyMe, RoleMiddleware(ROLE_DEVELOPER, ROLE_COMPANY))
	company.Get("/:id", h.GetCompanyByID, RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY))
	company.Patch("/", h.UpdateCompany, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY))
	company.Delete("/", h.DeleteCompany, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY))

	// === Company File routes ===
	companyFiles := company.Group("/:id/files")
	companyFiles.Post("/logo", h.UploadCompanyLogo, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY))
	companyFiles.Post("/documents", h.UploadCompanyDocument, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY))
	companyFiles.Delete("/logo", h.DeleteCompanyLogo, OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY))
}
