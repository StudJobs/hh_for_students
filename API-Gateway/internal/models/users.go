package models

import "github.com/google/uuid"

// Profile HTTP модель
type User struct {
	ID                 uuid.UUID `json:"id"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Age                int32     `json:"age"`
	Tg                 string    `json:"tg"`
	ResumeID           uuid.UUID `json:"resume_id"`
	Email              string    `json:"email"`
	Description        string    `json:"description"`
	ProfessionCategory string    `json:"profession_category"`
}

// ProfileList HTTP модель
type ProfileList struct {
	Profiles   []User             `json:"profiles"`
	Pagination PaginationResponse `json:"pagination"`
}
