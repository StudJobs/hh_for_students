package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

var sb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

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
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateUser(ctx context.Context, email, hashedPassword string, role int) (string, error) {
	query, args, err := sb.
		Insert("users").
		Columns("email", "password", "role").
		Values(email, hashedPassword, role).
		Suffix("RETURNING uuid").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build query: %w", err)
	}

	log.Printf("Creating user in database - email: %s, role: %d", email, role)

	var userUUID string
	err = r.db.QueryRow(ctx, query, args...).Scan(&userUUID)
	if err != nil {
		log.Printf("Failed to create user - email: %s, error: %v", email, err)
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User created successfully - email: %s, uuid: %s", email, userUUID)
	return userUUID, nil
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := sb.
		Select("uuid", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"email": email}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	log.Printf("Finding user by email: %s", email)

	var user User
	err = r.db.QueryRow(ctx, query, args...).Scan(
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
	query, args, err := sb.
		Select("uuid", "email", "password", "role", "created_at").
		From("users").
		Where(squirrel.Eq{"uuid": uuid}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	log.Printf("Finding user by uuid: %s", uuid)

	var user User
	err = r.db.QueryRow(ctx, query, args...).Scan(
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

func (r *AuthRepository) IsUserLoggedOut(ctx context.Context, userID string) (bool, error) {
	query, args, err := sb.
		Select("COUNT(*)").
		From("user_logouts").
		Where(squirrel.Eq{"user_id": userID}).
		Where("expires_at > ?", time.Now()).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build logout check query: %w", err)
	}

	var count int
	err = r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check logout status: %w", err)
	}

	return count > 0, nil
}

func (r *AuthRepository) DeleteUser(ctx context.Context, userID string) error {
	// Мягкое удаление - устанавливаем deleted_at
	query, args, err := sb.
		Update("users").
		Set("deleted_at", time.Now()).
		Where(squirrel.Eq{"uuid": userID}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete user query: %w", err)
	}

	log.Printf("Deleting user: %s", userID)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Failed to delete user: %s, error: %v", userID, err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		log.Printf("User not found or already deleted: %s", userID)
		return fmt.Errorf("user not found or already deleted")
	}

	log.Printf("User deleted successfully: %s", userID)
	return nil
}

// CleanupExpiredLogouts удаляет устаревшие записи logout
func (r *AuthRepository) CleanupExpiredLogouts(ctx context.Context) error {
	query, args, err := sb.
		Delete("user_logouts").
		Where("expires_at < ?", time.Now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build cleanup query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired logouts: %w", err)
	}

	log.Printf("Cleaned up expired logouts, count: %d", result.RowsAffected())
	return nil
}
