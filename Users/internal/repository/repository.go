package repository

import (
	"context"
	chatv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/chat/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Users interface {
	GetProfile(ctx context.Context, id string) (*usersv1.Profile, error)
	GetAllProfiles(ctx context.Context, professionCategory string, page, limit int32, role string) (*usersv1.ProfileList, error)
	CreateProfile(ctx context.Context, profile *usersv1.Profile) (*usersv1.Profile, error)
	UpdateProfile(ctx context.Context, id string, profile *usersv1.Profile) (*usersv1.Profile, error)
	DeleteProfile(ctx context.Context, id string) error
	AddVerifiedSkills(ctx context.Context, userID string, slugs []string) (*usersv1.Profile, error)
	AddExpertVerifiedSkills(ctx context.Context, userID string, slugs []string) error
}

type Chat interface {
	Insert(ctx context.Context, threadID, fromUser, body string) (*chatv1.Message, error)
	ListByThread(ctx context.Context, threadID string, page, limit int32) (*chatv1.MessageList, error)
	ListUserThreads(ctx context.Context, userID string, limit int32) ([]*chatv1.Thread, error)
}

type Repository struct {
	Users Users
	Chat  Chat
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Users: NewUsersRepository(db),
		Chat:  NewChatRepository(db),
	}
}
