package models

// Vacancy представляет вакансию
type Vacancy struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Experience     int32   `json:"experience"`
	Salary         int32   `json:"salary"`
	PositionStatus string  `json:"position_status"`
	Schedule       string  `json:"schedule"`
	WorkFormat     string  `json:"work_format"`
	CompanyID      string  `json:"company_id"`
	CreateAt       string  `json:"create_at"`
	AttachmentID   *string `json:"attachment_id,omitempty"`
	AttachmentURL  *string `json:"attachment_url,omitempty"`
}

// VacancyList представляет список вакансий с пагинацией
type VacancyList struct {
	Vacancies  []*Vacancy          `json:"vacancies"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
