package grpc

import (
	"context"
	"log"
	"time"

	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients содержит все gRPC клиенты
type Clients struct {
	Auth  authv1.AuthServiceClient
	Users usersv1.UsersServiceClient

	// Добавьте другие клиенты по мере необходимости
	// Vacancy vacancyv1.VacancyServiceClient
	// Achievement achievementv1.AchievementServiceClient
}

// Config конфигурация для gRPC подключений
type Config struct {
	AuthAddress  string
	UsersAddress string
	Timeout      time.Duration
}

// NewClients создает и возвращает все gRPC клиенты
func NewClients(cfg Config) (*Clients, error) {
	log.Printf("Initializing gRPC clients...")

	// Создаем соединение с Auth сервисом
	authConn, err := createConnection(cfg.AuthAddress, cfg.Timeout)
	if err != nil {
		return nil, err
	}

	// Создаем соединение с Users сервисом
	//usersConn, err := createConnection(cfg.UsersAddress, cfg.Timeout)
	//if err != nil {
	//	authConn.Close()
	//	return nil, err
	//}

	clients := &Clients{
		Auth: authv1.NewAuthServiceClient(authConn),
		//Users: usersv1.NewUsersServiceClient(usersConn),
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
