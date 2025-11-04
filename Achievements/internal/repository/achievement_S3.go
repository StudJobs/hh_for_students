package repository

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/minio/minio-go/v7"
)

type S3Repository struct {
	client     *minio.Client
	bucketName string
}

func NewS3Repository(client *minio.Client) *S3Repository {
	bucketName := getEnv("MINIO_BUCKET", "achievements")
	return &S3Repository{
		client:     client,
		bucketName: bucketName,
	}
}

// GenerateUploadURL генерирует presigned URL для загрузки файла
func (r *S3Repository) GenerateUploadURL(ctx context.Context, s3Key, fileType string, expiry int64) (string, error) {
	log.Printf("S3Repository: Generating upload URL for key: %s, type: %s", s3Key, fileType)

	expiryTime := time.Duration(expiry) * time.Minute
	url, err := r.client.PresignedPutObject(ctx, r.bucketName, s3Key, expiryTime)
	if err != nil {
		log.Printf("S3Repository: Failed to generate upload URL for %s: %v", s3Key, err)
		return "", status.Error(codes.Internal, "failed to generate upload URL")
	}

	log.Printf("S3Repository: Successfully generated upload URL for key: %s", s3Key)
	return url.String(), nil
}

// GenerateDownloadURL генерирует presigned URL для скачивания файла
func (r *S3Repository) GenerateDownloadURL(ctx context.Context, s3Key string, expiry int64) (string, error) {
	log.Printf("S3Repository: Generating download URL for key: %s", s3Key)

	expiryTime := time.Duration(expiry) * time.Minute
	url, err := r.client.PresignedGetObject(ctx, r.bucketName, s3Key, expiryTime, nil)
	if err != nil {
		log.Printf("S3Repository: Failed to generate download URL for %s: %v", s3Key, err)
		return "", status.Error(codes.Internal, "failed to generate download URL")
	}

	log.Printf("S3Repository: Successfully generated download URL for key: %s", s3Key)
	return url.String(), nil
}

// DeleteObject удаляет файл из S3
func (r *S3Repository) DeleteObject(ctx context.Context, s3Key string) error {
	log.Printf("S3Repository: Deleting object with key: %s", s3Key)

	err := r.client.RemoveObject(ctx, r.bucketName, s3Key, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("S3Repository: Failed to delete object %s: %v", s3Key, err)
		return status.Error(codes.Internal, "failed to delete file")
	}

	log.Printf("S3Repository: Successfully deleted object with key: %s", s3Key)
	return nil
}

// ObjectExists проверяет существование объекта в S3
func (r *S3Repository) ObjectExists(ctx context.Context, s3Key string) (bool, error) {
	_, err := r.client.StatObject(ctx, r.bucketName, s3Key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		log.Printf("S3Repository: Error checking object existence for %s: %v", s3Key, err)
		return false, err
	}
	return true, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
