package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

type VacancyRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewVacancyRepository(db *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *VacancyRepository) CreateVacancy(ctx context.Context, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	return nil, nil
}
func (r *VacancyRepository) UpdateVacancy(ctx context.Context, id string, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	return nil, nil
}
func (r *VacancyRepository) DeleteVacancy(ctx context.Context, id string) error {
	return nil
}
func (r *VacancyRepository) GetVacancy(ctx context.Context, id string) (*vacancyv1.Vacancy, error) {
	return nil, nil
}
func (r *VacancyRepository) GetAllVacancies(ctx context.Context, limit, offset int32) (*vacancyv1.VacancyList, error) {
	return nil, nil
}
