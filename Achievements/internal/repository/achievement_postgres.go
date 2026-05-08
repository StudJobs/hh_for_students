package repository

import (
	"context"
	"errors"
	"log"
	"strings"
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
	ID                 int64      `db:"id"`
	Name               string     `db:"name"`
	UserUUID           string     `db:"user_uuid"`
	FileName           string     `db:"file_name"`
	FileType           string     `db:"file_type"`
	FileSize           int64      `db:"file_size"`
	S3Key              string     `db:"s3_key"`
	Type               int32      `db:"type"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
	DeletedAt          *time.Time `db:"deleted_at"`
	VerificationStatus int32      `db:"verification_status"`
	ReviewedBy         *string    `db:"reviewed_by"`
	ReviewedAt         *time.Time `db:"reviewed_at"`
	ReviewComment      *string    `db:"review_comment"`
}

// Колонки для SELECT — единый источник правды.
var achievementSelectColumns = []string{
	"id", "name", "user_uuid", "file_name", "file_type", "file_size",
	"s3_key", "type", "created_at",
	"verification_status", "reviewed_by", "reviewed_at", "review_comment",
}

func scanAchievement(scanner interface {
	Scan(dest ...interface{}) error
}, a *AchievementDB) error {
	return scanner.Scan(
		&a.ID, &a.Name, &a.UserUUID, &a.FileName, &a.FileType, &a.FileSize,
		&a.S3Key, &a.Type, &a.CreatedAt,
		&a.VerificationStatus, &a.ReviewedBy, &a.ReviewedAt, &a.ReviewComment,
	)
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
		Columns("name", "user_uuid", "file_name", "file_type", "file_size", "s3_key", "type").
		Values(achievement.Name, achievement.UserUUID, achievement.FileName,
			achievement.FileType, achievement.FileSize, achievement.S3Key, achievement.Type).
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
		Select(achievementSelectColumns...).
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
		if err := scanAchievement(rows, &achievement); err != nil {
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
		Select(achievementSelectColumns...).
		From(ACHIEVEMENT_TABLE).
		Where(squirrel.Eq{"user_uuid": userUUID, "name": name}).
		Where("deleted_at IS NULL").
		ToSql()

	if err != nil {
		log.Printf("Repository: Failed to build get achievement query: %v", err)
		return nil, status.Error(codes.Internal, "failed to get achievement")
	}

	var achievement AchievementDB
	if err := scanAchievement(r.db.QueryRow(ctx, query, args...), &achievement); err != nil {
		log.Printf("Repository: Achievement not found for user %s: %s", userUUID, name)
		return nil, status.Errorf(codes.NotFound, "achievement '%s' not found", name)
	}

	log.Printf("Repository: Successfully retrieved achievement for user %s: %s", userUUID, name)
	return &achievement, nil
}

// GetAchievementByID возвращает достижение по числовому ID (для ревью эксперта).
func (r *AchievementRepository) GetAchievementByID(ctx context.Context, id int64) (*AchievementDB, error) {
	query, args, err := r.sb.
		Select(achievementSelectColumns...).
		From(ACHIEVEMENT_TABLE).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to build query")
	}

	var achievement AchievementDB
	if err := scanAchievement(r.db.QueryRow(ctx, query, args...), &achievement); err != nil {
		return nil, status.Errorf(codes.NotFound, "achievement %d not found", id)
	}
	return &achievement, nil
}

// SetVerificationStatus обновляет статус верификации (используется для submit и review).
// Если reviewerUUID пуст — поля reviewed_by/reviewed_at/review_comment не трогаются (submit).
// Если заполнен — выставляются NOW() и комментарий.
func (r *AchievementRepository) SetVerificationStatus(
	ctx context.Context,
	id int64,
	newStatus int32,
	expectedStatus int32,
	reviewerUUID, comment string,
) (*AchievementDB, error) {
	upd := r.sb.
		Update(ACHIEVEMENT_TABLE).
		Set("verification_status", newStatus).
		Where(squirrel.Eq{"id": id, "verification_status": expectedStatus}).
		Where("deleted_at IS NULL")
	if reviewerUUID != "" {
		upd = upd.
			Set("reviewed_by", reviewerUUID).
			Set("reviewed_at", squirrel.Expr("NOW()")).
			Set("review_comment", comment)
	}
	upd = upd.Suffix("RETURNING " + strings.Join(achievementSelectColumns, ", "))

	query, args, err := upd.ToSql()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to build update query")
	}

	var achievement AchievementDB
	if err := scanAchievement(r.db.QueryRow(ctx, query, args...), &achievement); err != nil {
		log.Printf("Repository: SetVerificationStatus failed for id=%d (expected status=%d): %v", id, expectedStatus, err)
		return nil, status.Errorf(codes.FailedPrecondition, "achievement %d not in expected status", id)
	}
	return &achievement, nil
}

// ListPending возвращает достижения в статусе PENDING (для очереди эксперта).
func (r *AchievementRepository) ListPending(ctx context.Context, page, limit int32) ([]*AchievementDB, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := uint64((page - 1) * limit)

	query, args, err := r.sb.
		Select(achievementSelectColumns...).
		From(ACHIEVEMENT_TABLE).
		Where(squirrel.Eq{"verification_status": 2}). // PENDING
		Where("deleted_at IS NULL").
		OrderBy("created_at ASC").
		Limit(uint64(limit)).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to build query")
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to query pending achievements")
	}
	defer rows.Close()

	var achievements []*AchievementDB
	for rows.Next() {
		var a AchievementDB
		if err := scanAchievement(rows, &a); err != nil {
			log.Printf("Repository: ListPending scan failed: %v", err)
			continue
		}
		achievements = append(achievements, &a)
	}
	return achievements, nil
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
