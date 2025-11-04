package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/minio/minio-go/v7"
)

// Achievement определяет методы для работы с достижениями в БД
type Achievement interface {
	CreateAchievement(ctx context.Context, achievement *AchievementDB) error
	GetAchievementsByUser(ctx context.Context, userUUID string) ([]*AchievementDB, error)
	GetAchievementByName(ctx context.Context, userUUID, name string) (*AchievementDB, error)
	DeleteAchievement(ctx context.Context, userUUID, name string) error
}

// S3 определяет методы для работы с файловым хранилищем
type S3 interface {
	GenerateUploadURL(ctx context.Context, s3Key, fileType string, expiry int64) (string, error)
	GenerateDownloadURL(ctx context.Context, s3Key string, expiry int64) (string, error)
	DeleteObject(ctx context.Context, s3Key string) error
	ObjectExists(ctx context.Context, s3Key string) (bool, error) // Добавлен новый метод
}

// Repository объединяет все репозитории
type Repository struct {
	Achievement Achievement
	S3          S3
}

// NewRepository создает новый экземпляр репозитория
func NewRepository(db *pgxpool.Pool, minioClient *minio.Client) *Repository {
	return &Repository{
		Achievement: NewAchievementRepository(db),
		S3:          NewS3Repository(minioClient),
	}
}
