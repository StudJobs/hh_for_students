package handlers

import (
	"github.com/studjobs/hh_for_students/users/internal/searchclient"
	"github.com/studjobs/hh_for_students/users/internal/service"
	"log"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
)

type UsersHandler struct {
	usersv1.UnimplementedUsersServiceServer
	service *service.Service
	search  *searchclient.Client
}

func NewUsersHandler(service *service.Service, search *searchclient.Client) *UsersHandler {
	log.Println("Handlers: Initializing UsersHandler")
	return &UsersHandler{
		service: service,
		search:  search,
	}
}
