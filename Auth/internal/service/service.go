package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/auth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
)

type TokenManager interface {
	GenerateToken(userUUID, email string, role authv1.Role) (string, error)
	ValidateToken(token string) (string, authv1.Role, error)
}

type AuthService struct {
	repo         *repository.AuthRepository // ИСПРАВЛЕНО: был *repository.Repository
	tokenManager TokenManager
}

func NewAuthService(repo *repository.AuthRepository, tokenManager TokenManager) *AuthService { // ИСПРАВЛЕНО
	return &AuthService{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) AuthenticateUser(ctx context.Context, email, password string, role authv1.Role) (*authv1.AuthResponse, error) {
	log.Printf("Authenticating user - email: %s, role: %v", email, role)

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		log.Printf("Authentication failed - user not found: %s", email)
		return nil, ErrInvalidCredentials
	}

	log.Printf("User found - uuid: %s, db_role: %d, requested_role: %d", user.UUID, user.Role, role)

	if int(role) != user.Role {
		log.Printf("Authentication failed - role mismatch for user %s: expected %d, got %d", email, user.Role, role)
		return nil, ErrInvalidCredentials
	}

	log.Printf("Verifying password for user: %s", email)
	if err = s.verifyPassword(user.Password, password); err != nil {
		log.Printf("Authentication failed - invalid password for user: %s", email)
		return nil, ErrInvalidCredentials
	}

	log.Printf("Generating token for user: %s", email)
	token, err := s.tokenManager.GenerateToken(user.UUID, user.Email, role)
	if err != nil {
		log.Printf("Token generation failed for user %s: %v", email, err)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("Authentication successful for user: %s", email)
	return &authv1.AuthResponse{
		Token:    token,
		UserUuid: user.UUID,
		Role:     role,
	}, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, email, password string, role authv1.Role) (*authv1.AuthResponse, error) {
	log.Printf("Registering user - email: %s, role: %v", email, role)

	existingUser, err := s.repo.FindUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		log.Printf("Registration failed - user already exists: %s", email)
		return nil, ErrUserAlreadyExists
	}

	log.Printf("Hashing password for user: %s", email)
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		log.Printf("Password hashing failed for user %s: %v", email, err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	log.Printf("Creating user in database: %s", email)
	userUUID, err := s.repo.CreateUser(ctx, email, hashedPassword, int(role))
	if err != nil {
		log.Printf("User creation failed for email %s: %v", email, err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Generating token for new user: %s", email)
	token, err := s.tokenManager.GenerateToken(userUUID, email, role)
	if err != nil {
		log.Printf("Token generation failed for new user %s: %v", email, err)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("Registration successful for user: %s, uuid: %s", email, userUUID)
	return &authv1.AuthResponse{
		Token:    token,
		UserUuid: userUUID,
		Role:     role,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*authv1.TokenValidation, error) {
	log.Printf("Validating token: %s...", token[:10])

	userUUID, role, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return &authv1.TokenValidation{Valid: false}, nil
	}

	log.Printf("Token validated - user_uuid: %s, checking user existence", userUUID)

	// Проверяем что пользователь все еще существует
	user, err := s.repo.FindUserByUUID(ctx, userUUID)
	if err != nil || user == nil {
		log.Printf("Token validation failed - user not found: %s", userUUID)
		return &authv1.TokenValidation{Valid: false}, nil
	}

	log.Printf("Token validation successful for user: %s", userUUID)
	return &authv1.TokenValidation{
		Valid:    true,
		UserUuid: userUUID,
		Role:     role,
	}, nil
}

func (s *AuthService) hashPassword(password string) (string, error) {
	log.Printf("Hashing password")
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
	}
	return string(hashedBytes), err
}

func (s *AuthService) verifyPassword(hashedPassword, password string) error {
	log.Printf("Verifying password")
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Printf("Password verification failed: %v", err)
	}
	return err
}
