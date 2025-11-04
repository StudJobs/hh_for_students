package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	_ "github.com/studjobs/hh_for_students/api-gateway/docs"
	"github.com/studjobs/hh_for_students/api-gateway/internal/grpc"
	"github.com/studjobs/hh_for_students/api-gateway/internal/handlers"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
	"github.com/studjobs/hh_for_students/api-gateway/server"
)

// @title StudJobs HH API Gateway
// @version 1.0
// @description API Gateway для платформы StudJobs HH
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@studjobs.ru

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате: "Bearer {token}"

// @securityDefinitions.apikey RoleAuth
// @in header
// @name X-User-Role
// @description Роль пользователя
func main() {
	log.Printf("=== API Gateway Starting ===")

	// Загрузка конфигурации
	if err := initConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Конфигурация gRPC клиентов из config.yaml
	grpcConfig := grpc.Config{
		AuthAddress:            viper.GetString("grpc.auth_address"),
		UsersAddress:           viper.GetString("grpc.users_address"),
		UserAchievementAddress: viper.GetString("grpc.user_ach_address"),
		VacancyAddress:         viper.GetString("grpc.vacancy_address"),
		CompanyAddress:         viper.GetString("grpc.company_address"),
		Timeout:                10 * time.Second,
	}

	// Остальной код без изменений...
	clients, err := grpc.NewClients(grpcConfig)
	if err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}

	apiGateway := services.NewApiGateway(clients.Auth, clients.Users, clients.Achievement, clients.Company, clients.Vacancy)
	handler := handlers.NewHandler(apiGateway)
	app := handler.Init()

	srv := server.NewServer(app)
	serverPort := viper.GetString("server.port")

	go func() {
		log.Printf("=== Starting HTTP Server on port %s ===", serverPort)
		if err := srv.Run(serverPort); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("✓ API Gateway started successfully")
	waitForShutdownSignal(srv)
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func waitForShutdownSignal(srv *server.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Server stopped")
}
