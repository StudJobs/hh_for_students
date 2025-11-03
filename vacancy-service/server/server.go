package server

import (
	"fmt"
	"log"
	"net"

	vacancyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/vacancy/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	port       string
}

func New(port string, vacancyService vacancyv1.VacancyServiceServer) *Server {
	grpcServer := grpc.NewServer()

	// Регистрация сервисов
	vacancyv1.RegisterVacancyServiceServer(grpcServer, vacancyService)
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		port:       port,
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
	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
