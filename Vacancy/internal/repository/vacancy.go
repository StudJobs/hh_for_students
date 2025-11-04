package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

type VacancyRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewVacancyRepository(db *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *VacancyRepository) GetVacancy(ctx context.Context, id string) (*vacancyv1.Vacancy, error) {
	log.Printf("Repository: Getting vacancy with ID: %s", id)

	query, args, err := r.buildVacancyQueryBuilder("", "", "", "", 0, 0, 0, 0, "").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build get vacancy query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	vacancy, err := scanVacancyRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		log.Printf("Repository: Failed to get vacancy with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get vacancy: %w", err)
	}

	log.Printf("Repository: Successfully retrieved vacancy with ID: %s", id)
	return vacancy, nil
}

func (r *VacancyRepository) GetAllVacancies(ctx context.Context, companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string, page, limit int32) (*vacancyv1.VacancyList, error) {

	log.Printf("Repository: Getting all vacancies - page: %d, limit: %d, company: '%s', status: '%s', work_format: '%s', schedule: '%s', salary: %d-%d, experience: %d-%d, search: '%s'",
		page, limit, companyID, positionStatus, workFormat, schedule, minSalary, maxSalary, minExperience, maxExperience, searchTitle)

	// Расчет offset
	offset := (page - 1) * limit

	query, args, err := r.buildVacancyQueryBuilder(companyID, positionStatus, workFormat, schedule,
		minSalary, maxSalary, minExperience, maxExperience, searchTitle).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build get all vacancies query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	log.Printf("Repository: Executing query: %s with args: %v", query, args)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to get vacancies: %v", err)
		return nil, fmt.Errorf("failed to get vacancies: %w", err)
	}
	defer rows.Close()

	var vacancies []*vacancyv1.Vacancy
	for rows.Next() {
		vacancy, err := scanVacancyRow(rows)
		if err != nil {
			log.Printf("Repository: Failed to scan vacancy row: %v", err)
			return nil, fmt.Errorf("failed to scan vacancy: %w", err)
		}
		vacancies = append(vacancies, vacancy)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Repository: Error iterating rows: %v", err)
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Get total count
	countQuery, countArgs, err := r.buildVacancyCountBuilder(companyID, positionStatus, workFormat, schedule,
		minSalary, maxSalary, minExperience, maxExperience, searchTitle).ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build count query: %v", err)
		return nil, fmt.Errorf("failed to build count query: %w", err)
	}

	var totalCount int32
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		log.Printf("Repository: Failed to get total count: %v", err)
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	log.Printf("Repository: Retrieved %d vacancies, total: %d", len(vacancies), totalCount)
	return &vacancyv1.VacancyList{
		Vacancies:  vacancies,
		Pagination: calculatePagination(totalCount, limit, page),
	}, nil
}

func (r *VacancyRepository) GetHRVacancies(ctx context.Context, companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string, page, limit int32) (*vacancyv1.VacancyList, error) {

	// Для HR вакансий используем ту же логику, но можно добавить дополнительные фильтры если нужно
	return r.GetAllVacancies(ctx, companyID, positionStatus, workFormat, schedule,
		minSalary, maxSalary, minExperience, maxExperience, searchTitle, page, limit)
}

func (r *VacancyRepository) CreateVacancy(ctx context.Context, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	log.Printf("Repository: Creating vacancy with title: %s", vacancy.Title)

	query, args, err := r.sb.
		Insert(VACANCY_TABLE).
		Columns("title", "experience", "salary", "position_status",
			"schedule", "work_format", "company_id", "attachment_id").
		Values(
			vacancy.Title,
			vacancy.Experience,
			vacancy.Salary,
			vacancy.PositionStatus,
			vacancy.Schedule,
			vacancy.WorkFormat,
			vacancy.CompanyId,
			vacancy.AttachmentId,
		).
		Suffix("RETURNING id, title, experience, salary, position_status, schedule, work_format, company_id, attachment_id, created_at").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build create vacancy query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	createdVacancy, err := scanVacancyRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		log.Printf("Repository: Failed to create vacancy with title %s: %v", vacancy.Title, err)
		return nil, fmt.Errorf("failed to create vacancy: %w", err)
	}

	log.Printf("Repository: Successfully created vacancy with ID: %s", createdVacancy.Id)
	return createdVacancy, nil
}

func (r *VacancyRepository) UpdateVacancy(ctx context.Context, id string, vacancy *vacancyv1.Vacancy) (*vacancyv1.Vacancy, error) {
	log.Printf("Repository: PATCH updating vacancy with ID: %s", id)

	// Строим UPDATE запрос только для заполненных полей
	updateBuilder := r.sb.
		Update(VACANCY_TABLE).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("updated_at", squirrel.Expr("NOW()"))

	// Добавляем только те поля, которые нужно обновить
	if vacancy.Title != "" {
		updateBuilder = updateBuilder.Set("title", vacancy.Title)
	}
	if vacancy.Experience > 0 {
		updateBuilder = updateBuilder.Set("experience", vacancy.Experience)
	}
	if vacancy.Salary > 0 {
		updateBuilder = updateBuilder.Set("salary", vacancy.Salary)
	}
	if vacancy.PositionStatus != "" {
		updateBuilder = updateBuilder.Set("position_status", vacancy.PositionStatus)
	}
	if vacancy.Schedule != "" {
		updateBuilder = updateBuilder.Set("schedule", vacancy.Schedule)
	}
	if vacancy.WorkFormat != "" {
		updateBuilder = updateBuilder.Set("work_format", vacancy.WorkFormat)
	}
	if vacancy.CompanyId != "" {
		updateBuilder = updateBuilder.Set("company_id", vacancy.CompanyId)
	}
	if vacancy.AttachmentId != "" {
		updateBuilder = updateBuilder.Set("attachment_id", vacancy.AttachmentId)
	}

	query, args, err := updateBuilder.
		Suffix("RETURNING id, title, experience, salary, position_status, schedule, work_format, company_id, attachment_id, created_at").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build update vacancy query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	updatedVacancy, err := scanVacancyRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		log.Printf("Repository: Failed to update vacancy with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to update vacancy: %w", err)
	}

	log.Printf("Repository: Successfully PATCH updated vacancy with ID: %s", updatedVacancy.Id)
	return updatedVacancy, nil
}

func (r *VacancyRepository) GetAllExistPositions(ctx context.Context) ([]string, error) {
	log.Printf("Repository: Getting all existing positions")

	query, args, err := r.sb.
		Select("DISTINCT position_status").
		From(VACANCY_TABLE).
		Where("deleted_at IS NULL").
		Where("position_status IS NOT NULL").
		Where("position_status != ''").
		OrderBy("position_status ASC").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build get positions query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to get positions: %v", err)
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	defer rows.Close()

	var positions []string
	for rows.Next() {
		var position string
		err := rows.Scan(&position)
		if err != nil {
			log.Printf("Repository: Failed to scan position row: %v", err)
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Repository: Error iterating position rows: %v", err)
		return nil, fmt.Errorf("error iterating position rows: %w", err)
	}

	log.Printf("Repository: Retrieved %d unique positions", len(positions))
	return positions, nil
}

func (r *VacancyRepository) DeleteVacancy(ctx context.Context, id string) error {
	log.Printf("Repository: Deleting vacancy with ID: %s", id)

	query, args, err := r.sb.
		Delete(VACANCY_TABLE).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build delete vacancy query: %v", err)
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to delete vacancy with ID %s: %v", id, err)
		return fmt.Errorf("failed to delete vacancy with ID %s: %w", id, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Repository: Vacancy not found for deletion with ID: %s", id)
		return ErrVacancyNotFound
	}

	log.Printf("Repository: Successfully deleted vacancy with ID: %s, rows affected: %d", id, rowsAffected)
	return nil
}

// scanVacancyRow сканирует строку из БД в Vacancy объект
func scanVacancyRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*vacancyv1.Vacancy, error) {
	var vacancy vacancyv1.Vacancy
	var workFormat sql.NullString
	var attachmentID sql.NullString
	var createdAt time.Time

	err := scanner.Scan(
		&vacancy.Id,
		&vacancy.Title,
		&vacancy.Experience,
		&vacancy.Salary,
		&vacancy.PositionStatus,
		&vacancy.Schedule,
		&workFormat,
		&vacancy.CompanyId,
		&attachmentID,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	// Обрабатываем nullable поля
	vacancy.WorkFormat = nullStringToString(workFormat)
	vacancy.AttachmentId = nullStringToString(attachmentID)
	vacancy.CreateAt = timeToString(createdAt)

	return &vacancy, nil
}

// nullStringToString конвертирует sql.NullString в string
func nullStringToString(nullStr sql.NullString) string {
	if nullStr.Valid {
		return nullStr.String
	}
	return ""
}

// timeToString конвертирует time.Time в RFC3339 строку
func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

// buildVacancyQueryBuilder создает базовый query builder для вакансий с фильтрами
func (r *VacancyRepository) buildVacancyQueryBuilder(companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string) squirrel.SelectBuilder {

	queryBuilder := r.sb.
		Select("id", "title", "experience", "salary", "position_status",
			"schedule", "work_format", "company_id", "attachment_id", "created_at").
		From(VACANCY_TABLE).
		Where("deleted_at IS NULL")

	if companyID != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"company_id": companyID})
	}
	if positionStatus != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"position_status": positionStatus})
	}
	if workFormat != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"work_format": workFormat})
	}
	if schedule != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"schedule": schedule})
	}
	if minSalary > 0 {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"salary": minSalary})
	}
	if maxSalary > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"salary": maxSalary})
	}
	if minExperience > 0 {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"experience": minExperience})
	}
	if maxExperience > 0 {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"experience": maxExperience})
	}
	if searchTitle != "" {
		queryBuilder = queryBuilder.Where(squirrel.ILike{"title": "%" + searchTitle + "%"})
	}

	return queryBuilder
}

// buildVacancyCountBuilder создает builder для подсчета вакансий с фильтрами
func (r *VacancyRepository) buildVacancyCountBuilder(companyID, positionStatus, workFormat, schedule string,
	minSalary, maxSalary, minExperience, maxExperience int32,
	searchTitle string) squirrel.SelectBuilder {

	countBuilder := r.sb.
		Select("COUNT(*)").
		From(VACANCY_TABLE).
		Where("deleted_at IS NULL")

	if companyID != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"company_id": companyID})
	}
	if positionStatus != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"position_status": positionStatus})
	}
	if workFormat != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"work_format": workFormat})
	}
	if schedule != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"schedule": schedule})
	}
	if minSalary > 0 {
		countBuilder = countBuilder.Where(squirrel.GtOrEq{"salary": minSalary})
	}
	if maxSalary > 0 {
		countBuilder = countBuilder.Where(squirrel.LtOrEq{"salary": maxSalary})
	}
	if minExperience > 0 {
		countBuilder = countBuilder.Where(squirrel.GtOrEq{"experience": minExperience})
	}
	if maxExperience > 0 {
		countBuilder = countBuilder.Where(squirrel.LtOrEq{"experience": maxExperience})
	}
	if searchTitle != "" {
		countBuilder = countBuilder.Where(squirrel.ILike{"title": "%" + searchTitle + "%"})
	}

	return countBuilder
}

// calculatePagination вычисляет пагинацию
func calculatePagination(totalCount, limit, currentPage int32) *commonv1.PaginationResponse {
	var pages int32
	if limit > 0 {
		pages = (totalCount + limit - 1) / limit
	}

	return &commonv1.PaginationResponse{
		Total:       totalCount,
		Pages:       pages,
		CurrentPage: currentPage,
	}
}
