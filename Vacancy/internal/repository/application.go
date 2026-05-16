package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	applicationv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/application/v1"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrApplicationNotFound = errors.New("application not found")
	APPLICATIONS_TABLE     = "vacancy_applications"
)

type Application interface {
	Apply(ctx context.Context, vacancyID, studentID, coverLetter string) (*applicationv1.Application, error)
	Withdraw(ctx context.Context, id, studentID string) error
	GetByID(ctx context.Context, id string) (*applicationv1.Application, error)
	ListByStudent(ctx context.Context, studentID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error)
	ListByVacancy(ctx context.Context, vacancyID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error)
	UpdateStatus(ctx context.Context, id string, status applicationv1.ApplicationStatus, hrComment string) (*applicationv1.Application, error)
	AssignHR(ctx context.Context, id, hrUserID string) (*applicationv1.Application, error)
}

type ApplicationRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Apply создаёт отклик или возвращает существующий активный (идемпотентность).
// Реализация ON CONFLICT не подходит: уникальный индекс — частичный (WHERE deleted_at IS NULL),
// а ON CONFLICT (vacancy_id, student_id) ловит конфликт по всему индексу. Делаем select-then-insert
// под транзакцией с serializable isolation: коротко и безопасно.
func (r *ApplicationRepository) Apply(ctx context.Context, vacancyID, studentID, coverLetter string) (*applicationv1.Application, error) {
	log.Printf("AppRepo: Apply vacancy=%s student=%s", vacancyID, studentID)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	selectQuery, selectArgs, err := r.sb.
		Select("id", "vacancy_id", "student_id", "cover_letter", "status",
			"hr_comment", "created_at", "updated_at", "COALESCE(hr_assignee_id::text, '')").
		From(APPLICATIONS_TABLE).
		Where(squirrel.Eq{"vacancy_id": vacancyID, "student_id": studentID}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select: %w", err)
	}

	app, err := scanApplicationRow(tx.QueryRow(ctx, selectQuery, selectArgs...))
	if err == nil {
		// Уже есть активный отклик — возвращаем его (идемпотентно).
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return nil, fmt.Errorf("commit: %w", commitErr)
		}
		return app, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("select existing: %w", err)
	}

	insertQuery, insertArgs, err := r.sb.
		Insert(APPLICATIONS_TABLE).
		Columns("vacancy_id", "student_id", "cover_letter").
		Values(vacancyID, studentID, coverLetter).
		Suffix("RETURNING id, vacancy_id, student_id, cover_letter, status, hr_comment, created_at, updated_at, COALESCE(hr_assignee_id::text, '')").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert: %w", err)
	}

	app, err = scanApplicationRow(tx.QueryRow(ctx, insertQuery, insertArgs...))
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return app, nil
}

func (r *ApplicationRepository) Withdraw(ctx context.Context, id, studentID string) error {
	log.Printf("AppRepo: Withdraw id=%s student=%s", id, studentID)

	query, args, err := r.sb.
		Update(APPLICATIONS_TABLE).
		Set("deleted_at", squirrel.Expr("NOW()")).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id, "student_id": studentID}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrApplicationNotFound
	}
	return nil
}

func (r *ApplicationRepository) GetByID(ctx context.Context, id string) (*applicationv1.Application, error) {
	query, args, err := r.sb.
		Select("id", "vacancy_id", "student_id", "cover_letter", "status",
			"hr_comment", "created_at", "updated_at", "COALESCE(hr_assignee_id::text, '')").
		From(APPLICATIONS_TABLE).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	app, err := scanApplicationRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("get: %w", err)
	}
	return app, nil
}

func (r *ApplicationRepository) ListByStudent(ctx context.Context, studentID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error) {
	return r.list(ctx, squirrel.Eq{"student_id": studentID}, status, page, limit)
}

func (r *ApplicationRepository) ListByVacancy(ctx context.Context, vacancyID string, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error) {
	return r.list(ctx, squirrel.Eq{"vacancy_id": vacancyID}, status, page, limit)
}

func (r *ApplicationRepository) list(ctx context.Context, filter squirrel.Sqlizer, status applicationv1.ApplicationStatus, page, limit int32) (*applicationv1.ApplicationList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	listQB := r.sb.
		Select("id", "vacancy_id", "student_id", "cover_letter", "status",
			"hr_comment", "created_at", "updated_at", "COALESCE(hr_assignee_id::text, '')").
		From(APPLICATIONS_TABLE).
		Where(filter).
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	if status != applicationv1.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED {
		listQB = listQB.Where(squirrel.Eq{"status": int(status)})
	}

	query, args, err := listQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var items []*applicationv1.Application
	for rows.Next() {
		app, err := scanApplicationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		items = append(items, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate: %w", err)
	}

	countQB := r.sb.
		Select("COUNT(*)").
		From(APPLICATIONS_TABLE).
		Where(filter).
		Where("deleted_at IS NULL")
	if status != applicationv1.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED {
		countQB = countQB.Where(squirrel.Eq{"status": int(status)})
	}
	countQuery, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build count: %w", err)
	}

	var total int32
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	pages := int32(0)
	if limit > 0 {
		pages = (total + limit - 1) / limit
	}

	return &applicationv1.ApplicationList{
		Applications: items,
		Pagination: &commonv1.PaginationResponse{
			Total:       total,
			Pages:       pages,
			CurrentPage: page,
		},
	}, nil
}

func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id string, status applicationv1.ApplicationStatus, hrComment string) (*applicationv1.Application, error) {
	log.Printf("AppRepo: UpdateStatus id=%s status=%d", id, status)

	if status == applicationv1.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED {
		return nil, fmt.Errorf("status is required")
	}

	query, args, err := r.sb.
		Update(APPLICATIONS_TABLE).
		Set("status", int(status)).
		Set("hr_comment", hrComment).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Suffix("RETURNING id, vacancy_id, student_id, cover_letter, status, hr_comment, created_at, updated_at, COALESCE(hr_assignee_id::text, '')").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update: %w", err)
	}

	app, err := scanApplicationRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("update: %w", err)
	}
	return app, nil
}

// AssignHR проставляет hr_assignee_id только если он ещё NULL (идемпотентно).
func (r *ApplicationRepository) AssignHR(ctx context.Context, id, hrUserID string) (*applicationv1.Application, error) {
	query := `
UPDATE vacancy_applications
SET hr_assignee_id = COALESCE(hr_assignee_id, $2::uuid),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, vacancy_id, student_id, cover_letter, status, hr_comment, created_at, updated_at, COALESCE(hr_assignee_id::text, '')`
	return scanApplicationRow(r.db.QueryRow(ctx, query, id, hrUserID))
}

func scanApplicationRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*applicationv1.Application, error) {
	var (
		app                  applicationv1.Application
		status               int16
		createdAt, updatedAt time.Time
		hrAssigneeID         string
	)
	if err := scanner.Scan(
		&app.Id,
		&app.VacancyId,
		&app.StudentId,
		&app.CoverLetter,
		&status,
		&app.HrComment,
		&createdAt,
		&updatedAt,
		&hrAssigneeID,
	); err != nil {
		return nil, err
	}
	app.Status = applicationv1.ApplicationStatus(status)
	app.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	app.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	app.HrAssigneeId = hrAssigneeID
	return &app, nil
}
