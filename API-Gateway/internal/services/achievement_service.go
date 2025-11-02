package services

import (
	"context"
	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
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

func (s *achievementService) GetAllAchievements(ctx context.Context, userID string) (*models.AchievementList, error) {
	req := &achievementv1.GetAllAchievementsRequest{
		UserUuid: userID,
	}

	protoResponse, err := s.client.GetAllAchievements(ctx, req)
	if err != nil {
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

	return &models.AchievementList{
		Achievements: achievements,
	}, nil
}

func (s *achievementService) GetAchievementDownloadUrl(ctx context.Context, userID, achieveName string) (*models.AchievementUrl, error) {
	req := &achievementv1.GetAchievementRequest{
		UserUuid:        userID,
		AchievementName: achieveName,
	}

	protoResponse, err := s.client.GetAchievementDownloadUrl(ctx, req)
	if err != nil {
		return nil, err
	}

	return &models.AchievementUrl{
		URL:     protoResponse.Url,
		Method:  protoResponse.Method,
		Headers: protoResponse.Headers,
	}, nil
}

func (s *achievementService) GetAchievementUploadUrl(ctx context.Context, userID, achieveName, fileName, fileType string) (*models.AchievementUrl, error) {
	req := &achievementv1.GetAchievementUploadRequest{
		UserUuid:        userID,
		AchievementName: achieveName,
		FileName:        fileName,
		FileType:        fileType,
	}

	protoResponse, err := s.client.GetAchievementUploadUrl(ctx, req)
	if err != nil {
		return nil, err
	}

	return &models.AchievementUrl{
		URL:     protoResponse.Url,
		Method:  protoResponse.Method,
		Headers: protoResponse.Headers,
	}, nil
}

func (s *achievementService) AddAchievementMeta(ctx context.Context, meta *models.AchievementMeta) error {
	protoMeta := &achievementv1.AchievementMeta{
		Name:      meta.Name,
		UserUuid:  meta.UserUUID,
		FileName:  meta.FileName,
		FileType:  meta.FileType,
		FileSize:  meta.FileSize,
		CreatedAt: meta.CreatedAt,
	}

	req := &achievementv1.AddAchievementMetaRequest{
		Meta: protoMeta,
	}

	_, err := s.client.AddAchievementMeta(ctx, req)
	return err
}

func (s *achievementService) DeleteAchievement(ctx context.Context, userID, achieveName string) error {
	req := &achievementv1.DeleteAchievementRequest{
		UserUuid:        userID,
		AchievementName: achieveName,
	}

	_, err := s.client.DeleteAchievement(ctx, req)
	return err
}
