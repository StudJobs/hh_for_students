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
		protoAchievements[i] = toProtoMeta(achievement)
	}

	return &achievementv1.AchievementList{
		Achievements: protoAchievements,
	}, nil
}

// toProtoMeta — единая конвертация service.AchievementResponse -> proto.AchievementMeta.
func toProtoMeta(a *service.AchievementResponse) *achievementv1.AchievementMeta {
	m := &achievementv1.AchievementMeta{
		Id:                 a.ID,
		Name:               a.Name,
		UserUuid:           a.UserUUID,
		FileName:           a.FileName,
		FileType:           a.FileType,
		FileSize:           a.FileSize,
		CreatedAt:          a.CreatedAt.Format(time.RFC3339),
		Type:               achievementv1.AchievementType(a.Type),
		VerificationStatus: achievementv1.VerificationStatus(a.VerificationStatus),
		ReviewedBy:         a.ReviewedBy,
		ReviewComment:      a.ReviewComment,
	}
	if !a.ReviewedAt.IsZero() {
		m.ReviewedAt = a.ReviewedAt.Format(time.RFC3339)
	}
	return m
}

// SubmitForReview — студент отправляет своё DRAFT-достижение на ревью.
func (h *Handler) SubmitForReview(ctx context.Context, req *achievementv1.SubmitForReviewRequest) (*commonv1.Empty, error) {
	if req.GetUserUuid() == "" || req.GetAchievementId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_uuid and achievement_id are required")
	}
	if err := h.service.Achievement.SubmitForReview(ctx, req.GetUserUuid(), req.GetAchievementId()); err != nil {
		return nil, err
	}
	return &commonv1.Empty{}, nil
}

// GetExpertQueue — список PENDING для эксперта.
func (h *Handler) GetExpertQueue(ctx context.Context, req *achievementv1.GetExpertQueueRequest) (*achievementv1.AchievementList, error) {
	page := req.GetPage()
	limit := req.GetLimit()
	achievements, err := h.service.Achievement.GetExpertQueue(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*achievementv1.AchievementMeta, len(achievements))
	for i, a := range achievements {
		out[i] = toProtoMeta(a)
	}
	return &achievementv1.AchievementList{Achievements: out}, nil
}

// ReviewAchievement — эксперт принимает решение.
func (h *Handler) ReviewAchievement(ctx context.Context, req *achievementv1.ReviewAchievementRequest) (*commonv1.Empty, error) {
	if req.GetAchievementId() <= 0 || req.GetReviewerUuid() == "" {
		return nil, status.Error(codes.InvalidArgument, "achievement_id and reviewer_uuid are required")
	}
	if err := h.service.Achievement.ReviewAchievement(
		ctx,
		req.GetAchievementId(),
		req.GetReviewerUuid(),
		int32(req.GetDecision()),
		req.GetComment(),
	); err != nil {
		return nil, err
	}
	return &commonv1.Empty{}, nil
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
		int32(meta.GetType()),
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
