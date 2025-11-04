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
	Name      string
	UserUUID  string
	FileName  string
	FileType  string
	FileSize  int64
	S3Key     string
	CreatedAt time.Time
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
	for i, achievement := range achievements {
		result[i] = &AchievementResponse{
			Name:      achievement.Name,
			UserUUID:  achievement.UserUUID,
			FileName:  achievement.FileName,
			FileType:  achievement.FileType,
			FileSize:  achievement.FileSize,
			S3Key:     achievement.S3Key,
			CreatedAt: achievement.CreatedAt,
		}
	}

	log.Printf("Service: Retrieved %d achievements for user %s", len(result), userUUID)
	return result, nil
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

// AddAchievementMeta добавляет метаданные достижения после успешной загрузки
func (s *AchievementService) AddAchievementMeta(userUUID, achievementName, fileName, fileType string, fileSize int64, s3Key string) error {
	log.Printf("Service: Adding achievement metadata for user %s, achievement: %s, S3 key: %s", userUUID, achievementName, s3Key)

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
		CreatedAt: time.Now(),
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
