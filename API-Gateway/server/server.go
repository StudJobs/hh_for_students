package server

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app *fiber.App
}

func NewServer(app *fiber.App) *Server {
	return &Server{
		app: app,
	}
}

func (s *Server) Run(port string) error {
	address := ":" + port

	log.Printf("Server starting on http://localhost%s", address)
	log.Printf("Server name: %s", s.app.Config().AppName)

	return s.app.Listen(address)
}

func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down server gracefully...")
	return s.app.ShutdownWithContext(ctx)
}
