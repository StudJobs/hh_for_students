package handlers

import (
	"context"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/studjobs/hh_for_students/company/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func (h *CompanyHandlers) NewCompany(ctx context.Context, req *companyv1.NewCompanyRequest) (*companyv1.Company, error) {
	log.Printf("Handlers: NewCompany request received for name: %s", req.Company.GetName())

	if req.Company == nil {
		log.Printf("Handlers: NewCompany failed - company is required")
		return nil, status.Error(codes.InvalidArgument, "company is required")
	}

	company, err := h.service.Company.CreateCompany(ctx, req.Company)
	if err != nil {
		log.Printf("Handlers: NewCompany failed for name %s: %v", req.Company.GetName(), err)
		switch err {
		case service.ErrInvalidCompanyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create company")
		}
	}

	log.Printf("Handlers: NewCompany completed successfully for ID: %s", company.Id)
	return company, nil
}

func (h *CompanyHandlers) UpdateCompany(ctx context.Context, req *companyv1.UpdateCompanyRequest) (*companyv1.Company, error) {
	log.Printf("Handlers: UpdateCompany request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: UpdateCompany failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Company == nil {
		log.Printf("Handlers: UpdateCompany failed - company is required")
		return nil, status.Error(codes.InvalidArgument, "company is required")
	}

	company, err := h.service.Company.UpdateCompany(ctx, req.Id, req.Company)
	if err != nil {
		log.Printf("Handlers: UpdateCompany failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrCompanyNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidCompanyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to update company")
		}
	}

	log.Printf("Handlers: UpdateCompany completed successfully for ID: %s", company.Id)
	return company, nil
}

func (h *CompanyHandlers) DeleteCompany(ctx context.Context, req *companyv1.DeleteCompanyRequest) (*commonv1.Empty, error) {
	log.Printf("Handlers: DeleteCompany request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: DeleteCompany failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := h.service.Company.DeleteCompany(ctx, req.Id)
	if err != nil {
		log.Printf("Handlers: DeleteCompany failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrCompanyNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidCompanyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to delete company")
		}
	}

	log.Printf("Handlers: DeleteCompany completed successfully for ID: %s", req.Id)
	return &commonv1.Empty{}, nil
}

func (h *CompanyHandlers) GetCompany(ctx context.Context, req *companyv1.GetCompanyRequest) (*companyv1.Company, error) {
	log.Printf("Handlers: GetCompany request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: GetCompany failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	company, err := h.service.Company.GetCompany(ctx, req.Id)
	if err != nil {
		log.Printf("Handlers: GetCompany failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrCompanyNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidCompanyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to get company")
		}
	}

	log.Printf("Handlers: GetCompany completed successfully for ID: %s", company.Id)
	return company, nil
}

func (h *CompanyHandlers) GetAllCompanies(ctx context.Context, req *companyv1.GetAllCompaniesRequest) (*companyv1.CompanyList, error) {
	log.Printf("Handlers: GetAllCompanies request received")

	var page, limit int32 = 1, 10

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
	}

	log.Printf("Handlers: GetAllCompanies - page: %d, limit: %d, city: %s, type: %s",
		page, limit, req.GetCity(), req.GetCompanyType())

	companies, err := h.service.Company.ListCompanies(ctx, req.City, req.CompanyType, page, limit)
	if err != nil {
		log.Printf("Handlers: GetAllCompanies failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to get companies")
	}

	log.Printf("Handlers: GetAllCompanies completed successfully, returned %d companies", len(companies.Companies))
	return companies, nil
}
