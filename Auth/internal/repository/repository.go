package repository

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Auth interface {
	CreateUser(ctx context.Context, email, hashedPassword string, role int) (string, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByUUID(ctx context.Context, uuid string) (*User, error)
}
type Repository struct {
	Auth Auth
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Auth: NewAuthRepository(db),
	}
}
