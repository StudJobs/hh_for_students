package handlers

import (
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/auth/internal/service"
)

type AuthHandlers struct {
	authv1.UnimplementedAuthServiceServer
	service *service.Service
}

func NewAuthHandlers(service *service.Service) *AuthHandlers {
	return &AuthHandlers{
		service: service,
	}
}
