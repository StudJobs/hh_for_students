package handlers

import (
	"context"
	"errors"
	"log"

	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/studjobs/hh_for_students/company/internal/repository"
)

func (h *CompanyHandlers) ApplyMembership(ctx context.Context, req *companyv1.ApplyMembershipRequest) (*companyv1.CompanyMember, error) {
	if req.GetCompanyId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "company_id and user_id required")
	}
	m, err := h.service.Membership.Apply(ctx, req.GetCompanyId(), req.GetUserId(), req.GetNote())
	if err != nil {
		log.Printf("ApplyMembership failed: %v", err)
		return nil, status.Error(codes.Internal, "apply failed")
	}
	return m, nil
}

func (h *CompanyHandlers) ReviewMembership(ctx context.Context, req *companyv1.ReviewMembershipRequest) (*companyv1.CompanyMember, error) {
	if req.GetMembershipId() == "" {
		return nil, status.Error(codes.InvalidArgument, "membership_id required")
	}
	m, err := h.service.Membership.Review(ctx, req.GetMembershipId(), req.GetStatus())
	if err != nil {
		if errors.Is(err, repository.ErrMembershipNotFound) {
			return nil, status.Error(codes.NotFound, "membership not found")
		}
		log.Printf("ReviewMembership failed: %v", err)
		return nil, status.Error(codes.Internal, "review failed")
	}
	return m, nil
}

func (h *CompanyHandlers) ListMembers(ctx context.Context, req *companyv1.ListMembersRequest) (*companyv1.CompanyMemberList, error) {
	if req.GetCompanyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "company_id required")
	}
	list, err := h.service.Membership.ListByCompany(ctx, req.GetCompanyId(), req.GetStatus())
	if err != nil {
		log.Printf("ListMembers failed: %v", err)
		return nil, status.Error(codes.Internal, "list failed")
	}
	return &companyv1.CompanyMemberList{Members: list}, nil
}

func (h *CompanyHandlers) GetMembershipByUser(ctx context.Context, req *companyv1.GetMembershipByUserRequest) (*companyv1.CompanyMember, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	m, err := h.service.Membership.GetByUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, repository.ErrMembershipNotFound) {
			return nil, status.Error(codes.NotFound, "no membership")
		}
		log.Printf("GetMembershipByUser failed: %v", err)
		return nil, status.Error(codes.Internal, "get failed")
	}
	return m, nil
}

func (h *CompanyHandlers) ListMembershipsByUser(ctx context.Context, req *companyv1.ListMembershipsByUserRequest) (*companyv1.CompanyMemberList, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	list, err := h.service.Membership.ListByUser(ctx, req.GetUserId(), req.GetStatus())
	if err != nil {
		log.Printf("ListMembershipsByUser failed: %v", err)
		return nil, status.Error(codes.Internal, "list failed")
	}
	return &companyv1.CompanyMemberList{Members: list}, nil
}
