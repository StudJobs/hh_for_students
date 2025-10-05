package services

import (
	"context"
	apigatewayV1 "github.com/StudJobs/proto_srtucture/gen/go/proto/apigateway/v1"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
)

type Users interface {
	CreateUser(ctx context.Context, req *usersv1.NewProfileRequest) (*usersv1.Profile, error)
	GetFullProfile(ctx context.Context, userID string) (*FullUserProfile, error)
	GetUsers(ctx context.Context, pagination *commonv1.Pagination) (*usersv1.ProfileList, error)
	UpdateUser(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error)
	DeleteUser(ctx context.Context, userID string) error
}
type Authentication interface {
	Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error)
	Register(ctx context.Context, req *authv1.SignUpRequest) (*authv1.AuthResponse, error)
}

type ServiceAPIGateway struct {
	User Users
	Auth Authentication
}

func NewServiceAPIGateway(client apigatewayV1.ApiGatewayServiceClient) *ServiceAPIGateway {
	return &ServiceAPIGateway{
		User: NewUserService(client),
		Auth: NewAuthService(client),
	}
}
