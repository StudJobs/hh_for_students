package handlers

import (
	"context"
	"errors"
	"log"

	applicationv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/application/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"hh_for_students/vacancy-service/internal/service"
)

// ApplicationHandler реализует gRPC ApplicationServiceServer.
// Регистрируется на том же gRPC-сервере, что и VacancyHandler (порт 50054).
type ApplicationHandler struct {
	applicationv1.UnimplementedApplicationServiceServer
	service *service.Service
}

func NewApplicationHandler(svc *service.Service) *ApplicationHandler {
	return &ApplicationHandler{service: svc}
}

func (h *ApplicationHandler) Apply(ctx context.Context, req *applicationv1.ApplyRequest) (*applicationv1.Application, error) {
	app, err := h.service.Application.Apply(ctx, req.GetVacancyId(), req.GetStudentId(), req.GetCoverLetter())
	if err != nil {
		log.Printf("AppHandlers: Apply failed: %v", err)
		if errors.Is(err, service.ErrInvalidApplicationData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to apply")
	}
	return app, nil
}

func (h *ApplicationHandler) Withdraw(ctx context.Context, req *applicationv1.WithdrawRequest) (*commonv1.Empty, error) {
	if err := h.service.Application.Withdraw(ctx, req.GetId(), req.GetStudentId()); err != nil {
		log.Printf("AppHandlers: Withdraw failed: %v", err)
		switch {
		case errors.Is(err, service.ErrApplicationNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, service.ErrInvalidApplicationData):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to withdraw")
		}
	}
	return &commonv1.Empty{}, nil
}

func (h *ApplicationHandler) ListMine(ctx context.Context, req *applicationv1.ListMineRequest) (*applicationv1.ApplicationList, error) {
	page, limit := paginationFrom(req.GetPagination())
	list, err := h.service.Application.ListByStudent(ctx, req.GetStudentId(), req.GetStatus(), page, limit)
	if err != nil {
		log.Printf("AppHandlers: ListMine failed: %v", err)
		if errors.Is(err, service.ErrInvalidApplicationData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to list applications")
	}
	return list, nil
}

func (h *ApplicationHandler) ListForVacancy(ctx context.Context, req *applicationv1.ListForVacancyRequest) (*applicationv1.ApplicationList, error) {
	page, limit := paginationFrom(req.GetPagination())
	list, err := h.service.Application.ListByVacancy(ctx, req.GetVacancyId(), req.GetStatus(), page, limit)
	if err != nil {
		log.Printf("AppHandlers: ListForVacancy failed: %v", err)
		if errors.Is(err, service.ErrInvalidApplicationData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to list applications")
	}
	return list, nil
}

func (h *ApplicationHandler) UpdateStatus(ctx context.Context, req *applicationv1.UpdateStatusRequest) (*applicationv1.Application, error) {
	app, err := h.service.Application.UpdateStatus(ctx, req.GetId(), req.GetStatus(), req.GetHrComment())
	if err != nil {
		log.Printf("AppHandlers: UpdateStatus failed: %v", err)
		switch {
		case errors.Is(err, service.ErrApplicationNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, service.ErrInvalidApplicationData):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to update status")
		}
	}
	return app, nil
}

func paginationFrom(p *commonv1.Pagination) (page, limit int32) {
	page, limit = 1, 20
	if p != nil {
		if p.Page > 0 {
			page = p.Page
		}
		if p.Limit > 0 {
			limit = p.Limit
		}
	}
	return
}
