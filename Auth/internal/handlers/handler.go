package handlers

import (
	"context"
	"log"

	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandlers struct {
	authv1.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthHandlers(authService *service.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

func (h *AuthHandlers) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	log.Printf("gRPC Login request - email: %s, role: %v", req.Email, req.Role)

	if req.Email == "" || req.Password == "" || req.Role == authv1.Role_ROLE_UNSPECIFIED {
		log.Printf("gRPC Login failed - missing required fields")
		return nil, status.Error(codes.InvalidArgument, "email, password and role are required")
	}

	authResponse, err := h.authService.AuthenticateUser(ctx, req.Email, req.Password, req.Role)
	if err != nil {
		log.Printf("gRPC Login failed for email %s: %v", req.Email, err)
		switch err {
		case service.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	log.Printf("gRPC Login successful for email: %s, user_uuid: %s", req.Email, authResponse.UserUuid)
	return authResponse, nil
}

func (h *AuthHandlers) SignUp(ctx context.Context, req *authv1.SignUpRequest) (*authv1.AuthResponse, error) {
	log.Printf("gRPC SignUp request - email: %s, role: %v", req.Email, req.Role)

	if req.Email == "" || req.Password == "" || req.Role == authv1.Role_ROLE_UNSPECIFIED {
		log.Printf("gRPC SignUp failed - missing required fields")
		return nil, status.Error(codes.InvalidArgument, "email, password and role are required")
	}

	if len(req.Password) < 6 {
		log.Printf("gRPC SignUp failed - password too short for email: %s", req.Email)
		return nil, status.Error(codes.InvalidArgument, "password must be at least 6 characters")
	}

	authResponse, err := h.authService.RegisterUser(ctx, req.Email, req.Password, req.Role)
	if err != nil {
		log.Printf("gRPC SignUp failed for email %s: %v", req.Email, err)
		switch err {
		case service.ErrUserAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	log.Printf("gRPC SignUp successful for email: %s, user_uuid: %s", req.Email, authResponse.UserUuid)
	return authResponse, nil
}

func (h *AuthHandlers) ParseToken(ctx context.Context, req *authv1.ParseTokenRequest) (*authv1.TokenValidation, error) {
	log.Printf("gRPC ParseToken request - token: %s...", req.AccessToken[:10])

	if req.AccessToken == "" {
		log.Printf("gRPC ParseToken failed - empty token")
		return nil, status.Error(codes.InvalidArgument, "access token is required")
	}

	tokenValidation, err := h.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		log.Printf("gRPC ParseToken failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to validate token")
	}

	log.Printf("gRPC ParseToken result - valid: %t, user_uuid: %s", tokenValidation.Valid, tokenValidation.UserUuid)
	return tokenValidation, nil
}
