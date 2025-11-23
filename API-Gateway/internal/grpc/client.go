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

// NewClients создаёт gRPC клиентов, пропуская отсутствующие адреса.
func NewClients(cfg Config) (*Clients, error) {
	log.Printf("Initializing gRPC clients...")

	clients := &Clients{}

	// подключаем каждую зависимость только если адрес указан
	if cfg.AuthAddress != "" {
		if conn := mustConn(cfg.AuthAddress, cfg.Timeout); conn != nil {
			clients.Auth = authv1.NewAuthServiceClient(conn)
		}
	}

	if cfg.UsersAddress != "" {
		if conn := mustConn(cfg.UsersAddress, cfg.Timeout); conn != nil {
			clients.Users = usersv1.NewUsersServiceClient(conn)
		}
	}

	if cfg.CompanyAddress != "" {
		if conn := mustConn(cfg.CompanyAddress, cfg.Timeout); conn != nil {
			clients.Company = companyv1.NewCompanyServiceClient(conn)
		}
	}

	if cfg.VacancyAddress != "" {
		if conn := mustConn(cfg.VacancyAddress, cfg.Timeout); conn != nil {
			clients.Vacancy = vacancyv1.NewVacancyServiceClient(conn)
		}
	}

	if cfg.UserAchievementAddress != "" {
		if conn := mustConn(cfg.UserAchievementAddress, cfg.Timeout); conn != nil {
			clients.Achievement = achievementv1.NewAchievementServiceClient(conn)
		}
	}
	cfg.Timeout = 30 * time.Second,
	log.Printf("✓ gRPC initialization completed (some services may be skipped)")
	return clients, nil
}

// mustConn — безопасное подключение: если адрес пустой или ошибка — вернёт nil.
func mustConn(address string, timeout time.Duration) *grpc.ClientConn {
	log.Printf("Connecting to gRPC service: %s", address)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
            log.Printf("⏰ Timeout connecting to %s: %v (skipping)", address, err)
        } else {
            log.Printf("⚠ Failed to connect to %s: %v (skipping)", address, err)
        }
        return nil
		log.Printf("⚠ Failed to connect to %s: %v (skipping)", address, err)
		return nil
	}

	log.Printf("✓ Connected to %s", address)
	return conn
}

// Close закрывает все соединения (если нужно будет добавить graceful shutdown)
func (c *Clients) Close() {
	// В будущем можно добавить закрытие соединений
	log.Printf("Closing gRPC clients...")
}
