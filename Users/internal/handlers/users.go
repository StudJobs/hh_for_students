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
	h.search.IndexProfile(ctx, profile)
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
	h.search.IndexProfile(ctx, profile)
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
	h.search.DeleteProfile(ctx, req.Id)
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

func (h *UsersHandler) GetExpertiseTest(ctx context.Context, req *usersv1.GetExpertiseTestRequest) (*usersv1.ExpertiseTest, error) {
	out, err := h.service.User.GetExpertiseTest(ctx, req.GetSkillSlug())
	if err != nil {
		log.Printf("Handlers: GetExpertiseTest slug=%s failed: %v", req.GetSkillSlug(), err)
		return nil, status.Error(codes.InvalidArgument, "invalid skill slug")
	}
	return out, nil
}

func (h *UsersHandler) SubmitExpertiseTest(ctx context.Context, req *usersv1.SubmitExpertiseTestRequest) (*usersv1.SubmitExpertiseTestResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	out, err := h.service.User.SubmitExpertiseTest(ctx, req.GetUserId(), req.GetSkillSlug(), req.GetAnswerIndices())
	if err != nil {
		log.Printf("Handlers: SubmitExpertiseTest user=%s slug=%s failed: %v", req.GetUserId(), req.GetSkillSlug(), err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return out, nil
}

func (h *UsersHandler) AddVerifiedSkills(ctx context.Context, req *usersv1.AddVerifiedSkillsRequest) (*usersv1.Profile, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	p, err := h.service.User.AddVerifiedSkills(ctx, req.GetUserId(), req.GetSkillSlugs())
	if err != nil {
		log.Printf("Handlers: AddVerifiedSkills user=%s failed: %v", req.GetUserId(), err)
		return nil, status.Error(codes.Internal, "failed to add verified skills")
	}
	if h.search != nil {
		h.search.IndexProfile(ctx, p) // переиндексировать чтобы новые навыки попали в Search
	}
	return p, nil
}
