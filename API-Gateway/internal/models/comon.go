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

// SuccessResponse универсальный успешный ответ
// @Description Успешный ответ API
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// VacancyAttachmentResponse ответ с вложением вакансии
// @Description Ответ с информацией о вложении вакансии
type VacancyAttachmentResponse struct {
	AttachmentID  *string `json:"attachment_id,omitempty" example:"attach-123"`
	AttachmentURL *string `json:"attachment_url,omitempty" example:"https://cdn.example.com/vacancies/attach-123.pdf"`
}

// FileValidationRequest запрос на валидацию файла
// @Description Запрос на проверку возможности загрузки файла
type FileValidationRequest struct {
	FileName string `json:"file_name" example:"document.pdf"`
	FileSize int64  `json:"file_size" example:"1048576"`
	FileType string `json:"file_type" example:"application/pdf"`
}

// FileValidationResponse ответ валидации файла
// @Description Ответ с результатом проверки файла
type FileValidationResponse struct {
	Valid   bool     `json:"valid" example:"true"`
	Message string   `json:"message,omitempty" example:"File validation successful"`
	Errors  []string `json:"errors,omitempty" example:"File size too large"`
}
