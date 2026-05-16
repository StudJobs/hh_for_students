package services

import (
	"context"
	"log"

	applicationv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/application/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

type applicationService struct {
	client applicationv1.ApplicationServiceClient
}

func NewApplicationService(client applicationv1.ApplicationServiceClient) ApplicationService {
	return &applicationService{client: client}
}

func (s *applicationService) Available() bool {
	return s.client != nil
}

func (s *applicationService) Apply(ctx context.Context, vacancyID, studentID, coverLetter string) (*models.Application, error) {
	resp, err := s.client.Apply(ctx, &applicationv1.ApplyRequest{
		VacancyId:   vacancyID,
		StudentId:   studentID,
		CoverLetter: coverLetter,
	})
	if err != nil {
		log.Printf("ApplicationService: Apply failed: %v", err)
		return nil, err
	}
	return protoToModel(resp), nil
}

func (s *applicationService) Withdraw(ctx context.Context, id, studentID string) error {
	_, err := s.client.Withdraw(ctx, &applicationv1.WithdrawRequest{
		Id:        id,
		StudentId: studentID,
	})
	return err
}

func (s *applicationService) ListMine(ctx context.Context, studentID string, status int32, page, limit int32) (*models.ApplicationList, error) {
	resp, err := s.client.ListMine(ctx, &applicationv1.ListMineRequest{
		StudentId:  studentID,
		Pagination: &commonv1.Pagination{Page: page, Limit: limit},
		Status:     applicationv1.ApplicationStatus(status),
	})
	if err != nil {
		log.Printf("ApplicationService: ListMine failed: %v", err)
		return nil, err
	}
	return protoListToModel(resp), nil
}

func (s *applicationService) ListForVacancy(ctx context.Context, vacancyID string, status int32, page, limit int32) (*models.ApplicationList, error) {
	resp, err := s.client.ListForVacancy(ctx, &applicationv1.ListForVacancyRequest{
		VacancyId:  vacancyID,
		Pagination: &commonv1.Pagination{Page: page, Limit: limit},
		Status:     applicationv1.ApplicationStatus(status),
	})
	if err != nil {
		log.Printf("ApplicationService: ListForVacancy failed: %v", err)
		return nil, err
	}
	return protoListToModel(resp), nil
}

func (s *applicationService) UpdateStatus(ctx context.Context, id string, status int32, hrComment string) (*models.Application, error) {
	resp, err := s.client.UpdateStatus(ctx, &applicationv1.UpdateStatusRequest{
		Id:        id,
		Status:    applicationv1.ApplicationStatus(status),
		HrComment: hrComment,
	})
	if err != nil {
		log.Printf("ApplicationService: UpdateStatus failed: %v", err)
		return nil, err
	}
	return protoToModel(resp), nil
}

func protoToModel(a *applicationv1.Application) *models.Application {
	if a == nil {
		return nil
	}
	return &models.Application{
		ID:           a.Id,
		VacancyID:    a.VacancyId,
		StudentID:    a.StudentId,
		CoverLetter:  a.CoverLetter,
		Status:       int32(a.Status),
		HRComment:    a.HrComment,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		HRAssigneeID: a.HrAssigneeId,
	}
}

func (s *applicationService) Get(ctx context.Context, id string) (*models.Application, error) {
	resp, err := s.client.Get(ctx, &applicationv1.GetRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protoToModel(resp), nil
}

func (s *applicationService) AssignHR(ctx context.Context, id, hrUserID string) (*models.Application, error) {
	resp, err := s.client.AssignHR(ctx, &applicationv1.AssignHRRequest{Id: id, HrUserId: hrUserID})
	if err != nil {
		return nil, err
	}
	return protoToModel(resp), nil
}

func protoListToModel(l *applicationv1.ApplicationList) *models.ApplicationList {
	if l == nil {
		return nil
	}
	out := &models.ApplicationList{
		Applications: make([]*models.Application, 0, len(l.Applications)),
	}
	for _, a := range l.Applications {
		out.Applications = append(out.Applications, protoToModel(a))
	}
	if l.Pagination != nil {
		out.Pagination = &models.PaginationResponse{
			Total:       l.Pagination.Total,
			Pages:       l.Pagination.Pages,
			CurrentPage: l.Pagination.CurrentPage,
		}
	}
	return out
}
