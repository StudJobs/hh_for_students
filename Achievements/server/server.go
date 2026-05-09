package server

import (
	"fmt"
	"log"
	"net"

	achievementv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/achievement/v1"
	"github.com/studjobs/hh_for_students/achievments/internal/metrics"
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

func New(port string, achievementService achievementv1.AchievementServiceServer) *Server {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(metrics.UnaryInterceptor()))

	// Регистрация сервисов
	achievementv1.RegisterAchievementServiceServer(grpcServer, achievementService)

	// Создание и настройка health сервера
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Установка статусов сервисов
	healthServer.SetServingStatus("achievement.v1", healthpb.HealthCheckResponse_SERVING)
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
		s.healthServer.SetServingStatus("achievement.v1", healthpb.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}

	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}

// Shutdown немедленная остановка сервера
func (s *Server) Shutdown() {
	log.Println("Shutting down gRPC server immediately...")

	if s.healthServer != nil {
		s.healthServer.SetServingStatus("achievement.v1", healthpb.HealthCheckResponse_NOT_SERVING)
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
