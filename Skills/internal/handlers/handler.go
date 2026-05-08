package handlers

import (
	"context"
	"log"

	skillsv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/skills/v1"

	"github.com/studjobs/hh_for_students/skills/internal/models"
	"github.com/studjobs/hh_for_students/skills/internal/service"
)

type Handler struct {
	skillsv1.UnimplementedSkillsServiceServer
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Search(ctx context.Context, req *skillsv1.SearchSkillsRequest) (*skillsv1.SkillList, error) {
	log.Printf("Handler: Search query=%q category=%d limit=%d",
		req.GetQuery(), req.GetCategory(), req.GetLimit())

	skills, err := h.service.Skills.Search(ctx, req.GetQuery(), int32(req.GetCategory()), int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	return &skillsv1.SkillList{Skills: toProto(skills)}, nil
}

func (h *Handler) Popular(ctx context.Context, req *skillsv1.PopularSkillsRequest) (*skillsv1.SkillList, error) {
	log.Printf("Handler: Popular category=%d limit=%d", req.GetCategory(), req.GetLimit())

	skills, err := h.service.Skills.Popular(ctx, int32(req.GetCategory()), int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	return &skillsv1.SkillList{Skills: toProto(skills)}, nil
}

func (h *Handler) Bulk(ctx context.Context, req *skillsv1.BulkSkillsRequest) (*skillsv1.SkillList, error) {
	log.Printf("Handler: Bulk len(slugs)=%d", len(req.GetSlugs()))

	skills, err := h.service.Skills.Bulk(ctx, req.GetSlugs())
	if err != nil {
		return nil, err
	}
	return &skillsv1.SkillList{Skills: toProto(skills)}, nil
}

func toProto(in []models.Skill) []*skillsv1.Skill {
	out := make([]*skillsv1.Skill, len(in))
	for i, s := range in {
		out[i] = &skillsv1.Skill{
			Id:         s.ID,
			Slug:       s.Slug,
			Name:       s.Name,
			Category:   skillsv1.SkillCategory(s.Category),
			Popularity: s.Popularity,
		}
	}
	return out
}
