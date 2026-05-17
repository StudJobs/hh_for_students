package service

import (
	"context"

	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"

	"github.com/studjobs/hh_for_students/company/internal/repository"
)

type MembershipInterface interface {
	Apply(ctx context.Context, companyID, userID, note string) (*companyv1.CompanyMember, error)
	Review(ctx context.Context, membershipID string, status companyv1.MembershipStatus) (*companyv1.CompanyMember, error)
	ListByCompany(ctx context.Context, companyID string, status companyv1.MembershipStatus) ([]*companyv1.CompanyMember, error)
	GetByUser(ctx context.Context, userID string) (*companyv1.CompanyMember, error)
	ListByUser(ctx context.Context, userID string, status companyv1.MembershipStatus) ([]*companyv1.CompanyMember, error)
}

type membershipService struct {
	repo repository.Membership
}

func NewMembershipService(repo repository.Membership) MembershipInterface {
	return &membershipService{repo: repo}
}

func (s *membershipService) Apply(ctx context.Context, companyID, userID, note string) (*companyv1.CompanyMember, error) {
	return s.repo.Apply(ctx, companyID, userID, note)
}

func (s *membershipService) Review(ctx context.Context, id string, status companyv1.MembershipStatus) (*companyv1.CompanyMember, error) {
	return s.repo.Review(ctx, id, status)
}

func (s *membershipService) ListByCompany(ctx context.Context, companyID string, status companyv1.MembershipStatus) ([]*companyv1.CompanyMember, error) {
	return s.repo.ListByCompany(ctx, companyID, status)
}

func (s *membershipService) GetByUser(ctx context.Context, userID string) (*companyv1.CompanyMember, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *membershipService) ListByUser(ctx context.Context, userID string, status companyv1.MembershipStatus) ([]*companyv1.CompanyMember, error) {
	return s.repo.ListByUser(ctx, userID, status)
}
