package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAchievementNotFound = errors.New("achievement not found")
	ErrAchievementExists   = errors.New("achievement already exists")
	ACHIEVEMENT_TABLE      = "achievements"
)

// AchievementDB представляет модель достижения для БД
type AchievementDB struct {
	ID        uint       `db:"id"`
	Name      string     `db:"name"`
	UserUUID  string     `db:"user_uuid"`
	FileName  string     `db:"file_name"`
	FileType  string     `db:"file_type"`
	FileSize  int64      `db:"file_size"`
	S3Key     string     `db:"s3_key"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type AchievementRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewAchievementRepository(db *pgxpool.Pool) *AchievementRepository {
	return &AchievementRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// CreateAchievement создает новую запись о достижении
func (r *AchievementRepository) CreateAchievement(ctx context.Context, achievement *AchievementDB) error {
	log.Printf("Repository: Creating achievement for user %s: %s", achievement.UserUUID, achievement.Name)

	// Проверяем существование достижения
	var existing AchievementDB
	checkQuery, checkArgs, err := r.sb.
		Select("id", "name", "user_uuid").
		From(ACHIEVEMENT_TABLE).
		Where(squirrel.Eq{"user_uuid": achievement.UserUUID, "name": achievement.Name}).
		Where("deleted_at IS NULL").
		ToSql()

	if err == nil {
		err = r.db.QueryRow(ctx, checkQuery, checkArgs...).Scan(
			&existing.ID, &existing.Name, &existing.UserUUID,
		)
		if err == nil {
			return status.Errorf(codes.AlreadyExists, "achievement '%s' already exists for user %s", achievement.Name, achievement.UserUUID)
		}
	}

	// Создаем новое достижение
	query, args, err := r.sb.
		Insert(ACHIEVEMENT_TABLE).
		Columns("name", "user_uuid", "file_name", "file_type", "file_size", "s3_key").
		Values(achievement.Name, achievement.UserUUID, achievement.FileName,
			achievement.FileType, achievement.FileSize, achievement.S3Key).
		ToSql()

	if err != nil {
		log.Printf("Repository: Failed to build create achievement query: %v", err)
		return status.Error(codes.Internal, "failed to create achievement")
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to create achievement: %v", err)
		return status.Error(codes.Internal, "failed to create achievement")
	}

	log.Printf("Repository: Successfully created achievement for user %s: %s", achievement.UserUUID, achievement.Name)
	return nil
}

// GetAchievementsByUser возвращает все достижения пользователя
func (r *AchievementRepository) GetAchievementsByUser(ctx context.Context, userUUID string) ([]*AchievementDB, error) {
	log.Printf("Repository: Getting achievements for user: %s", userUUID)

	query, args, err := r.sb.
		Select("id", "name", "user_uuid", "file_name", "file_type", "file_size", "s3_key", "created_at").
		From(ACHIEVEMENT_TABLE).
		Where(squirrel.Eq{"user_uuid": userUUID}).
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC").
		ToSql()

	if err != nil {
		log.Printf("Repository: Failed to build get achievements query: %v", err)
		return nil, status.Error(codes.Internal, "failed to get achievements")
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to query achievements: %v", err)
		return nil, status.Error(codes.Internal, "failed to get achievements")
	}
	defer rows.Close()

	var achievements []*AchievementDB
	for rows.Next() {
		var achievement AchievementDB
		err := rows.Scan(
			&achievement.ID,
			&achievement.Name,
			&achievement.UserUUID,
			&achievement.FileName,
			&achievement.FileType,
			&achievement.FileSize,
			&achievement.S3Key,
			&achievement.CreatedAt,
		)
		if err != nil {
			log.Printf("Repository: Failed to scan achievement: %v", err)
			continue
		}
		achievements = append(achievements, &achievement)
	}

	log.Printf("Repository: Retrieved %d achievements for user %s", len(achievements), userUUID)
	return achievements, nil
}

// GetAchievementByName возвращает конкретное достижение по имени
func (r *AchievementRepository) GetAchievementByName(ctx context.Context, userUUID, name string) (*AchievementDB, error) {
	log.Printf("Repository: Getting achievement for user %s: %s", userUUID, name)

	query, args, err := r.sb.
		Select("id", "name", "user_uuid", "file_name", "file_type", "file_size", "s3_key", "created_at").
		From(ACHIEVEMENT_TABLE).
		Where(squirrel.Eq{"user_uuid": userUUID, "name": name}).
		Where("deleted_at IS NULL").
		ToSql()

	if err != nil {
		log.Printf("Repository: Failed to build get achievement query: %v", err)
		return nil, status.Error(codes.Internal, "failed to get achievement")
	}

	var achievement AchievementDB
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&achievement.ID,
		&achievement.Name,
		&achievement.UserUUID,
		&achievement.FileName,
		&achievement.FileType,
		&achievement.FileSize,
		&achievement.S3Key,
		&achievement.CreatedAt,
	)

	if err != nil {
		log.Printf("Repository: Achievement not found for user %s: %s", userUUID, name)
		return nil, status.Errorf(codes.NotFound, "achievement '%s' not found", name)
	}

	log.Printf("Repository: Successfully retrieved achievement for user %s: %s", userUUID, name)
	return &achievement, nil
}

// DeleteAchievement удаляет достижение (мягкое удаление)
func (r *AchievementRepository) DeleteAchievement(ctx context.Context, userUUID, name string) error {
	log.Printf("Repository: Deleting achievement for user %s: %s", userUUID, name)

	query, args, err := r.sb.
		Update(ACHIEVEMENT_TABLE).
		Set("deleted_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"user_uuid": userUUID, "name": name}).
		Where("deleted_at IS NULL").
		ToSql()

	if err != nil {
		log.Printf("Repository: Failed to build delete achievement query: %v", err)
		return status.Error(codes.Internal, "failed to delete achievement")
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to delete achievement: %v", err)
		return status.Error(codes.Internal, "failed to delete achievement")
	}

	if result.RowsAffected() == 0 {
		log.Printf("Repository: Achievement not found for deletion: user %s, name %s", userUUID, name)
		return status.Errorf(codes.NotFound, "achievement '%s' not found", name)
	}

	log.Printf("Repository: Successfully deleted achievement for user %s: %s", userUUID, name)
	return nil
}
