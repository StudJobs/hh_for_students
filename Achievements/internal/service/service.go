package service

import "github.com/studjobs/hh_for_students/achievments/internal/repository"

// Achievement определяет методы бизнес-логики для работы с достижениями
type Achievement interface {
	GetAllAchievements(userUUID string) ([]*AchievementResponse, error)
	GetAchievementDownloadURL(userUUID, achievementName string) (string, error)
	GetAchievementUploadURL(userUUID, achievementName, fileName, fileType string, fileSize int64) (string, string, error)
	AddAchievementMeta(userUUID, achievementName, fileName, fileType string, fileSize int64, s3Key string) error
	DeleteAchievement(userUUID, achievementName string) error
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
