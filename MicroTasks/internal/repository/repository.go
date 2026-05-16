package repository

import (
	"context"
	"errors"

	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrTaskNotFound       = errors.New("microtask not found")
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrTaskNotOpen        = errors.New("microtask is not open for apply")
	ErrAlreadyAssigned    = errors.New("microtask already has an assignee")
)

type MicroTasks interface {
	Create(ctx context.Context, t *microtaskv1.MicroTask) (*microtaskv1.MicroTask, error)
	Update(ctx context.Context, id string, t *microtaskv1.MicroTask) (*microtaskv1.MicroTask, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*microtaskv1.MicroTask, error)
	List(ctx context.Context, status microtaskv1.MicroTaskStatus, skillSlugs []string, page, limit int32) (*microtaskv1.MicroTaskList, error)
	ListByCompany(ctx context.Context, companyID string, page, limit int32) (*microtaskv1.MicroTaskList, error)
	ListByStudent(ctx context.Context, studentID string, status microtaskv1.MicroTaskStatus, page, limit int32) (*microtaskv1.MicroTaskList, error)

	Apply(ctx context.Context, taskID, studentID string) (*microtaskv1.MicroTask, error)
	SetStatus(ctx context.Context, id string, status microtaskv1.MicroTaskStatus) (*microtaskv1.MicroTask, error)
}

type Submissions interface {
	Create(ctx context.Context, s *microtaskv1.Submission) (*microtaskv1.Submission, error)
	Get(ctx context.Context, id string) (*microtaskv1.Submission, error)
	ListByTask(ctx context.Context, taskID string, page, limit int32) (*microtaskv1.SubmissionList, error)
	ListByStudent(ctx context.Context, studentID string, page, limit int32) (*microtaskv1.SubmissionList, error)
	Review(ctx context.Context, id string, status microtaskv1.SubmissionStatus, reviewComment string) (*microtaskv1.Submission, error)
}

type Repository struct {
	Tasks       MicroTasks
	Submissions Submissions
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Tasks:       NewMicroTaskRepository(db),
		Submissions: NewSubmissionRepository(db),
	}
}
