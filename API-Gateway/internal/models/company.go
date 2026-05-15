package models

type CompanyType struct {
	Value string `json:"value"`
}

type Company struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	City        string       `json:"city"`
	Site        string       `json:"site"`
	Type        *CompanyType `json:"type,omitempty"`

	// Auto-cleanup policy: каждые N часов фоновый воркер Gateway soft-удаляет
	// closed-вакансии (старше CleanupVacanciesAfterDays дней) и completed-микрозадачи
	// (старше CleanupTasksAfterDays). 0 = не чистить.
	CleanupVacanciesAfterDays int32 `json:"cleanup_vacancies_after_days,omitempty"`
	CleanupTasksAfterDays     int32 `json:"cleanup_tasks_after_days,omitempty"`

	// Ссылки на файлы
	LogoURL *string `json:"logo_url,omitempty"` // Ссылка на логотип
	LogoID  *string `json:"logo_id,omitempty"`  // ID логотипа в achievements
}

type CompanyList struct {
	Companies  []*Company          `json:"companies"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
