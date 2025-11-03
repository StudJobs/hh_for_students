package service

import (
	"context"
	"errors"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	"github.com/studjobs/hh_for_students/auth/internal/repository"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrUserLoggedOut      = errors.New("user has been logged out")
)

type ITokenManager interface {
	GenerateToken(userUUID, email string, role authv1.Role) (string, error)
	ValidateToken(token string) (string, authv1.Role, error)
}

type IAuthService interface {
	AuthenticateUser(ctx context.Context, email, password string, role authv1.Role) (*authv1.AuthResponse, error)
	RegisterUser(ctx context.Context, email, password string, role authv1.Role) (*authv1.AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (*authv1.TokenValidation, error)
	Logout(ctx context.Context, userID string) (*commonv1.Empty, error)
	hashPassword(password string) (string, error)
	verifyPassword(hashedPassword, password string) error
	DeleteUser(ctx context.Context, userID string) error
}

type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

type Service struct {
	Auth IAuthService
}

func NewService(repo *repository.Repository, cfg JWTConfig) *Service {
	return &Service{
		Auth: NewAuthService(repo, NewJWTManager(cfg)),
	}
}
