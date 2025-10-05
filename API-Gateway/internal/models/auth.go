package models

// LoginRequest HTTP модель
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "ROLE_STUDENT", "ROLE_DEVELOPER", "ROLE_EMPLOYER"
}

// SignUpRequest HTTP модель
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// AuthResponse HTTP модель
type AuthResponse struct {
	Token    string `json:"token"`
	UserUUID string `json:"user_uuid"`
	Role     string `json:"role"`
}
