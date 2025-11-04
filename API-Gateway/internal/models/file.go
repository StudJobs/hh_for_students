package models

import "github.com/google/uuid"

// FileInfo информация о файле
// @Description Детальная информация о загруженном файле
type FileInfo struct {
	ID        *uuid.UUID `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string     `json:"name" example:"avatar_12345.jpg"`
	URL       *string    `json:"url,omitempty" example:"https://api.example.com/files/user123/avatar_12345.jpg"`
	DirectURL *string    `json:"direct_url,omitempty" example:"https://cdn.example.com/avatars/avatar_12345.jpg"`
	Type      string     `json:"type" example:"image" enums:"image,document,other"`
	Category  string     `json:"category" example:"avatar" enums:"avatar,resume,logo,document,attachment"`
}

// FileUploadResponse ответ после загрузки файла
// @Description Ответ после успешной загрузки файла
type FileUploadResponse struct {
	FileInfo *FileInfo `json:"file_info"`
	Message  string    `json:"message" example:"File uploaded successfully"`
}

// FileDownloadResponse ответ с URL для скачивания
// @Description Ответ с URL для скачивания файла
type FileDownloadResponse struct {
	URL       string `json:"url" example:"https://s3.amazonaws.com/bucket/file.pdf?signature=..."`
	ExpiresAt int64  `json:"expires_at" example:"1696523400"`
}

// FileListResponse список файлов
// @Description Список файлов с пагинацией
type FileListResponse struct {
	Files      []FileInfo         `json:"files"`
	Pagination PaginationResponse `json:"pagination"`
}
