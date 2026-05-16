package service

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/studjobs/hh_for_students/achievments/internal/repository"
)

// AchievementResponse представляет ответ с данными достижения
type AchievementResponse struct {
	ID                 int64
	Name               string
	UserUUID           string
	FileName           string
	FileType           string
	FileSize           int64
	S3Key              string
	Type               int32
	CreatedAt          time.Time
	VerificationStatus int32
	ReviewedBy         string
	ReviewedAt         time.Time
	ReviewComment      string
	ExternalURL        string
	Description        string
	SkillSlug          string
}

func toResponse(a *repository.AchievementDB) *AchievementResponse {
	r := &AchievementResponse{
		ID:                 a.ID,
		Name:               a.Name,
		UserUUID:           a.UserUUID,
		FileName:           a.FileName,
		FileType:           a.FileType,
		FileSize:           a.FileSize,
		S3Key:              a.S3Key,
		Type:               a.Type,
		CreatedAt:          a.CreatedAt,
		VerificationStatus: a.VerificationStatus,
	}
	if a.ReviewedBy != nil {
		r.ReviewedBy = *a.ReviewedBy
	}
	if a.ReviewedAt != nil {
		r.ReviewedAt = *a.ReviewedAt
	}
	if a.ReviewComment != nil {
		r.ReviewComment = *a.ReviewComment
	}
	if a.ExternalURL != nil {
		r.ExternalURL = *a.ExternalURL
	}
	if a.Description != nil {
		r.Description = *a.Description
	}
	r.SkillSlug = a.SkillSlug
	return r
}

type AchievementService struct {
	repo *repository.Repository
}

func NewAchievementService(repo *repository.Repository) *AchievementService {
	return &AchievementService{repo: repo}
}

// GetAllAchievements возвращает все достижения пользователя
func (s *AchievementService) GetAllAchievements(userUUID string) ([]*AchievementResponse, error) {
	log.Printf("Service: Getting all achievements for user: %s", userUUID)

	if userUUID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_uuid is required")
	}

	ctx := context.Background()
	achievements, err := s.repo.Achievement.GetAchievementsByUser(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	result := make([]*AchievementResponse, len(achievements))
	for i, a := range achievements {
		result[i] = toResponse(a)
	}

	log.Printf("Service: Retrieved %d achievements for user %s", len(result), userUUID)
	return result, nil
}

// SubmitForReview — студент переводит DRAFT -> PENDING.
func (s *AchievementService) SubmitForReview(ctx context.Context, userUUID string, achievementID int64) error {
	if userUUID == "" || achievementID <= 0 {
		return status.Error(codes.InvalidArgument, "user_uuid and achievement_id are required")
	}

	a, err := s.repo.Achievement.GetAchievementByID(ctx, achievementID)
	if err != nil {
		return err
	}
	if a.UserUUID != userUUID {
		return status.Error(codes.PermissionDenied, "not your achievement")
	}
	// Разрешён переход из DRAFT(1) и REJECTED(4) — повторная отправка после правок.
	if a.VerificationStatus != 1 && a.VerificationStatus != 4 {
		return status.Errorf(codes.FailedPrecondition, "cannot submit from status %d", a.VerificationStatus)
	}

	if _, err := s.repo.Achievement.SetVerificationStatus(ctx, achievementID, 2, a.VerificationStatus, "", ""); err != nil {
		return err
	}
	log.Printf("Service: SubmitForReview ok: id=%d, user=%s", achievementID, userUUID)
	return nil
}

// GetExpertQueue — все PENDING-достижения с пагинацией.
func (s *AchievementService) GetExpertQueue(ctx context.Context, page, limit int32) ([]*AchievementResponse, error) {
	achievements, err := s.repo.Achievement.ListPending(ctx, page, limit)
	if err != nil {
		return nil, err
	}
	result := make([]*AchievementResponse, len(achievements))
	for i, a := range achievements {
		result[i] = toResponse(a)
	}
	return result, nil
}

// CreateMicrotaskAchievement создаёт сразу APPROVED-ачивку после approve микрозадачи.
// Идемпотентно: повторный вызов с тем же microtask_id+user_uuid просто скипается.
func (s *AchievementService) CreateMicrotaskAchievement(
	ctx context.Context,
	userUUID, microtaskID, microtaskTitle, solutionURL, reviewerUUID, reviewComment string,
) error {
	if userUUID == "" || microtaskID == "" {
		return status.Error(codes.InvalidArgument, "user_uuid and microtask_id are required")
	}

	short := microtaskID
	if len(short) > 6 {
		short = short[:6]
	}
	titleClean := strings.TrimSpace(microtaskTitle)
	var name string
	if titleClean != "" {
		name = fmt.Sprintf("Микрозадача: %s · #%s", titleClean, short)
	} else {
		name = fmt.Sprintf("Микрозадача #%s", short)
	}

	s3Key := fmt.Sprintf("microtask:%s:%s", microtaskID, userUUID)
	reviewedBy := reviewerUUID
	commentVal := reviewComment

	a := &repository.AchievementDB{
		Name:               name,
		UserUUID:           userUUID,
		FileName:           solutionURL,
		FileType:           "external/url",
		FileSize:           0,
		S3Key:              s3Key,
		Type:               5, // ACHIEVEMENT_TYPE_MICROTASK_RESULT
		VerificationStatus: 3, // VERIFICATION_STATUS_APPROVED
		ReviewedBy:         &reviewedBy,
		ReviewComment:      &commentVal,
	}

	if err := s.repo.Achievement.CreateMicrotaskAchievement(ctx, a); err != nil {
		return err
	}
	log.Printf("Service: CreateMicrotaskAchievement ok: user=%s, microtask=%s", userUUID, microtaskID)
	return nil
}

// GetAchievement — простой read-by-id, нужен handler'у для проверки прав ревью.
func (s *AchievementService) GetAchievement(ctx context.Context, achievementID int64) (*repository.AchievementDB, error) {
	if achievementID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "achievement_id is required")
	}
	return s.repo.Achievement.GetAchievementByID(ctx, achievementID)
}

// ReviewAchievement — эксперт принимает решение APPROVED(3) или REJECTED(4).
// Возвращает обновлённый AchievementDB — handler использует его для side-effect'ов
// (например, AddVerifiedSkills при approve SKILL_VERIFICATION).
func (s *AchievementService) ReviewAchievement(ctx context.Context, achievementID int64, reviewerUUID string, decision int32, comment string) (*repository.AchievementDB, error) {
	if achievementID <= 0 || reviewerUUID == "" {
		return nil, status.Error(codes.InvalidArgument, "achievement_id and reviewer_uuid are required")
	}
	if decision != 3 && decision != 4 {
		return nil, status.Error(codes.InvalidArgument, "decision must be APPROVED or REJECTED")
	}
	// Переход разрешён только из PENDING(2).
	a, err := s.repo.Achievement.SetVerificationStatus(ctx, achievementID, decision, 2, reviewerUUID, comment)
	if err != nil {
		return nil, err
	}
	log.Printf("Service: ReviewAchievement ok: id=%d, reviewer=%s, decision=%d, type=%d, slug=%q", achievementID, reviewerUUID, decision, a.Type, a.SkillSlug)
	return a, nil
}

// GetAchievementDownloadURL генерирует URL для скачивания достижения
func (s *AchievementService) GetAchievementDownloadURL(userUUID, achievementName string) (string, error) {
	log.Printf("Service: Generating download URL for user %s, achievement: %s", userUUID, achievementName)

	if userUUID == "" || achievementName == "" {
		return "", status.Error(codes.InvalidArgument, "user_uuid and achievement_name are required")
	}

	ctx := context.Background()

	// Получаем достижение чтобы получить S3 ключ
	achievement, err := s.repo.Achievement.GetAchievementByName(ctx, userUUID, achievementName)
	if err != nil {
		return "", err
	}

	// Проверяем существование файла в S3 перед генерацией URL
	exists, err := s.repo.S3.ObjectExists(ctx, achievement.S3Key)
	if err != nil {
		log.Printf("Service: Error checking file existence in S3: %v", err)
		return "", status.Error(codes.Internal, "error checking file")
	}
	if !exists {
		log.Printf("Service: File not found in S3: %s", achievement.S3Key)
		return "", status.Error(codes.NotFound, "file not found in storage")
	}

	// Генерируем presigned URL с временем жизни 15 минут
	url, err := s.repo.S3.GenerateDownloadURL(ctx, achievement.S3Key, 15)
	if err != nil {
		return "", err
	}

	log.Printf("Service: Generated download URL for achievement %s of user %s", achievementName, userUUID)
	return url, nil
}

// GetAchievementUploadURL генерирует URL для загрузки файла достижения
func (s *AchievementService) GetAchievementUploadURL(userUUID, achievementName, fileName, fileType string, fileSize int64) (string, string, error) {
	log.Printf("Service: Generating upload URL for user %s, achievement: %s", userUUID, achievementName)

	if userUUID == "" || achievementName == "" || fileName == "" || fileType == "" {
		return "", "", status.Error(codes.InvalidArgument, "all fields are required")
	}

	if fileSize <= 0 {
		return "", "", status.Error(codes.InvalidArgument, "file_size must be positive")
	}

	// Генерируем уникальный S3 ключ
	s3Key := s.generateS3Key(userUUID, achievementName, fileName)

	ctx := context.Background()

	// Проверяем, не существует ли уже достижение с таким именем
	_, err := s.repo.Achievement.GetAchievementByName(ctx, userUUID, achievementName)
	if err == nil {
		return "", "", status.Errorf(codes.AlreadyExists, "achievement '%s' already exists for user %s", achievementName, userUUID)
	}

	// Генерируем presigned URL для загрузки с временем жизни 30 минут
	url, err := s.repo.S3.GenerateUploadURL(ctx, s3Key, fileType, 30)
	if err != nil {
		return "", "", err
	}

	log.Printf("Service: Generated upload URL for achievement %s of user %s, S3 key: %s", achievementName, userUUID, s3Key)
	return url, s3Key, nil
}

// AddAchievementMeta добавляет метаданные достижения после успешной загрузки.
// externalURL и description — опциональные ссылки/контекст работы, эксперт их
// видит на странице проверки.
func (s *AchievementService) AddAchievementMeta(userUUID, achievementName, fileName, fileType string, fileSize int64, s3Key string, achievementType int32, externalURL, description, skillSlug string) error {
	log.Printf("Service: Adding achievement metadata for user %s, achievement: %s, S3 key: %s, type: %d", userUUID, achievementName, s3Key, achievementType)

	if userUUID == "" || achievementName == "" || fileName == "" || fileType == "" || s3Key == "" {
		return status.Error(codes.InvalidArgument, "all fields are required")
	}

	ctx := context.Background()

	// Проверяем, что файл действительно загружен в S3
	exists, err := s.repo.S3.ObjectExists(ctx, s3Key)
	if err != nil {
		log.Printf("Service: Error checking file existence: %v", err)
		return status.Error(codes.Internal, "error verifying file upload")
	}
	if !exists {
		log.Printf("Service: File not found in S3: %s", s3Key)
		return status.Error(codes.FailedPrecondition, "file not uploaded to storage")
	}

	achievement := &repository.AchievementDB{
		Name:      achievementName,
		UserUUID:  userUUID,
		FileName:  fileName,
		FileType:  fileType,
		FileSize:  fileSize,
		S3Key:     s3Key,
		Type:      achievementType,
		CreatedAt: time.Now(),
	}
	if externalURL != "" {
		achievement.ExternalURL = &externalURL
	}
	if description != "" {
		achievement.Description = &description
	}
	if skillSlug != "" {
		achievement.SkillSlug = skillSlug
	}

	if err := s.repo.Achievement.CreateAchievement(ctx, achievement); err != nil {
		return err
	}

	log.Printf("Service: Added achievement metadata for %s of user %s with S3 key: %s", achievementName, userUUID, s3Key)
	return nil
}

// DeleteAchievement удаляет достижение
func (s *AchievementService) DeleteAchievement(userUUID, achievementName string) error {
	log.Printf("Service: Deleting achievement for user %s: %s", userUUID, achievementName)

	if userUUID == "" || achievementName == "" {
		return status.Error(codes.InvalidArgument, "user_uuid and achievement_name are required")
	}

	ctx := context.Background()

	// Сначала получаем достижение чтобы узнать S3 ключ для удаления файла
	achievement, err := s.repo.Achievement.GetAchievementByName(ctx, userUUID, achievementName)
	if err != nil {
		return err
	}

	log.Printf("Service: Found achievement with S3 key: %s", achievement.S3Key)

	// Удаляем файл из S3
	if err := s.repo.S3.DeleteObject(ctx, achievement.S3Key); err != nil {
		log.Printf("Service: Failed to delete file from S3 (key: %s): %v", achievement.S3Key, err)
		// Продолжаем удаление метаданных даже если файл не найден
	} else {
		log.Printf("Service: Successfully deleted file from S3: %s", achievement.S3Key)
	}

	// Удаляем метаданные из БД
	if err := s.repo.Achievement.DeleteAchievement(ctx, userUUID, achievementName); err != nil {
		return err
	}

	log.Printf("Service: Deleted achievement %s of user %s", achievementName, userUUID)
	return nil
}

// generateS3Key создает уникальный ключ для хранения в S3
func (s *AchievementService) generateS3Key(userUUID, achievementName, fileName string) string {
	timestamp := time.Now().UnixNano() // Используем наносекунды для большей уникальности

	// Очищаем имя файла от небезопасных символов
	safeFileName := strings.ReplaceAll(fileName, " ", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "/", "_")
	safeFileName = strings.ReplaceAll(safeFileName, "\\", "_")

	// Сохраняем расширение файла
	fileExt := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(safeFileName, fileExt)

	// Создаем безопасный ключ (формат: user_uuid/achievement_name_baseName_timestamp.ext)
	safeKey := fmt.Sprintf("%s/%s_%s_%d%s",
		userUUID,
		achievementName,
		baseName,
		timestamp,
		fileExt)

	log.Printf("Service: Generated S3 key: %s", safeKey)
	return safeKey
}
