package service

import (
	"context"
	"hh_for_students/vacancy-service/internal/repository"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
)

type VacancyService struct {
	repo *repository.Repository
}

func NewVacancyService(repo *repository.Repository) *VacancyService {
	return &VacancyService{
		repo: repo,
	}
}

func (s *VacancyService) CreateVacancy(ctx context.Context, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	return nil, nil
}

func (s *VacancyService) UpdateVacancy(ctx context.Context, id string, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	return nil, nil
}

func (s *VacancyService) DeleteVacancy(ctx context.Context, id string) error {
	return nil
}

func (s *VacancyService) GetVacancy(ctx context.Context, id string) (*vacancyv1.Vacancy, error) {
	return nil, nil
}

func (s *VacancyService) GetAllVacancies(ctx context.Context, limit, offset int32) (*vacancyv1.VacancyList, error) {
	return nil, nil
}
