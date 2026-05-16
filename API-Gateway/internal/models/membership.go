package models

// CompanyMember — HTTP-модель HR-сотрудника компании.
type CompanyMember struct {
	ID         string `json:"id"`
	CompanyID  string `json:"company_id"`
	UserID     string `json:"user_id"`
	Status     int32  `json:"status"` // 1=PENDING, 2=APPROVED, 3=REJECTED
	Note       string `json:"note,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	ReviewedAt string `json:"reviewed_at,omitempty"`
}
