package models

// AchievementMeta HTTP модель метаданных достижения
// @Description Метаданные достижения пользователя
type AchievementMeta struct {
	Name      string `json:"name" example:"Диплом бакалавра"`
	UserUUID  string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	FileName  string `json:"file_name" example:"diploma.pdf"`
	FileType  string `json:"file_type" example:"application/pdf"`
	FileSize  int64  `json:"file_size" example:"2048576"`
	CreatedAt string `json:"created_at" example:"2023-10-05T14:30:00Z"`
}

// AchievementList HTTP модель списка достижений
// @Description Список достижений пользователя
type AchievementList struct {
	Achievements []AchievementMeta `json:"achievements"`
}

// AchievementUrl HTTP модель URL для скачивания
// @Description URL для скачивания файла достижения
type AchievementUrl struct {
	URL       string `json:"url" example:"https://s3.amazonaws.com/bucket/file.pdf?signature=..."`
	ExpiresAt int64  `json:"expires_at" example:"1696523400"`
}

// UploadUrlResponse HTTP модель URL для загрузки
// @Description URL для загрузки файла в S3
type UploadUrlResponse struct {
	UploadURL string `json:"upload_url" example:"https://s3.amazonaws.com/bucket/upload?signature=..."`
	S3Key     string `json:"s3_key" example:"achievements/user123/diploma.pdf"`
	ExpiresAt int64  `json:"expires_at" example:"1696523400"`
}

// AchievementUploadRequest HTTP модель запроса на создание достижения
// @Description Запрос на подготовку загрузки достижения
type AchievementUploadRequest struct {
	Name     string `json:"name" example:"Диплом бакалавра" validate:"required"`
	FileName string `json:"file_name" example:"diploma.pdf" validate:"required"`
	FileType string `json:"file_type" example:"application/pdf" validate:"required"`
	FileSize int64  `json:"file_size" example:"2048576" validate:"required,min=1"`
}

// AchievementCreateResponse HTTP модель ответа при создании достижения
// @Description Ответ с данными для загрузки файла достижения
type AchievementCreateResponse struct {
	Meta      AchievementMeta   `json:"meta"`
	UploadURL UploadUrlResponse `json:"upload_url"`
}
