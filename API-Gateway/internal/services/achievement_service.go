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
		achievements[i] = protoToModelMeta(protoAchievement)
	}

	log.Printf("AchievementService: Successfully retrieved %d achievements for user: %s", len(achievements), userID)
	return &models.AchievementList{
		Achievements: achievements,
	}, nil
}

// protoToModelMeta — единая конвертация AchievementMeta из proto в HTTP-модель.
func protoToModelMeta(p *achievementv1.AchievementMeta) models.AchievementMeta {
	return models.AchievementMeta{
		ID:                 p.GetId(),
		Name:               p.GetName(),
		UserUUID:           p.GetUserUuid(),
		FileName:           p.GetFileName(),
		FileType:           p.GetFileType(),
		FileSize:           p.GetFileSize(),
		Type:               int32(p.GetType()),
		CreatedAt:          p.GetCreatedAt(),
		VerificationStatus: int32(p.GetVerificationStatus()),
		ReviewedBy:         p.GetReviewedBy(),
		ReviewedAt:         p.GetReviewedAt(),
		ReviewComment:      p.GetReviewComment(),
		ExternalURL:        p.GetExternalUrl(),
		Description:        p.GetDescription(),
	}
}

// SubmitForReview — отправка достижения студентом на экспертную проверку.
func (s *achievementService) SubmitForReview(ctx context.Context, userUUID string, achievementID int64) error {
	_, err := s.client.SubmitForReview(ctx, &achievementv1.SubmitForReviewRequest{
		UserUuid:      userUUID,
		AchievementId: achievementID,
	})
	if err != nil {
		log.Printf("AchievementService: SubmitForReview failed: %v", err)
		return err
	}
	return nil
}

// GetExpertQueue — список достижений в статусе PENDING.
func (s *achievementService) GetExpertQueue(ctx context.Context, page, limit int32) (*models.AchievementList, error) {
	resp, err := s.client.GetExpertQueue(ctx, &achievementv1.GetExpertQueueRequest{
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		log.Printf("AchievementService: GetExpertQueue failed: %v", err)
		return nil, err
	}
	out := make([]models.AchievementMeta, len(resp.Achievements))
	for i, p := range resp.Achievements {
		out[i] = protoToModelMeta(p)
	}
	return &models.AchievementList{Achievements: out}, nil
}

// ReviewAchievement — эксперт принимает решение по достижению.
func (s *achievementService) ReviewAchievement(ctx context.Context, achievementID int64, reviewerUUID string, decision int32, comment string) error {
	_, err := s.client.ReviewAchievement(ctx, &achievementv1.ReviewAchievementRequest{
		AchievementId: achievementID,
		ReviewerUuid:  reviewerUUID,
		Decision:      achievementv1.VerificationStatus(decision),
		Comment:       comment,
	})
	if err != nil {
		log.Printf("AchievementService: ReviewAchievement failed: %v", err)
		return err
	}
	return nil
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
		Name:               meta.Name,
		UserUuid:           meta.UserUUID,
		FileName:           meta.FileName,
		FileType:           meta.FileType,
		FileSize:           meta.FileSize,
		Type:               achievementv1.AchievementType(meta.Type),
		CreatedAt:          meta.CreatedAt,
		VerificationStatus: achievementv1.VerificationStatus(meta.VerificationStatus),
		ExternalUrl:        meta.ExternalURL,
		Description:        meta.Description,
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
