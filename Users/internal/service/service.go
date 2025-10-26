// internal/service/users_service.go
package service

import (
	"context"
	"errors"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/studjobs/hh_for_students/users/internal/repository"
	"log"
)

var (
	ErrProfileNotFound    = errors.New("profile not found")
	ErrInvalidProfileData = errors.New("invalid profile data")
)

type User interface {
	CreateProfile(ctx context.Context, profile *usersv1.Profile) (*usersv1.Profile, error)
	UpdateProfile(ctx context.Context, id string, profile *usersv1.Profile) (*usersv1.Profile, error)
	DeleteProfile(ctx context.Context, id string) error
	GetProfile(ctx context.Context, id string) (*usersv1.Profile, error)
	ListProfiles(ctx context.Context, professionCategory string, page, limit int32) (*usersv1.ProfileList, error)
}

type Service struct {
	User User
}

func NewService(repo *repository.Repository) *Service {
	log.Println("Service: Initializing UsersService")
	return &Service{
		User: NewUsersService(repo),
	}
}
