package repository

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/studjobs/hh_for_students/skills/internal/models"
)

type Skills interface {
	Search(ctx context.Context, query string, category int32, limit int) ([]models.Skill, error)
	Popular(ctx context.Context, category int32, limit int) ([]models.Skill, error)
	Bulk(ctx context.Context, slugs []string) ([]models.Skill, error)
}

type Repository struct {
	Skills Skills
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Skills: NewSkillsRepository(db),
	}
}
