package DB

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

// DBConfig содержит конфигурацию для подключения к PostgreSQL
type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgres создает новое подключение к базе данных PostgreSQL и выполняет миграции
func NewPostgres(cfg DBConfig) (*pgxpool.Pool, error) {
	// Формируем строку подключения
	strCfg := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Подключаемся к базе данных через pgxpool
	dbPool, err := pgxpool.Connect(context.Background(), strCfg)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}
	logrus.Printf("database is connected")

	// Проверяем доступность базы данных
	if err := dbPool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error when pinging the database: %w", err)
	}

	// Запускаем миграции
	if err := runMigrations(dbPool); err != nil {
		return nil, fmt.Errorf("migration execution error: %w", err)
	}
	logrus.Printf("migration is created")
	return dbPool, nil
}

// runMigrations запускает миграции для базы данных
func runMigrations(dbPool *pgxpool.Pool) error {

	sqlDB := stdlib.OpenDB(*dbPool.Config().ConnConfig)
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка при создании драйвера миграции: %w", err)
	}

	// Получение текущего рабочего каталога
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("ошибка при получении текущего рабочего каталога: %w", err)
	}

	// Формирование абсолютного пути
	absoluteMigrationPath := filepath.Join(currentDir, "schema")

	// Преобразуем путь для Windows
	absoluteMigrationPath = strings.ReplaceAll(absoluteMigrationPath, "\\", "/")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+absoluteMigrationPath,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("ошибка при создании мигратора: %w", err)
	}

	// Проверяем, что мигратор не nil
	if m == nil {
		return fmt.Errorf("мигратор не инициализирован")
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка при выполнении миграций: %w", err)
	}

	return nil
}
