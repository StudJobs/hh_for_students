// server/server.go
package server

import (
	"fmt"
	"google.golang.org/grpc/reflection"
	"log"
	"net"

	companyv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/company/v1"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	port       string
}

func New(port string, companyService companyv1.CompanyServiceServer) *Server {
	grpcServer := grpc.NewServer()

	// Регистрируем сервисы
	companyv1.RegisterCompanyServiceServer(grpcServer, companyService)
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
