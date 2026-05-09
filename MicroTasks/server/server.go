package server

import (
	"fmt"
	"log"
	"net"

	microtaskv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/microtask/v1"
	"github.com/studjobs/hh_for_students/microtasks/internal/metrics"
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

func New(port string, srv microtaskv1.MicroTaskServiceServer) *Server {
	gs := grpc.NewServer(grpc.UnaryInterceptor(metrics.UnaryInterceptor()))
	microtaskv1.RegisterMicroTaskServiceServer(gs, srv)

	hs := health.NewServer()
	healthpb.RegisterHealthServer(gs, hs)
	hs.SetServingStatus("microtask.v1", healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(gs)

	return &Server{grpcServer: gs, port: port, healthServer: hs}
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
		s.healthServer.SetServingStatus("microtask.v1", healthpb.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}
	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
