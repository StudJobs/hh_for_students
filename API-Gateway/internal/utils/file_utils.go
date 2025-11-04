package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
)

type FileHandler struct {
	apiService *services.ApiGateway
}

func NewFileHandler(apiService *services.ApiGateway) *FileHandler {
	return &FileHandler{
		apiService: apiService,
	}
}

// GetFileInfo возвращает информацию о файле для отдачи в API
func (fh *FileHandler) GetFileInfo(
	ctx context.Context,
	entityID string,
	fileName string,
	category string,
) (*models.FileInfo, error) {
	if fileName == "" {
		return nil, nil
	}

	// Получаем URL для скачивания
	downloadURL, err := fh.apiService.Achievement.GetAchievementDownloadUrl(ctx, entityID, fileName)
	if err != nil {
		log.Printf("GetFileInfo: Failed to get download URL for %s: %v", fileName, err)
		return nil, err
	}

	// Определяем тип файла
	fileType := fh.detectFileType(fileName)

	fileInfo := &models.FileInfo{
		Name:     fileName,
		Type:     fileType,
		Category: category,
		URL:      &downloadURL.URL, // Всегда отдаем presigned URL
	}

	// Для изображений аватаров и логотипов добавляем прямую ссылку
	if fileType == "image" && (category == "avatar" || category == "logo") {
		directURL := fmt.Sprintf("/api/v1/files/%s/%s", entityID, fileName)
		fileInfo.DirectURL = &directURL
	}

	return fileInfo, nil
}

// UploadFileDirect загружает файл напрямую (объединяет создание и подтверждение)
func (fh *FileHandler) UploadFileDirect(
	ctx context.Context,
	entityID string,
	entityType string, // "user", "company"
	category string,
	fileHeader *multipart.FileHeader,
) (*models.FileInfo, error) {
	// Создаем уникальное имя для файла
	fileName := fmt.Sprintf("%s_%s_%s_%d", entityType, category, entityID, time.Now().Unix())

	// Получаем URL для загрузки
	uploadResponse, err := fh.apiService.Achievement.GetAchievementUploadUrl(
		ctx,
		entityID,
		fileName,
		fileHeader.Filename,
		fileHeader.Header.Get("Content-Type"),
		fileHeader.Size,
	)
	if err != nil {
		log.Printf("UploadFileDirect: Failed to get upload URL for %s: %v", fileName, err)
		return nil, err
	}

	// Загружаем файл в S3
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Загружаем файл по presigned URL
	err = fh.uploadToPresignedURL(uploadResponse.UploadURL, fileData, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("UploadFileDirect: Failed to upload file to S3 for %s: %v", fileName, err)
		return nil, err
	}

	// Создаем метаданные
	achievementMeta := &models.AchievementMeta{
		Name:      fileName,
		UserUUID:  entityID,
		FileName:  fileHeader.Filename,
		FileType:  fileHeader.Header.Get("Content-Type"),
		FileSize:  fileHeader.Size,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Подтверждаем загрузку
	err = fh.apiService.Achievement.AddAchievementMeta(ctx, achievementMeta, uploadResponse.S3Key)
	if err != nil {
		log.Printf("UploadFileDirect: Failed to add achievement meta for %s: %v", fileName, err)
		return nil, err
	}

	// Получаем информацию о файле для ответа
	return fh.GetFileInfo(ctx, entityID, fileName, category)
}

// uploadToPresignedURL загружает файл по presigned URL
func (fh *FileHandler) uploadToPresignedURL(presignedURL string, fileData []byte, contentType string) error {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("PUT", presignedURL, strings.NewReader(string(fileData)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(fileData))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	return nil
}

// detectFileType определяет тип файла по имени
func (fh *FileHandler) detectFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		return "image"
	case ".pdf", ".doc", ".docx", ".txt", ".rtf":
		return "document"
	default:
		return "other"
	}
}

// ShouldServeDirect определяет, нужно ли отдавать файл напрямую
func (fh *FileHandler) ShouldServeDirect(fileType, category string) bool {
	return fileType == "image" && (category == "avatar" || category == "logo")
}

// DeleteFile удаляет файл
func (fh *FileHandler) DeleteFile(
	ctx context.Context,
	entityID string,
	fileName string,
) error {
	return fh.apiService.Achievement.DeleteAchievement(ctx, entityID, fileName)
}
