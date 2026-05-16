package repository

import (
	"context"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Company interface {
	GetCompany(ctx context.Context, id string) (*companyv1.Company, error)
	GetAllCompanies(ctx context.Context, city, companyType, query string, page, limit int32) (*companyv1.CompanyList, error)
	CreateCompany(ctx context.Context, company *companyv1.Company) (*companyv1.Company, error)
	UpdateCompany(ctx context.Context, id string, company *companyv1.Company) (*companyv1.Company, error)
	DeleteCompany(ctx context.Context, id string) error
}

type Membership interface {
	Apply(ctx context.Context, companyID, userID, note string) (*companyv1.CompanyMember, error)
	Review(ctx context.Context, membershipID string, status companyv1.MembershipStatus) (*companyv1.CompanyMember, error)
	ListByCompany(ctx context.Context, companyID string, status companyv1.MembershipStatus) ([]*companyv1.CompanyMember, error)
	GetByUser(ctx context.Context, userID string) (*companyv1.CompanyMember, error)
}

type Repository struct {
	Company    Company
	Membership Membership
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Company:    NewCompanyRepository(db),
		Membership: NewMembershipRepository(db),
	}
}
