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

	// Поля квеста от эксперта.
	IsSkillQuest    bool   `json:"is_skill_quest,omitempty"`
	TargetStudentID string `json:"target_student_id,omitempty"`
	TargetSkillSlug string `json:"target_skill_slug,omitempty"`
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
	SolutionURL      string `json:"solution_url"`
	Comment          string `json:"comment,omitempty"`
	SolutionFileName string `json:"solution_file_name,omitempty"`
}

type SolutionUploadInitRequest struct {
	FileName string `json:"file_name"`
}

type SolutionUploadInitResponse struct {
	FileID    string `json:"file_id"`
	UploadURL string `json:"upload_url"`
}

type SolutionUploadConfirmRequest struct {
	FileID string `json:"file_id"`
}

// Submission — HTTP-модель присланного решения.
type Submission struct {
	ID               string `json:"id"`
	MicrotaskID      string `json:"microtask_id"`
	StudentID        string `json:"student_id"`
	SolutionURL      string `json:"solution_url"`
	Comment          string `json:"comment,omitempty"`
	Status           int32  `json:"status"` // 1=PENDING, 2=APPROVED, 3=REJECTED
	ReviewComment    string `json:"review_comment,omitempty"`
	SubmittedAt      string `json:"submitted_at,omitempty"`
	ReviewedAt       string `json:"reviewed_at,omitempty"`
	SolutionFileName string `json:"solution_file_name,omitempty"`
	SolutionFileURL  string `json:"solution_file_url,omitempty"`
}

type SubmissionList struct {
	Submissions []*Submission       `json:"submissions"`
	Pagination  *PaginationResponse `json:"pagination,omitempty"`
}

type ReviewRequest struct {
	Status        int32  `json:"status"` // 2=APPROVED, 3=REJECTED
	ReviewComment string `json:"review_comment,omitempty"`
}

type CreateSkillQuestRequest struct {
	TargetStudentID string `json:"target_student_id"`
	TargetSkillSlug string `json:"target_skill_slug"`
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	Deadline        string `json:"deadline,omitempty"`
}
