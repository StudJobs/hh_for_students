package models

// MicroTask — HTTP-модель микрозадачи.
type MicroTask struct {
	ID          string   `json:"id"`
	CompanyID   string   `json:"company_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Reward      int32    `json:"reward"`
	Deadline    string   `json:"deadline,omitempty"`
	SkillSlugs  []string `json:"skill_slugs,omitempty"`
	Status      int32    `json:"status"` // 1=OPEN, 2=ASSIGNED, 3=COMPLETED, 4=CANCELLED
	AssignedTo  string   `json:"assigned_to,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

type MicroTaskList struct {
	Tasks      []*MicroTask        `json:"tasks"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type MicroTaskCreateRequest struct {
	CompanyID   string   `json:"company_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Reward      int32    `json:"reward"`
	Deadline    string   `json:"deadline,omitempty"`
	SkillSlugs  []string `json:"skill_slugs,omitempty"`
}

type MicroTaskUpdateRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Reward      *int32   `json:"reward,omitempty"`
	Deadline    *string  `json:"deadline,omitempty"`
	SkillSlugs  []string `json:"skill_slugs,omitempty"`
}

type SubmitRequest struct {
	SolutionURL string `json:"solution_url"`
	Comment     string `json:"comment,omitempty"`
}

// Submission — HTTP-модель присланного решения.
type Submission struct {
	ID            string `json:"id"`
	MicrotaskID   string `json:"microtask_id"`
	StudentID     string `json:"student_id"`
	SolutionURL   string `json:"solution_url"`
	Comment       string `json:"comment,omitempty"`
	Status        int32  `json:"status"` // 1=PENDING, 2=APPROVED, 3=REJECTED
	ReviewComment string `json:"review_comment,omitempty"`
	SubmittedAt   string `json:"submitted_at,omitempty"`
	ReviewedAt    string `json:"reviewed_at,omitempty"`
}

type SubmissionList struct {
	Submissions []*Submission       `json:"submissions"`
	Pagination  *PaginationResponse `json:"pagination,omitempty"`
}

type ReviewRequest struct {
	Status        int32  `json:"status"` // 2=APPROVED, 3=REJECTED
	ReviewComment string `json:"review_comment,omitempty"`
}
