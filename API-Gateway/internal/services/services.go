package services

import (
	"context"
	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
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

// UsersService интерфейс для работы с пользователями
type UsersService interface {
	CreateUser(ctx context.Context, req *usersv1.NewProfileRequest) (*usersv1.Profile, error)
	GetUser(ctx context.Context, userID string) (*usersv1.Profile, error)
	GetUsers(ctx context.Context, req *usersv1.GetAllProfilesRequest) (*usersv1.ProfileList, error)
	UpdateUser(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error)
	DeleteUser(ctx context.Context, userID string) error
}

type AchievementService interface {
	GetAllAchievements(ctx context.Context, userID string) (*models.AchievementList, error)
	GetAchievementDownloadUrl(ctx context.Context, userID, achieveName string) (*models.AchievementUrl, error)
	GetAchievementUploadUrl(ctx context.Context, userID, achieveName, fileName, fileType string, fileSize int64) (*models.UploadUrlResponse, error)
	AddAchievementMeta(ctx context.Context, meta *models.AchievementMeta, s3Key string) error
	DeleteAchievement(ctx context.Context, userID, achieveName string) error
}

type CompanyService interface {
	CreateCompany(ctx context.Context, company *models.Company) (*models.Company, error)
	GetCompany(ctx context.Context, id string) (*models.Company, error)
	GetAllCompanies(ctx context.Context, pagination *models.Pagination, city, companyType string) (*models.CompanyList, error)
	UpdateCompany(ctx context.Context, id string, company *models.Company) (*models.Company, error)
	DeleteCompany(ctx context.Context, id string) error
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
}

// ApiGateway объединяет все сервисы
type ApiGateway struct {
	Auth        AuthService
	User        UsersService
	Achievement AchievementService
	Company     CompanyService
	Vacancy     VacancyService
}

// NewApiGateway создает новый экземпляр ApiGateway
func NewApiGateway(
	authClient authv1.AuthServiceClient,
	usersClient usersv1.UsersServiceClient,
	achievementClient achievementv1.AchievementServiceClient,
	companyClient companyv1.CompanyServiceClient,
	vacancyClient vacancyv1.VacancyServiceClient,
) *ApiGateway {
	return &ApiGateway{
		Auth:        NewAuthService(authClient),
		User:        NewUsersService(usersClient),
		Achievement: NewAchievementService(achievementClient),
		Company:     NewCompanyService(companyClient),
		Vacancy:     NewVacancyService(vacancyClient),
	}
}
