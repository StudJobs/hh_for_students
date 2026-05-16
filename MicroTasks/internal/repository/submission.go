package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const submissionCols = "id, microtask_id, student_id, COALESCE(solution_url, ''), comment, status, review_comment, submitted_at, reviewed_at, solution_file_name"

type SubmissionRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewSubmissionRepository(db *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *SubmissionRepository) Create(ctx context.Context, s *microtaskv1.Submission) (*microtaskv1.Submission, error) {
	// solution_url теперь nullable — если пустое, передаём NULL, чтобы CHECK / индексы
	// потом могли отличать «нет ссылки» от «пустая строка».
	var solutionURL interface{}
	if u := s.GetSolutionUrl(); u != "" {
		solutionURL = u
	} else {
		solutionURL = nil
	}
	query, args, err := r.sb.
		Insert("microtask_submissions").
		Columns("microtask_id", "student_id", "solution_url", "comment", "status", "solution_file_name").
		Values(s.GetMicrotaskId(), s.GetStudentId(), solutionURL, s.GetComment(),
			int16(microtaskv1.SubmissionStatus_SUBMISSION_STATUS_PENDING), s.GetSolutionFileName()).
		Suffix("RETURNING " + submissionCols).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build create-submission query: %w", err)
	}
	return scanSubmission(r.db.QueryRow(ctx, query, args...))
}

func (r *SubmissionRepository) Get(ctx context.Context, id string) (*microtaskv1.Submission, error) {
	query, args, err := r.sb.
		Select(submissionCols).
		From("microtask_submissions").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get-submission query: %w", err)
	}
	s, err := scanSubmission(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSubmissionNotFound
	}
	return s, err
}

func (r *SubmissionRepository) ListByTask(ctx context.Context, taskID string, page, limit int32) (*microtaskv1.SubmissionList, error) {
	return r.list(ctx, squirrel.Eq{"microtask_id": taskID}, page, limit)
}

func (r *SubmissionRepository) ListByStudent(ctx context.Context, studentID string, page, limit int32) (*microtaskv1.SubmissionList, error) {
	return r.list(ctx, squirrel.Eq{"student_id": studentID}, page, limit)
}

func (r *SubmissionRepository) list(ctx context.Context, where squirrel.Eq, page, limit int32) (*microtaskv1.SubmissionList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query, args, err := r.sb.
		Select(submissionCols).
		From("microtask_submissions").
		Where(where).
		OrderBy("submitted_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list-submissions query: %w", err)
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list-submissions: %w", err)
	}
	defer rows.Close()

	var subs []*microtaskv1.Submission
	for rows.Next() {
		s, err := scanSubmission(rows)
		if err != nil {
			return nil, fmt.Errorf("scan submission: %w", err)
		}
		subs = append(subs, s)
	}

	countQuery, countArgs, err := r.sb.
		Select("COUNT(*)").
		From("microtask_submissions").
		Where(where).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build count-submissions query: %w", err)
	}
	var total int32
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count submissions: %w", err)
	}

	return &microtaskv1.SubmissionList{
		Submissions: subs,
		Pagination:  paginationResponse(total, limit, page),
	}, nil
}

func (r *SubmissionRepository) Review(ctx context.Context, id string, status microtaskv1.SubmissionStatus, reviewComment string) (*microtaskv1.Submission, error) {
	query, args, err := r.sb.
		Update("microtask_submissions").
		Set("status", int16(status)).
		Set("review_comment", reviewComment).
		Set("reviewed_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING " + submissionCols).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build review query: %w", err)
	}
	s, err := scanSubmission(r.db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSubmissionNotFound
	}
	return s, err
}

func scanSubmission(scanner interface {
	Scan(dest ...interface{}) error
}) (*microtaskv1.Submission, error) {
	var (
		s           microtaskv1.Submission
		statusInt   int16
		submittedAt time.Time
		reviewedAt  sql.NullTime
	)
	err := scanner.Scan(
		&s.Id,
		&s.MicrotaskId,
		&s.StudentId,
		&s.SolutionUrl,
		&s.Comment,
		&statusInt,
		&s.ReviewComment,
		&submittedAt,
		&reviewedAt,
		&s.SolutionFileName,
	)
	if err != nil {
		return nil, err
	}
	s.Status = microtaskv1.SubmissionStatus(statusInt)
	s.SubmittedAt = submittedAt.Format(time.RFC3339)
	if reviewedAt.Valid {
		s.ReviewedAt = reviewedAt.Time.Format(time.RFC3339)
	}
	return &s, nil
}
