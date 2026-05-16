package storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config — параметры подключения к MinIO/S3.
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

// Solutions хранит файлы-решения микрозадач в подпрефиксе бакета.
// Объект кладётся по пути `microtask-solutions/<microtask_id>/<student_id>/<file_id>`.
type Solutions struct {
	internal *minio.Client // PUT/HEAD/DELETE — host доступный из docker-network
	public   *minio.Client // PresignedPUT/PresignedGET — host доступный из браузера
	bucket   string
	prefix   string
}

// NewInternalClient валидирует подключение и создаёт бакет, если отсутствует.
func NewInternalClient(cfg S3Config) (*minio.Client, error) {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:    cfg.UseSSL,
		Transport: transport,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := cli.ListBuckets(ctx); err != nil {
		return nil, fmt.Errorf("connect minio: %w", err)
	}
	exists, err := cli.BucketExists(ctx, cfg.Bucket)
	if err == nil && !exists {
		if err := cli.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			log.Printf("storage: failed to create bucket %s: %v", cfg.Bucket, err)
		}
	}
	log.Printf("storage: connected to MinIO endpoint=%s bucket=%s", cfg.Endpoint, cfg.Bucket)
	return cli, nil
}

// NewPublicClient — клиент только для presigned URL (HMAC локально, network к нему не нужен).
func NewPublicClient(cfg S3Config) (*minio.Client, error) {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:    cfg.UseSSL,
		Region:    "us-east-1",
		Transport: transport,
	})
	if err != nil {
		return nil, fmt.Errorf("create public minio client: %w", err)
	}
	log.Printf("storage: public presigner created for endpoint=%s", cfg.Endpoint)
	return cli, nil
}

func NewSolutions(internal, public *minio.Client, bucket string) *Solutions {
	if public == nil {
		public = internal
	}
	return &Solutions{
		internal: internal,
		public:   public,
		bucket:   bucket,
		prefix:   "microtask-solutions",
	}
}

// Key собирает S3-ключ для файла. fileID — это slug (uuid+filename), он же возвращается
// клиенту, а потом сохраняется в БД в Submission.SolutionFileName.
func (s *Solutions) Key(microtaskID, studentID, fileID string) string {
	return fmt.Sprintf("%s/%s/%s/%s", s.prefix, microtaskID, studentID, fileID)
}

func (s *Solutions) PresignedPut(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.public.PresignedPutObject(ctx, s.bucket, key, ttl)
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return u.String(), nil
}

func (s *Solutions) PresignedGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.public.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}
	return u.String(), nil
}

func (s *Solutions) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.internal.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
