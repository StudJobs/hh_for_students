package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type User struct {
	UUID      string    `db:"uuid"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Role      int       `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(ctx context.Context, email, hashedPassword string, role int) (string, error) {
	const query = `
		INSERT INTO users (email, password, role) 
		VALUES ($1, $2, $3) 
		RETURNING uuid
	`

	log.Printf("Creating user in database - email: %s, role: %d", email, role)

	var userUUID string
	err := r.db.QueryRow(ctx, query, email, hashedPassword, role).Scan(&userUUID)
	if err != nil {
		log.Printf("Failed to create user - email: %s, error: %v", email, err)
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User created successfully - email: %s, uuid: %s", email, userUUID)
	return userUUID, nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT uuid, email, password, role, created_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`

	log.Printf("Finding user by email: %s", email)

	var user User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.UUID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		log.Printf("User not found by email: %s, error: %v", email, err)
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	log.Printf("User found by email: %s, uuid: %s", email, user.UUID)
	return &user, nil
}

func (r *AuthRepository) FindUserByUUID(ctx context.Context, uuid string) (*User, error) {
	const query = `
		SELECT uuid, email, password, role, created_at
		FROM users 
		WHERE uuid = $1 AND deleted_at IS NULL
	`

	log.Printf("Finding user by uuid: %s", uuid)

	var user User
	err := r.db.QueryRow(ctx, query, uuid).Scan(
		&user.UUID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		log.Printf("User not found by uuid: %s, error: %v", uuid, err)
		return nil, fmt.Errorf("failed to find user by uuid: %w", err)
	}

	log.Printf("User found by uuid: %s, email: %s", uuid, user.Email)
	return &user, nil
}
