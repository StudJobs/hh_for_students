package service

import (
	"context"
	"errors"
	"fmt"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/google/uuid"
	"github.com/studjobs/hh_for_students/company/internal/repository"
	"log"
)

type CompanyService struct {
	repo *repository.Repository
}

func NewCompanyService(repo *repository.Repository) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) CreateCompany(ctx context.Context, company *companyv1.Company) (*companyv1.Company, error) {
	log.Printf("Service: Creating new company with name: %s", company.Name)

	// Валидация данных компании
	if err := s.validateCompany(company); err != nil {
		log.Printf("Service: Company validation failed for name %s: %v", company.Name, err)
		return nil, err
	}

	log.Printf("Service: Creating company in repository for name: %s", company.Name)
	createdCompany, err := s.repo.Company.CreateCompany(ctx, company)
	if err != nil {
		log.Printf("Service: Failed to create company for name %s: %v", company.Name, err)
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	log.Printf("Service: Company created successfully with ID: %s", createdCompany.Id)
	return createdCompany, nil
}

func (s *CompanyService) UpdateCompany(ctx context.Context, id string, company *companyv1.Company) (*companyv1.Company, error) {
	log.Printf("Service: Updating company with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return nil, fmt.Errorf("%w: invalid uuid format", ErrInvalidCompanyData)
	}

	log.Printf("Service: Updating company in repository for ID: %s", id)
	updatedCompany, err := s.repo.Company.UpdateCompany(ctx, id, company)
	if err != nil {
		log.Printf("Service: Failed to update company with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	log.Printf("Service: Company updated successfully with ID: %s", updatedCompany.Id)
	return updatedCompany, nil
}

func (s *CompanyService) DeleteCompany(ctx context.Context, id string) error {
	log.Printf("Service: Deleting company with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return fmt.Errorf("%w: invalid uuid format", ErrInvalidCompanyData)
	}

	log.Printf("Service: Deleting company in repository for ID: %s", id)
	err := s.repo.Company.DeleteCompany(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCompanyNotFound) {
			log.Printf("Service: Company not found for deletion with ID: %s", id)
			return ErrCompanyNotFound
		}
		log.Printf("Service: Failed to delete company with ID %s: %v", id, err)
		return fmt.Errorf("failed to delete company: %w", err)
	}

	log.Printf("Service: Company deleted successfully with ID: %s", id)
	return nil
}

func (s *CompanyService) GetCompany(ctx context.Context, id string) (*companyv1.Company, error) {
	log.Printf("Service: Getting company with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return nil, fmt.Errorf("%w: invalid uuid format", ErrInvalidCompanyData)
	}

	log.Printf("Service: Getting company from repository for ID: %s", id)
	company, err := s.repo.Company.GetCompany(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCompanyNotFound) {
			log.Printf("Service: Company not found for ID: %s", id)
			return nil, ErrCompanyNotFound
		}
		log.Printf("Service: Failed to get company with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	log.Printf("Service: Company retrieved successfully with ID: %s", company.Id)
	return company, nil
}

func (s *CompanyService) ListCompanies(ctx context.Context, city, companyType string, page, limit int32) (*companyv1.CompanyList, error) {
	log.Printf("Service: Listing companies - page: %d, limit: %d, city: %s, type: %s",
		page, limit, city, companyType)

	if page < 1 {
		page = 1
		log.Printf("Service: Page adjusted to default: %d", page)
	}
	if limit < 1 || limit > 100 {
		limit = 10
		log.Printf("Service: Limit adjusted to default: %d", limit)
	}

	log.Printf("Service: Getting companies from repository")
	companies, err := s.repo.Company.GetAllCompanies(ctx, city, companyType, page, limit)
	if err != nil {
		log.Printf("Service: Failed to list companies: %v", err)
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}

	log.Printf("Service: Retrieved %d companies successfully", len(companies.Companies))
	return companies, nil
}

// validateCompany выполняет валидацию данных компании
func (s *CompanyService) validateCompany(company *companyv1.Company) error {
	log.Printf("Service: Validating company data for name: %s", company.Name)

	if company.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidCompanyData)
	}
	if len(company.Name) > 255 {
		return fmt.Errorf("%w: name must be less than 255 characters", ErrInvalidCompanyData)
	}
	if len(company.Description) > 1000 {
		return fmt.Errorf("%w: description must be less than 1000 characters", ErrInvalidCompanyData)
	}
	if len(company.City) > 100 {
		return fmt.Errorf("%w: city must be less than 100 characters", ErrInvalidCompanyData)
	}
	if company.Site != "" && len(company.Site) > 255 {
		return fmt.Errorf("%w: site must be less than 255 characters", ErrInvalidCompanyData)
	}

	log.Printf("Service: Company validation successful for name: %s", company.Name)
	return nil
}
