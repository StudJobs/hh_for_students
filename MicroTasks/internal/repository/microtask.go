package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const taskCols = "id, company_id, title, description, reward, deadline, skill_slugs, status, assigned_to, created_at, updated_at"

type MicroTaskRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewMicroTaskRepository(db *pgxpool.Pool) *MicroTaskRepository {
	return &MicroTaskRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *MicroTaskRepository) Create(ctx context.Context, t *microtaskv1.MicroTask) (*microtaskv1.MicroTask, error) {
	cols := []string{"company_id", "title", "description", "reward", "skill_slugs", "status"}
	vals := []interface{}{
		t.GetCompanyId(),
		t.GetTitle(),
		t.GetDescription(),
		t.GetReward(),
		stringSlice(t.GetSkillSlugs()),
		statusToInt(t.GetStatus(), int16(microtaskv1.MicroTaskStatus_MICROTASK_STATUS_OPEN)),
	}
	if d := t.GetDeadline(); d != "" {
		cols = append(cols, "deadline")
		vals = append(vals, d)
	}

	query, args, err := r.sb.
		Insert("microtasks").
		Columns(cols...).
		Values(vals...).
		Suffix("RETURNING " + taskCols).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build create query: %w", err)
	}

	return scanTask(r.db.QueryRow(ctx, query, args...))
}

func (r *MicroTaskRepository) Update(ctx context.Context, id string, t *microtaskv1.MicroTask) (*microtaskv1.MicroTask, error) {
	ub := r.sb.Update("microtasks").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("updated_at", squirrel.Expr("NOW()"))

	if t.GetTitle() != "" {
		ub = ub.Set("title", t.GetTitle())
	}
	if t.GetDescription() != "" {
		ub = ub.Set("description", t.GetDescription())
	}
	if t.GetReward() > 0 {
		ub = ub.Set("reward", t.GetReward())
	}
	if t.GetDeadline() != "" {
		ub = ub.Set("deadline", t.GetDeadline())
	}
	if len(t.GetSkillSlugs()) > 0 {
		ub = ub.Set("skill_slugs", stringSlice(t.GetSkillSlugs()))
	}
	if t.GetStatus() != microtaskv1.MicroTaskStatus_MICROTASK_STATUS_UNSPECIFIED {
		ub = ub.Set("status", int16(t.GetStatus()))
	}

	query, args, err := ub.Suffix("RETURNING " + taskCols).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update query: %w", err)
	}

	task, err := scanTask(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	return task, err
}

func (r *MicroTaskRepository) Delete(ctx context.Context, id string) error {
	query, args, err := r.sb.
		Update("microtasks").
		Set("deleted_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *MicroTaskRepository) Get(ctx context.Context, id string) (*microtaskv1.MicroTask, error) {
	query, args, err := r.sb.
		Select(taskCols).
		From("microtasks").
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get query: %w", err)
	}
	t, err := scanTask(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	return t, err
}

func (r *MicroTaskRepository) List(ctx context.Context, status microtaskv1.MicroTaskStatus, skillSlugs []string, page, limit int32) (*microtaskv1.MicroTaskList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	qb := r.sb.
		Select(taskCols).
		From("microtasks").
		Where("deleted_at IS NULL")
	cb := r.sb.
		Select("COUNT(*)").
		From("microtasks").
		Where("deleted_at IS NULL")

	if status != microtaskv1.MicroTaskStatus_MICROTASK_STATUS_UNSPECIFIED {
		qb = qb.Where(squirrel.Eq{"status": int16(status)})
		cb = cb.Where(squirrel.Eq{"status": int16(status)})
	}
	if len(skillSlugs) > 0 {
		qb = qb.Where("skill_slugs @> ?", stringSlice(skillSlugs))
		cb = cb.Where("skill_slugs @> ?", stringSlice(skillSlugs))
	}

	query, args, err := qb.OrderBy("created_at DESC").Limit(uint64(limit)).Offset(uint64(offset)).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list query: %w", err)
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var tasks []*microtaskv1.MicroTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan list row: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iter list rows: %w", err)
	}

	countQuery, countArgs, err := cb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build count query: %w", err)
	}
	var total int32
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	return &microtaskv1.MicroTaskList{
		Tasks:      tasks,
		Pagination: paginationResponse(total, limit, page),
	}, nil
}

func (r *MicroTaskRepository) ListByCompany(ctx context.Context, companyID string, page, limit int32) (*microtaskv1.MicroTaskList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query, args, err := r.sb.
		Select(taskCols).
		From("microtasks").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"company_id": companyID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list-by-company query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list-by-company: %w", err)
	}
	defer rows.Close()

	var tasks []*microtaskv1.MicroTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tasks = append(tasks, t)
	}

	countQuery, countArgs, err := r.sb.
		Select("COUNT(*)").
		From("microtasks").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"company_id": companyID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build count query: %w", err)
	}
	var total int32
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	return &microtaskv1.MicroTaskList{
		Tasks:      tasks,
		Pagination: paginationResponse(total, limit, page),
	}, nil
}

func (r *MicroTaskRepository) ListByStudent(ctx context.Context, studentID string, status microtaskv1.MicroTaskStatus, page, limit int32) (*microtaskv1.MicroTaskList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	qb := r.sb.
		Select(taskCols).
		From("microtasks").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"assigned_to": studentID})
	cb := r.sb.
		Select("COUNT(*)").
		From("microtasks").
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"assigned_to": studentID})

	if status != microtaskv1.MicroTaskStatus_MICROTASK_STATUS_UNSPECIFIED {
		qb = qb.Where(squirrel.Eq{"status": int16(status)})
		cb = cb.Where(squirrel.Eq{"status": int16(status)})
	}

	query, args, err := qb.OrderBy("updated_at DESC").Limit(uint64(limit)).Offset(uint64(offset)).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list-by-student query: %w", err)
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list-by-student: %w", err)
	}
	defer rows.Close()

	var tasks []*microtaskv1.MicroTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		tasks = append(tasks, t)
	}

	countQuery, countArgs, err := cb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build count query: %w", err)
	}
	var total int32
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	return &microtaskv1.MicroTaskList{
		Tasks:      tasks,
		Pagination: paginationResponse(total, limit, page),
	}, nil
}

// Apply переводит задачу из OPEN в ASSIGNED атомарно. Возвращает ErrTaskNotOpen, если статус не OPEN.
func (r *MicroTaskRepository) Apply(ctx context.Context, taskID, studentID string) (*microtaskv1.MicroTask, error) {
	query, args, err := r.sb.
		Update("microtasks").
		Set("status", int16(microtaskv1.MicroTaskStatus_MICROTASK_STATUS_ASSIGNED)).
		Set("assigned_to", studentID).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": taskID}).
		Where("deleted_at IS NULL").
		Where(squirrel.Eq{"status": int16(microtaskv1.MicroTaskStatus_MICROTASK_STATUS_OPEN)}).
		Suffix("RETURNING " + taskCols).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build apply query: %w", err)
	}
	t, err := scanTask(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		// Либо задачи нет, либо она не в статусе OPEN. Уточняем.
		existing, getErr := r.Get(ctx, taskID)
		if getErr != nil {
			return nil, ErrTaskNotFound
		}
		if existing.GetStatus() != microtaskv1.MicroTaskStatus_MICROTASK_STATUS_OPEN {
			return nil, ErrTaskNotOpen
		}
		return nil, ErrAlreadyAssigned
	}
	return t, err
}

func (r *MicroTaskRepository) SetStatus(ctx context.Context, id string, status microtaskv1.MicroTaskStatus) (*microtaskv1.MicroTask, error) {
	query, args, err := r.sb.
		Update("microtasks").
		Set("status", int16(status)).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Suffix("RETURNING " + taskCols).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build set-status query: %w", err)
	}
	t, err := scanTask(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	return t, err
}

func scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*microtaskv1.MicroTask, error) {
	var (
		t           microtaskv1.MicroTask
		deadline    sql.NullTime
		assignedTo  sql.NullString
		statusInt   int16
		createdAt   time.Time
		updatedAt   time.Time
		skillSlugs  []string
	)
	err := scanner.Scan(
		&t.Id,
		&t.CompanyId,
		&t.Title,
		&t.Description,
		&t.Reward,
		&deadline,
		&skillSlugs,
		&statusInt,
		&assignedTo,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.SkillSlugs = skillSlugs
	t.Status = microtaskv1.MicroTaskStatus(statusInt)
	if deadline.Valid {
		t.Deadline = deadline.Time.Format("2006-01-02")
	}
	if assignedTo.Valid {
		t.AssignedTo = assignedTo.String
	}
	t.CreatedAt = createdAt.Format(time.RFC3339)
	t.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &t, nil
}

func stringSlice(s []string) interface{} {
	if s == nil {
		return []string{}
	}
	return s
}

func statusToInt(s microtaskv1.MicroTaskStatus, fallback int16) int16 {
	if s == microtaskv1.MicroTaskStatus_MICROTASK_STATUS_UNSPECIFIED {
		return fallback
	}
	return int16(s)
}

func paginationResponse(total, limit, page int32) *commonv1.PaginationResponse {
	pages := int32(0)
	if limit > 0 {
		pages = (total + limit - 1) / limit
	}
	return &commonv1.PaginationResponse{
		Total:       total,
		Pages:       pages,
		CurrentPage: page,
	}
}
