package services

import (
	"context"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
)

type vacancyService struct {
	client vacancyv1.VacancyServiceClient
}

func NewVacancyService(client vacancyv1.VacancyServiceClient) VacancyService {
	log.Printf("Creating new VacancyService")
	return &vacancyService{
		client: client,
	}
}

func (s *vacancyService) CreateVacancy(ctx context.Context, vacancy *models.Vacancy) (*models.Vacancy, error) {
	log.Printf("VacancyService: CreateVacancy attempt for title: %s", vacancy.Title)

	protoVacancy := &vacancyv1.Vacancy{
		Title:          vacancy.Title,
		Experience:     vacancy.Experience,
		Salary:         vacancy.Salary,
		PositionStatus: vacancy.PositionStatus,
		Schedule:       vacancy.Schedule,
		WorkFormat:     vacancy.WorkFormat,
		CompanyId:      vacancy.CompanyID,
	}

	// Добавляем AttachmentID если он есть
	if vacancy.AttachmentID != nil {
		protoVacancy.AttachmentId = *vacancy.AttachmentID
	}

	resp, err := s.client.NewVacancy(ctx, &vacancyv1.NewVacancyRequest{
		Vacancy: protoVacancy,
	})
	if err != nil {
		log.Printf("VacancyService: CreateVacancy failed for title %s: %v", vacancy.Title, err)
		return nil, err
	}

	result := &models.Vacancy{
		ID:             resp.Id,
		Title:          resp.Title,
		Experience:     resp.Experience,
		Salary:         resp.Salary,
		PositionStatus: resp.PositionStatus,
		Schedule:       resp.Schedule,
		WorkFormat:     resp.WorkFormat,
		CompanyID:      resp.CompanyId,
		CreateAt:       resp.CreateAt,
	}

	// Сохраняем AttachmentID из ответа
	if resp.AttachmentId != "" {
		attachmentID := resp.AttachmentId
		result.AttachmentID = &attachmentID
	}

	log.Printf("VacancyService: CreateVacancy successful for id: %s", resp.Id)
	return result, nil
}

func (s *vacancyService) GetVacancy(ctx context.Context, id string) (*models.Vacancy, error) {
	log.Printf("VacancyService: GetVacancy attempt for id: %s", id)

	resp, err := s.client.GetVacancy(ctx, &vacancyv1.GetVacancyRequest{
		Id: id,
	})
	if err != nil {
		log.Printf("VacancyService: GetVacancy failed for id %s: %v", id, err)
		return nil, err
	}

	vacancy := &models.Vacancy{
		ID:             resp.Id,
		Title:          resp.Title,
		Experience:     resp.Experience,
		Salary:         resp.Salary,
		PositionStatus: resp.PositionStatus,
		Schedule:       resp.Schedule,
		WorkFormat:     resp.WorkFormat,
		CompanyID:      resp.CompanyId,
		CreateAt:       resp.CreateAt,
	}

	// Добавляем AttachmentID из ответа
	if resp.AttachmentId != "" {
		attachmentID := resp.AttachmentId
		vacancy.AttachmentID = &attachmentID
	}

	log.Printf("VacancyService: GetVacancy successful for id: %s", id)
	return vacancy, nil
}

func (s *vacancyService) GetAllVacancies(ctx context.Context, pagination *models.Pagination,
	companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string) (*models.VacancyList, error) {

	log.Printf("VacancyService: GetAllVacancies attempt with filters - company: %s, positionStatus: %s, workFormat: %s, schedule: %s, salary: %d-%d, experience: %d-%d, search: %s",
		companyID, positionStatus, workFormat, schedule, minSalary, maxSalary, minExperience, maxExperience, searchTitle)

	req := &vacancyv1.GetAllVacanciesRequest{
		CompanyId:      companyID,
		PositionStatus: positionStatus,
		WorkFormat:     workFormat,
		Schedule:       schedule,
		MinSalary:      minSalary,
		MaxSalary:      maxSalary,
		MinExperience:  minExperience,
		MaxExperience:  maxExperience,
		SearchTitle:    searchTitle,
	}

	if pagination != nil {
		req.Pagination = &commonv1.Pagination{
			Page:  pagination.Page,
			Limit: pagination.Limit,
		}
	}

	resp, err := s.client.GetAllVacancies(ctx, req)
	if err != nil {
		log.Printf("VacancyService: GetAllVacancies failed: %v", err)
		return nil, err
	}

	vacancies := make([]*models.Vacancy, len(resp.Vacancies))
	for i, protoVacancy := range resp.Vacancies {
		vacancy := &models.Vacancy{
			ID:             protoVacancy.Id,
			Title:          protoVacancy.Title,
			Experience:     protoVacancy.Experience,
			Salary:         protoVacancy.Salary,
			PositionStatus: protoVacancy.PositionStatus,
			Schedule:       protoVacancy.Schedule,
			WorkFormat:     protoVacancy.WorkFormat,
			CompanyID:      protoVacancy.CompanyId,
			CreateAt:       protoVacancy.CreateAt,
		}

		// Добавляем AttachmentID для каждой вакансии
		if protoVacancy.AttachmentId != "" {
			attachmentID := protoVacancy.AttachmentId
			vacancy.AttachmentID = &attachmentID
		}

		vacancies[i] = vacancy
	}

	result := &models.VacancyList{
		Vacancies: vacancies,
	}

	if resp.Pagination != nil {
		result.Pagination = &models.PaginationResponse{
			Total:       resp.Pagination.Total,
			Pages:       resp.Pagination.Pages,
			CurrentPage: resp.Pagination.CurrentPage,
		}
	}

	log.Printf("VacancyService: GetAllVacancies successful, found %d vacancies", len(vacancies))
	return result, nil
}

func (s *vacancyService) GetHRVacancies(ctx context.Context, pagination *models.Pagination,
	companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string) (*models.VacancyList, error) {

	log.Printf("VacancyService: GetHRVacancies attempt with filters - company: %s, positionStatus: %s, workFormat: %s, schedule: %s, salary: %d-%d, experience: %d-%d, search: %s",
		companyID, positionStatus, workFormat, schedule, minSalary, maxSalary, minExperience, maxExperience, searchTitle)

	req := &vacancyv1.GetHRVacanciesRequest{
		CompanyId:      companyID,
		PositionStatus: positionStatus,
		WorkFormat:     workFormat,
		Schedule:       schedule,
		MinSalary:      minSalary,
		MaxSalary:      maxSalary,
		MinExperience:  minExperience,
		MaxExperience:  maxExperience,
		SearchTitle:    searchTitle,
	}

	if pagination != nil {
		req.Pagination = &commonv1.Pagination{
			Page:  pagination.Page,
			Limit: pagination.Limit,
		}
	}

	resp, err := s.client.GetHRVacancies(ctx, req)
	if err != nil {
		log.Printf("VacancyService: GetHRVacancies failed: %v", err)
		return nil, err
	}

	vacancies := make([]*models.Vacancy, len(resp.Vacancies))
	for i, protoVacancy := range resp.Vacancies {
		vacancy := &models.Vacancy{
			ID:             protoVacancy.Id,
			Title:          protoVacancy.Title,
			Experience:     protoVacancy.Experience,
			Salary:         protoVacancy.Salary,
			PositionStatus: protoVacancy.PositionStatus,
			Schedule:       protoVacancy.Schedule,
			WorkFormat:     protoVacancy.WorkFormat,
			CompanyID:      protoVacancy.CompanyId,
			CreateAt:       protoVacancy.CreateAt,
		}

		// Добавляем AttachmentID для каждой вакансии
		if protoVacancy.AttachmentId != "" {
			attachmentID := protoVacancy.AttachmentId
			vacancy.AttachmentID = &attachmentID
		}

		vacancies[i] = vacancy
	}

	result := &models.VacancyList{
		Vacancies: vacancies,
	}

	if resp.Pagination != nil {
		result.Pagination = &models.PaginationResponse{
			Total:       resp.Pagination.Total,
			Pages:       resp.Pagination.Pages,
			CurrentPage: resp.Pagination.CurrentPage,
		}
	}

	log.Printf("VacancyService: GetHRVacancies successful, found %d vacancies", len(vacancies))
	return result, nil
}

func (s *vacancyService) UpdateVacancy(ctx context.Context, id string, vacancy *models.Vacancy) (*models.Vacancy, error) {
	log.Printf("VacancyService: UpdateVacancy attempt for id: %s", id)

	protoVacancy := &vacancyv1.Vacancy{
		Title:          vacancy.Title,
		Experience:     vacancy.Experience,
		Salary:         vacancy.Salary,
		PositionStatus: vacancy.PositionStatus,
		Schedule:       vacancy.Schedule,
		WorkFormat:     vacancy.WorkFormat,
		CompanyId:      vacancy.CompanyID,
	}

	// Добавляем AttachmentID если он есть
	if vacancy.AttachmentID != nil {
		protoVacancy.AttachmentId = *vacancy.AttachmentID
	}

	resp, err := s.client.UpdateVacancy(ctx, &vacancyv1.UpdateVacancyRequest{
		Id:      id,
		Vacancy: protoVacancy,
	})
	if err != nil {
		log.Printf("VacancyService: UpdateVacancy failed for id %s: %v", id, err)
		return nil, err
	}

	result := &models.Vacancy{
		ID:             resp.Id,
		Title:          resp.Title,
		Experience:     resp.Experience,
		Salary:         resp.Salary,
		PositionStatus: resp.PositionStatus,
		Schedule:       resp.Schedule,
		WorkFormat:     resp.WorkFormat,
		CompanyID:      resp.CompanyId,
		CreateAt:       resp.CreateAt,
	}

	// Сохраняем AttachmentID из ответа
	if resp.AttachmentId != "" {
		attachmentID := resp.AttachmentId
		result.AttachmentID = &attachmentID
	}

	log.Printf("VacancyService: UpdateVacancy successful for id: %s", id)
	return result, nil
}

func (s *vacancyService) DeleteVacancy(ctx context.Context, id string) error {
	log.Printf("VacancyService: DeleteVacancy attempt for id: %s", id)

	_, err := s.client.DeleteVacancy(ctx, &vacancyv1.DeleteVacancyRequest{
		Id: id,
	})
	if err != nil {
		log.Printf("VacancyService: DeleteVacancy failed for id %s: %v", id, err)
		return err
	}

	log.Printf("VacancyService: DeleteVacancy successful for id: %s", id)
	return nil
}

func (s *vacancyService) GetAllPositions(ctx context.Context) ([]string, error) {
	log.Printf("VacancyService: GetAllPositions attempt")

	resp, err := s.client.GetAllExistPositions(ctx, &vacancyv1.PositionsRequest{})
	if err != nil {
		log.Printf("VacancyService: GetAllPositions failed: %v", err)
		return nil, err
	}

	log.Printf("VacancyService: GetAllPositions successful, found %d positions", len(resp.Position))
	return resp.Position, nil
}
