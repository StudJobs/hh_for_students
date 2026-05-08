package clients

import (
	"fmt"
	"log"

	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Users      usersv1.UsersServiceClient
	Vacancy    vacancyv1.VacancyServiceClient
	MicroTasks microtaskv1.MicroTaskServiceClient

	usersConn      *grpc.ClientConn
	vacancyConn    *grpc.ClientConn
	microtasksConn *grpc.ClientConn
}

// New создаёт upstream-клиенты. microtasksAddr — необязательный (если пуст, MicroTasks-клиент = nil,
// reindex для микрозадач пропускается).
func New(usersAddr, vacancyAddr, microtasksAddr string) (*Clients, error) {
	uc, err := grpc.NewClient(usersAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("clients: dial users: %w", err)
	}
	vc, err := grpc.NewClient(vacancyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = uc.Close()
		return nil, fmt.Errorf("clients: dial vacancy: %w", err)
	}
	c := &Clients{
		Users:       usersv1.NewUsersServiceClient(uc),
		Vacancy:     vacancyv1.NewVacancyServiceClient(vc),
		usersConn:   uc,
		vacancyConn: vc,
	}
	if microtasksAddr != "" {
		mc, err := grpc.NewClient(microtasksAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("clients: dial microtasks (%s) failed: %v (microtask reindex disabled)", microtasksAddr, err)
		} else {
			c.MicroTasks = microtaskv1.NewMicroTaskServiceClient(mc)
			c.microtasksConn = mc
		}
	}
	return c, nil
}

func (c *Clients) Close() {
	if c.usersConn != nil {
		_ = c.usersConn.Close()
	}
	if c.vacancyConn != nil {
		_ = c.vacancyConn.Close()
	}
	if c.microtasksConn != nil {
		_ = c.microtasksConn.Close()
	}
}
