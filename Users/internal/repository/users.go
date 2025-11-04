package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Masterminds/squirrel"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
	PROFILE_TABLE      = "profiles"
)

type UsersRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewUsersRepository(db *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *UsersRepository) GetProfile(ctx context.Context, id string) (*usersv1.Profile, error) {
	log.Printf("Repository: Getting profile with ID: %s", id)

	query, args, err := r.sb.
		Select("id", "first_name", "last_name", "age", "tg", "resume_id", "avatar_id", "email", "description", "profession_category").
		From(PROFILE_TABLE).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build get profile query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var profile usersv1.Profile
	var resumeId, avatarId *string // Используем указатели для обработки NULL

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&profile.Id,
		&profile.FirstName,
		&profile.LastName,
		&profile.Age,
		&profile.Tg,
		&resumeId, // Сканируем в указатель
		&avatarId, // Сканируем в указатель
		&profile.Email,
		&profile.Description,
		&profile.ProfessionCategory,
	)
	if err != nil {
		log.Printf("Repository: Failed to get profile with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Обрабатываем NULL resume_id
	if resumeId != nil {
		profile.ResumeId = *resumeId
	}

	// Обрабатываем NULL avatar_id
	if avatarId != nil {
		profile.AvatarId = *avatarId
	}

	log.Printf("Repository: Successfully retrieved profile with ID: %s", id)
	return &profile, nil
}

func (r *UsersRepository) GetAllProfiles(ctx context.Context, professionCategory string, page, limit int32) (*usersv1.ProfileList, error) {
	log.Printf("Repository: Getting all profiles - page: %d, limit: %d, category: '%s'", page, limit, professionCategory)

	// Расчет offset
	offset := (page - 1) * limit

	// Базовый запрос
	queryBuilder := r.sb.
		Select("id", "first_name", "last_name", "age", "tg", "resume_id", "avatar_id", "email", "description", "profession_category").
		From(PROFILE_TABLE).
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	// Добавляем фильтр по категории если нужно
	if professionCategory != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"profession_category": professionCategory})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build get all profiles query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	log.Printf("Repository: Executing query: %s with args: %v", query, args)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to get profiles: %v", err)
		return nil, fmt.Errorf("failed to get profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*usersv1.Profile
	for rows.Next() {
		var profile usersv1.Profile
		var resumeId, avatarId *string // Используем указатели для обработки NULL

		err := rows.Scan(
			&profile.Id,
			&profile.FirstName,
			&profile.LastName,
			&profile.Age,
			&profile.Tg,
			&resumeId, // Сканируем в указатель
			&avatarId, // Сканируем в указатель
			&profile.Email,
			&profile.Description,
			&profile.ProfessionCategory,
		)
		if err != nil {
			log.Printf("Repository: Failed to scan profile row: %v", err)
			return nil, fmt.Errorf("failed to scan profile: %w", err)
		}

		// Обрабатываем NULL resume_id
		if resumeId != nil {
			profile.ResumeId = *resumeId
		}

		// Обрабатываем NULL avatar_id
		if avatarId != nil {
			profile.AvatarId = *avatarId
		}

		profiles = append(profiles, &profile)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Repository: Error iterating rows: %v", err)
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Get total count for pagination
	countBuilder := r.sb.
		Select("COUNT(*)").
		From(PROFILE_TABLE).
		Where("deleted_at IS NULL")

	if professionCategory != "" {
		countBuilder = countBuilder.Where(squirrel.Eq{"profession_category": professionCategory})
	}

	countQuery, countArgs, err := countBuilder.ToSql()
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

	// Calculate pages
	var pages int32
	if limit > 0 {
		pages = (totalCount + limit - 1) / limit
	}

	log.Printf("Repository: Retrieved %d profiles, total: %d, pages: %d", len(profiles), totalCount, pages)
	return &usersv1.ProfileList{
		Profiles: profiles,
		Pagination: &commonv1.PaginationResponse{
			Total:       totalCount,
			Pages:       pages,
			CurrentPage: page,
		},
	}, nil
}

func (r *UsersRepository) CreateProfile(ctx context.Context, profile *usersv1.Profile) (*usersv1.Profile, error) {
	log.Printf("Repository: Creating profile for email: %s", profile.Email)

	// Строим INSERT запрос
	insertBuilder := r.sb.
		Insert(PROFILE_TABLE).
		Columns("id", "first_name", "last_name", "age", "tg", "email", "description", "profession_category")

	// Добавляем resume_id и avatar_id только если они не пустые
	values := []interface{}{
		profile.Id,
		profile.FirstName,
		profile.LastName,
		profile.Age,
		profile.Tg,
		profile.Email,
		profile.Description,
		profile.ProfessionCategory,
	}

	if profile.ResumeId != "" {
		insertBuilder = insertBuilder.Columns("resume_id")
		values = append(values, profile.ResumeId)
	}

	if profile.AvatarId != "" {
		insertBuilder = insertBuilder.Columns("avatar_id")
		values = append(values, profile.AvatarId)
	}

	query, args, err := insertBuilder.
		Values(values...).
		Suffix("RETURNING id, first_name, last_name, age, tg, resume_id, avatar_id, email, description, profession_category").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build create profile query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var createdProfile usersv1.Profile
	var resumeId, avatarId *string // Используем указатели для обработки NULL

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&createdProfile.Id,
		&createdProfile.FirstName,
		&createdProfile.LastName,
		&createdProfile.Age,
		&createdProfile.Tg,
		&resumeId, // Сканируем в указатель
		&avatarId, // Сканируем в указатель
		&createdProfile.Email,
		&createdProfile.Description,
		&createdProfile.ProfessionCategory,
	)
	if err != nil {
		log.Printf("Repository: Failed to create profile for email %s: %v", profile.Email, err)
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Обрабатываем NULL resume_id
	if resumeId != nil {
		createdProfile.ResumeId = *resumeId
	}

	// Обрабатываем NULL avatar_id
	if avatarId != nil {
		createdProfile.AvatarId = *avatarId
	}

	log.Printf("Repository: Successfully created profile with ID: %s", createdProfile.Id)
	return &createdProfile, nil
}

func (r *UsersRepository) UpdateProfile(ctx context.Context, id string, profile *usersv1.Profile) (*usersv1.Profile, error) {
	log.Printf("Repository: PATCH updating profile with ID: %s", id)

	// Строим UPDATE запрос только для заполненных полей
	updateBuilder := r.sb.
		Update(PROFILE_TABLE).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		Set("updated_at", squirrel.Expr("NOW()"))

	// Добавляем только те поля, которые нужно обновить
	if profile.FirstName != "" {
		updateBuilder = updateBuilder.Set("first_name", profile.FirstName)
	}
	if profile.LastName != "" {
		updateBuilder = updateBuilder.Set("last_name", profile.LastName)
	}
	if profile.Age > 0 {
		updateBuilder = updateBuilder.Set("age", profile.Age)
	}
	if profile.Tg != "" {
		updateBuilder = updateBuilder.Set("tg", profile.Tg)
	}
	// resume_id может быть пустой строкой - это нормально
	if profile.ResumeId != "" {
		updateBuilder = updateBuilder.Set("resume_id", profile.ResumeId)
	}
	// avatar_id может быть пустой строкой - это нормально
	if profile.AvatarId != "" {
		updateBuilder = updateBuilder.Set("avatar_id", profile.AvatarId)
	}
	if profile.Email != "" {
		updateBuilder = updateBuilder.Set("email", profile.Email)
	}
	if profile.Description != "" {
		updateBuilder = updateBuilder.Set("description", profile.Description)
	}
	if profile.ProfessionCategory != "" {
		updateBuilder = updateBuilder.Set("profession_category", profile.ProfessionCategory)
	}

	query, args, err := updateBuilder.
		Suffix("RETURNING id, first_name, last_name, age, tg, resume_id, avatar_id, email, description, profession_category").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build update profile query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var updatedProfile usersv1.Profile
	var resumeId, avatarId *string // Используем указатели для обработки NULL

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&updatedProfile.Id,
		&updatedProfile.FirstName,
		&updatedProfile.LastName,
		&updatedProfile.Age,
		&updatedProfile.Tg,
		&resumeId, // Сканируем в указатель
		&avatarId, // Сканируем в указатель
		&updatedProfile.Email,
		&updatedProfile.Description,
		&updatedProfile.ProfessionCategory,
	)
	if err != nil {
		log.Printf("Repository: Failed to update profile with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Обрабатываем NULL resume_id
	if resumeId != nil {
		updatedProfile.ResumeId = *resumeId
	}

	// Обрабатываем NULL avatar_id
	if avatarId != nil {
		updatedProfile.AvatarId = *avatarId
	}

	log.Printf("Repository: Successfully PATCH updated profile with ID: %s", updatedProfile.Id)
	return &updatedProfile, nil
}

func (r *UsersRepository) DeleteProfile(ctx context.Context, id string) error {
	log.Printf("Repository: Deleting profile with ID: %s", id)

	query, args, err := r.sb.
		Update(PROFILE_TABLE).
		Set("deleted_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		log.Printf("Repository: Failed to build delete profile query: %v", err)
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Repository: Failed to delete profile with ID %s: %v", id, err)
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Repository: Profile not found for deletion with ID: %s", id)
		return ErrProfileNotFound
	}

	log.Printf("Repository: Successfully deleted profile with ID: %s, rows affected: %d", id, rowsAffected)
	return nil
}
