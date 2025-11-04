package models

// S3Config содержит конфигурацию для подключения к MinIO/S3
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

// UploadRequest содержит данные для загрузки файла
type UploadRequest struct {
	UserUUID        string
	AchievementName string
	FileName        string
	FileType        string
	FileSize        int64
}

// DownloadRequest содержит данные для скачивания файла
type DownloadRequest struct {
	UserUUID        string
	AchievementName string
}
