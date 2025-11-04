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

// Error HTTP модель ошибки
// @Description Стандартная модель ошибки API
type Error struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid request parameters"`
}

// ValidationError HTTP модель ошибки валидации
// @Description Детализированная ошибка валидации
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"Email is required"`
	Code    string `json:"code" example:"REQUIRED_FIELD"`
}

// ErrorResponse HTTP модель ответа с ошибкой
// @Description Расширенный ответ с ошибкой
type ErrorResponse struct {
	Error   string            `json:"error" example:"Bad Request"`
	Message string            `json:"message" example:"Invalid input data"`
	Code    string            `json:"code" example:"INVALID_INPUT"`
	Details []ValidationError `json:"details,omitempty"`
}
