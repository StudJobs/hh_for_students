package services

import (
	"context"
	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	apigatewayV1 "github.com/StudJobs/proto_srtucture/gen/go/proto/apigateway/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"golang.org/x/sync/errgroup"
)

type UserService struct {
	gatewayClient apigatewayV1.ApiGatewayServiceClient
}

func NewUserService(gatewayClient apigatewayV1.ApiGatewayServiceClient) *UserService {
	return &UserService{
		gatewayClient: gatewayClient,
	}
}

// FullUserProfile агрегированные данные пользователя
type FullUserProfile struct {
	Profile      *usersv1.Profile
	Achievements *achievementv1.AchievementList
}

func (s *UserService) GetFullProfile(ctx context.Context, userID string) (*FullUserProfile, error) {
	var profile *usersv1.Profile
	var achievements *achievementv1.AchievementList

	g, gCtx := errgroup.WithContext(ctx)

	// Параллельно запрашиваем данные через ApiGateway
	g.Go(func() error {
		var err error
		profile, err = s.gatewayClient.GetProfile(gCtx, &usersv1.GetProfileRequest{
			Id: userID,
		})
		return err
	})

	g.Go(func() error {
		var err error
		achievements, err = s.gatewayClient.GetAllAchievements(gCtx, &achievementv1.GetAllAchievementsRequest{
			UserUuid: userID,
		})
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &FullUserProfile{
		Profile:      profile,
		Achievements: achievements,
	}, nil
}

func (s *UserService) GetUsers(ctx context.Context, pagination *commonv1.Pagination) (*usersv1.ProfileList, error) {
	return s.gatewayClient.GetAllProfiles(ctx, &usersv1.GetAllProfilesRequest{
		Pagination: pagination,
	})
}

func (s *UserService) UpdateUser(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error) {
	return s.gatewayClient.UpdateProfile(ctx, req)
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.gatewayClient.DeleteProfile(ctx, &usersv1.DeleteProfileRequest{
		Id: userID,
	})
	return err
}

func (s *UserService) CreateUser(ctx context.Context, req *usersv1.NewProfileRequest) (*usersv1.Profile, error) {
	return s.gatewayClient.NewProfile(ctx, req)
}
