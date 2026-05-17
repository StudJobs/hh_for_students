package services

import (
	"context"
	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	applicationv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/application/v1"
	chatv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/chat/v1"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	skillsv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/skills/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"

	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
)

// AuthService интерфейс для работы с аутентификацией
type AuthService interface {
	Login(ctx context.Context, email, password, role string) (*models.AuthResponse, error)
	Register(ctx context.Context, email, password, role string) (*models.AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (bool, string, string, error)
	DeleteUser(ctx context.Context, userID string) error
}

// ExpertiseTest — облёгчённая HTTP-модель теста для проброса в Gateway.
type ExpertiseTestQuestion = struct {
	ID      int32    `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
}
type ExpertiseTestResp = struct {
	SkillSlug        string                  `json:"skill_slug"`
	Available        bool                    `json:"available"`
	Reason           string                  `json:"reason,omitempty"`
	Questions        []ExpertiseTestQuestion `json:"questions"`
	PassThresholdPct int32                   `json:"pass_threshold_pct"`
}
type ExpertiseSubmitResp = struct {
	Passed   bool   `json:"passed"`
	Correct  int32  `json:"correct"`
	Total    int32  `json:"total"`
	ScorePct int32  `json:"score_pct"`
	Message  string `json:"message"`
}

// UsersService интерфейс для работы с пользователями
type UsersService interface {
	CreateUser(ctx context.Context, req *usersv1.NewProfileRequest) (*usersv1.Profile, error)
	GetUser(ctx context.Context, userID string) (*usersv1.Profile, error)
	GetUsers(ctx context.Context, req *usersv1.GetAllProfilesRequest) (*usersv1.ProfileList, error)
	UpdateUser(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error)
	DeleteUser(ctx context.Context, userID string) error
	GetExpertiseTest(ctx context.Context, slug string) (*usersv1.ExpertiseTest, error)
	SubmitExpertiseTest(ctx context.Context, userID, slug string, answers []int32) (*usersv1.SubmitExpertiseTestResponse, error)
}

type AchievementService interface {
	GetAllAchievements(ctx context.Context, userID string) (*models.AchievementList, error)
	GetAchievementDownloadUrl(ctx context.Context, userID, achieveName string) (*models.AchievementUrl, error)
	GetAchievementUploadUrl(ctx context.Context, userID, achieveName, fileName, fileType string, fileSize int64) (*models.UploadUrlResponse, error)
	AddAchievementMeta(ctx context.Context, meta *models.AchievementMeta, s3Key string) error
	DeleteAchievement(ctx context.Context, userID, achieveName string) error
	SubmitForReview(ctx context.Context, userUUID string, achievementID int64) error
	GetExpertQueue(ctx context.Context, page, limit int32) (*models.AchievementList, error)
	ReviewAchievement(ctx context.Context, achievementID int64, reviewerUUID string, decision int32, comment string) error
}

type CompanyService interface {
	CreateCompany(ctx context.Context, company *models.Company) (*models.Company, error)
	GetCompany(ctx context.Context, id string) (*models.Company, error)
	GetAllCompanies(ctx context.Context, pagination *models.Pagination, city, companyType, query string) (*models.CompanyList, error)
	UpdateCompany(ctx context.Context, id string, company *models.Company) (*models.Company, error)
	DeleteCompany(ctx context.Context, id string) error

	// HR-membership
	ApplyMembership(ctx context.Context, companyID, userID, note string) (*models.CompanyMember, error)
	ReviewMembership(ctx context.Context, membershipID string, status int32) (*models.CompanyMember, error)
	ListMembers(ctx context.Context, companyID string, status int32) ([]*models.CompanyMember, error)
	GetMembershipByUser(ctx context.Context, userID string) (*models.CompanyMember, error)
	ListMembershipsByUser(ctx context.Context, userID string, status int32) ([]*models.CompanyMember, error)
}

type SkillsService interface {
	Search(ctx context.Context, query string, category int32, limit int32) ([]*models.Skill, error)
	Popular(ctx context.Context, category int32, limit int32) ([]*models.Skill, error)
	Bulk(ctx context.Context, slugs []string) ([]*models.Skill, error)
}

// SearchService — фасад над Elasticsearch-сервисом.
// Available() возвращает false, если Search-сервис не сконфигурирован — Gateway упадёт обратно на SQL-фильтр.
type SearchService interface {
	Available() bool
	SearchProfiles(ctx context.Context, query string, skillSlugs []string, professionCategory string, page, limit int32) (*usersv1.ProfileList, error)
	SearchVacancies(ctx context.Context, query string, skillSlugs []string, salaryMin, experienceMax int32, companyID string, page, limit int32) (*vacancyv1.VacancyList, error)
	// SearchVacanciesAsModel — то же, что SearchVacancies, но возвращает HTTP-модель.
	SearchVacanciesAsModel(ctx context.Context, query string, skillSlugs []string, salaryMin, experienceMax int32, companyID string, page, limit int32) (*models.VacancyList, error)
	// SearchMicroTasksAsModel ищет микрозадачи в ES и возвращает HTTP-модель.
	SearchMicroTasksAsModel(ctx context.Context, query string, skillSlugs []string, rewardMin int32, status int32, companyID string, page, limit int32) (*models.MicroTaskList, error)
}

// MicroTaskService — обёртка над gRPC-клиентом микросервиса MicroTasks.
type MicroTaskService interface {
	Available() bool
	Create(ctx context.Context, t *models.MicroTask) (*models.MicroTask, error)
	Update(ctx context.Context, id string, t *models.MicroTask) (*models.MicroTask, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*models.MicroTask, error)
	List(ctx context.Context, status int32, skillSlugs []string, page, limit int32) (*models.MicroTaskList, error)
	ListByCompany(ctx context.Context, companyID string, page, limit int32) (*models.MicroTaskList, error)
	ListByStudent(ctx context.Context, studentID string, status int32, page, limit int32) (*models.MicroTaskList, error)
	Apply(ctx context.Context, taskID, studentID string) (*models.MicroTask, error)
	Submit(ctx context.Context, taskID, studentID, solutionURL, comment, fileName string) (*models.Submission, error)
	SolutionUploadInit(ctx context.Context, taskID, studentID, fileName string) (fileID, uploadURL string, err error)
	SolutionUploadConfirm(ctx context.Context, taskID, studentID, fileID string) error
	CreateSkillQuest(ctx context.Context, expertID, studentID, slug, title, description, deadline string) (*models.MicroTask, error)
	ListSubmissions(ctx context.Context, taskID, studentID string, page, limit int32) (*models.SubmissionList, error)
	Review(ctx context.Context, submissionID string, status int32, reviewComment string) (*models.Submission, error)
}

type ApplicationService interface {
	Available() bool
	Apply(ctx context.Context, vacancyID, studentID, coverLetter string) (*models.Application, error)
	Withdraw(ctx context.Context, id, studentID string) error
	ListMine(ctx context.Context, studentID string, status int32, page, limit int32) (*models.ApplicationList, error)
	ListForVacancy(ctx context.Context, vacancyID string, status int32, page, limit int32) (*models.ApplicationList, error)
	UpdateStatus(ctx context.Context, id string, status int32, hrComment string) (*models.Application, error)
	Get(ctx context.Context, id string) (*models.Application, error)
	AssignHR(ctx context.Context, id, hrUserID string) (*models.Application, error)
}

type VacancyService interface {
	CreateVacancy(ctx context.Context, vacancy *models.Vacancy) (*models.Vacancy, error)
	GetVacancy(ctx context.Context, id string) (*models.Vacancy, error)
	GetAllVacancies(ctx context.Context, pagination *models.Pagination,
		companyID, positionStatus, workFormat, schedule string,
		minSalary, maxSalary, minExperience, maxExperience int32,
		searchTitle string) (*models.VacancyList, error)
	GetHRVacancies(ctx context.Context, pagination *models.Pagination,
		companyID, positionStatus, workFormat, schedule string,
		minSalary, maxSalary, minExperience, maxExperience int32,
		searchTitle string) (*models.VacancyList, error)
	UpdateVacancy(ctx context.Context, id string, vacancy *models.Vacancy) (*models.Vacancy, error)
	DeleteVacancy(ctx context.Context, id string) error
	GetAllPositions(ctx context.Context) ([]string, error)
	ModerateVacancy(ctx context.Context, id string, status int32, comment string) (*models.Vacancy, error)
}

type ChatService interface {
	SendMessage(ctx context.Context, threadID, fromUser, body string) (*models.ChatMessage, error)
	ListMessages(ctx context.Context, threadID string, page, limit int32) (*models.ChatMessageList, error)
	ListUserThreads(ctx context.Context, userID string, limit int32) ([]*models.ChatThread, error)
	EditMessage(ctx context.Context, messageID, fromUser, body string) (*models.ChatMessage, error)
	HideThread(ctx context.Context, userID, threadID string) error
	ListHiddenThreads(ctx context.Context, userID string) ([]string, error)
}

// ApiGateway объединяет все сервисы
type ApiGateway struct {
	Auth        AuthService
	User        UsersService
	Achievement AchievementService
	Company     CompanyService
	Vacancy     VacancyService
	Application ApplicationService
	Skills      SkillsService
	Search      SearchService
	MicroTasks  MicroTaskService
	Chat        ChatService
}

// NewApiGateway создает новый экземпляр ApiGateway
func NewApiGateway(
	authClient authv1.AuthServiceClient,
	usersClient usersv1.UsersServiceClient,
	achievementClient achievementv1.AchievementServiceClient,
	companyClient companyv1.CompanyServiceClient,
	vacancyClient vacancyv1.VacancyServiceClient,
	applicationClient applicationv1.ApplicationServiceClient,
	skillsClient skillsv1.SkillsServiceClient,
	searchClient searchv1.SearchServiceClient,
	microtasksClient microtaskv1.MicroTaskServiceClient,
	chatClient chatv1.ChatServiceClient,
) *ApiGateway {
	return &ApiGateway{
		Auth:        NewAuthService(authClient),
		User:        NewUsersService(usersClient),
		Achievement: NewAchievementService(achievementClient),
		Company:     NewCompanyService(companyClient),
		Vacancy:     NewVacancyService(vacancyClient),
		Application: NewApplicationService(applicationClient),
		Skills:      NewSkillsService(skillsClient),
		Search:      NewSearchService(searchClient),
		MicroTasks:  NewMicroTaskService(microtasksClient),
		Chat:        NewChatService(chatClient),
	}
}
