package services

import (
	"context"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"log"
)

type companyService struct {
	client companyv1.CompanyServiceClient
}

func NewCompanyService(client companyv1.CompanyServiceClient) CompanyService {
	log.Printf("Creating new CompanyService")
	return &companyService{
		client: client,
	}
}

func (s *companyService) CreateCompany(ctx context.Context, company *models.Company) (*models.Company, error) {
	log.Printf("CompanyService: CreateCompany attempt for name: %s", company.Name)

	protoCompany := &companyv1.Company{
		Name:        company.Name,
		Description: company.Description,
		City:        company.City,
		Site:        company.Site,
	}

	if company.Type != nil {
		protoCompany.Type = &companyv1.CompanyType{
			Value: company.Type.Value,
		}
	}

	resp, err := s.client.NewCompany(ctx, &companyv1.NewCompanyRequest{
		Company: protoCompany,
	})
	if err != nil {
		log.Printf("CompanyService: CreateCompany failed for name %s: %v", company.Name, err)
		return nil, err
	}

	result := &models.Company{
		ID:          resp.Id,
		Name:        resp.Name,
		Description: resp.Description,
		City:        resp.City,
		Site:        resp.Site,
	}

	if resp.Type != nil {
		result.Type = &models.CompanyType{
			Value: resp.Type.Value,
		}
	}

	log.Printf("CompanyService: CreateCompany successful for id: %s", resp.Id)
	return result, nil
}

func (s *companyService) GetCompany(ctx context.Context, id string) (*models.Company, error) {
	log.Printf("CompanyService: GetCompany attempt for id: %s", id)

	resp, err := s.client.GetCompany(ctx, &companyv1.GetCompanyRequest{
		Id: id,
	})
	if err != nil {
		log.Printf("CompanyService: GetCompany failed for id %s: %v", id, err)
		return nil, err
	}

	company := &models.Company{
		ID:          resp.Id,
		Name:        resp.Name,
		Description: resp.Description,
		City:        resp.City,
		Site:        resp.Site,
	}

	if resp.Type != nil {
		company.Type = &models.CompanyType{
			Value: resp.Type.Value,
		}
	}

	log.Printf("CompanyService: GetCompany successful for id: %s", id)
	return company, nil
}

func (s *companyService) GetAllCompanies(ctx context.Context, pagination *models.Pagination, city, companyType string) (*models.CompanyList, error) {
	log.Printf("CompanyService: GetAllCompanies attempt")

	req := &companyv1.GetAllCompaniesRequest{
		City:        city,
		CompanyType: companyType,
	}

	if pagination != nil {
		req.Pagination = &commonv1.Pagination{
			Page:  pagination.Page,
			Limit: pagination.Limit,
		}
	}

	resp, err := s.client.GetAllCompanies(ctx, req)
	if err != nil {
		log.Printf("CompanyService: GetAllCompanies failed: %v", err)
		return nil, err
	}

	companies := make([]*models.Company, len(resp.Companies))
	for i, protoCompany := range resp.Companies {
		company := &models.Company{
			ID:          protoCompany.Id,
			Name:        protoCompany.Name,
			Description: protoCompany.Description,
			City:        protoCompany.City,
			Site:        protoCompany.Site,
		}

		if protoCompany.Type != nil {
			company.Type = &models.CompanyType{
				Value: protoCompany.Type.Value,
			}
		}
		companies[i] = company
	}

	result := &models.CompanyList{
		Companies: companies,
	}

	if resp.Pagination != nil {
		result.Pagination = &models.PaginationResponse{
			Total:       resp.Pagination.Total,
			Pages:       resp.Pagination.Pages,
			CurrentPage: resp.Pagination.CurrentPage,
		}
	}

	log.Printf("CompanyService: GetAllCompanies successful, found %d companies", len(companies))
	return result, nil
}

func (s *companyService) UpdateCompany(ctx context.Context, id string, company *models.Company) (*models.Company, error) {
	log.Printf("CompanyService: UpdateCompany attempt for id: %s", id)

	protoCompany := &companyv1.Company{
		Name:        company.Name,
		Description: company.Description,
		City:        company.City,
		Site:        company.Site,
	}

	if company.Type != nil {
		protoCompany.Type = &companyv1.CompanyType{
			Value: company.Type.Value,
		}
	}

	resp, err := s.client.UpdateCompany(ctx, &companyv1.UpdateCompanyRequest{
		Id:      id,
		Company: protoCompany,
	})
	if err != nil {
		log.Printf("CompanyService: UpdateCompany failed for id %s: %v", id, err)
		return nil, err
	}

	result := &models.Company{
		ID:          resp.Id,
		Name:        resp.Name,
		Description: resp.Description,
		City:        resp.City,
		Site:        resp.Site,
	}

	if resp.Type != nil {
		result.Type = &models.CompanyType{
			Value: resp.Type.Value,
		}
	}

	log.Printf("CompanyService: UpdateCompany successful for id: %s", id)
	return result, nil
}

func (s *companyService) DeleteCompany(ctx context.Context, id string) error {
	log.Printf("CompanyService: DeleteCompany attempt for id: %s", id)

	_, err := s.client.DeleteCompany(ctx, &companyv1.DeleteCompanyRequest{
		Id: id,
	})
	if err != nil {
		log.Printf("CompanyService: DeleteCompany failed for id %s: %v", id, err)
		return err
	}

	log.Printf("CompanyService: DeleteCompany successful for id: %s", id)
	return nil
}
