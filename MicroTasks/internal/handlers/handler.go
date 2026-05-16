package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/studjobs/hh_for_students/microtasks/internal/achievementclient"
	"github.com/studjobs/hh_for_students/microtasks/internal/searchclient"
	"github.com/studjobs/hh_for_students/microtasks/internal/service"
	"github.com/studjobs/hh_for_students/microtasks/internal/storage"
	"github.com/studjobs/hh_for_students/microtasks/internal/usersclient"
)

type Handler struct {
	microtaskv1.UnimplementedMicroTaskServiceServer

	svc          *service.Service
	search       *searchclient.Client
	achievements *achievementclient.Client
	users        *usersclient.Client
	solutions    *storage.Solutions
}

func New(svc *service.Service, search *searchclient.Client, achievements *achievementclient.Client, users *usersclient.Client, solutions *storage.Solutions) *Handler {
	return &Handler{svc: svc, search: search, achievements: achievements, users: users, solutions: solutions}
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

func (h *Handler) ListByStudent(ctx context.Context, req *microtaskv1.ListByStudentRequest) (*microtaskv1.MicroTaskList, error) {
	page, limit := normalizePagination(req.GetPagination())
	list, err := h.svc.Tasks.ListByStudent(ctx, req.GetStudentId(), req.GetStatus(), page, limit)
	if err != nil {
		return nil, mapErr(err, "list-by-student")
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
	s, err := h.svc.Submissions.Submit(ctx, req.GetMicrotaskId(), req.GetStudentId(), req.GetSolutionUrl(), req.GetComment(), req.GetSolutionFileName())
	if err != nil {
		return nil, mapErr(err, "submit")
	}
	log.Printf("Handlers: Submit microtask=%s student=%s submission=%s file=%q", s.GetMicrotaskId(), s.GetStudentId(), s.GetId(), s.GetSolutionFileName())
	return h.enrichSubmission(ctx, s), nil
}

func (h *Handler) SolutionUploadInit(ctx context.Context, req *microtaskv1.SolutionUploadInitRequest) (*microtaskv1.SolutionUploadInitResponse, error) {
	if req.GetMicrotaskId() == "" || req.GetStudentId() == "" || req.GetFileName() == "" {
		return nil, status.Error(codes.InvalidArgument, "microtask_id, student_id, file_name are required")
	}
	// Студент должен быть assigned_to этой задачи.
	t, err := h.svc.Tasks.Get(ctx, req.GetMicrotaskId())
	if err != nil {
		return nil, mapErr(err, "solution-init-get")
	}
	if t.GetAssignedTo() != req.GetStudentId() {
		return nil, status.Error(codes.PermissionDenied, "task is not assigned to this student")
	}
	if h.solutions == nil {
		return nil, status.Error(codes.Unavailable, "solutions storage is not configured")
	}
	rnd := make([]byte, 8)
	if _, err := rand.Read(rnd); err != nil {
		return nil, status.Error(codes.Internal, "rand failed")
	}
	fileID := hex.EncodeToString(rnd) + "-" + req.GetFileName()
	key := h.solutions.Key(req.GetMicrotaskId(), req.GetStudentId(), fileID)
	url, err := h.solutions.PresignedPut(ctx, key, 15*time.Minute)
	if err != nil {
		log.Printf("Handlers: SolutionUploadInit presign failed: %v", err)
		return nil, status.Error(codes.Internal, "presign failed")
	}
	return &microtaskv1.SolutionUploadInitResponse{FileId: fileID, UploadUrl: url}, nil
}

func (h *Handler) SolutionUploadConfirm(ctx context.Context, req *microtaskv1.SolutionUploadConfirmRequest) (*commonv1.Empty, error) {
	if req.GetMicrotaskId() == "" || req.GetStudentId() == "" || req.GetFileId() == "" {
		return nil, status.Error(codes.InvalidArgument, "microtask_id, student_id, file_id are required")
	}
	if h.solutions == nil {
		return nil, status.Error(codes.Unavailable, "solutions storage is not configured")
	}
	key := h.solutions.Key(req.GetMicrotaskId(), req.GetStudentId(), req.GetFileId())
	exists, err := h.solutions.Exists(ctx, key)
	if err != nil {
		log.Printf("Handlers: SolutionUploadConfirm Exists failed: %v", err)
		return nil, status.Error(codes.Internal, "head failed")
	}
	if !exists {
		return nil, status.Error(codes.FailedPrecondition, "file is not uploaded yet")
	}
	return &commonv1.Empty{}, nil
}

func (h *Handler) CreateSkillQuest(ctx context.Context, req *microtaskv1.CreateSkillQuestRequest) (*microtaskv1.MicroTask, error) {
	if req.GetExpertId() == "" || req.GetTargetStudentId() == "" || req.GetTargetSkillSlug() == "" || req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "expert_id, target_student_id, target_skill_slug, title are required")
	}
	t, err := h.svc.Tasks.CreateSkillQuest(ctx, req.GetExpertId(), req.GetTargetStudentId(), req.GetTargetSkillSlug(), req.GetTitle(), req.GetDescription(), req.GetDeadline())
	if err != nil {
		return nil, mapErr(err, "create-quest")
	}
	// Квесты НЕ индексируем в Search — они не в публичной выдаче.
	log.Printf("Handlers: CreateSkillQuest id=%s expert=%s student=%s skill=%s", t.GetId(), req.GetExpertId(), req.GetTargetStudentId(), req.GetTargetSkillSlug())
	return t, nil
}

func (h *Handler) ListSubmissions(ctx context.Context, req *microtaskv1.ListSubmissionsRequest) (*microtaskv1.SubmissionList, error) {
	page, limit := normalizePagination(req.GetPagination())
	var list *microtaskv1.SubmissionList
	var err error
	if mt := req.GetMicrotaskId(); mt != "" {
		list, err = h.svc.Submissions.ListByTask(ctx, mt, page, limit)
		if err != nil {
			return nil, mapErr(err, "list-submissions-by-task")
		}
	} else if st := req.GetStudentId(); st != "" {
		list, err = h.svc.Submissions.ListByStudent(ctx, st, page, limit)
		if err != nil {
			return nil, mapErr(err, "list-submissions-by-student")
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "microtask_id or student_id is required")
	}
	for i, s := range list.GetSubmissions() {
		list.Submissions[i] = h.enrichSubmission(ctx, s)
	}
	return list, nil
}

func (h *Handler) Review(ctx context.Context, req *microtaskv1.ReviewRequest) (*microtaskv1.Submission, error) {
	sub, task, err := h.svc.Submissions.Review(ctx, req.GetSubmissionId(), req.GetStatus(), req.GetReviewComment())
	if err != nil {
		return nil, mapErr(err, "review")
	}
	log.Printf("Handlers: Review submission=%s status=%v", sub.GetId(), sub.GetStatus())
	if task != nil {
		// При approve задача перешла в COMPLETED — переиндексируем (квесты не индексируем).
		if !task.GetIsSkillQuest() {
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

		// B4/C: добавить навыки задачи в verified_skill_slugs студента.
		// Для обычной микрозадачи — все её skill_slugs; для квеста — только target_skill_slug.
		var skills []string
		if task.GetIsSkillQuest() {
			if s := task.GetTargetSkillSlug(); s != "" {
				skills = []string{s}
			}
		} else {
			skills = task.GetSkillSlugs()
		}
		if len(skills) > 0 {
			h.users.AddVerifiedSkills(ctx, sub.GetStudentId(), skills)
		}
	}
	return sub, nil
}

// enrichSubmission добавляет presigned GET URL для solution_file_name (если задан).
// Срок жизни URL — 15 минут, достаточно для UI-сессии.
func (h *Handler) enrichSubmission(ctx context.Context, s *microtaskv1.Submission) *microtaskv1.Submission {
	if s == nil || h.solutions == nil || s.GetSolutionFileName() == "" {
		return s
	}
	key := h.solutions.Key(s.GetMicrotaskId(), s.GetStudentId(), s.GetSolutionFileName())
	url, err := h.solutions.PresignedGet(ctx, key, 15*time.Minute)
	if err == nil {
		s.SolutionFileUrl = url
	}
	return s
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
