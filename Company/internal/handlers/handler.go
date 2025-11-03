package handlers

import (
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"github.com/studjobs/hh_for_students/company/internal/service"
	"log"
)

type CompanyHandlers struct {
	companyv1.UnimplementedCompanyServiceServer
	service *service.Service
}

func NewCompanyHandlers(service *service.Service) *CompanyHandlers {
	log.Println("Handlers: Initializing CompanyHandlers")
	return &CompanyHandlers{
		service: service,
	}
}
