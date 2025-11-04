package server

import (
	"context"
	"fmt"
	"log"
	"net"

	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/studjobs/hh_for_students/auth/internal/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer   *grpc.Server
	port         string
	healthServer *health.Server
}

func New(port string, authHandlers *handlers.AuthHandlers) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// Регистрация сервисов
	authv1.RegisterAuthServiceServer(grpcServer, authHandlers)

	// Создание и настройка health сервера
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Установка статусов сервисов
	healthServer.SetServingStatus("auth.v1", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING) // Общий статус сервера

	// Включение reflection для тестирования
	reflection.Register(grpcServer)

	return &Server{
		grpcServer:   grpcServer,
		port:         port,
		healthServer: healthServer,
	}
}

func (s *Server) Run() error {
	// Используем 0.0.0.0 чтобы слушать на всех интерфейсах
	addr := fmt.Sprintf("0.0.0.0:%s", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	log.Printf("gRPC server listening on %s", addr)

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) GracefulStop() {
	log.Println("Shutting down gRPC server gracefully...")

	// Установка статуса NOT_SERVING перед остановкой
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("auth.v1", healthpb.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}

	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}

// Shutdown немедленная остановка сервера
func (s *Server) Shutdown() {
	log.Println("Shutting down gRPC server immediately...")

	if s.healthServer != nil {
		s.healthServer.SetServingStatus("auth.v1", healthpb.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}

	s.grpcServer.Stop()
	log.Println("gRPC server stopped")
}

// SetServiceStatus позволяет динамически менять статус сервиса
func (s *Server) SetServiceStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	if s.healthServer != nil {
		s.healthServer.SetServingStatus(service, status)
	}
}

// GetHealthServer возвращает health server для кастомной логики
func (s *Server) GetHealthServer() *health.Server {
	return s.healthServer
}

// loggingInterceptor - простой интерцептор для логирования запросов
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("gRPC method: %s", info.FullMethod)
	return handler(ctx, req)
}
