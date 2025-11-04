package models

// LoginRequest HTTP модель для входа
// @Description Запрос на аутентификацию пользователя
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=6"`
	Role     string `json:"role" example:"ROLE_STUDENT" validate:"required,oneof=ROLE_STUDENT ROLE_DEVELOPER ROLE_HR ROLE_COMPANY"`
}

// SignUpRequest HTTP модель для регистрации
// @Description Запрос на регистрацию нового пользователя
type SignUpRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=6"`
	Role     string `json:"role" example:"ROLE_STUDENT" validate:"required,oneof=ROLE_STUDENT ROLE_DEVELOPER ROLE_HR ROLE_COMPANY"`
}

// AuthResponse HTTP модель ответа аутентификации
// @Description Ответ с данными аутентификации
type AuthResponse struct {
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserUUID string `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role     string `json:"role" example:"ROLE_STUDENT"`
}
