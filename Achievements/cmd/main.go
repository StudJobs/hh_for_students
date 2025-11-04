package main

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/studjobs/hh_for_students/achievments/internal/handlers"
	"github.com/studjobs/hh_for_students/achievments/internal/repository"
	"github.com/studjobs/hh_for_students/achievments/internal/repository/DB"
	"github.com/studjobs/hh_for_students/achievments/internal/service"
	"github.com/studjobs/hh_for_students/achievments/server"

	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	// Загрузка конфигурации
	if err := initConfig(); err != nil {
		log.Fatalf("Ошибка инициализации конфигурации: %s", err.Error())
	}

	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Printf("Предупреждение: ошибка загрузки .env файла: %s", err.Error())
	}

	// Проверка обязательных переменных окружения
	dbPassword := os.Getenv("DB_PASS")
	if dbPassword == "" {
		log.Fatal("Переменная окружения DB_PASS обязательна для установки")
	}

	// Инициализация базы данных PostgreSQL
	db, err := DB.NewPostgres(DB.DBConfig{
		Host:     getEnv("DB_HOST", viper.GetString("database.host")),
		Port:     getEnv("DB_PORT", viper.GetString("database.port")),
		Username: getEnv("DB_USER", viper.GetString("database.username")),
		Password: dbPassword,
		DBName:   getEnv("DB_NAME", viper.GetString("database.name")),
		SSLMode:  getEnv("DB_SSLMODE", viper.GetString("database.sslmode")),
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %s", err.Error())
	}
	defer db.Close()
	log.Println("✓ Успешное подключение к PostgreSQL")

	// Инициализация MinIO/S3 клиента
	minioClient, err := DB.NewMinioClient(DB.S3Config{
		Endpoint:  getEnv("MINIO_ENDPOINT", viper.GetString("minio.endpoint")),
		AccessKey: getEnv("MINIO_ACCESS_KEY", viper.GetString("minio.access_key")),
		SecretKey: getEnv("MINIO_SECRET_KEY", viper.GetString("minio.secret_key")),
		UseSSL:    getEnvAsBool("MINIO_USE_SSL", viper.GetBool("minio.use_ssl")),
		Bucket:    getEnv("MINIO_BUCKET", viper.GetString("minio.bucket")),
	})
	if err != nil {
		log.Fatalf("Ошибка подключения к MinIO/S3: %s", err.Error())
	}
	log.Println("✓ Успешное подключение к MinIO/S3 хранилищу")

	// Инициализация репозитория с зависимостями от БД и S3
	repo := repository.NewRepository(db, minioClient)

	// Инициализация сервисного слоя
	services := service.NewService(repo)

	// Инициализация gRPC обработчиков
	handler := handlers.NewHandler(services)

	// Получение порта для gRPC сервера
	grpcPort := getEnv("GRPC_PORT", viper.GetString("grpc.port"))
	if grpcPort == "" {
		grpcPort = "50052" // значение по умолчанию для сервиса достижений
		log.Printf("Предупреждение: используется порт gRPC по умолчанию: %s", grpcPort)
	}

	log.Printf("Запуск сервиса достижений на gRPC порту: %s", grpcPort)

	// Инициализация и запуск gRPC сервера
	grpcServer := server.New(grpcPort, handler)

	// Graceful shutdown
	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Fatalf("Ошибка запуска gRPC сервера: %s", err.Error())
		}
	}()

	// Логирование успешного старта сервиса
	log.Printf("✓ Сервис достижений успешно запущен на порту %s", grpcPort)
	log.Printf("✓ База данных подключена: %s:%s",
		getEnv("DB_HOST", viper.GetString("database.host")),
		getEnv("DB_PORT", viper.GetString("database.port")))
	log.Printf("✓ MinIO/S3 подключен: %s",
		getEnv("MINIO_ENDPOINT", viper.GetString("minio.endpoint")))

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал остановки, выполнение graceful shutdown...")
	grpcServer.GracefulStop()
	log.Println("Сервис достижений остановлен")
}

// initConfig инициализирует конфигурацию приложения из YAML файла
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsBool возвращает boolean значение переменной окружения
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
