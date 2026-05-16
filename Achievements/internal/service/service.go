package service

import (
	"context"

	"github.com/studjobs/hh_for_students/achievments/internal/repository"
)

// Achievement определяет методы бизнес-логики для работы с достижениями
type Achievement interface {
	GetAllAchievements(userUUID string) ([]*AchievementResponse, error)
	GetAchievementDownloadURL(userUUID, achievementName string) (string, error)
	GetAchievementUploadURL(userUUID, achievementName, fileName, fileType string, fileSize int64) (string, string, error)
	AddAchievementMeta(userUUID, achievementName, fileName, fileType string, fileSize int64, s3Key string, achievementType int32, externalURL, description, skillSlug string) error
	DeleteAchievement(userUUID, achievementName string) error
	SubmitForReview(ctx context.Context, userUUID string, achievementID int64) error
	GetExpertQueue(ctx context.Context, page, limit int32) ([]*AchievementResponse, error)
	ReviewAchievement(ctx context.Context, achievementID int64, reviewerUUID string, decision int32, comment string) (*repository.AchievementDB, error)
	CreateMicrotaskAchievement(ctx context.Context, userUUID, microtaskID, microtaskTitle, solutionURL, reviewerUUID, reviewComment string) error
}

// Service объединяет все сервисы
type Service struct {
	Achievement Achievement
}

// NewService создает новый экземпляр сервиса
func NewService(repo *repository.Repository) *Service {
	return &Service{
		Achievement: NewAchievementService(repo),
	}
}
