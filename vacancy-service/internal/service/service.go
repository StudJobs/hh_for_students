package service

import (
	"context"
	"hh_for_students/vacancy-service/internal/repository"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
)

type Vacancy interface {
	CreateVacancy(ctx context.Context, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error)
	UpdateVacancy(ctx context.Context, id string, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error)
	DeleteVacancy(ctx context.Context, id string) error
	GetVacancy(ctx context.Context, id string) (*vacancyv1.Vacancy, error)
	GetAllVacancies(ctx context.Context, limit, offset int32) (*vacancyv1.VacancyList, error)
}

type Service struct {
	Vacancy Vacancy
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		Vacancy: NewVacancyService(repo),
	}
}
