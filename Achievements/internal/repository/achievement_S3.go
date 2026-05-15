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
	client       *minio.Client // internal: PUT, RemoveObject, StatObject
	publicClient *minio.Client // public: PresignedGetObject (URL отдаётся клиенту-браузеру)
	bucketName   string
}

// NewS3Repository принимает оба клиента; publicClient может быть nil — тогда
// для presigned GET используется тот же клиент, что и для PUT.
func NewS3Repository(client, publicClient *minio.Client) *S3Repository {
	bucketName := getEnv("MINIO_BUCKET", "achievements")
	if publicClient == nil {
		publicClient = client
	}
	return &S3Repository{
		client:       client,
		publicClient: publicClient,
		bucketName:   bucketName,
	}
}

// GenerateUploadURL генерирует presigned PUT URL для загрузки файла.
// URL уходит к клиенту в браузер (через 3-step flow: meta → PUT → confirm),
// поэтому host должен быть browser-reachable. Используем publicClient — тот
// же, что и для GET. Gateway, делающий PUT из docker-контейнера для avatar/
// resume, подменяет host обратно на internal в момент connect, но Host header
// шлёт public — это удовлетворяет AWS Sig V4 верификацию (см. uploadToPresignedURL).
func (r *S3Repository) GenerateUploadURL(ctx context.Context, s3Key, fileType string, expiry int64) (string, error) {
	log.Printf("S3Repository: Generating upload URL for key: %s, type: %s", s3Key, fileType)

	expiryTime := time.Duration(expiry) * time.Minute
	presignedURL, err := r.publicClient.PresignedPutObject(ctx, r.bucketName, s3Key, expiryTime)
	if err != nil {
		log.Printf("S3Repository: Failed to generate upload URL for %s: %v", s3Key, err)
		return "", status.Error(codes.Internal, "failed to generate upload URL")
	}

	log.Printf("S3Repository: Successfully generated upload URL for key: %s", s3Key)
	return presignedURL.String(), nil
}

// GenerateDownloadURL генерирует presigned URL для скачивания файла.
// URL отдаётся клиенту-браузеру, поэтому используется publicClient с host'ом,
// доступным из браузера (например `localhost:9000`).
func (r *S3Repository) GenerateDownloadURL(ctx context.Context, s3Key string, expiry int64) (string, error) {
	log.Printf("S3Repository: Generating download URL for key: %s", s3Key)

	expiryTime := time.Duration(expiry) * time.Minute
	presignedURL, err := r.publicClient.PresignedGetObject(ctx, r.bucketName, s3Key, expiryTime, nil)
	if err != nil {
		log.Printf("S3Repository: Failed to generate download URL for %s: %v", s3Key, err)
		return "", status.Error(codes.Internal, "failed to generate download URL")
	}

	log.Printf("S3Repository: Successfully generated download URL for key: %s", s3Key)
	return presignedURL.String(), nil
}

// DeleteObject удаляет файл из S3 (internal-клиент).
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

// ObjectExists проверяет существование объекта в S3 (internal-клиент).
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
