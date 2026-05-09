package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/studjobs/hh_for_students/skills/internal/handlers"
	"github.com/studjobs/hh_for_students/skills/internal/metrics"
	"github.com/studjobs/hh_for_students/skills/internal/repository"
	"github.com/studjobs/hh_for_students/skills/internal/service"
	"github.com/studjobs/hh_for_students/skills/server"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("warning: error loading .env file: %s", err.Error())
	}

	dbPassword := os.Getenv("DB_PASS")
	if dbPassword == "" {
		log.Fatal("DB_PASS environment variable is required")
	}

	db, err := repository.NewPostgres(repository.Config{
		Host:     getEnv("DB_HOST", viper.GetString("database.host")),
		Port:     getEnv("DB_PORT", viper.GetString("database.port")),
		Username: getEnv("DB_USER", viper.GetString("database.username")),
		Password: dbPassword,
		DBName:   getEnv("DB_NAME", viper.GetString("database.name")),
		SSLMode:  getEnv("DB_SSLMODE", viper.GetString("database.sslmode")),
	})
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	services := service.NewService(repo)
	handler := handlers.NewHandler(services)

	grpcPort := getEnv("GRPC_PORT", viper.GetString("grpc.port"))
	if grpcPort == "" {
		grpcPort = "50056"
		log.Printf("warning: using default gRPC port: %s", grpcPort)
	}

	metrics.ServeMetrics(getEnv("METRICS_ADDR", ":9097"))

	log.Printf("Starting Skills Service on gRPC port: %s", grpcPort)
	grpcServer := server.New(grpcPort, handler)

	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Fatalf("failed to run gRPC server: %s", err.Error())
		}
	}()

	log.Printf("✓ Skills service started successfully on port %s", grpcPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
	log.Println("Skills service stopped")
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
