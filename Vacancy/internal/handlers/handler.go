package handlers

import (
	"hh_for_students/vacancy-service/internal/searchclient"
	"hh_for_students/vacancy-service/internal/service"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
)

type VacancyHandler struct {
	vacancyv1.UnimplementedVacancyServiceServer
	service *service.Service
	search  *searchclient.Client
}

func NewVacancyHandler(service *service.Service, search *searchclient.Client) *VacancyHandler {
	return &VacancyHandler{
		service: service,
		search:  search,
	}
}
