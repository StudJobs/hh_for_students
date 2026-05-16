package service

import (
	"context"
	"errors"

	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"

	"github.com/studjobs/hh_for_students/microtasks/internal/repository"
)

var (
	ErrInvalidArg     = errors.New("invalid argument")
	ErrTaskNotFound   = repository.ErrTaskNotFound
	ErrSubmissionNF   = repository.ErrSubmissionNotFound
	ErrTaskNotOpen    = repository.ErrTaskNotOpen
)

type Service struct {
	Tasks       *MicroTaskService
	Submissions *SubmissionService
}

func NewService(repo *repository.Repository) *Service {
	tasks := NewMicroTaskService(repo)
	subs := NewSubmissionService(repo)
	return &Service{Tasks: tasks, Submissions: subs}
}

type MicroTaskService struct {
	repo *repository.Repository
}

func NewMicroTaskService(repo *repository.Repository) *MicroTaskService {
	return &MicroTaskService{repo: repo}
}

func (s *MicroTaskService) Create(ctx context.Context, t *microtaskv1.MicroTask) (*microtaskv1.MicroTask, error) {
	if t == nil || t.GetCompanyId() == "" || t.GetTitle() == "" {
		return nil, ErrInvalidArg
	}
	t.Status = microtaskv1.MicroTaskStatus_MICROTASK_STATUS_OPEN
	t.AssignedTo = ""
	return s.repo.Tasks.Create(ctx, t)
}

func (s *MicroTaskService) Update(ctx context.Context, id string, t *microtaskv1.MicroTask) (*microtaskv1.MicroTask, error) {
	if id == "" || t == nil {
		return nil, ErrInvalidArg
	}
	return s.repo.Tasks.Update(ctx, id, t)
}

func (s *MicroTaskService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidArg
	}
	return s.repo.Tasks.Delete(ctx, id)
}

func (s *MicroTaskService) Get(ctx context.Context, id string) (*microtaskv1.MicroTask, error) {
	if id == "" {
		return nil, ErrInvalidArg
	}
	return s.repo.Tasks.Get(ctx, id)
}

func (s *MicroTaskService) List(ctx context.Context, status microtaskv1.MicroTaskStatus, skillSlugs []string, page, limit int32) (*microtaskv1.MicroTaskList, error) {
	return s.repo.Tasks.List(ctx, status, skillSlugs, page, limit)
}

func (s *MicroTaskService) ListByCompany(ctx context.Context, companyID string, page, limit int32) (*microtaskv1.MicroTaskList, error) {
	if companyID == "" {
		return nil, ErrInvalidArg
	}
	return s.repo.Tasks.ListByCompany(ctx, companyID, page, limit)
}

func (s *MicroTaskService) ListByStudent(ctx context.Context, studentID string, status microtaskv1.MicroTaskStatus, page, limit int32) (*microtaskv1.MicroTaskList, error) {
	if studentID == "" {
		return nil, ErrInvalidArg
	}
	return s.repo.Tasks.ListByStudent(ctx, studentID, status, page, limit)
}

func (s *MicroTaskService) CreateSkillQuest(ctx context.Context, expertID, studentID, slug, title, description, deadline string) (*microtaskv1.MicroTask, error) {
	if expertID == "" || studentID == "" || slug == "" || title == "" {
		return nil, ErrInvalidArg
	}
	return s.repo.Tasks.CreateSkillQuest(ctx, expertID, studentID, slug, title, description, deadline)
}

func (s *MicroTaskService) Apply(ctx context.Context, taskID, studentID string) (*microtaskv1.MicroTask, error) {
	if taskID == "" || studentID == "" {
		return nil, ErrInvalidArg
	}
	return s.repo.Tasks.Apply(ctx, taskID, studentID)
}

type SubmissionService struct {
	repo *repository.Repository
}

func NewSubmissionService(repo *repository.Repository) *SubmissionService {
	return &SubmissionService{repo: repo}
}

func (s *SubmissionService) Submit(ctx context.Context, taskID, studentID, solutionURL, comment, fileName string) (*microtaskv1.Submission, error) {
	if taskID == "" || studentID == "" {
		return nil, ErrInvalidArg
	}
	// Хотя бы один из (solutionURL, fileName) должен быть заполнен.
	if solutionURL == "" && fileName == "" {
		return nil, ErrInvalidArg
	}
	// Проверяем, что задача существует и студент имеет к ней доступ (он её взял).
	task, err := s.repo.Tasks.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.GetAssignedTo() != "" && task.GetAssignedTo() != studentID {
		return nil, ErrInvalidArg
	}
	return s.repo.Submissions.Create(ctx, &microtaskv1.Submission{
		MicrotaskId:      taskID,
		StudentId:        studentID,
		SolutionUrl:      solutionURL,
		Comment:          comment,
		SolutionFileName: fileName,
	})
}

func (s *SubmissionService) ListByTask(ctx context.Context, taskID string, page, limit int32) (*microtaskv1.SubmissionList, error) {
	if taskID == "" {
		return nil, ErrInvalidArg
	}
	return s.repo.Submissions.ListByTask(ctx, taskID, page, limit)
}

func (s *SubmissionService) ListByStudent(ctx context.Context, studentID string, page, limit int32) (*microtaskv1.SubmissionList, error) {
	if studentID == "" {
		return nil, ErrInvalidArg
	}
	return s.repo.Submissions.ListByStudent(ctx, studentID, page, limit)
}

// Review выполняет approve/reject. При APPROVE задача переходит в COMPLETED.
// Возвращает обновлённую submission и (опционально) обновлённую задачу.
func (s *SubmissionService) Review(ctx context.Context, submissionID string, status microtaskv1.SubmissionStatus, reviewComment string) (*microtaskv1.Submission, *microtaskv1.MicroTask, error) {
	if submissionID == "" || status == microtaskv1.SubmissionStatus_SUBMISSION_STATUS_UNSPECIFIED {
		return nil, nil, ErrInvalidArg
	}
	sub, err := s.repo.Submissions.Review(ctx, submissionID, status, reviewComment)
	if err != nil {
		return nil, nil, err
	}
	if status == microtaskv1.SubmissionStatus_SUBMISSION_STATUS_APPROVED {
		task, err := s.repo.Tasks.SetStatus(ctx, sub.GetMicrotaskId(), microtaskv1.MicroTaskStatus_MICROTASK_STATUS_COMPLETED)
		if err != nil {
			return sub, nil, err
		}
		return sub, task, nil
	}
	return sub, nil, nil
}
