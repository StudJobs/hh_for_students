package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"github.com/google/uuid"
	"hh_for_students/vacancy-service/internal/repository"
)

type VacancyService struct {
	repo *repository.Repository
}

func NewVacancyService(repo *repository.Repository) *VacancyService {
	return &VacancyService{
		repo: repo,
	}
}

func (s *VacancyService) CreateVacancy(ctx context.Context, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	log.Printf("Service: Creating new vacancy with title: %s", vacancy.Title)

	// Валидация данных вакансии
	if err := s.validateVacancy(vacancy); err != nil {
		log.Printf("Service: Vacancy validation failed for title %s: %v", vacancy.Title, err)
		return nil, err
	}

	log.Printf("Service: Creating vacancy in repository for title: %s", vacancy.Title)
	createdVacancy, err := s.repo.Vacancy.CreateVacancy(ctx, vacancy)
	if err != nil {
		log.Printf("Service: Failed to create vacancy for title %s: %v", vacancy.Title, err)
		return nil, fmt.Errorf("failed to create vacancy: %w", err)
	}

	log.Printf("Service: Vacancy created successfully with ID: %s", createdVacancy.Id)
	return createdVacancy, nil
}

func (s *VacancyService) UpdateVacancy(ctx context.Context, id string, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	log.Printf("Service: Updating vacancy with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return nil, fmt.Errorf("%w: invalid uuid format", ErrInvalidVacancyData)
	}

	log.Printf("Service: Updating vacancy in repository for ID: %s", id)
	updatedVacancy, err := s.repo.Vacancy.UpdateVacancy(ctx, id, vacancy)
	if err != nil {
		log.Printf("Service: Failed to update vacancy with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to update vacancy: %w", err)
	}

	log.Printf("Service: Vacancy updated successfully with ID: %s", updatedVacancy.Id)
	return updatedVacancy, nil
}

func (s *VacancyService) DeleteVacancy(ctx context.Context, id string) error {
	log.Printf("Service: Deleting vacancy with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return fmt.Errorf("%w: invalid uuid format", ErrInvalidVacancyData)
	}

	log.Printf("Service: Deleting vacancy in repository for ID: %s", id)
	err := s.repo.Vacancy.DeleteVacancy(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrVacancyNotFound) {
			log.Printf("Service: Vacancy not found for deletion with ID: %s", id)
			return ErrVacancyNotFound
		}
		log.Printf("Service: Failed to delete vacancy with ID %s: %v", id, err)
		return fmt.Errorf("failed to delete vacancy: %w", err)
	}

	log.Printf("Service: Vacancy deleted successfully with ID: %s", id)
	return nil
}

func (s *VacancyService) GetVacancy(ctx context.Context, id string) (*vacancyv1.Vacancy, error) {
	log.Printf("Service: Getting vacancy with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return nil, fmt.Errorf("%w: invalid uuid format", ErrInvalidVacancyData)
	}

	log.Printf("Service: Getting vacancy from repository for ID: %s", id)
	vacancy, err := s.repo.Vacancy.GetVacancy(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrVacancyNotFound) {
			log.Printf("Service: Vacancy not found for ID: %s", id)
			return nil, ErrVacancyNotFound
		}
		log.Printf("Service: Failed to get vacancy with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get vacancy: %w", err)
	}

	log.Printf("Service: Vacancy retrieved successfully with ID: %s", vacancy.Id)
	return vacancy, nil
}

func (s *VacancyService) GetAllVacancies(ctx context.Context, companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string, page, limit int32) (*vacancyv1.VacancyList, error) {

	log.Printf("Service: Listing vacancies - page: %d, limit: %d, company: %s, status: %s, work_format: %s, schedule: %s, salary: %d-%d, experience: %d-%d, search: %s",
		page, limit, companyID, positionStatus, workFormat, schedule, minSalary, maxSalary, minExperience, maxExperience, searchTitle)

	if limit < 1 || limit > 100 {
		limit = 10
		log.Printf("Service: Limit adjusted to default: %d", limit)
	}

	log.Printf("Service: Getting vacancies from repository")
	vacancies, err := s.repo.Vacancy.GetAllVacancies(ctx, companyID, positionStatus, workFormat, schedule,
		minSalary, maxSalary, minExperience, maxExperience, searchTitle, page, limit)
	if err != nil {
		log.Printf("Service: Failed to list vacancies: %v", err)
		return nil, fmt.Errorf("failed to list vacancies: %w", err)
	}

	log.Printf("Service: Retrieved %d vacancies successfully", len(vacancies.Vacancies))
	return vacancies, nil
}

func (s *VacancyService) GetHRVacancies(ctx context.Context, companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string, page, limit int32) (*vacancyv1.VacancyList, error) {

	log.Printf("Service: Listing HR vacancies - page: %d, limit: %d, company: %s, status: %s, work_format: %s, schedule: %s, salary: %d-%d, experience: %d-%d, search: %s",
		page, limit, companyID, positionStatus, workFormat, schedule, minSalary, maxSalary, minExperience, maxExperience, searchTitle)

	if limit < 1 || limit > 100 {
		limit = 10
		log.Printf("Service: Limit adjusted to default: %d", limit)
	}

	log.Printf("Service: Getting HR vacancies from repository")
	vacancies, err := s.repo.Vacancy.GetHRVacancies(ctx, companyID, positionStatus, workFormat, schedule,
		minSalary, maxSalary, minExperience, maxExperience, searchTitle, page, limit)
	if err != nil {
		log.Printf("Service: Failed to list HR vacancies: %v", err)
		return nil, fmt.Errorf("failed to list HR vacancies: %w", err)
	}

	log.Printf("Service: Retrieved %d HR vacancies successfully", len(vacancies.Vacancies))
	return vacancies, nil
}

// validateVacancy выполняет валидацию данных вакансии при создании
func (s *VacancyService) validateVacancy(vacancy *vacancyv1.Vacancy) error {
	log.Printf("Service: Validating vacancy data for title: %s", vacancy.Title)

	if vacancy.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidVacancyData)
	}
	if len(vacancy.Title) > 128 {
		return fmt.Errorf("%w: title must be less than 128 characters", ErrInvalidVacancyData)
	}
	if vacancy.Experience < 0 {
		return fmt.Errorf("%w: experience cannot be negative", ErrInvalidVacancyData)
	}
	if vacancy.Salary < 0 {
		return fmt.Errorf("%w: salary cannot be negative", ErrInvalidVacancyData)
	}
	if vacancy.PositionStatus == "" {
		return fmt.Errorf("%w: position_status is required", ErrInvalidVacancyData)
	}
	if vacancy.Schedule == "" {
		return fmt.Errorf("%w: schedule is required", ErrInvalidVacancyData)
	}
	if len(vacancy.WorkFormat) > 64 {
		return fmt.Errorf("%w: work_format must be less than 64 characters", ErrInvalidVacancyData)
	}
	if vacancy.CompanyId == "" {
		return fmt.Errorf("%w: company_id is required", ErrInvalidVacancyData)
	}
	if _, err := uuid.Parse(vacancy.CompanyId); err != nil {
		return fmt.Errorf("%w: invalid company_id uuid format", ErrInvalidVacancyData)
	}

	log.Printf("Service: Vacancy validation successful for title: %s", vacancy.Title)
	return nil
}

func (s *VacancyService) GetAllExistPositions(ctx context.Context, req *vacancyv1.PositionsRequest) (*vacancyv1.PositionsResponse, error) {
	log.Printf("Service: Getting all existing positions")
	req.ProtoMessage()

	positions, err := s.repo.Vacancy.GetAllExistPositions(ctx)
	if err != nil {
		log.Printf("Service: Failed to get positions: %v", err)
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	log.Printf("Service: Retrieved %d positions", len(positions))
	return &vacancyv1.PositionsResponse{
		Position: positions,
	}, nil
}
