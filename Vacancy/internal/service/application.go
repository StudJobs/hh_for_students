package service

import (
	"context"
	"errors"
	"log"

	applicationv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/application/v1"
	"hh_for_students/vacancy-service/internal/repository"
)

var (
	ErrApplicationNotFound     = errors.New("application not found")
	ErrInvalidApplicationData  = errors.New("invalid application data")
	ErrApplicationForbidden    = errors.New("not allowed for this application")
)

type Application interface {
	Apply(ctx context.Context, vacancyID, studentID, coverLetter string) (*applicationv1.Application, error)
	Withdraw(ctx context.Context, id, studentID string) error
	ListByStudent(ctx context.Context, studentID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error)
	ListByVacancy(ctx context.Context, vacancyID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error)
	UpdateStatus(ctx context.Context, id string, status applicationv1.ApplicationStatus, hrComment string) (*applicationv1.Application, error)
}

type ApplicationService struct {
	repo repository.Application
}

func NewApplicationService(repo repository.Application) *ApplicationService {
	return &ApplicationService{repo: repo}
}

func (s *ApplicationService) Apply(ctx context.Context, vacancyID, studentID, coverLetter string) (*applicationv1.Application, error) {
	if vacancyID == "" || studentID == "" {
		return nil, ErrInvalidApplicationData
	}
	app, err := s.repo.Apply(ctx, vacancyID, studentID, coverLetter)
	if err != nil {
		log.Printf("AppService: Apply failed: %v", err)
		return nil, err
	}
	return app, nil
}

func (s *ApplicationService) Withdraw(ctx context.Context, id, studentID string) error {
	if id == "" || studentID == "" {
		return ErrInvalidApplicationData
	}
	if err := s.repo.Withdraw(ctx, id, studentID); err != nil {
		if errors.Is(err, repository.ErrApplicationNotFound) {
			return ErrApplicationNotFound
		}
		return err
	}
	return nil
}

func (s *ApplicationService) ListByStudent(ctx context.Context, studentID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error) {
	if studentID == "" {
		return nil, ErrInvalidApplicationData
	}
	return s.repo.ListByStudent(ctx, studentID, status, page, limit)
}

func (s *ApplicationService) ListByVacancy(ctx context.Context, vacancyID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error) {
	if vacancyID == "" {
		return nil, ErrInvalidApplicationData
	}
	return s.repo.ListByVacancy(ctx, vacancyID, status, page, limit)
}

func (s *ApplicationService) UpdateStatus(ctx context.Context, id string, status applicationv1.ApplicationStatus, hrComment string) (*applicationv1.Application, error) {
	if id == "" {
		return nil, ErrInvalidApplicationData
	}
	app, err := s.repo.UpdateStatus(ctx, id, status, hrComment)
	if err != nil {
		if errors.Is(err, repository.ErrApplicationNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	return app, nil
}
