package services

import (
	"context"
	"log"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
)

type usersService struct {
	client usersv1.UsersServiceClient
}

// NewUsersService создает новый экземпляр UsersService
func NewUsersService(client usersv1.UsersServiceClient) UsersService {
	log.Printf("Creating new UsersService")
	return &usersService{
		client: client,
	}
}

func (s *usersService) CreateUser(ctx context.Context, req *usersv1.NewProfileRequest) (*usersv1.Profile, error) {
	log.Printf("UsersService: CreateUser attempt for email: %s", req.Profile.Email)

	resp, err := s.client.NewProfile(ctx, req)
	if err != nil {
		log.Printf("UsersService: CreateUser failed for email %s: %v", req.Profile.Email, err)
		return nil, err
	}

	log.Printf("UsersService: CreateUser successful for user_id: %s", resp.Id)
	return resp, nil
}

func (s *usersService) GetUser(ctx context.Context, userID string) (*usersv1.Profile, error) {
	log.Printf("UsersService: GetUser attempt for user_id: %s", userID)

	resp, err := s.client.GetProfile(ctx, &usersv1.GetProfileRequest{
		Id: userID,
	})
	if err != nil {
		log.Printf("UsersService: GetUser failed for user_id %s: %v", userID, err)
		return nil, err
	}

	log.Printf("UsersService: GetUser successful for user_id: %s", resp.Id)
	return resp, nil
}

func (s *usersService) GetUsers(ctx context.Context, req *usersv1.GetAllProfilesRequest) (*usersv1.ProfileList, error) {

	resp, err := s.client.GetAllProfiles(ctx, req)
	if err != nil {
		log.Printf("UsersService: GetUsers failed: %v", err)
		return nil, err
	}

	log.Printf("UsersService: GetUsers successful, found %d users", len(resp.Profiles))
	return resp, nil
}

func (s *usersService) UpdateUser(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error) {
	log.Printf("UsersService: UpdateUser attempt for user_id: %s", req.Id)

	resp, err := s.client.UpdateProfile(ctx, req)
	if err != nil {
		log.Printf("UsersService: UpdateUser failed for user_id %s: %v", req.Id, err)
		return nil, err
	}

	log.Printf("UsersService: UpdateUser successful for user_id: %s", resp.Id)
	return resp, nil
}

func (s *usersService) DeleteUser(ctx context.Context, userID string) error {
	log.Printf("UsersService: DeleteUser attempt for user_id: %s", userID)

	_, err := s.client.DeleteProfile(ctx, &usersv1.DeleteProfileRequest{
		Id: userID,
	})
	if err != nil {
		log.Printf("UsersService: DeleteUser failed for user_id %s: %v", userID, err)
		return err
	}

	log.Printf("UsersService: DeleteUser successful for user_id: %s", userID)
	return nil
}
