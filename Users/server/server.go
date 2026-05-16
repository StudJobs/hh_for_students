package server

import (
	"fmt"
	"log"
	"net"

	chatv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/chat/v1"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"
	"github.com/studjobs/hh_for_students/users/internal/metrics"
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

func New(port string, usersService usersv1.UsersServiceServer, chatService chatv1.ChatServiceServer) *Server {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(metrics.UnaryInterceptor()))

	// Регистрация сервисов
	usersv1.RegisterUsersServiceServer(grpcServer, usersService)
	if chatService != nil {
		chatv1.RegisterChatServiceServer(grpcServer, chatService)
	}

	// Создание и настройка health сервера
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Установка статусов сервисов
	healthServer.SetServingStatus("users.v1", healthpb.HealthCheckResponse_SERVING)
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	log.Printf("gRPC server listening on port %s", s.port)

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

func (s *Server) GracefulStop() {
	log.Println("Shutting down gRPC server gracefully...")

	// Установка статуса NOT_SERVING перед остановкой
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("users.v1", healthpb.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}

	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}

// SetServiceStatus позволяет динамически менять статус сервиса
func (s *Server) SetServiceStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	if s.healthServer != nil {
		s.healthServer.SetServingStatus(service, status)
	}
}

// Shutdown немедленная остановка сервера
func (s *Server) Shutdown() {
	log.Println("Shutting down gRPC server immediately...")

	if s.healthServer != nil {
		s.healthServer.SetServingStatus("users.v1", healthpb.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}

	s.grpcServer.Stop()
	log.Println("gRPC server stopped")
}

// GetHealthServer возвращает health server для кастомной логики
func (s *Server) GetHealthServer() *health.Server {
	return s.healthServer
}
