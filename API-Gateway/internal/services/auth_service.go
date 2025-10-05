package services

import (
	"context"
	apigatewayV1 "github.com/StudJobs/proto_srtucture/gen/go/proto/apigateway/v1"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
)

type AuthService struct {
	gatewayClient apigatewayV1.ApiGatewayServiceClient
}

func NewAuthService(gatewayClient apigatewayV1.ApiGatewayServiceClient) *AuthService {
	return &AuthService{
		gatewayClient: gatewayClient,
	}
}

func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	return s.gatewayClient.Login(ctx, req)
}

func (s *AuthService) Register(ctx context.Context, req *authv1.SignUpRequest) (*authv1.AuthResponse, error) {
	return s.gatewayClient.SignUp(ctx, req)
}
