package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrMembershipNotFound = errors.New("membership not found")

type MembershipRepository struct {
	db *pgxpool.Pool
}

func NewMembershipRepository(db *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{db: db}
}

const membershipCols = "id, company_id, user_id, status, note, created_at, reviewed_at"

func scanMember(row pgx.Row) (*companyv1.CompanyMember, error) {
	var m companyv1.CompanyMember
	var status int16
	var createdAt time.Time
	var reviewedAt sql.NullTime
	err := row.Scan(&m.Id, &m.CompanyId, &m.UserId, &status, &m.Note, &createdAt, &reviewedAt)
	if err != nil {
		return nil, err
	}
	m.Status = companyv1.MembershipStatus(status)
	m.CreatedAt = createdAt.Format(time.RFC3339)
	if reviewedAt.Valid {
		m.ReviewedAt = reviewedAt.Time.Format(time.RFC3339)
	}
	return &m, nil
}

// Apply создаёт PENDING-заявку. Идемпотентно по UNIQUE(company_id, user_id):
// повторный вызов возвращает уже существующую запись.
func (r *MembershipRepository) Apply(ctx context.Context, companyID, userID, note string) (*companyv1.CompanyMember, error) {
	query := `
INSERT INTO company_members (company_id, user_id, status, note)
VALUES ($1, $2, 1, $3)
ON CONFLICT (company_id, user_id) DO UPDATE SET note = EXCLUDED.note
RETURNING ` + membershipCols
	row := r.db.QueryRow(ctx, query, companyID, userID, note)
	m, err := scanMember(row)
	if err != nil {
		return nil, fmt.Errorf("apply membership: %w", err)
	}
	return m, nil
}

// Review меняет статус заявки на APPROVED/REJECTED.
func (r *MembershipRepository) Review(ctx context.Context, membershipID string, status companyv1.MembershipStatus) (*companyv1.CompanyMember, error) {
	query := `
UPDATE company_members
SET status = $2, reviewed_at = NOW()
WHERE id = $1
RETURNING ` + membershipCols
	row := r.db.QueryRow(ctx, query, membershipID, int16(status))
	m, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMembershipNotFound
	}
	return m, err
}

func (r *MembershipRepository) ListByCompany(ctx context.Context, companyID string, status companyv1.MembershipStatus) ([]*companyv1.CompanyMember, error) {
	q := "SELECT " + membershipCols + " FROM company_members WHERE company_id = $1"
	args := []interface{}{companyID}
	if status != companyv1.MembershipStatus_MEMBERSHIP_STATUS_UNSPECIFIED {
		q += " AND status = $2"
		args = append(args, int16(status))
	}
	q += " ORDER BY created_at DESC"
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()
	var out []*companyv1.CompanyMember
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// GetByUser возвращает membership пользователя: предпочтительно APPROVED;
// если APPROVED нет — возвращает самый свежий PENDING (чтобы HR видел статус заявки).
// REJECTED не возвращается.
func (r *MembershipRepository) GetByUser(ctx context.Context, userID string) (*companyv1.CompanyMember, error) {
	q := "SELECT " + membershipCols + ` FROM company_members
WHERE user_id = $1 AND status IN (1, 2)
ORDER BY (status = 2) DESC, created_at DESC
LIMIT 1`
	row := r.db.QueryRow(ctx, q, userID)
	m, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrMembershipNotFound
	}
	return m, err
}
