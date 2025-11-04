package grpc

import (
	"context"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"log"
	"time"

	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients содержит все gRPC клиенты
type Clients struct {
	Auth        authv1.AuthServiceClient
	Users       usersv1.UsersServiceClient
	Achievement achievementv1.AchievementServiceClient
	Company     companyv1.CompanyServiceClient
	Vacancy     vacancyv1.VacancyServiceClient
}

// Config конфигурация для gRPC подключений
type Config struct {
	AuthAddress            string
	UsersAddress           string
	UserAchievementAddress string
	VacancyAddress         string
	CompanyAddress         string
	Timeout                time.Duration
}

// NewClients создает и возвращает все gRPC клиенты
func NewClients(cfg Config) (*Clients, error) {
	log.Printf("Initializing gRPC clients...")

	//// Создаем соединение с Auth сервисом
	authConn, err := createConnection(cfg.AuthAddress, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	// Создаем соединение с Users сервисом
	usersConn, err := createConnection(cfg.UsersAddress, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	companyConn, err := createConnection(cfg.CompanyAddress, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	vacanyConn, err := createConnection(cfg.VacancyAddress, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	// Создаем соединение с Achievement сервисом
	achievementConn, err := createConnection(cfg.UserAchievementAddress, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	clients := &Clients{
		Auth:        authv1.NewAuthServiceClient(authConn),
		Users:       usersv1.NewUsersServiceClient(usersConn),
		Vacancy:     vacancyv1.NewVacancyServiceClient(vacanyConn),
		Company:     companyv1.NewCompanyServiceClient(companyConn),
		Achievement: achievementv1.NewAchievementServiceClient(achievementConn),
	}

	log.Printf("✓ All gRPC clients initialized successfully")
	return clients, nil
}

// createConnection создает gRPC соединение
func createConnection(address string, timeout time.Duration) (*grpc.ClientConn, error) {
	log.Printf("Connecting to gRPC service at: %s", address)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", address, err)
		return nil, err
	}

	log.Printf("✓ Connected to %s", address)
	return conn, nil
}

// Close закрывает все соединения (если нужно будет добавить graceful shutdown)
func (c *Clients) Close() {
	// В будущем можно добавить закрытие соединений
	log.Printf("Closing gRPC clients...")
}
