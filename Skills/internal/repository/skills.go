package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/studjobs/hh_for_students/skills/internal/models"
)

var sb = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type SkillsRepository struct {
	db *pgxpool.Pool
}

func NewSkillsRepository(db *pgxpool.Pool) *SkillsRepository {
	return &SkillsRepository{db: db}
}

func (r *SkillsRepository) Search(ctx context.Context, query string, category int32, limit int) ([]models.Skill, error) {
	q := sb.
		Select("id", "slug", "name", "category", "popularity", "created_at").
		From("skills").
		Where(squirrel.Eq{"deleted_at": nil}).
		OrderBy("popularity DESC", "name ASC").
		Limit(uint64(limit))

	if trimmed := strings.TrimSpace(query); trimmed != "" {
		pattern := "%" + strings.ToLower(trimmed) + "%"
		q = q.Where(squirrel.Or{
			squirrel.Expr("LOWER(slug) LIKE ?", pattern),
			squirrel.Expr("LOWER(name) LIKE ?", pattern),
		})
	}
	if category > 0 {
		q = q.Where(squirrel.Eq{"category": category})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build search query: %w", err)
	}
	return r.scanSkills(ctx, sql, args)
}

func (r *SkillsRepository) Popular(ctx context.Context, category int32, limit int) ([]models.Skill, error) {
	q := sb.
		Select("id", "slug", "name", "category", "popularity", "created_at").
		From("skills").
		Where(squirrel.Eq{"deleted_at": nil}).
		OrderBy("popularity DESC", "name ASC").
		Limit(uint64(limit))

	if category > 0 {
		q = q.Where(squirrel.Eq{"category": category})
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build popular query: %w", err)
	}
	return r.scanSkills(ctx, sql, args)
}

func (r *SkillsRepository) Bulk(ctx context.Context, slugs []string) ([]models.Skill, error) {
	if len(slugs) == 0 {
		return nil, nil
	}

	sql, args, err := sb.
		Select("id", "slug", "name", "category", "popularity", "created_at").
		From("skills").
		Where(squirrel.Eq{"slug": slugs}).
		Where(squirrel.Eq{"deleted_at": nil}).
		OrderBy("popularity DESC", "name ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build bulk query: %w", err)
	}
	return r.scanSkills(ctx, sql, args)
}

func (r *SkillsRepository) scanSkills(ctx context.Context, sql string, args []interface{}) ([]models.Skill, error) {
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query skills: %w", err)
	}
	defer rows.Close()

	var result []models.Skill
	for rows.Next() {
		var s models.Skill
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Category, &s.Popularity, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan skill row: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate skills: %w", err)
	}
	return result, nil
}
