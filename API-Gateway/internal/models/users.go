package models

import "github.com/google/uuid"

// User HTTP модель пользователя
// @Description Модель данных пользователя
type User struct {
	ID                 uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	FirstName          string    `json:"first_name" example:"Иван"`
	LastName           string    `json:"last_name" example:"Иванов"`
	Age                int32     `json:"age" example:"25"`
	Tg                 string    `json:"tg" example:"@ivanov"`
	Email              string    `json:"email" example:"ivan@example.com"`
	Description        string    `json:"description" example:"Опытный разработчик"`
	ProfessionCategory string    `json:"profession_category" example:"Backend Developer"`

	// Ссылки на файлы
	ResumeURL *string    `json:"resume_url,omitempty" example:"https://example.com/files/user/resume.pdf"`
	AvatarURL *string    `json:"avatar_url,omitempty" example:"https://example.com/files/user/avatar.jpg"`
	ResumeID  *uuid.UUID `json:"resume_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
}

// ProfileList HTTP модель списка пользователей
// @Description Список пользователей с пагинацией
type ProfileList struct {
	Profiles   []User             `json:"profiles"`
	Pagination PaginationResponse `json:"pagination"`
}

// UserUpdateRequest HTTP модель для обновления пользователя
// @Description Запрос на обновление данных пользователя
type UserUpdateRequest struct {
	FirstName          *string `json:"first_name,omitempty" example:"Петр"`
	LastName           *string `json:"last_name,omitempty" example:"Петров"`
	Age                *int32  `json:"age,omitempty" example:"26"`
	Tg                 *string `json:"tg,omitempty" example:"@petrov"`
	Email              *string `json:"email,omitempty" example:"petr@example.com"`
	Description        *string `json:"description,omitempty" example:"Fullstack разработчик"`
	ProfessionCategory *string `json:"profession_category,omitempty" example:"Fullstack Developer"`
	ResumeID           *string `json:"resume_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	AvatarID           *string `json:"avatar_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
}
