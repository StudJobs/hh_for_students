package services

import (
	"context"
	"log"

	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authService struct {
	client authv1.AuthServiceClient
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(client authv1.AuthServiceClient) AuthService {
	log.Printf("Creating new AuthService")
	return &authService{
		client: client,
	}
}

func (s *authService) Login(ctx context.Context, email, password, role string) (*models.AuthResponse, error) {
	log.Printf("AuthService: Login attempt for email: %s", email)

	// Конвертируем роль
	grpcRole, err := convertRoleToGRPC(role)
	if err != nil {
		log.Printf("AuthService: Login failed - invalid role: %s", role)
		return nil, err
	}

	// Вызываем auth сервис
	resp, err := s.client.Login(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
		Role:     grpcRole,
	})
	if err != nil {
		log.Printf("AuthService: Login failed for email %s: %v", email, err)
		return nil, err
	}

	// Конвертируем ответ
	authResp := &models.AuthResponse{
		Token:    resp.Token,
		UserUUID: resp.UserUuid,
		Role:     convertRoleFromGRPC(resp.Role),
	}

	log.Printf("AuthService: Login successful for email: %s", email)
	return authResp, nil
}

func (s *authService) Register(ctx context.Context, email, password, role string) (*models.AuthResponse, error) {
	log.Printf("AuthService: Register attempt for email: %s", email)

	// Конвертируем роль
	grpcRole, err := convertRoleToGRPC(role)
	if err != nil {
		log.Printf("AuthService: Register failed - invalid role: %s", role)
		return nil, err
	}

	// Вызываем auth сервис
	resp, err := s.client.SignUp(ctx, &authv1.SignUpRequest{
		Email:    email,
		Password: password,
		Role:     grpcRole,
	})
	if err != nil {
		log.Printf("AuthService: Register failed for email %s: %v", email, err)
		return nil, err
	}

	// Конвертируем ответ
	authResp := &models.AuthResponse{
		Token:    resp.Token,
		UserUUID: resp.UserUuid,
		Role:     convertRoleFromGRPC(resp.Role),
	}

	log.Printf("AuthService: Register successful for email: %s", email)
	return authResp, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (bool, string, string, error) {
	log.Printf("AuthService: ValidateToken attempt")

	resp, err := s.client.ParseToken(ctx, &authv1.ParseTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			log.Printf("AuthService: ValidateToken failed - invalid token")
			return false, "", "", nil
		}
		log.Printf("AuthService: ValidateToken failed: %v", err)
		return false, "", "", err
	}

	log.Printf("AuthService: ValidateToken successful for user_uuid: %s", resp.UserUuid)
	return resp.Valid, resp.UserUuid, convertRoleFromGRPC(resp.Role), nil
}

func (s *authService) Logout(ctx context.Context, userID string) error {
	log.Printf("AuthService: Logout attempt for userID: %s", userID)

	if _, err := s.client.Logout(ctx, &authv1.LogoutRequest{UserUuid: userID}); err != nil {
		log.Printf("AuthService: Logout failed for user %s: %v", userID, err)
		return err
	}

	return nil
}

func (s *authService) DeleteUser(ctx context.Context, userID string) error {
	log.Printf("AuthService: DeleteUser attempt for userID: %s", userID)

	if _, err := s.client.Delete(ctx, &authv1.DeleteRequest{UserUuid: userID}); err != nil {
		log.Printf("AuthService: DeleteUser failed for user %s: %v", userID, err)
		return err
	}

	return nil
}
