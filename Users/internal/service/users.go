package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/google/uuid"
	"github.com/studjobs/hh_for_students/users/internal/repository"
)

type UsersService struct {
	repo *repository.Repository
}

func NewUsersService(repo *repository.Repository) *UsersService {
	return &UsersService{repo: repo}
}

func (s *UsersService) CreateProfile(ctx context.Context, profile *usersv1.Profile) (*usersv1.Profile, error) {
	log.Printf("Service: Creating new profile for email: %s", profile.Email)

	// Генерируем UUID если не указан
	if profile.Id == "" {
		profile.Id = uuid.New().String()
		log.Printf("Service: Generated UUID for new profile: %s", profile.Id)
	}

	log.Printf("Service: Creating profile in repository for email: %s", profile.Email)
	createdProfile, err := s.repo.Users.CreateProfile(ctx, profile)
	if err != nil {
		log.Printf("Service: Failed to create profile for email %s: %v", profile.Email, err)
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	log.Printf("Service: Profile created successfully with ID: %s", createdProfile.Id)
	return createdProfile, nil
}

func (s *UsersService) UpdateProfile(ctx context.Context, id string, profile *usersv1.Profile) (*usersv1.Profile, error) {
	log.Printf("Service: Updating profile with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return nil, fmt.Errorf("%w: invalid uuid format", ErrInvalidProfileData)
	}

	log.Printf("Service: Updating profile in repository for ID: %s", id)
	updatedProfile, err := s.repo.Users.UpdateProfile(ctx, id, profile)
	if err != nil {
		log.Printf("Service: Failed to update profile with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	log.Printf("Service: Profile updated successfully with ID: %s", updatedProfile.Id)
	return updatedProfile, nil
}

func (s *UsersService) DeleteProfile(ctx context.Context, id string) error {
	log.Printf("Service: Deleting profile with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return fmt.Errorf("%w: invalid uuid format", ErrInvalidProfileData)
	}

	log.Printf("Service: Deleting profile in repository for ID: %s", id)
	err := s.repo.Users.DeleteProfile(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			log.Printf("Service: Profile not found for deletion with ID: %s", id)
			return ErrProfileNotFound
		}
		log.Printf("Service: Failed to delete profile with ID %s: %v", id, err)
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	log.Printf("Service: Profile deleted successfully with ID: %s", id)
	return nil
}

func (s *UsersService) GetProfile(ctx context.Context, id string) (*usersv1.Profile, error) {
	log.Printf("Service: Getting profile with ID: %s", id)

	// Валидация UUID
	if _, err := uuid.Parse(id); err != nil {
		log.Printf("Service: Invalid UUID format for ID: %s", id)
		return nil, fmt.Errorf("%w: invalid uuid format", ErrInvalidProfileData)
	}

	log.Printf("Service: Getting profile from repository for ID: %s", id)
	profile, err := s.repo.Users.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			log.Printf("Service: Profile not found for ID: %s", id)
			return nil, ErrProfileNotFound
		}
		log.Printf("Service: Failed to get profile with ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	log.Printf("Service: Profile retrieved successfully with ID: %s", profile.Id)
	return profile, nil
}

func (s *UsersService) ListProfiles(ctx context.Context, professionCategory string, page, limit int32) (*usersv1.ProfileList, error) {
	log.Printf("Service: Listing profiles - page: %d, limit: %d, category: %s", page, limit, professionCategory)

	if page < 1 {
		page = 1
		log.Printf("Service: Page adjusted to default: %d", page)
	}
	if limit < 1 || limit > 100 {
		limit = 10
		log.Printf("Service: Limit adjusted to default: %d", limit)
	}

	log.Printf("Service: Getting profiles from repository")
	profiles, err := s.repo.Users.GetAllProfiles(ctx, professionCategory, page, limit)
	if err != nil {
		log.Printf("Service: Failed to list profiles: %v", err)
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	log.Printf("Service: Retrieved %d profiles successfully", len(profiles.Profiles))
	return profiles, nil
}
