package server

import (
	"fmt"
	"log"
	"net"

	searchv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/search/v1"
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

func New(port string, searchServer searchv1.SearchServiceServer) *Server {
	grpcServer := grpc.NewServer()
	searchv1.RegisterSearchServiceServer(grpcServer, searchServer)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("search.v1", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	return &Server{
		grpcServer:   grpcServer,
		port:         port,
		healthServer: healthServer,
	}
}

func (s *Server) Run() error {
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
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("search.v1", healthpb.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}
	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
