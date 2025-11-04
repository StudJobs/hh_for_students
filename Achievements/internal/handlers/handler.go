package handlers

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"

	"github.com/studjobs/hh_for_students/achievments/internal/service"
)

type Handler struct {
	achievementv1.UnimplementedAchievementServiceServer
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetAllAchievements возвращает все достижения пользователя
func (h *Handler) GetAllAchievements(ctx context.Context, req *achievementv1.GetAllAchievementsRequest) (*achievementv1.AchievementList, error) {
	log.Printf("Handler: GetAllAchievements called for user: %s", req.GetUserUuid())

	achievements, err := h.service.Achievement.GetAllAchievements(req.GetUserUuid())
	if err != nil {
		return nil, err
	}

	protoAchievements := make([]*achievementv1.AchievementMeta, len(achievements))
	for i, achievement := range achievements {
		protoAchievements[i] = &achievementv1.AchievementMeta{
			Name:      achievement.Name,
			UserUuid:  achievement.UserUUID,
			FileName:  achievement.FileName,
			FileType:  achievement.FileType,
			FileSize:  achievement.FileSize,
			CreatedAt: achievement.CreatedAt.Format(time.RFC3339),
		}
	}

	return &achievementv1.AchievementList{
		Achievements: protoAchievements,
	}, nil
}

// GetAchievementDownloadUrl возвращает URL для скачивания достижения
func (h *Handler) GetAchievementDownloadUrl(ctx context.Context, req *achievementv1.GetAchievementRequest) (*achievementv1.AchievementUrl, error) {
	log.Printf("Handler: GetAchievementDownloadUrl called for user: %s, achievement: %s",
		req.GetUserUuid(), req.GetAchievementName())

	url, err := h.service.Achievement.GetAchievementDownloadURL(req.GetUserUuid(), req.GetAchievementName())
	if err != nil {
		return nil, err
	}

	return &achievementv1.AchievementUrl{
		Url:       url,
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

// GetAchievementUploadUrl возвращает URL для загрузки файла достижения
func (h *Handler) GetAchievementUploadUrl(ctx context.Context, req *achievementv1.GetAchievementUploadRequest) (*achievementv1.UploadUrlResponse, error) {
	log.Printf("Handler: GetAchievementUploadUrl called for user: %s, achievement: %s",
		req.GetUserUuid(), req.GetAchievementName())

	url, s3Key, err := h.service.Achievement.GetAchievementUploadURL(
		req.GetUserUuid(),
		req.GetAchievementName(),
		req.GetFileName(),
		req.GetFileType(),
		req.GetFileSize(),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Handler: Returning upload URL and S3 key: %s", s3Key)

	return &achievementv1.UploadUrlResponse{
		UploadUrl: url,
		S3Key:     s3Key,
		ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
	}, nil
}

// AddAchievementMeta добавляет метаданные достижения
func (h *Handler) AddAchievementMeta(ctx context.Context, req *achievementv1.AddAchievementMetaRequest) (*commonv1.Empty, error) {
	log.Printf("Handler: AddAchievementMeta called")

	if req.GetMeta() == nil {
		return nil, status.Error(codes.InvalidArgument, "metadata is required")
	}

	if req.GetS3Key() == "" {
		return nil, status.Error(codes.InvalidArgument, "s3_key is required")
	}

	meta := req.GetMeta()

	err := h.service.Achievement.AddAchievementMeta(
		meta.GetUserUuid(),
		meta.GetName(),
		meta.GetFileName(),
		meta.GetFileType(),
		meta.GetFileSize(),
		req.GetS3Key(),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Handler: Successfully added achievement metadata with S3 key: %s", req.GetS3Key())
	return &commonv1.Empty{}, nil
}

// DeleteAchievement удаляет достижение
func (h *Handler) DeleteAchievement(ctx context.Context, req *achievementv1.DeleteAchievementRequest) (*commonv1.Empty, error) {
	log.Printf("Handler: DeleteAchievement called for user: %s, achievement: %s",
		req.GetUserUuid(), req.GetAchievementName())

	if err := h.service.Achievement.DeleteAchievement(req.GetUserUuid(), req.GetAchievementName()); err != nil {
		return nil, err
	}

	return &commonv1.Empty{}, nil
}
