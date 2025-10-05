package main

import (
	apigatewayV1 "github.com/StudJobs/proto_srtucture/gen/go/proto/apigateway/v1"
	"github.com/studjobs/hh_for_students/api-gateway/internal/handlers"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
	"github.com/studjobs/hh_for_students/api-gateway/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	grpcConn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcConn.Close()

	grpcClient := apigatewayV1.NewApiGatewayServiceClient(grpcConn)
	apiService := services.NewServiceAPIGateway(grpcClient)

	handler := handlers.NewHandler(grpcClient, apiService)
	app := handler.Init()

	srv := server.NewServer(app)

	go func() {
		if err := srv.Run("8080"); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal...")

}
