package models

import "github.com/google/uuid"

// FileInfo HTTP модель информации о файле
type FileInfo struct {
	ID        *uuid.UUID `json:"id,omitempty"`
	Name      string     `json:"name"`
	URL       *string    `json:"url,omitempty"`        // Presigned URL для фронта
	DirectURL *string    `json:"direct_url,omitempty"` // Прямая ссылка через бекенд
	Type      string     `json:"type"`                 // "image", "document", "other"
	Category  string     `json:"category"`             // "avatar", "resume", "logo", etc.
}

// FileUploadResponse HTTP модель ответа на загрузку файла
type FileUploadResponse struct {
	FileInfo *FileInfo `json:"file_info"`
	Message  string    `json:"message"`
}
