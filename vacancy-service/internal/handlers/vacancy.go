package handlers

import (
	"context"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"hh_for_students/vacancy-service/internal/service"
	"log"
)

func (h *VacancyHandler) NewVacancy(ctx context.Context, req *vacancyv1.NewVacancyRequest) (*vacancyv1.Vacancy, error) {
	log.Printf("Handlers: NewVacancy request received for title: %s", req.Vacancy.GetTitle())

	if req.Vacancy == nil {
		log.Printf("Handlers: NewVacancy failed - vacancy is required")
		return nil, status.Error(codes.InvalidArgument, "vacancy is required")
	}

	vacancy, err := h.service.Vacancy.CreateVacancy(ctx, req.Vacancy)
	if err != nil {
		log.Printf("Handlers: NewVacancy failed for title %s: %v", req.Vacancy.GetTitle(), err)
		switch err {
		case service.ErrInvalidVacancyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create vacancy")
		}
	}

	log.Printf("Handlers: NewVacancy completed successfully for ID: %s", vacancy.Id)
	return vacancy, nil
}

func (h *VacancyHandler) UpdateVacancy(ctx context.Context, req *vacancyv1.UpdateVacancyRequest) (*vacancyv1.Vacancy, error) {
	log.Printf("Handlers: UpdateVacancy request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: UpdateVacancy failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Vacancy == nil {
		log.Printf("Handlers: UpdateVacancy failed - vacancy is required")
		return nil, status.Error(codes.InvalidArgument, "vacancy is required")
	}

	vacancy, err := h.service.Vacancy.UpdateVacancy(ctx, req.Id, req.Vacancy)
	if err != nil {
		log.Printf("Handlers: UpdateVacancy failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrVacancyNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidVacancyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to update vacancy")
		}
	}

	log.Printf("Handlers: UpdateVacancy completed successfully for ID: %s", vacancy.Id)
	return vacancy, nil
}

func (h *VacancyHandler) DeleteVacancy(ctx context.Context, req *vacancyv1.DeleteVacancyRequest) (*commonv1.Empty, error) {
	log.Printf("Handlers: DeleteVacancy request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: DeleteVacancy failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := h.service.Vacancy.DeleteVacancy(ctx, req.Id)
	if err != nil {
		log.Printf("Handlers: DeleteVacancy failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrVacancyNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidVacancyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to delete vacancy")
		}
	}

	log.Printf("Handlers: DeleteVacancy completed successfully for ID: %s", req.Id)
	return &commonv1.Empty{}, nil
}

func (h *VacancyHandler) GetVacancy(ctx context.Context, req *vacancyv1.GetVacancyRequest) (*vacancyv1.Vacancy, error) {
	log.Printf("Handlers: GetVacancy request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: GetVacancy failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	vacancy, err := h.service.Vacancy.GetVacancy(ctx, req.Id)
	if err != nil {
		log.Printf("Handlers: GetVacancy failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrVacancyNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidVacancyData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to get vacancy")
		}
	}

	log.Printf("Handlers: GetVacancy completed successfully for ID: %s", vacancy.Id)
	return vacancy, nil
}

func (h *VacancyHandler) GetAllVacancies(ctx context.Context, req *vacancyv1.GetAllVacanciesRequest) (*vacancyv1.VacancyList, error) {
	log.Printf("Handlers: GetAllVacancies request received with filters - company: %s, status: %s, work_format: %s, schedule: %s, salary: %d-%d, experience: %d-%d, search: %s",
		req.GetCompanyId(), req.GetPositionStatus(), req.GetWorkFormat(), req.GetSchedule(),
		req.GetMinSalary(), req.GetMaxSalary(), req.GetMinExperience(), req.GetMaxExperience(),
		req.GetSearchTitle())

	var page, limit int32 = 1, 10

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
	}

	vacancies, err := h.service.Vacancy.GetAllVacancies(ctx,
		req.CompanyId,
		req.PositionStatus,
		req.WorkFormat,
		req.Schedule,
		req.MinSalary,
		req.MaxSalary,
		req.MinExperience,
		req.MaxExperience,
		req.SearchTitle,
		page,
		limit)
	if err != nil {
		log.Printf("Handlers: GetAllVacancies failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to get vacancies")
	}

	log.Printf("Handlers: GetAllVacancies completed successfully, returned %d vacancies", len(vacancies.Vacancies))
	return vacancies, nil
}

func (h *VacancyHandler) GetHRVacancies(ctx context.Context, req *vacancyv1.GetHRVacanciesRequest) (*vacancyv1.VacancyList, error) {
	log.Printf("Handlers: GetHRVacancies request received with filters - company: %s, status: %s, work_format: %s, schedule: %s, salary: %d-%d, experience: %d-%d, search: %s",
		req.GetCompanyId(), req.GetPositionStatus(), req.GetWorkFormat(), req.GetSchedule(),
		req.GetMinSalary(), req.GetMaxSalary(), req.GetMinExperience(), req.GetMaxExperience(),
		req.GetSearchTitle())

	var page, limit int32 = 1, 10

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
	}

	vacancies, err := h.service.Vacancy.GetHRVacancies(ctx,
		req.CompanyId,
		req.PositionStatus,
		req.WorkFormat,
		req.Schedule,
		req.MinSalary,
		req.MaxSalary,
		req.MinExperience,
		req.MaxExperience,
		req.SearchTitle,
		page,
		limit)
	if err != nil {
		log.Printf("Handlers: GetHRVacancies failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to get HR vacancies")
	}

	log.Printf("Handlers: GetHRVacancies completed successfully, returned %d vacancies", len(vacancies.Vacancies))
	return vacancies, nil
}

func (h *VacancyHandler) GetAllExistPositions(ctx context.Context, req *vacancyv1.PositionsRequest) (*vacancyv1.PositionsResponse, error) {
	log.Printf("gRPC GetAllExistPositions request")

	response, err := h.service.Vacancy.GetAllExistPositions(ctx, req)
	if err != nil {
		log.Printf("gRPC GetAllExistPositions failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to get positions")
	}

	log.Printf("gRPC GetAllExistPositions successful - found %d positions", len(response.Position))
	return response, nil
}
