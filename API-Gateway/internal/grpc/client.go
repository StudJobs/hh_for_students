package grpc

import (
	"context"
	applicationv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/application/v1"
	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
	skillsv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/skills/v1"
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
	Application applicationv1.ApplicationServiceClient
	Skills      skillsv1.SkillsServiceClient
	Search      searchv1.SearchServiceClient
	MicroTasks  microtaskv1.MicroTaskServiceClient
}

// Config конфигурация для gRPC подключений
type Config struct {
	AuthAddress            string
	UsersAddress           string
	UserAchievementAddress string
	VacancyAddress         string
	CompanyAddress         string
	SkillsAddress          string
	SearchAddress          string
	MicroTasksAddress      string
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
		// ApplicationService живёт на том же gRPC-сервере, что и VacancyService (порт 50054),
		// и обслуживается тем же подключением — экономим коннект.
		if conn := mustConn(cfg.VacancyAddress, cfg.Timeout); conn != nil {
			clients.Vacancy = vacancyv1.NewVacancyServiceClient(conn)
			clients.Application = applicationv1.NewApplicationServiceClient(conn)
		}
	}

	if cfg.UserAchievementAddress != "" {
		if conn := mustConn(cfg.UserAchievementAddress, cfg.Timeout); conn != nil {
			clients.Achievement = achievementv1.NewAchievementServiceClient(conn)
		}
	}

	if cfg.SkillsAddress != "" {
		if conn := mustConn(cfg.SkillsAddress, cfg.Timeout); conn != nil {
			clients.Skills = skillsv1.NewSkillsServiceClient(conn)
		}
	}

	if cfg.SearchAddress != "" {
		if conn := mustConn(cfg.SearchAddress, cfg.Timeout); conn != nil {
			clients.Search = searchv1.NewSearchServiceClient(conn)
		}
	}

	if cfg.MicroTasksAddress != "" {
		if conn := mustConn(cfg.MicroTasksAddress, cfg.Timeout); conn != nil {
			clients.MicroTasks = microtaskv1.NewMicroTaskServiceClient(conn)
		}
	}
	log.Println("✓ gRPC initialization completed (some services may be skipped)")

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
	}

	log.Printf("✓ Connected to %s", address)
	return conn
}

// Close закрывает все соединения (если нужно будет добавить graceful shutdown)
func (c *Clients) Close() {
	// В будущем можно добавить закрытие соединений
	log.Printf("Closing gRPC clients...")
}
