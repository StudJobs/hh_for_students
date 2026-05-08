package searcher

import (
	"context"
	"encoding/json"
	"fmt"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"

	"github.com/studjobs/hh_for_students/search/internal/esclient"
)

const (
	defaultLimit = 10
	maxLimit     = 100
)

type Searcher struct {
	es *esclient.Client
}

func New(es *esclient.Client) *Searcher {
	return &Searcher{es: es}
}

func (s *Searcher) SearchProfiles(ctx context.Context, req *searchv1.SearchProfilesRequest) (*usersv1.ProfileList, error) {
	page, limit := normalizePagination(req.GetPagination())

	must := []map[string]any{}
	if q := req.GetQuery(); q != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"first_name^2", "last_name^2", "profession_category", "education_institution", "description"},
			},
		})
	}
	for _, slug := range req.GetSkillSlugs() {
		must = append(must, map[string]any{
			"term": map[string]any{"skill_slugs": slug},
		})
	}
	if cat := req.GetProfessionCategory(); cat != "" {
		must = append(must, map[string]any{
			"term": map[string]any{"profession_category.keyword": cat},
		})
	}

	query := buildQuery(must, page, limit)
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("searcher: marshal profiles query: %w", err)
	}

	raw, err := s.es.Search(ctx, esclient.IndexProfiles, body)
	if err != nil {
		return nil, err
	}

	var resp esResponse[profileSource]
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("searcher: unmarshal profiles: %w", err)
	}

	out := make([]*usersv1.Profile, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		out = append(out, &usersv1.Profile{
			Id:                   h.Source.ID,
			FirstName:            h.Source.FirstName,
			LastName:             h.Source.LastName,
			ProfessionCategory:   h.Source.ProfessionCategory,
			EducationInstitution: h.Source.EducationInstitution,
			Description:          h.Source.Description,
			Role:                 h.Source.Role,
			SkillSlugs:           h.Source.SkillSlugs,
			Age:                  h.Source.Age,
			Email:                h.Source.Email,
			Tg:                   h.Source.Tg,
			AvatarId:             h.Source.AvatarID,
			ResumeId:             h.Source.ResumeID,
		})
	}

	return &usersv1.ProfileList{
		Profiles:   out,
		Pagination: paginationResponse(resp.Hits.Total.Value, page, limit),
	}, nil
}

func (s *Searcher) SearchVacancies(ctx context.Context, req *searchv1.SearchVacanciesRequest) (*vacancyv1.VacancyList, error) {
	page, limit := normalizePagination(req.GetPagination())

	must := []map[string]any{}
	filter := []map[string]any{}

	if q := req.GetQuery(); q != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"title^2", "position_status"},
			},
		})
	}
	for _, slug := range req.GetSkillSlugs() {
		filter = append(filter, map[string]any{
			"term": map[string]any{"skill_slugs": slug},
		})
	}
	if min := req.GetSalaryMin(); min > 0 {
		filter = append(filter, map[string]any{
			"range": map[string]any{"salary": map[string]any{"gte": min}},
		})
	}
	if exp := req.GetExperienceMax(); exp > 0 {
		filter = append(filter, map[string]any{
			"range": map[string]any{"experience": map[string]any{"lte": exp}},
		})
	}
	if cid := req.GetCompanyId(); cid != "" {
		filter = append(filter, map[string]any{
			"term": map[string]any{"company_id": cid},
		})
	}

	query := map[string]any{
		"from": (page - 1) * limit,
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filter,
			},
		},
	}
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("searcher: marshal vacancies query: %w", err)
	}

	raw, err := s.es.Search(ctx, esclient.IndexVacancies, body)
	if err != nil {
		return nil, err
	}

	var resp esResponse[vacancySource]
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("searcher: unmarshal vacancies: %w", err)
	}

	out := make([]*vacancyv1.Vacancy, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		out = append(out, &vacancyv1.Vacancy{
			Id:             h.Source.ID,
			Title:          h.Source.Title,
			PositionStatus: h.Source.PositionStatus,
			Schedule:       h.Source.Schedule,
			WorkFormat:     h.Source.WorkFormat,
			CompanyId:      h.Source.CompanyID,
			SkillSlugs:     h.Source.SkillSlugs,
			Experience:     h.Source.Experience,
			Salary:         h.Source.Salary,
			CreateAt:       h.Source.CreateAt,
			AttachmentId:   h.Source.AttachmentID,
		})
	}

	return &vacancyv1.VacancyList{
		Vacancies:  out,
		Pagination: paginationResponse(resp.Hits.Total.Value, page, limit),
	}, nil
}

func buildQuery(must []map[string]any, page, limit int32) map[string]any {
	if len(must) == 0 {
		return map[string]any{
			"from":  (page - 1) * limit,
			"size":  limit,
			"query": map[string]any{"match_all": map[string]any{}},
		}
	}
	return map[string]any{
		"from": (page - 1) * limit,
		"size": limit,
		"query": map[string]any{
			"bool": map[string]any{"must": must},
		},
	}
}

func normalizePagination(p *commonv1.Pagination) (int32, int32) {
	page := int32(1)
	limit := int32(defaultLimit)
	if p != nil {
		if p.GetPage() > 0 {
			page = p.GetPage()
		}
		if p.GetLimit() > 0 {
			limit = p.GetLimit()
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

func paginationResponse(total, page, limit int32) *commonv1.PaginationResponse {
	pages := int32(0)
	if limit > 0 {
		pages = (total + limit - 1) / limit
	}
	return &commonv1.PaginationResponse{
		Total:       total,
		Pages:       pages,
		CurrentPage: page,
	}
}

type esResponse[T any] struct {
	Hits struct {
		Total struct {
			Value int32 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source T `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type profileSource struct {
	ID                   string   `json:"id"`
	FirstName            string   `json:"first_name"`
	LastName             string   `json:"last_name"`
	ProfessionCategory   string   `json:"profession_category"`
	EducationInstitution string   `json:"education_institution"`
	Description          string   `json:"description"`
	Role                 string   `json:"role"`
	SkillSlugs           []string `json:"skill_slugs"`
	Age                  int32    `json:"age"`
	Email                string   `json:"email"`
	Tg                   string   `json:"tg"`
	AvatarID             string   `json:"avatar_id"`
	ResumeID             string   `json:"resume_id"`
}

type vacancySource struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	PositionStatus string   `json:"position_status"`
	Schedule       string   `json:"schedule"`
	WorkFormat     string   `json:"work_format"`
	CompanyID      string   `json:"company_id"`
	SkillSlugs     []string `json:"skill_slugs"`
	Experience     int32    `json:"experience"`
	Salary         int32    `json:"salary"`
	CreateAt       string   `json:"create_at"`
	AttachmentID   string   `json:"attachment_id"`
}
