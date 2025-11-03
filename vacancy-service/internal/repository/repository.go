package repository

import (
	"context"
	"errors"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrVacancyNotFound = errors.New("vacancy not found")
	VACANCY_TABLE      = "vacancies"
)

type Vacancy interface {
	CreateVacancy(ctx context.Context, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error)
	UpdateVacancy(ctx context.Context, id string, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error)
	DeleteVacancy(ctx context.Context, id string) error
	GetVacancy(ctx context.Context, id string) (*vacancyv1.Vacancy, error)
	GetAllVacancies(ctx context.Context, companyID, positionStatus string, limit, offset int32) (*vacancyv1.VacancyList, error)
	GetAllExistPositions(ctx context.Context) ([]string, error)
}

// Repository — основная хранилка
type Repository struct {
	Vacancy Vacancy
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Vacancy: NewVacancyRepository(db),
	}
}
