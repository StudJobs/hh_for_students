package DB

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"time"
)

// S3Config содержит конфигурацию для подключения к MinIO/S3
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

func NewMinioClient(config S3Config) (*minio.Client, error) {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания MinIO клиента: %w", err)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MinIO: %w", err)
	}

	// Создаем бакет если не существует
	exists, err := minioClient.BucketExists(ctx, config.Bucket)
	if err == nil && !exists {
		err = minioClient.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("Предупреждение: не удалось создать бакет %s: %v", config.Bucket, err)
		} else {
			log.Printf("Бакет %s успешно создан", config.Bucket)
		}
	}

	return minioClient, nil
}
