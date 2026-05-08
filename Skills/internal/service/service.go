package service

import (
	"context"

	"github.com/studjobs/hh_for_students/skills/internal/models"
	"github.com/studjobs/hh_for_students/skills/internal/repository"
)

type ISkillsService interface {
	Search(ctx context.Context, query string, category int32, limit int) ([]models.Skill, error)
	Popular(ctx context.Context, category int32, limit int) ([]models.Skill, error)
	Bulk(ctx context.Context, slugs []string) ([]models.Skill, error)
}

type Service struct {
	Skills ISkillsService
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		Skills: NewSkillsService(repo),
	}
}
