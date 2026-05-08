package indexer

import (
	"context"
	"encoding/json"
	"fmt"

	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"

	"github.com/studjobs/hh_for_students/search/internal/esclient"
)

type Indexer struct {
	es *esclient.Client
}

func New(es *esclient.Client) *Indexer {
	return &Indexer{es: es}
}

func (i *Indexer) IndexProfile(ctx context.Context, p *usersv1.Profile) error {
	if p == nil || p.GetId() == "" {
		return fmt.Errorf("indexer: empty profile")
	}
	doc := profileDoc{
		ID:                   p.GetId(),
		FirstName:            p.GetFirstName(),
		LastName:             p.GetLastName(),
		ProfessionCategory:   p.GetProfessionCategory(),
		EducationInstitution: p.GetEducationInstitution(),
		Description:          p.GetDescription(),
		Role:                 p.GetRole(),
		SkillSlugs:           p.GetSkillSlugs(),
		Age:                  p.GetAge(),
		Email:                p.GetEmail(),
		Tg:                   p.GetTg(),
		AvatarID:             p.GetAvatarId(),
		ResumeID:             p.GetResumeId(),
	}
	if doc.SkillSlugs == nil {
		doc.SkillSlugs = []string{}
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("indexer: marshal profile: %w", err)
	}
	return i.es.Index(ctx, esclient.IndexProfiles, p.GetId(), body)
}

func (i *Indexer) IndexVacancy(ctx context.Context, v *vacancyv1.Vacancy) error {
	if v == nil || v.GetId() == "" {
		return fmt.Errorf("indexer: empty vacancy")
	}
	doc := vacancyDoc{
		ID:             v.GetId(),
		Title:          v.GetTitle(),
		PositionStatus: v.GetPositionStatus(),
		Schedule:       v.GetSchedule(),
		WorkFormat:     v.GetWorkFormat(),
		CompanyID:      v.GetCompanyId(),
		SkillSlugs:     v.GetSkillSlugs(),
		Experience:     v.GetExperience(),
		Salary:         v.GetSalary(),
		CreateAt:       v.GetCreateAt(),
		AttachmentID:   v.GetAttachmentId(),
	}
	if doc.SkillSlugs == nil {
		doc.SkillSlugs = []string{}
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("indexer: marshal vacancy: %w", err)
	}
	return i.es.Index(ctx, esclient.IndexVacancies, v.GetId(), body)
}

func (i *Indexer) DeleteProfile(ctx context.Context, id string) error {
	return i.es.Delete(ctx, esclient.IndexProfiles, id)
}

func (i *Indexer) DeleteVacancy(ctx context.Context, id string) error {
	return i.es.Delete(ctx, esclient.IndexVacancies, id)
}

func (i *Indexer) IndexMicroTask(ctx context.Context, t *microtaskv1.MicroTask) error {
	if t == nil || t.GetId() == "" {
		return fmt.Errorf("indexer: empty microtask")
	}
	doc := microTaskDoc{
		ID:          t.GetId(),
		CompanyID:   t.GetCompanyId(),
		Title:       t.GetTitle(),
		Description: t.GetDescription(),
		Reward:      t.GetReward(),
		Deadline:    t.GetDeadline(),
		SkillSlugs:  t.GetSkillSlugs(),
		Status:      int32(t.GetStatus()),
		AssignedTo:  t.GetAssignedTo(),
		CreatedAt:   t.GetCreatedAt(),
		UpdatedAt:   t.GetUpdatedAt(),
	}
	if doc.SkillSlugs == nil {
		doc.SkillSlugs = []string{}
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("indexer: marshal microtask: %w", err)
	}
	return i.es.Index(ctx, esclient.IndexMicroTasks, t.GetId(), body)
}

func (i *Indexer) DeleteMicroTask(ctx context.Context, id string) error {
	return i.es.Delete(ctx, esclient.IndexMicroTasks, id)
}

type profileDoc struct {
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

type vacancyDoc struct {
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

type microTaskDoc struct {
	ID          string   `json:"id"`
	CompanyID   string   `json:"company_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Reward      int32    `json:"reward"`
	Deadline    string   `json:"deadline"`
	SkillSlugs  []string `json:"skill_slugs"`
	Status      int32    `json:"status"`
	AssignedTo  string   `json:"assigned_to"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}
