package services

import (
	"context"
	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
)

// AuthService интерфейс для работы с аутентификацией
type AuthService interface {
	Login(ctx context.Context, email, password, role string) (*models.AuthResponse, error)
	Register(ctx context.Context, email, password, role string) (*models.AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (bool, string, string, error)
}

// UsersService интерфейс для работы с пользователями
type UsersService interface {
	CreateUser(ctx context.Context, req *usersv1.NewProfileRequest) (*usersv1.Profile, error)
	GetUser(ctx context.Context, userID string) (*usersv1.Profile, error)
	GetUsers(ctx context.Context, pagination *usersv1.GetAllProfilesRequest) (*usersv1.ProfileList, error)
	UpdateUser(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error)
	DeleteUser(ctx context.Context, userID string) error
}

type AchievementService interface {
	GetAllAchievements(ctx context.Context, userID string) (*models.AchievementList, error)
	GetAchievementDownloadUrl(ctx context.Context, userID, achieveName string) (*models.AchievementUrl, error)
	GetAchievementUploadUrl(ctx context.Context, userID, achieveName, fileName, fileType string) (*models.AchievementUrl, error)
	AddAchievementMeta(ctx context.Context, meta *models.AchievementMeta) error
	DeleteAchievement(ctx context.Context, userID, achieveName string) error
}

// ApiGateway объединяет все сервисы
type ApiGateway struct {
	Auth        AuthService
	User        UsersService
	Achievement AchievementService
}

// NewApiGateway создает новый экземпляр ApiGateway
func NewApiGateway(
	authClient authv1.AuthServiceClient,
	usersClient usersv1.UsersServiceClient,
	achievementClient achievementv1.AchievementServiceClient,
) *ApiGateway {
	return &ApiGateway{
		Auth:        NewAuthService(authClient),
		User:        NewUsersService(usersClient),
		Achievement: NewAchievementService(achievementClient),
	}
}
