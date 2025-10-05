package models

// Company HTTP модель
type Company struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	City        string `json:"city"`
	Site        string `json:"site"`
	Type        string `json:"type"`
}

// Vacancy HTTP модель
type Vacancy struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Experience     int32  `json:"experience"`
	Salary         int32  `json:"salary"`
	PositionStatus string `json:"position_status"`
	UsageFormat    string `json:"usage_format"`
	Schedule       int32  `json:"schedule"`
	WorkFormat     string `json:"work_format"`
	CompanyID      string `json:"company_id"`
	DateStart      string `json:"date_start"`
}

// VacancyList HTTP модель
type VacancyList struct {
	Vacancies  []Vacancy          `json:"vacancies"`
	Pagination PaginationResponse `json:"pagination"`
}
