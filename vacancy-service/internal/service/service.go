package service

import (
	"context"
	"errors"
	"hh_for_students/vacancy-service/internal/repository"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
)

var (
	ErrVacancyNotFound    = errors.New("vacancy not found")
	ErrInvalidVacancyData = errors.New("invalid vacancy data")
)

type Vacancy interface {
	CreateVacancy(ctx context.Context, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error)
	UpdateVacancy(ctx context.Context, id string, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error)
	DeleteVacancy(ctx context.Context, id string) error
	GetVacancy(ctx context.Context, id string) (*vacancyv1.Vacancy, error)
	GetAllVacancies(ctx context.Context, companyID, positionStatus, workFormat, schedule string,
		minSalary, maxSalary, minExperience, maxExperience int32,
		searchTitle string, page, limit int32) (*vacancyv1.VacancyList, error)
	GetHRVacancies(ctx context.Context, companyID, positionStatus, workFormat, schedule string,
		minSalary, maxSalary, minExperience, maxExperience int32,
		searchTitle string, page, limit int32) (*vacancyv1.VacancyList, error)
	GetAllExistPositions(ctx context.Context, req *vacancyv1.PositionsRequest) (*vacancyv1.PositionsResponse, error)
}

type Service struct {
	Vacancy Vacancy
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		Vacancy: NewVacancyService(repo),
	}
}
