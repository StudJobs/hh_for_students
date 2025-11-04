package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrCompanyNotFound = errors.New("company not found")
	COMPANY_TABLE      = "companies"
)

type CompanyRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewCompanyRepository(db *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// GetCompany возвращает компанию по ID
func (r *CompanyRepository) GetCompany(ctx context.Context, id string) (*companyv1.Company, error) {
	log.Printf("Repository: Getting company with ID: %s", id)

	query, args, err := r.buildCompanyQueryBuilder("", "").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build get company query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	company, err := scanCompanyRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		log.Printf("Repository: Failed to get company with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	log.Printf("Repository: Successfully retrieved company with ID: %s", id)
	return company, nil
}

// GetAllCompanies возвращает список компаний с фильтрацией и пагинацией
func (r *CompanyRepository) GetAllCompanies(ctx context.Context, city, companyType string, page, limit int32) (*companyv1.CompanyList, error) {
	log.Printf("Repository: Getting all companies - page: %d, limit: %d, city: '%s', type: '%s'",
		page, limit, city, companyType)

	// Расчет offset
	offset := (page - 1) * limit

	query, args, err := r.buildCompanyQueryBuilder(city, companyType).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build get all companies query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	log.Printf("Repository: Executing query: %s with args: %v", query, args)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to get companies: %v", err)
		return nil, fmt.Errorf("failed to get companies: %w", err)
	}
	defer rows.Close()

	var companies []*companyv1.Company
	for rows.Next() {
		company, err := scanCompanyRow(rows)
		if err != nil {
			log.Printf("Repository: Failed to scan company row: %v", err)
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Repository: Error iterating rows: %v", err)
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Получаем общее количество
	countQuery, countArgs, err := r.buildCompanyCountBuilder(city, companyType).ToSql()
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

	log.Printf("Repository: Retrieved %d companies, total: %d", len(companies), totalCount)
	return &companyv1.CompanyList{
		Companies:  companies,
		Pagination: calculatePagination(totalCount, limit, page),
	}, nil
}

// CreateCompany создает новую компанию
func (r *CompanyRepository) CreateCompany(ctx context.Context, company *companyv1.Company) (*companyv1.Company, error) {
	log.Printf("Repository: Creating company with name: %s", company.Name)

	query, args, err := r.sb.
		Insert(COMPANY_TABLE).
		Columns("id", "name", "description", "city", "site", "company_type", "logo_id").
		Values(
			company.Id,
			company.Name,
			company.Description,
			company.City,
			company.Site,
			getCompanyTypeValue(company.Type),
			company.LogoId,
		).
		Suffix("RETURNING id, name, description, city, site, company_type, logo_id, created_at").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build create company query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	createdCompany, err := scanCompanyRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		log.Printf("Repository: Failed to create company with name %s: %v", company.Name, err)
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	log.Printf("Repository: Successfully created company with ID: %s", createdCompany.Id)
	return createdCompany, nil
}

// UpdateCompany обновляет компанию (PATCH)
func (r *CompanyRepository) UpdateCompany(ctx context.Context, id string, company *companyv1.Company) (*companyv1.Company, error) {
	log.Printf("Repository: PATCH updating company with ID: %s", id)

	// Строим UPDATE запрос только для заполненных полей
	updateBuilder := r.sb.
		Update(COMPANY_TABLE).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("updated_at", squirrel.Expr("NOW()"))

	// Добавляем только те поля, которые нужно обновить
	if company.Name != "" {
		updateBuilder = updateBuilder.Set("name", company.Name)
	}
	if company.Description != "" {
		updateBuilder = updateBuilder.Set("description", company.Description)
	}
	if company.City != "" {
		updateBuilder = updateBuilder.Set("city", company.City)
	}
	if company.Site != "" {
		updateBuilder = updateBuilder.Set("site", company.Site)
	}
	if company.Type != nil && company.Type.Value != "" {
		updateBuilder = updateBuilder.Set("company_type", company.Type.Value)
	}
	// Обновляем logo_id если он передан
	if company.LogoId != "" {
		updateBuilder = updateBuilder.Set("logo_id", company.LogoId)
	}

	query, args, err := updateBuilder.
		Suffix("RETURNING id, name, description, city, site, company_type, logo_id, created_at").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build update company query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	updatedCompany, err := scanCompanyRow(r.db.QueryRow(ctx, query, args...))
	if err != nil {
		log.Printf("Repository: Failed to update company with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	log.Printf("Repository: Successfully PATCH updated company with ID: %s", updatedCompany.Id)
	return updatedCompany, nil
}

// DeleteCompany удаляет компанию (soft delete)
func (r *CompanyRepository) DeleteCompany(ctx context.Context, id string) error {
	log.Printf("Repository: Deleting company with ID: %s", id)

	query, args, err := r.sb.
		Update(COMPANY_TABLE).
		Set("deleted_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build delete company query: %v", err)
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to delete company with ID %s: %v", id, err)
		return fmt.Errorf("failed to delete company: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Repository: Company not found for deletion with ID: %s", id)
		return ErrCompanyNotFound
	}

	log.Printf("Repository: Successfully deleted company with ID: %s, rows affected: %d", id, rowsAffected)
	return nil
}

// Вспомогательные приватные функции

// scanCompanyRow сканирует строку из БД в Company объект
func scanCompanyRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*companyv1.Company, error) {
	var company companyv1.Company
	var description, city, site, companyType, logoId sql.NullString
	var createdAt time.Time

	err := scanner.Scan(
		&company.Id,
		&company.Name,
		&description,
		&city,
		&site,
		&companyType,
		&logoId,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	// Обрабатываем nullable поля
	company.Description = nullStringToString(description)
	company.City = nullStringToString(city)
	company.Site = nullStringToString(site)
	company.LogoId = nullStringToString(logoId)

	// Обрабатываем company_type
	if companyType.Valid {
		company.Type = &companyv1.CompanyType{Value: companyType.String}
	}

	return &company, nil
}

// getCompanyTypeValue возвращает значение company_type
func getCompanyTypeValue(companyType *companyv1.CompanyType) string {
	if companyType != nil {
		return companyType.Value
	}
	return ""
}

// nullStringToString конвертирует sql.NullString в string
func nullStringToString(nullStr sql.NullString) string {
	if nullStr.Valid {
		return nullStr.String
	}
	return ""
}

// buildCompanyQueryBuilder создает базовый query builder для компаний
func (r *CompanyRepository) buildCompanyQueryBuilder(city, companyType string) squirrel.SelectBuilder {
	queryBuilder := r.sb.
		Select("id", "name", "description", "city", "site", "company_type", "logo_id", "created_at").
		From(COMPANY_TABLE).
		Where("deleted_at IS NULL")

	if city != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"city": city})
	}
	if companyType != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"company_type": companyType})
	}

	return queryBuilder
}

// buildCompanyCountBuilder создает builder для подсчета компаний
func (r *CompanyRepository) buildCompanyCountBuilder(city, companyType string) squirrel.SelectBuilder {
	countBuilder := r.sb.
		Select("COUNT(*)").
		From(COMPANY_TABLE).
		Where("deleted_at IS NULL")

	if city != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"city": city})
	}
	if companyType != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"company_type": companyType})
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
