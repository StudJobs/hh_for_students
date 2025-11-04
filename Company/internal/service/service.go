// internal/service/users_service.go
package service

import (
	"context"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/studjobs/hh_for_students/company/internal/repository"
	"log"
)

// CompanyInterface определяет интерфейс для работы с компаниями
type CompanyInterface interface {
	CreateCompany(ctx context.Context, company *companyv1.Company) (*companyv1.Company, error)
	GetCompany(ctx context.Context, id string) (*companyv1.Company, error)
	ListCompanies(ctx context.Context, city, companyType string, page, limit int32) (*companyv1.CompanyList, error)
	UpdateCompany(ctx context.Context, id string, company *companyv1.Company) (*companyv1.Company, error)
	DeleteCompany(ctx context.Context, id string) error
}

type Service struct {
	Company CompanyInterface
}

func NewService(repo *repository.Repository) *Service {
	log.Println("Service: Initializing Service")
	return &Service{
		Company: NewCompanyService(repo.Company),
	}
}
