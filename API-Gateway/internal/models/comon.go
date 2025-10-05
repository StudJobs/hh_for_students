package models

// Pagination HTTP модель
type Pagination struct {
	Page  int32 `json:"page"`
	Limit int32 `json:"limit"`
}

// PaginationResponse HTTP модель
type PaginationResponse struct {
	Total       int32 `json:"total"`
	Pages       int32 `json:"pages"`
	CurrentPage int32 `json:"current_page"`
}

// Error HTTP модель
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
