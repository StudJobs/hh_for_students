package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/studjobs/hh_for_students/microtasks/internal/achievementclient"
	"github.com/studjobs/hh_for_students/microtasks/internal/handlers"
	"github.com/studjobs/hh_for_students/microtasks/internal/metrics"
	"github.com/studjobs/hh_for_students/microtasks/internal/repository"
	"github.com/studjobs/hh_for_students/microtasks/internal/searchclient"
	"github.com/studjobs/hh_for_students/microtasks/internal/service"
	"github.com/studjobs/hh_for_students/microtasks/internal/storage"
	"github.com/studjobs/hh_for_students/microtasks/internal/usersclient"
	"github.com/studjobs/hh_for_students/microtasks/server"
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

	// MinIO: внутренний клиент для HEAD/упаковки, public — для presigned URL клиенту-браузеру.
	s3Bucket := getEnv("MINIO_BUCKET", "achievements")
	s3UseSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))
	// MinIO опционален: если недоступен — file-upload отключён, но сервис стартует.
	var solutionsStore *storage.Solutions
	s3Internal, err := storage.NewInternalClient(storage.S3Config{
		Endpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
		AccessKey: getEnv("MINIO_ACCESS_KEY", ""),
		SecretKey: getEnv("MINIO_SECRET_KEY", ""),
		UseSSL:    s3UseSSL,
		Bucket:    s3Bucket,
	})
	if err != nil {
		log.Printf("warning: MinIO init failed: %v — file-upload disabled", err)
	} else {
		s3Public := s3Internal
		if pubEP := getEnv("MINIO_PUBLIC_ENDPOINT", ""); pubEP != "" {
			pubCli, pubErr := storage.NewPublicClient(storage.S3Config{
				Endpoint:  pubEP,
				AccessKey: getEnv("MINIO_ACCESS_KEY", ""),
				SecretKey: getEnv("MINIO_SECRET_KEY", ""),
				UseSSL:    s3UseSSL,
				Bucket:    s3Bucket,
			})
			if pubErr != nil {
				log.Printf("warning: MinIO public client init failed: %v — using internal endpoint for presign", pubErr)
			} else {
				s3Public = pubCli
			}
		}
		solutionsStore = storage.NewSolutions(s3Internal, s3Public, s3Bucket)
	}

	svc := service.NewService(repo)

	searchCli := searchclient.New(getEnv("SEARCH_GRPC_ADDR", viper.GetString("clients.search_addr")))
	defer searchCli.Close()

	achievementsCli := achievementclient.New(getEnv("ACHIEVEMENTS_GRPC_ADDR", viper.GetString("clients.achievements_addr")))
	defer achievementsCli.Close()

	usersCli := usersclient.New(getEnv("USERS_GRPC_ADDR", "user:50052"))
	defer usersCli.Close()

	handler := handlers.New(svc, searchCli, achievementsCli, usersCli, solutionsStore)

	grpcPort := getEnv("GRPC_PORT", viper.GetString("grpc.port"))
	if grpcPort == "" {
		grpcPort = "50058"
	}

	metrics.ServeMetrics(getEnv("METRICS_ADDR", ":9099"))

	log.Printf("Starting MicroTasks Service on gRPC port: %s", grpcPort)
	grpcServer := server.New(grpcPort, handler)

	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Fatalf("failed to run gRPC server: %s", err.Error())
		}
	}()

	log.Printf("✓ MicroTasks service started successfully on port %s", grpcPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
	log.Println("MicroTasks service stopped")
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
