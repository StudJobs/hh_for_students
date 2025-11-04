package models

import (
	"time"
)

// Achievement представляет модель достижения в базе данных
type Achievement struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`       // Уникальное имя достижения в рамках пользователя
	UserUUID  string     `json:"user_uuid"`  // UUID владельца достижения
	FileName  string     `json:"file_name"`  // Оригинальное имя файла
	FileType  string     `json:"file_type"`  // MIME-тип файла
	FileSize  int64      `json:"file_size"`  // Размер файла в байтах
	S3Key     string     `json:"s3_key"`     // Ключ файла в S3
	CreatedAt time.Time  `json:"created_at"` // Время создания
	UpdatedAt time.Time  `json:"updated_at"` // Время обновления
	DeletedAt *time.Time `json:"deleted_at"` // Время мягкого удаления
}
