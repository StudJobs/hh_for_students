package models

// Application — HTTP-модель отклика студента на вакансию.
// Числовые значения status совпадают с ApplicationStatus в proto:
//   1 = PENDING (ожидает решения HR)
//   2 = ACCEPTED
//   3 = REJECTED
type Application struct {
	ID          string `json:"id"`
	VacancyID   string `json:"vacancy_id"`
	StudentID   string `json:"student_id"`
	CoverLetter string `json:"cover_letter"`
	Status      int32  `json:"status"`
	HRComment   string `json:"hr_comment"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ApplicationList — список откликов с пагинацией.
type ApplicationList struct {
	Applications []*Application      `json:"applications"`
	Pagination   *PaginationResponse `json:"pagination,omitempty"`
}

// ApplyRequest — payload для POST /vacancy/:id/respond
type ApplyRequest struct {
	CoverLetter string `json:"cover_letter"`
}

// ApplicationReviewRequest — payload для PATCH /hr/applications/:id
type ApplicationReviewRequest struct {
	Decision int32  `json:"decision"` // 2 = ACCEPT, 3 = REJECT
	Comment  string `json:"comment,omitempty"`
}
