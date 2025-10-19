package services

import (
	"context"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
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
	GetUsers(ctx context.Context, pagination *commonv1.Pagination) (*usersv1.ProfileList, error)
	UpdateUser(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error)
	DeleteUser(ctx context.Context, userID string) error
}

// ApiGateway объединяет все сервисы
type ApiGateway struct {
	Auth AuthService
	User UsersService
}

// NewApiGateway создает новый экземпляр ApiGateway
func NewApiGateway(authClient authv1.AuthServiceClient, usersClient usersv1.UsersServiceClient) *ApiGateway {
	return &ApiGateway{
		Auth: NewAuthService(authClient),
		User: NewUsersService(usersClient),
	}
}
