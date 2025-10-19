package server

import (
	"fmt"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"log"
	"net"

	"context"
	"github.com/studjobs/hh_for_students/auth/internal/handlers"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	port       string
}

func New(port string, authHandlers *handlers.AuthHandlers) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// Регистрируем сервисы
	authv1.RegisterAuthServiceServer(grpcServer, authHandlers)

	return &Server{
		grpcServer: grpcServer,
		port:       port,
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
	log.Println("Shutting down gRPC server...")
	s.grpcServer.GracefulStop()
}

// loggingInterceptor - простой интерцептор для логирования запросов
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("gRPC method: %s", info.FullMethod)
	return handler(ctx, req)
}
