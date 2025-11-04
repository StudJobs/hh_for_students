package service

import (
	"context"
	"errors"
	"log"

	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/studjobs/hh_for_students/company/internal/repository"
)

var (
	ErrCompanyNotFound    = errors.New("company not found")
	ErrInvalidCompanyData = errors.New("invalid company data")
)

// CompanyService реализует CompanyInterface
type CompanyService struct {
	repo repository.Company
}

// NewCompanyService создает новый экземпляр CompanyService
func NewCompanyService(repo repository.Company) *CompanyService {
	log.Printf("Service: Initializing CompanyService")
	return &CompanyService{
		repo: repo,
	}
}

// CreateCompany создает новую компанию
func (s *CompanyService) CreateCompany(ctx context.Context, company *companyv1.Company) (*companyv1.Company, error) {
	log.Printf("Service: Creating company with name: %s", company.Name)

	// Валидация обязательных полей
	if company.Name == "" {
		log.Printf("Service: Company name is required")
		return nil, ErrInvalidCompanyData
	}

	createdCompany, err := s.repo.CreateCompany(ctx, company)
	if err != nil {
		log.Printf("Service: Failed to create company: %v", err)
		return nil, err
	}

	log.Printf("Service: Successfully created company with ID: %s", createdCompany.Id)
	return createdCompany, nil
}

// GetCompany возвращает компанию по ID
func (s *CompanyService) GetCompany(ctx context.Context, id string) (*companyv1.Company, error) {
	log.Printf("Service: Getting company with ID: %s", id)

	company, err := s.repo.GetCompany(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCompanyNotFound) {
			log.Printf("Service: Company not found with ID: %s", id)
			return nil, ErrCompanyNotFound
		}
		log.Printf("Service: Failed to get company with ID %s: %v", id, err)
		return nil, err
	}

	log.Printf("Service: Successfully retrieved company with ID: %s", id)
	return company, nil
}

// ListCompanies возвращает список компаний с фильтрацией и пагинацией
func (s *CompanyService) ListCompanies(ctx context.Context, city, companyType string, page, limit int32) (*companyv1.CompanyList, error) {
	log.Printf("Service: Listing companies - city: %s, type: %s, page: %d, limit: %d",
		city, companyType, page, limit)

	companies, err := s.repo.GetAllCompanies(ctx, city, companyType, page, limit)
	if err != nil {
		log.Printf("Service: Failed to list companies: %v", err)
		return nil, err
	}

	log.Printf("Service: Successfully listed %d companies", len(companies.Companies))
	return companies, nil
}

// UpdateCompany обновляет компанию
func (s *CompanyService) UpdateCompany(ctx context.Context, id string, company *companyv1.Company) (*companyv1.Company, error) {
	log.Printf("Service: Updating company with ID: %s", id)

	// Проверяем существование компании
	_, err := s.repo.GetCompany(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCompanyNotFound) {
			log.Printf("Service: Company not found for update with ID: %s", id)
			return nil, ErrCompanyNotFound
		}
		log.Printf("Service: Failed to get company for update with ID %s: %v", id, err)
		return nil, err
	}

	updatedCompany, err := s.repo.UpdateCompany(ctx, id, company)
	if err != nil {
		log.Printf("Service: Failed to update company with ID %s: %v", id, err)
		return nil, err
	}

	log.Printf("Service: Successfully updated company with ID: %s", id)
	return updatedCompany, nil
}

// DeleteCompany удаляет компанию
func (s *CompanyService) DeleteCompany(ctx context.Context, id string) error {
	log.Printf("Service: Deleting company with ID: %s", id)

	err := s.repo.DeleteCompany(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCompanyNotFound) {
			log.Printf("Service: Company not found for deletion with ID: %s", id)
			return ErrCompanyNotFound
		}
		log.Printf("Service: Failed to delete company with ID %s: %v", id, err)
		return err
	}

	log.Printf("Service: Successfully deleted company with ID: %s", id)
	return nil
}
