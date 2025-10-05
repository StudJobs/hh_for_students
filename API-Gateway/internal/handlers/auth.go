package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// Login(ctx context.Context, in *v14.LoginRequest, opts ...grpc.CallOption) (*v14.AuthResponse, error)
func (h *Handler) Login(c *fiber.Ctx) error {

	return nil
}

// SignUp(ctx context.Context, in *v14.SignUpRequest, opts ...grpc.CallOption) (*v14.AuthResponse, error)
func (h *Handler) Register(c *fiber.Ctx) error { return nil }

// ParseToken(ctx context.Context, in *v14.ParseTokenRequest, opts ...grpc.CallOption) (*v14.TokenValidation, error)
func (h *Handler) parseToken(c *fiber.Ctx) error { return nil }
