package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/cache"
	"github.com/studjobs/hh_for_students/api-gateway/internal/metrics"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
	"github.com/studjobs/hh_for_students/api-gateway/internal/utils"
	"log"
	"strings"
	"time"

	swagger "github.com/arsmn/fiber-swagger/v2"
)

type Handler struct {
	app          *fiber.App
	apiService   *services.ApiGateway
	fileHandler  *utils.FileHandler
	cacheClient  *cache.Client
	rateLimiter  *RateLimiter
}

// NewHandler создает новый экземпляр Handler.
// cacheClient — может быть nil (тогда middleware no-op'ит).
// rateLimiter — может быть nil (тогда не применяется).
func NewHandler(apiService *services.ApiGateway, cacheClient *cache.Client, rateLimiter *RateLimiter) *Handler {
	log.Printf("Creating new Handler")
	return &Handler{
		apiService:  apiService,
		fileHandler: utils.NewFileHandler(apiService),
		cacheClient: cacheClient,
		rateLimiter: rateLimiter,
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
	h.app.Use(metrics.HTTPMiddleware())
	// Rate limit идёт ПЕРЕД auth: на /auth/login тоже распространяется. Иначе
	// botnet может бесконечно проверять пароли без ограничений.
	if h.rateLimiter != nil {
		h.app.Use(h.rateLimiter.Middleware())
	}
	h.app.Use(AuthMiddleware(h.apiService))
	// Cache идёт ПОСЛЕ auth — чтобы middleware видел уже-проверенные запросы и
	// 401/403 не попадали в кэш как валидные ответы.
	if h.cacheClient != nil && h.cacheClient.Enabled() {
		h.app.Use(CacheMiddleware(h.cacheClient))
	}

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
	files.Get("/:entity_id/:file_name", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.ServeFileDirect)

	// === User routes ===
	users := api.Group("/users")
	// ПРАВИЛЬНО: Middleware идут до обработчика
	users.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.GetUsers)
	users.Get("/me", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.GetMe)
	users.Get("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.GetUser)
	users.Get("/:id/achievements", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.GetUserAchievementsByID)
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
	userAchievement.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_EXPERT), h.GetUserAchievements)
	userAchievement.Post("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.CreateUserAchievement)                                                // Нет :id
	userAchievement.Post("/:id/confirm", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT), h.ConfirmAchievementUpload)                       // Есть :id (имя достижения)
	userAchievement.Get("/:id/download", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_EXPERT), h.GetAchievementDownloadUrl) // Есть :id (имя)
	userAchievement.Delete("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_STUDENT), h.DeleteAchievement)                                    // Есть :id (имя)
	userAchievement.Post("/:id/submit", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.SubmitAchievementForReview)                                 // :id — числовой achievement ID

	// === Expert review routes ===
	expert := api.Group("/expert")
	expert.Get("/queue", RoleMiddleware(ROLE_DEVELOPER, ROLE_EXPERT), h.GetExpertQueue)
	expert.Post("/achievements/:id/review", RoleMiddleware(ROLE_DEVELOPER, ROLE_EXPERT), h.ReviewAchievement)
	expert.Post("/quests", RoleMiddleware(ROLE_DEVELOPER, ROLE_EXPERT), h.CreateSkillQuest)

	// === Chat (минимальный polling, без WS) ===
	// thread_id строится из двух path-сегментов: /chat/<kind>/<resource_uuid>.
	// Двоеточие в URL Fiber не декодирует надёжно — поэтому kind и id отдельно.
	chat := api.Group("/chat")
	chat.Get("/:kind/:rid", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.GetChatMessages)
	chat.Post("/:kind/:rid", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.SendChatMessage)

	// === HR routes ===
	profileHR := api.Group("/hr")
	profileHR.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.GetUsers)
	profileHR.Get("/me", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.GetMe)
	profileHR.Patch("/edit", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.UpdateUser)
	profileHR.Delete("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR), h.DeleteUser)

	// === HR vacancy routes ===
	// До появления HR-membership flow единственный «аутентифицированный HR» — это владелец
	// компании (ROLE_COMPANY). ROLE_HR пускаем в whitelist, но в самом обработчике
	// блокируем (см. CreateHRVacancy) — иначе любой HR публиковал бы вакансии под чужой
	// company_id (была B4: дыра доступа).
	HRVacancy := profileHR.Group("/vacancy")
	HRVacancy.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.GetHRVacancies)
	HRVacancy.Get("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.GetVacancy)
	HRVacancy.Post("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.CreateHRVacancy)
	HRVacancy.Patch("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.UpdateVacancy)
	HRVacancy.Delete("/:id", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.DeleteVacancy)

	api.Get("/positions", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.GetPositions)

	// === Vacancy routes ===
	vacancy := api.Group("/vacancy")
	vacancy.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.GetVacancies)
	vacancy.Get("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR), h.GetVacancy)
	// Студент откликается на вакансию (cover_letter опционален).
	vacancy.Post("/:id/respond", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.RespondToVacancy)

	// === Vacancy File routes ===
	vacancyFiles := vacancy.Group("/:id/files")
	vacancyFiles.Post("/attachment", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.UploadVacancyAttachment)
	vacancyFiles.Delete("/attachment", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR), h.DeleteVacancyAttachment)

	// === Применения (отклики на вакансии) ===
	// Студенческая часть.
	userApplications := api.Group("/user/applications")
	userApplications.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.ListMyApplications)
	userApplications.Delete("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.WithdrawApplication)

	// HR-часть: список откликов на конкретную вакансию + accept/reject.
	HRVacancy.Get("/:id/applications", OwnerOrRoleMiddleware(ID, ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.ListVacancyApplications)
	profileHR.Patch("/applications/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.ReviewApplication)

	// === Skills (справочник тегов компетенций) ===
	skills := api.Group("/skills")
	skills.Get("/search", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.SearchSkills)
	skills.Get("/popular", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.PopularSkills)
	skills.Get("/bulk", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY, ROLE_EXPERT), h.BulkSkills)

	// === MicroTasks: студенческие операции ===
	tasks := api.Group("/tasks")
	tasks.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.GetTasks)
	tasks.Get("/mine", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.GetMyTasks)
	tasks.Get("/my-submissions", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.ListMySubmissions)
	tasks.Get("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT, ROLE_HR, ROLE_COMPANY), h.GetTask)
	tasks.Post("/:id/apply", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.ApplyToTask)
	tasks.Post("/:id/submit", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.SubmitTask)
	tasks.Post("/:id/solution-upload-init", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.SolutionUploadInit)
	tasks.Post("/:id/solution-upload-confirm", RoleMiddleware(ROLE_DEVELOPER, ROLE_STUDENT), h.SolutionUploadConfirm)

	// === MicroTasks: HR-операции ===
	hrTasks := profileHR.Group("/tasks")
	hrTasks.Get("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.GetHRTasks)
	hrTasks.Post("/", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.CreateHRTask)
	hrTasks.Patch("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.UpdateHRTask)
	hrTasks.Delete("/:id", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.DeleteHRTask)
	hrTasks.Get("/:id/submissions", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.ListTaskSubmissions)
	hrTasks.Post("/submissions/:submission_id/review", RoleMiddleware(ROLE_DEVELOPER, ROLE_HR, ROLE_COMPANY), h.ReviewSubmission)

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

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// normalizePagination clamps page/limit query params to safe ranges.
func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}
	return page, limit
}

// splitCSV разбивает строку через запятую на список slug-ов, удаляя пустые элементы.
func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
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
	case ROLE_EXPERT:
		return "ROLE_EXPERT"
	default:
		return "ROLE_STUDENT"
	}
}
