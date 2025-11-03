// internal/service/users_service.go
package service

import (
	"context"
	"errors"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/studjobs/hh_for_students/company/internal/repository"
	"log"
)

var (
	ErrCompanyNotFound    = errors.New("company not found")
	ErrInvalidCompanyData = errors.New("invalid company data")
)

type Company interface {
	CreateCompany(ctx context.Context, company *companyv1.Company) (*companyv1.Company, error)
	UpdateCompany(ctx context.Context, id string, company *companyv1.Company) (*companyv1.Company, error)
	DeleteCompany(ctx context.Context, id string) error
	GetCompany(ctx context.Context, id string) (*companyv1.Company, error)
	ListCompanies(ctx context.Context, city, companyType string, page, limit int32) (*companyv1.CompanyList, error)
}

type Service struct {
	Company Company
}

func NewService(repo *repository.Repository) *Service {
	log.Println("Service: Initializing CompanyService")
	return &Service{
		Company: NewCompanyService(repo),
	}
}
