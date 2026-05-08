package services

import (
	"context"
	"log"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"

	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
)

type searchService struct {
	client searchv1.SearchServiceClient
}

func NewSearchService(client searchv1.SearchServiceClient) SearchService {
	log.Printf("Creating new SearchService")
	return &searchService{client: client}
}

func (s *searchService) Available() bool {
	return s.client != nil
}

func (s *searchService) SearchProfiles(ctx context.Context, query string, skillSlugs []string, professionCategory string, page, limit int32) (*usersv1.ProfileList, error) {
	return s.client.SearchProfiles(ctx, &searchv1.SearchProfilesRequest{
		Query:              query,
		SkillSlugs:         skillSlugs,
		ProfessionCategory: professionCategory,
		Pagination:         &commonv1.Pagination{Page: page, Limit: limit},
	})
}

func (s *searchService) SearchVacancies(ctx context.Context, query string, skillSlugs []string, salaryMin, experienceMax int32, companyID string, page, limit int32) (*vacancyv1.VacancyList, error) {
	return s.client.SearchVacancies(ctx, &searchv1.SearchVacanciesRequest{
		Query:         query,
		SkillSlugs:    skillSlugs,
		SalaryMin:     salaryMin,
		ExperienceMax: experienceMax,
		CompanyId:     companyID,
		Pagination:    &commonv1.Pagination{Page: page, Limit: limit},
	})
}

func (s *searchService) SearchMicroTasksAsModel(ctx context.Context, query string, skillSlugs []string, rewardMin int32, status int32, companyID string, page, limit int32) (*models.MicroTaskList, error) {
	resp, err := s.client.SearchMicroTasks(ctx, &searchv1.SearchMicroTasksRequest{
		Query:      query,
		SkillSlugs: skillSlugs,
		RewardMin:  rewardMin,
		Status:     microtaskv1.MicroTaskStatus(status),
		CompanyId:  companyID,
		Pagination: &commonv1.Pagination{Page: page, Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	out := make([]*models.MicroTask, 0, len(resp.GetTasks()))
	for _, pt := range resp.GetTasks() {
		out = append(out, &models.MicroTask{
			ID:          pt.GetId(),
			CompanyID:   pt.GetCompanyId(),
			Title:       pt.GetTitle(),
			Description: pt.GetDescription(),
			Reward:      pt.GetReward(),
			Deadline:    pt.GetDeadline(),
			SkillSlugs:  pt.GetSkillSlugs(),
			Status:      int32(pt.GetStatus()),
			AssignedTo:  pt.GetAssignedTo(),
			CreatedAt:   pt.GetCreatedAt(),
			UpdatedAt:   pt.GetUpdatedAt(),
		})
	}
	list := &models.MicroTaskList{Tasks: out}
	if p := resp.GetPagination(); p != nil {
		list.Pagination = &models.PaginationResponse{Total: p.GetTotal(), Pages: p.GetPages(), CurrentPage: p.GetCurrentPage()}
	}
	return list, nil
}

func (s *searchService) SearchVacanciesAsModel(ctx context.Context, query string, skillSlugs []string, salaryMin, experienceMax int32, companyID string, page, limit int32) (*models.VacancyList, error) {
	resp, err := s.SearchVacancies(ctx, query, skillSlugs, salaryMin, experienceMax, companyID, page, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*models.Vacancy, 0, len(resp.GetVacancies()))
	for _, pv := range resp.GetVacancies() {
		v := &models.Vacancy{
			ID:             pv.GetId(),
			Title:          pv.GetTitle(),
			Experience:     pv.GetExperience(),
			Salary:         pv.GetSalary(),
			PositionStatus: pv.GetPositionStatus(),
			Schedule:       pv.GetSchedule(),
			WorkFormat:     pv.GetWorkFormat(),
			CompanyID:      pv.GetCompanyId(),
			CreateAt:       pv.GetCreateAt(),
			SkillSlugs:     pv.GetSkillSlugs(),
		}
		if pv.GetAttachmentId() != "" {
			id := pv.GetAttachmentId()
			v.AttachmentID = &id
		}
		out = append(out, v)
	}
	list := &models.VacancyList{Vacancies: out}
	if p := resp.GetPagination(); p != nil {
		list.Pagination = &models.PaginationResponse{
			Total:       p.GetTotal(),
			Pages:       p.GetPages(),
			CurrentPage: p.GetCurrentPage(),
		}
	}
	return list, nil
}
