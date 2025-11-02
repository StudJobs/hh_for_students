package main

import (
	"hh_for_students/vacancy-service/internal/handlers"
	"hh_for_students/vacancy-service/internal/repository"
	"hh_for_students/vacancy-service/internal/service"
	"hh_for_students/vacancy-service/server"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Загрузка конфигурации
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

	// Инициализация базы данных
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

	// Инициализация зависимостей
	repo := repository.NewRepository(db)
	serv := service.NewService(repo)
	vacancyHandlers := handlers.NewVacancyHandler(serv)

	// Получаем порт из конфигурации!
	grpcPort := getEnv("GRPC_PORT", viper.GetString("grpc.port"))
	if grpcPort == "" {
		grpcPort = "50052" // значение по умолчанию
		log.Printf("warning: using default gRPC port: %s", grpcPort)
	}

	log.Printf("Starting Auth Service on gRPC port: %s", grpcPort)

	// Запуск gRPC сервера
	grpcServer := server.New(grpcPort, vacancyHandlers)

	// Graceful shutdown
	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Fatalf("failed to run gRPC server: %s", err.Error())
		}
	}()

	log.Printf("✓ User service started successfully on port %s", grpcPort)
	log.Printf("✓ Database connected: %s:%s",
		getEnv("DB_HOST", viper.GetString("database.host")),
		getEnv("DB_PORT", viper.GetString("database.port")))

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
	log.Println("Vacancy service stopped")
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
