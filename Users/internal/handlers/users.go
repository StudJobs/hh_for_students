package handlers

import (
	"context"
	commonv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/common/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/studjobs/hh_for_students/users/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func (h *UsersHandler) NewProfile(ctx context.Context, req *usersv1.NewProfileRequest) (*usersv1.Profile, error) {
	log.Printf("Handlers: NewProfile request received for email: %s", req.Profile.GetEmail())

	if req.Profile == nil {
		log.Printf("Handlers: NewProfile failed - profile is required")
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}

	profile, err := h.service.User.CreateProfile(ctx, req.Profile)
	if err != nil {
		log.Printf("Handlers: NewProfile failed for email %s: %v", req.Profile.GetEmail(), err)
		switch err {
		case service.ErrInvalidProfileData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create profile")
		}
	}

	log.Printf("Handlers: NewProfile completed successfully for ID: %s", profile.Id)
	return profile, nil
}

func (h *UsersHandler) UpdateProfile(ctx context.Context, req *usersv1.UpdateProfileRequest) (*usersv1.Profile, error) {
	log.Printf("Handlers: UpdateProfile request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: UpdateProfile failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Profile == nil {
		log.Printf("Handlers: UpdateProfile failed - profile is required")
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}

	profile, err := h.service.User.UpdateProfile(ctx, req.Id, req.Profile)
	if err != nil {
		log.Printf("Handlers: UpdateProfile failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrProfileNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidProfileData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to update profile")
		}
	}

	log.Printf("Handlers: UpdateProfile completed successfully for ID: %s", profile.Id)
	return profile, nil
}

func (h *UsersHandler) DeleteProfile(ctx context.Context, req *usersv1.DeleteProfileRequest) (*commonv1.Empty, error) {
	log.Printf("Handlers: DeleteProfile request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: DeleteProfile failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := h.service.User.DeleteProfile(ctx, req.Id)
	if err != nil {
		log.Printf("Handlers: DeleteProfile failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrProfileNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidProfileData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to delete profile")
		}
	}

	log.Printf("Handlers: DeleteProfile completed successfully for ID: %s", req.Id)
	return &commonv1.Empty{}, nil
}

func (h *UsersHandler) GetProfile(ctx context.Context, req *usersv1.GetProfileRequest) (*usersv1.Profile, error) {
	log.Printf("Handlers: GetProfile request received for ID: %s", req.Id)

	if req.Id == "" {
		log.Printf("Handlers: GetProfile failed - id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	profile, err := h.service.User.GetProfile(ctx, req.Id)
	if err != nil {
		log.Printf("Handlers: GetProfile failed for ID %s: %v", req.Id, err)
		switch err {
		case service.ErrProfileNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidProfileData:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to get profile")
		}
	}

	log.Printf("Handlers: GetProfile completed successfully for ID: %s", profile.Id)
	return profile, nil
}

func (h *UsersHandler) GetAllProfiles(ctx context.Context, req *usersv1.GetAllProfilesRequest) (*usersv1.ProfileList, error) {
	log.Printf("Handlers: GetAllProfiles request received")

	var page, limit int32 = 1, 10

	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
	}

	log.Printf("Handlers: GetAllProfiles - page: %d, limit: %d, category: %s, role: %s",
		page, limit, req.GetProfessionCategory(), req.GetRole())

	profiles, err := h.service.User.ListProfiles(ctx, req.ProfessionCategory, page, limit, req.Role)
	if err != nil {
		log.Printf("Handlers: GetAllProfiles failed: %v", err)
		return nil, status.Error(codes.Internal, "failed to get profiles")
	}

	log.Printf("Handlers: GetAllProfiles completed successfully, returned %d profiles", len(profiles.Profiles))
	return profiles, nil
}
