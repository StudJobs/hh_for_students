package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/studjobs/hh_for_students/search/internal/clients"
	"github.com/studjobs/hh_for_students/search/internal/esclient"
	"github.com/studjobs/hh_for_students/search/internal/handlers"
	"github.com/studjobs/hh_for_students/search/internal/indexer"
	"github.com/studjobs/hh_for_students/search/internal/reindexer"
	"github.com/studjobs/hh_for_students/search/internal/searcher"
	"github.com/studjobs/hh_for_students/search/server"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("warning: error loading .env file: %s", err.Error())
	}

	esURL := getEnv("ELASTICSEARCH_URL", viper.GetString("elasticsearch.url"))
	if esURL == "" {
		esURL = "http://elasticsearch:9200"
	}
	usersAddr := getEnv("USERS_GRPC_ADDR", viper.GetString("clients.users_addr"))
	vacancyAddr := getEnv("VACANCY_GRPC_ADDR", viper.GetString("clients.vacancy_addr"))
	grpcPort := getEnv("GRPC_PORT", viper.GetString("grpc.port"))
	if grpcPort == "" {
		grpcPort = "50057"
	}

	es, err := esclient.New(esURL)
	if err != nil {
		log.Fatalf("failed to init elasticsearch client: %s", err.Error())
	}

	c, err := clients.New(usersAddr, vacancyAddr)
	if err != nil {
		log.Fatalf("failed to init upstream gRPC clients: %s", err.Error())
	}
	defer c.Close()

	idx := indexer.New(es)
	srch := searcher.New(es)
	rx := reindexer.New(es, idx, c)

	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := rx.EnsureIndices(startupCtx, false); err != nil {
		log.Printf("warning: ensure indices: %s", err.Error())
	}
	cancel()

	handler := handlers.New(srch, idx, rx)

	log.Printf("Starting Search Service on gRPC port: %s (es=%s, users=%s, vacancy=%s)", grpcPort, esURL, usersAddr, vacancyAddr)
	grpcServer := server.New(grpcPort, handler)

	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Fatalf("failed to run gRPC server: %s", err.Error())
		}
	}()

	log.Printf("Search service started successfully on port %s", grpcPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
	log.Println("Search service stopped")
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
