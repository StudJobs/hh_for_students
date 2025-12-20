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
	files.Get("/:entity_id/:file_name", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.ServeFileDirect)

	// === User routes ===
	users := api.Group("/users")
	// ПРАВИЛЬНО: Middleware идут до обработчика
	users.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.GetUsers)
	users.Get("/me", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.GetMe)
	users.Get("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.GetUser)
	// Для /edit нет параметра :id, поэтому используем RoleMiddleware.
	// Проверка, что юзер меняет именно себя, должна быть внутри h.UpdateUser.
	users.Patch("/edit", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.UpdateUser)
	users.Delete("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.DeleteUser)

	// === User File routes ===
	userFiles := users.Group("/files")
	// Здесь нет параметра :id, подразумевается, что юзер загружает файлы для себя.
	userFiles.Post("/avatar", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.UploadUserAvatar)
	userFiles.Post("/resume", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.UploadUserResume)
	userFiles.Delete("/avatar", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.DeleteUserAvatar)
	userFiles.Delete("/resume", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.DeleteUserResume)

	// === User Achievement routes ===
	userAchievement := api.Group("/user/achievements")
	userAchievement.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.GetUserAchievements)
	userAchievement.Post("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.CreateUserAchievement)                                    // Нет :id
	userAchievement.Post("/:id/confirm", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT), h.ConfirmAchievementUpload)           // Есть :id
	userAchievement.Get("/:id/download", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.GetAchievementDownloadUrl) // Есть :id
	userAchievement.Delete("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT), h.DeleteAchievement)                        // Есть :id

	// === HR routes ===
	profileHR := api.Group("/hr")
	profileHR.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.GetUsers)
	profileHR.Get("/me", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.GetMe)
	profileHR.Patch("/edit", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.UpdateUser)
	profileHR.Delete("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.DeleteUser)

	// === HR vacancy routes ===
	HRVacancy := profileHR.Group("/vacancy")
	// Подразумевается, что HR работает со своими вакансиями. Проверка на владение - внутри обработчиков.
	HRVacancy.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.GetHRVacancies)
	HRVacancy.Get("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.GetVacancy)
	HRVacancy.Post("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.CreateHRVacancy)
	HRVacancy.Patch("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.UpdateVacancy)
	HRVacancy.Delete("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.DeleteVacancy)

	api.Get("/positions", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.GetPositions)

	// === Vacancy routes ===
	vacancy := api.Group("/vacancy")
	vacancy.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.GetVacancies)
	vacancy.Get("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.GetVacancy)

	// === Vacancy File routes ===
	vacancyFiles := vacancy.Group("/:id/files")
	vacancyFiles.Post("/attachment", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.UploadVacancyAttachment)
	vacancyFiles.Delete("/attachment", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.DeleteVacancyAttachment)

	// === Company ===
	company := api.Group("/company")
	company.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.GetCompanies)
	company.Get("/me", RoleMiddleware(ROLE_DEVELOPER, ROLE_COMPANY), h.GetCompanyMe)
	company.Get("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.GetCompanyByID)
	company.Patch("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_COMPANY), h.UpdateCompany)  // Нет :id
	company.Delete("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_COMPANY), h.DeleteCompany) // Нет :id

	// === Company File routes ===
	companyFiles := company.Group("/:id/files")
	companyFiles.Post("/logo", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY), h.UploadCompanyLogo)
	companyFiles.Post("/documents", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY), h.UploadCompanyDocument)
	companyFiles.Delete("/logo", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_COMPANY), h.DeleteCompanyLogo)
}

func (h *Handler) roleConvert(userRole Role) string {
	switch userRole {
	case ROLE_COMPANY:
		return "ROLE_COMPANY_OWNER"
	case ROLE_DEVELOPER:
		return "ROLE_DEVELOPER"
	case ROLE_STUDENT:
		return "ROLE_STUDENT"
	case ROLE_HR:
		return "ROLE_EMPLOYER"
	default:
		return "ROLE_STUDENT"
	}
}
