package service

import (
	"context"

	"github.com/studjobs/hh_for_students/skills/internal/models"
	"github.com/studjobs/hh_for_students/skills/internal/repository"
)

const (
	defaultLimit = 20
	maxLimit     = 100
	maxBulkSlugs = 500
)

type SkillsService struct {
	repo *repository.Repository
}

func NewSkillsService(repo *repository.Repository) *SkillsService {
	return &SkillsService{repo: repo}
}

func (s *SkillsService) Search(ctx context.Context, query string, category int32, limit int) ([]models.Skill, error) {
	return s.repo.Skills.Search(ctx, query, category, normalizeLimit(limit))
}

func (s *SkillsService) Popular(ctx context.Context, category int32, limit int) ([]models.Skill, error) {
	return s.repo.Skills.Popular(ctx, category, normalizeLimit(limit))
}

func (s *SkillsService) Bulk(ctx context.Context, slugs []string) ([]models.Skill, error) {
	cleaned := dedupeSlugs(slugs)
	if len(cleaned) > maxBulkSlugs {
		cleaned = cleaned[:maxBulkSlugs]
	}
	return s.repo.Skills.Bulk(ctx, cleaned)
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func dedupeSlugs(slugs []string) []string {
	seen := make(map[string]struct{}, len(slugs))
	out := make([]string, 0, len(slugs))
	for _, s := range slugs {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
