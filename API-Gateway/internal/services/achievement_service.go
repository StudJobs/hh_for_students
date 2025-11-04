package services

import (
	"context"
	"log"

	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

type achievementService struct {
	client achievementv1.AchievementServiceClient
}

func NewAchievementService(client achievementv1.AchievementServiceClient) AchievementService {
	log.Printf("Creating new AchievementService")
	return &achievementService{
		client: client,
	}
}

// GetAllAchievements возвращает все достижения пользователя
func (s *achievementService) GetAllAchievements(ctx context.Context, userID string) (*models.AchievementList, error) {
	log.Printf("AchievementService: Getting all achievements for user: %s", userID)

	req := &achievementv1.GetAllAchievementsRequest{
		UserUuid: userID,
	}

	protoResponse, err := s.client.GetAllAchievements(ctx, req)
	if err != nil {
		log.Printf("AchievementService: Failed to get achievements for user %s: %v", userID, err)
		return nil, err
	}

	achievements := make([]models.AchievementMeta, len(protoResponse.Achievements))
	for i, protoAchievement := range protoResponse.Achievements {
		achievements[i] = models.AchievementMeta{
			Name:      protoAchievement.Name,
			UserUUID:  protoAchievement.UserUuid,
			FileName:  protoAchievement.FileName,
			FileType:  protoAchievement.FileType,
			FileSize:  protoAchievement.FileSize,
			CreatedAt: protoAchievement.CreatedAt,
		}
	}

	log.Printf("AchievementService: Successfully retrieved %d achievements for user: %s", len(achievements), userID)
	return &models.AchievementList{
		Achievements: achievements,
	}, nil
}

// GetAchievementDownloadUrl возвращает URL для скачивания достижения
func (s *achievementService) GetAchievementDownloadUrl(ctx context.Context, userID, achieveName string) (*models.AchievementUrl, error) {
	log.Printf("AchievementService: Getting download URL for user %s, achievement: %s", userID, achieveName)

	req := &achievementv1.GetAchievementRequest{
		UserUuid:        userID,
		AchievementName: achieveName,
	}

	protoResponse, err := s.client.GetAchievementDownloadUrl(ctx, req)
	if err != nil {
		log.Printf("AchievementService: Failed to get download URL for user %s, achievement %s: %v", userID, achieveName, err)
		return nil, err
	}

	url := &models.AchievementUrl{
		URL:       protoResponse.Url,
		ExpiresAt: protoResponse.ExpiresAt,
	}

	log.Printf("AchievementService: Successfully retrieved download URL for user %s, achievement: %s", userID, achieveName)
	return url, nil
}

// GetAchievementUploadUrl возвращает URL для загрузки файла достижения
func (s *achievementService) GetAchievementUploadUrl(ctx context.Context, userID, achieveName, fileName, fileType string, fileSize int64) (*models.UploadUrlResponse, error) {
	log.Printf("AchievementService: Getting upload URL for user %s, achievement: %s", userID, achieveName)

	req := &achievementv1.GetAchievementUploadRequest{
		UserUuid:        userID,
		AchievementName: achieveName,
		FileName:        fileName,
		FileType:        fileType,
		FileSize:        fileSize,
	}

	protoResponse, err := s.client.GetAchievementUploadUrl(ctx, req)
	if err != nil {
		log.Printf("AchievementService: Failed to get upload URL for user %s, achievement %s: %v", userID, achieveName, err)
		return nil, err
	}

	response := &models.UploadUrlResponse{
		UploadURL: protoResponse.UploadUrl,
		S3Key:     protoResponse.S3Key,
		ExpiresAt: protoResponse.ExpiresAt,
	}

	log.Printf("AchievementService: Successfully retrieved upload URL for user %s, achievement: %s, S3 key: %s",
		userID, achieveName, protoResponse.S3Key)
	return response, nil
}

// AddAchievementMeta добавляет метаданные достижения после успешной загрузки
func (s *achievementService) AddAchievementMeta(ctx context.Context, meta *models.AchievementMeta, s3Key string) error {
	log.Printf("AchievementService: Adding achievement metadata for user %s, achievement: %s", meta.UserUUID, meta.Name)

	protoMeta := &achievementv1.AchievementMeta{
		Name:      meta.Name,
		UserUuid:  meta.UserUUID,
		FileName:  meta.FileName,
		FileType:  meta.FileType,
		FileSize:  meta.FileSize,
		CreatedAt: meta.CreatedAt,
	}

	req := &achievementv1.AddAchievementMetaRequest{
		Meta:  protoMeta,
		S3Key: s3Key,
	}

	_, err := s.client.AddAchievementMeta(ctx, req)
	if err != nil {
		log.Printf("AchievementService: Failed to add achievement metadata for user %s, achievement %s: %v",
			meta.UserUUID, meta.Name, err)
		return err
	}

	log.Printf("AchievementService: Successfully added achievement metadata for user %s, achievement: %s",
		meta.UserUUID, meta.Name)
	return nil
}

// DeleteAchievement удаляет достижение
func (s *achievementService) DeleteAchievement(ctx context.Context, userID, achieveName string) error {
	log.Printf("AchievementService: Deleting achievement for user %s: %s", userID, achieveName)

	req := &achievementv1.DeleteAchievementRequest{
		UserUuid:        userID,
		AchievementName: achieveName,
	}

	_, err := s.client.DeleteAchievement(ctx, req)
	if err != nil {
		log.Printf("AchievementService: Failed to delete achievement for user %s: %s: %v", userID, achieveName, err)
		return err
	}

	log.Printf("AchievementService: Successfully deleted achievement for user %s: %s", userID, achieveName)
	return nil
}
