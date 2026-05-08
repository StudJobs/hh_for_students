package services

import (
	"context"
	"fmt"

	skillsv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/skills/v1"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

type skillsServiceImpl struct {
	client skillsv1.SkillsServiceClient
}

func NewSkillsService(client skillsv1.SkillsServiceClient) SkillsService {
	return &skillsServiceImpl{client: client}
}

func (s *skillsServiceImpl) Search(ctx context.Context, query string, category int32, limit int32) ([]*models.Skill, error) {
	if s.client == nil {
		return nil, fmt.Errorf("skills service is not available")
	}
	resp, err := s.client.Search(ctx, &skillsv1.SearchSkillsRequest{
		Query:    query,
		Limit:    limit,
		Category: skillsv1.SkillCategory(category),
	})
	if err != nil {
		return nil, fmt.Errorf("skills.Search: %w", err)
	}
	return mapSkills(resp.GetSkills()), nil
}

func (s *skillsServiceImpl) Popular(ctx context.Context, category int32, limit int32) ([]*models.Skill, error) {
	if s.client == nil {
		return nil, fmt.Errorf("skills service is not available")
	}
	resp, err := s.client.Popular(ctx, &skillsv1.PopularSkillsRequest{
		Limit:    limit,
		Category: skillsv1.SkillCategory(category),
	})
	if err != nil {
		return nil, fmt.Errorf("skills.Popular: %w", err)
	}
	return mapSkills(resp.GetSkills()), nil
}

func (s *skillsServiceImpl) Bulk(ctx context.Context, slugs []string) ([]*models.Skill, error) {
	if s.client == nil {
		return nil, fmt.Errorf("skills service is not available")
	}
	resp, err := s.client.Bulk(ctx, &skillsv1.BulkSkillsRequest{Slugs: slugs})
	if err != nil {
		return nil, fmt.Errorf("skills.Bulk: %w", err)
	}
	return mapSkills(resp.GetSkills()), nil
}

func mapSkills(in []*skillsv1.Skill) []*models.Skill {
	out := make([]*models.Skill, len(in))
	for i, s := range in {
		out[i] = &models.Skill{
			ID:         s.GetId(),
			Slug:       s.GetSlug(),
			Name:       s.GetName(),
			Category:   int32(s.GetCategory()),
			Popularity: s.GetPopularity(),
		}
	}
	return out
}
