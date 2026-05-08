package handlers

import (
	"context"
	"errors"
	"log"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/studjobs/hh_for_students/microtasks/internal/achievementclient"
	"github.com/studjobs/hh_for_students/microtasks/internal/searchclient"
	"github.com/studjobs/hh_for_students/microtasks/internal/service"
)

type Handler struct {
	microtaskv1.UnimplementedMicroTaskServiceServer

	svc          *service.Service
	search       *searchclient.Client
	achievements *achievementclient.Client
}

func New(svc *service.Service, search *searchclient.Client, achievements *achievementclient.Client) *Handler {
	return &Handler{svc: svc, search: search, achievements: achievements}
}

func (h *Handler) Create(ctx context.Context, req *microtaskv1.CreateMicroTaskRequest) (*microtaskv1.MicroTask, error) {
	t, err := h.svc.Tasks.Create(ctx, req.GetTask())
	if err != nil {
		return nil, mapErr(err, "create")
	}
	log.Printf("Handlers: Create microtask id=%s company=%s", t.GetId(), t.GetCompanyId())
	h.search.IndexTask(ctx, t)
	return t, nil
}

func (h *Handler) Update(ctx context.Context, req *microtaskv1.UpdateMicroTaskRequest) (*microtaskv1.MicroTask, error) {
	t, err := h.svc.Tasks.Update(ctx, req.GetId(), req.GetTask())
	if err != nil {
		return nil, mapErr(err, "update")
	}
	h.search.IndexTask(ctx, t)
	return t, nil
}

func (h *Handler) Delete(ctx context.Context, req *microtaskv1.DeleteMicroTaskRequest) (*commonv1.Empty, error) {
	if err := h.svc.Tasks.Delete(ctx, req.GetId()); err != nil {
		return nil, mapErr(err, "delete")
	}
	h.search.DeleteTask(ctx, req.GetId())
	return &commonv1.Empty{}, nil
}

func (h *Handler) Get(ctx context.Context, req *microtaskv1.GetMicroTaskRequest) (*microtaskv1.MicroTask, error) {
	t, err := h.svc.Tasks.Get(ctx, req.GetId())
	if err != nil {
		return nil, mapErr(err, "get")
	}
	return t, nil
}

func (h *Handler) List(ctx context.Context, req *microtaskv1.ListMicroTasksRequest) (*microtaskv1.MicroTaskList, error) {
	page, limit := normalizePagination(req.GetPagination())
	list, err := h.svc.Tasks.List(ctx, req.GetStatus(), req.GetSkillSlugs(), page, limit)
	if err != nil {
		return nil, mapErr(err, "list")
	}
	return list, nil
}

func (h *Handler) ListByCompany(ctx context.Context, req *microtaskv1.ListByCompanyRequest) (*microtaskv1.MicroTaskList, error) {
	page, limit := normalizePagination(req.GetPagination())
	list, err := h.svc.Tasks.ListByCompany(ctx, req.GetCompanyId(), page, limit)
	if err != nil {
		return nil, mapErr(err, "list-by-company")
	}
	return list, nil
}

func (h *Handler) Apply(ctx context.Context, req *microtaskv1.ApplyRequest) (*microtaskv1.MicroTask, error) {
	t, err := h.svc.Tasks.Apply(ctx, req.GetMicrotaskId(), req.GetStudentId())
	if err != nil {
		return nil, mapErr(err, "apply")
	}
	log.Printf("Handlers: Apply microtask=%s student=%s", t.GetId(), req.GetStudentId())
	h.search.IndexTask(ctx, t)
	return t, nil
}

func (h *Handler) Submit(ctx context.Context, req *microtaskv1.SubmitRequest) (*microtaskv1.Submission, error) {
	s, err := h.svc.Submissions.Submit(ctx, req.GetMicrotaskId(), req.GetStudentId(), req.GetSolutionUrl(), req.GetComment())
	if err != nil {
		return nil, mapErr(err, "submit")
	}
	log.Printf("Handlers: Submit microtask=%s student=%s submission=%s", s.GetMicrotaskId(), s.GetStudentId(), s.GetId())
	return s, nil
}

func (h *Handler) ListSubmissions(ctx context.Context, req *microtaskv1.ListSubmissionsRequest) (*microtaskv1.SubmissionList, error) {
	page, limit := normalizePagination(req.GetPagination())
	if mt := req.GetMicrotaskId(); mt != "" {
		list, err := h.svc.Submissions.ListByTask(ctx, mt, page, limit)
		if err != nil {
			return nil, mapErr(err, "list-submissions-by-task")
		}
		return list, nil
	}
	if st := req.GetStudentId(); st != "" {
		list, err := h.svc.Submissions.ListByStudent(ctx, st, page, limit)
		if err != nil {
			return nil, mapErr(err, "list-submissions-by-student")
		}
		return list, nil
	}
	return nil, status.Error(codes.InvalidArgument, "microtask_id or student_id is required")
}

func (h *Handler) Review(ctx context.Context, req *microtaskv1.ReviewRequest) (*microtaskv1.Submission, error) {
	sub, task, err := h.svc.Submissions.Review(ctx, req.GetSubmissionId(), req.GetStatus(), req.GetReviewComment())
	if err != nil {
		return nil, mapErr(err, "review")
	}
	log.Printf("Handlers: Review submission=%s status=%v", sub.GetId(), sub.GetStatus())
	if task != nil {
		// При approve задача перешла в COMPLETED — переиндексируем.
		h.search.IndexTask(ctx, task)

		// F5: автопополнение портфолио. Best-effort, ошибки не пробрасываем.
		h.achievements.CreateMicrotaskAchievement(
			ctx,
			sub.GetStudentId(),
			task.GetId(),
			task.GetTitle(),
			sub.GetSolutionUrl(),
			task.GetCompanyId(),
			req.GetReviewComment(),
		)
	}
	return sub, nil
}

func mapErr(err error, op string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, service.ErrInvalidArg):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrTaskNotFound):
		return status.Error(codes.NotFound, "microtask not found")
	case errors.Is(err, service.ErrSubmissionNF):
		return status.Error(codes.NotFound, "submission not found")
	case errors.Is(err, service.ErrTaskNotOpen):
		return status.Error(codes.FailedPrecondition, "microtask is not open")
	default:
		log.Printf("Handlers: %s failed: %v", op, err)
		return status.Error(codes.Internal, "internal error")
	}
}

func normalizePagination(p *commonv1.Pagination) (int32, int32) {
	page := int32(1)
	limit := int32(10)
	if p != nil {
		if p.GetPage() > 0 {
			page = p.GetPage()
		}
		if p.GetLimit() > 0 {
			limit = p.GetLimit()
		}
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
