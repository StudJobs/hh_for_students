package services

import (
	"context"
	"log"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

type microTaskService struct {
	client microtaskv1.MicroTaskServiceClient
}

func NewMicroTaskService(client microtaskv1.MicroTaskServiceClient) MicroTaskService {
	log.Printf("Creating new MicroTaskService")
	return &microTaskService{client: client}
}

func (s *microTaskService) Available() bool {
	return s.client != nil
}

func (s *microTaskService) Create(ctx context.Context, t *models.MicroTask) (*models.MicroTask, error) {
	resp, err := s.client.Create(ctx, &microtaskv1.CreateMicroTaskRequest{Task: toProto(t)})
	if err != nil {
		return nil, err
	}
	return fromProto(resp), nil
}

func (s *microTaskService) Update(ctx context.Context, id string, t *models.MicroTask) (*models.MicroTask, error) {
	resp, err := s.client.Update(ctx, &microtaskv1.UpdateMicroTaskRequest{Id: id, Task: toProto(t)})
	if err != nil {
		return nil, err
	}
	return fromProto(resp), nil
}

func (s *microTaskService) Delete(ctx context.Context, id string) error {
	_, err := s.client.Delete(ctx, &microtaskv1.DeleteMicroTaskRequest{Id: id})
	return err
}

func (s *microTaskService) Get(ctx context.Context, id string) (*models.MicroTask, error) {
	resp, err := s.client.Get(ctx, &microtaskv1.GetMicroTaskRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return fromProto(resp), nil
}

func (s *microTaskService) List(ctx context.Context, status int32, skillSlugs []string, page, limit int32) (*models.MicroTaskList, error) {
	resp, err := s.client.List(ctx, &microtaskv1.ListMicroTasksRequest{
		Pagination: &commonv1.Pagination{Page: page, Limit: limit},
		Status:     microtaskv1.MicroTaskStatus(status),
		SkillSlugs: skillSlugs,
	})
	if err != nil {
		return nil, err
	}
	return listFromProto(resp), nil
}

func (s *microTaskService) ListByCompany(ctx context.Context, companyID string, page, limit int32) (*models.MicroTaskList, error) {
	resp, err := s.client.ListByCompany(ctx, &microtaskv1.ListByCompanyRequest{
		CompanyId:  companyID,
		Pagination: &commonv1.Pagination{Page: page, Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	return listFromProto(resp), nil
}

func (s *microTaskService) Apply(ctx context.Context, taskID, studentID string) (*models.MicroTask, error) {
	resp, err := s.client.Apply(ctx, &microtaskv1.ApplyRequest{MicrotaskId: taskID, StudentId: studentID})
	if err != nil {
		return nil, err
	}
	return fromProto(resp), nil
}

func (s *microTaskService) Submit(ctx context.Context, taskID, studentID, solutionURL, comment string) (*models.Submission, error) {
	resp, err := s.client.Submit(ctx, &microtaskv1.SubmitRequest{
		MicrotaskId: taskID, StudentId: studentID, SolutionUrl: solutionURL, Comment: comment,
	})
	if err != nil {
		return nil, err
	}
	return submissionFromProto(resp), nil
}

func (s *microTaskService) ListSubmissions(ctx context.Context, taskID, studentID string, page, limit int32) (*models.SubmissionList, error) {
	resp, err := s.client.ListSubmissions(ctx, &microtaskv1.ListSubmissionsRequest{
		MicrotaskId: taskID,
		StudentId:   studentID,
		Pagination:  &commonv1.Pagination{Page: page, Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	subs := make([]*models.Submission, 0, len(resp.GetSubmissions()))
	for _, ps := range resp.GetSubmissions() {
		subs = append(subs, submissionFromProto(ps))
	}
	out := &models.SubmissionList{Submissions: subs}
	if p := resp.GetPagination(); p != nil {
		out.Pagination = &models.PaginationResponse{Total: p.GetTotal(), Pages: p.GetPages(), CurrentPage: p.GetCurrentPage()}
	}
	return out, nil
}

func (s *microTaskService) Review(ctx context.Context, submissionID string, status int32, reviewComment string) (*models.Submission, error) {
	resp, err := s.client.Review(ctx, &microtaskv1.ReviewRequest{
		SubmissionId:  submissionID,
		Status:        microtaskv1.SubmissionStatus(status),
		ReviewComment: reviewComment,
	})
	if err != nil {
		return nil, err
	}
	return submissionFromProto(resp), nil
}

func toProto(m *models.MicroTask) *microtaskv1.MicroTask {
	if m == nil {
		return nil
	}
	return &microtaskv1.MicroTask{
		Id:          m.ID,
		CompanyId:   m.CompanyID,
		Title:       m.Title,
		Description: m.Description,
		Reward:      m.Reward,
		Deadline:    m.Deadline,
		SkillSlugs:  m.SkillSlugs,
		Status:      microtaskv1.MicroTaskStatus(m.Status),
		AssignedTo:  m.AssignedTo,
	}
}

func fromProto(p *microtaskv1.MicroTask) *models.MicroTask {
	if p == nil {
		return nil
	}
	return &models.MicroTask{
		ID:          p.GetId(),
		CompanyID:   p.GetCompanyId(),
		Title:       p.GetTitle(),
		Description: p.GetDescription(),
		Reward:      p.GetReward(),
		Deadline:    p.GetDeadline(),
		SkillSlugs:  p.GetSkillSlugs(),
		Status:      int32(p.GetStatus()),
		AssignedTo:  p.GetAssignedTo(),
		CreatedAt:   p.GetCreatedAt(),
		UpdatedAt:   p.GetUpdatedAt(),
	}
}

func listFromProto(p *microtaskv1.MicroTaskList) *models.MicroTaskList {
	if p == nil {
		return &models.MicroTaskList{Tasks: nil}
	}
	tasks := make([]*models.MicroTask, 0, len(p.GetTasks()))
	for _, pt := range p.GetTasks() {
		tasks = append(tasks, fromProto(pt))
	}
	out := &models.MicroTaskList{Tasks: tasks}
	if pg := p.GetPagination(); pg != nil {
		out.Pagination = &models.PaginationResponse{Total: pg.GetTotal(), Pages: pg.GetPages(), CurrentPage: pg.GetCurrentPage()}
	}
	return out
}

func submissionFromProto(p *microtaskv1.Submission) *models.Submission {
	if p == nil {
		return nil
	}
	return &models.Submission{
		ID:            p.GetId(),
		MicrotaskID:   p.GetMicrotaskId(),
		StudentID:     p.GetStudentId(),
		SolutionURL:   p.GetSolutionUrl(),
		Comment:       p.GetComment(),
		Status:        int32(p.GetStatus()),
		ReviewComment: p.GetReviewComment(),
		SubmittedAt:   p.GetSubmittedAt(),
		ReviewedAt:    p.GetReviewedAt(),
	}
}
