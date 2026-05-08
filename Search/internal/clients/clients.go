package clients

import (
	"fmt"

	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Users   usersv1.UsersServiceClient
	Vacancy vacancyv1.VacancyServiceClient

	usersConn   *grpc.ClientConn
	vacancyConn *grpc.ClientConn
}

func New(usersAddr, vacancyAddr string) (*Clients, error) {
	uc, err := grpc.NewClient(usersAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("clients: dial users: %w", err)
	}
	vc, err := grpc.NewClient(vacancyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = uc.Close()
		return nil, fmt.Errorf("clients: dial vacancy: %w", err)
	}
	return &Clients{
		Users:       usersv1.NewUsersServiceClient(uc),
		Vacancy:     vacancyv1.NewVacancyServiceClient(vc),
		usersConn:   uc,
		vacancyConn: vc,
	}, nil
}

func (c *Clients) Close() {
	if c.usersConn != nil {
		_ = c.usersConn.Close()
	}
	if c.vacancyConn != nil {
		_ = c.vacancyConn.Close()
	}
}
