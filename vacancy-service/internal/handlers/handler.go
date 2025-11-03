package handlers

import (
	"hh_for_students/vacancy-service/internal/service"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
)

type VacancyHandler struct {
	vacancyv1.UnimplementedVacancyServiceServer
	service *service.Service
}

func NewVacancyHandler(service *service.Service) *VacancyHandler {
	return &VacancyHandler{
		service: service,
	}
}
