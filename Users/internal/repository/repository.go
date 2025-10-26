package repository

import (
	"context"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Users interface {
	GetProfile(ctx context.Context, id string) (*usersv1.Profile, error)
	GetAllProfiles(ctx context.Context, professionCategory string, page, limit int32) (*usersv1.ProfileList, error)
	CreateProfile(ctx context.Context, profile *usersv1.Profile) (*usersv1.Profile, error)
	UpdateProfile(ctx context.Context, id string, profile *usersv1.Profile) (*usersv1.Profile, error)
	DeleteProfile(ctx context.Context, id string) error
}

type Repository struct {
	Users Users
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Users: NewUsersRepository(db),
	}
}
