package handlers

import (
	"context"
	usersv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/users/v1"

	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// Login обрабатывает вход пользователя
func (h *Handler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	// Парсим тело запроса
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Login failed - body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	log.Printf("Login attempt for email: %s", req.Email)

	// Валидация обязательных полей
	if req.Email == "" || req.Password == "" || req.Role == "" {
		log.Printf("Login failed - missing fields for email: %s", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "MISSING_FIELDS",
			Message: "Email, password and role are required",
		})
	}

	log.Printf("Calling API Gateway Login for email: %s, role: %s", req.Email, req.Role)

	// Вызываем сервис (конвертация роли теперь в сервисе)
	ctx := context.Background()
	resp, err := h.apiService.Auth.Login(ctx, req.Email, req.Password, req.Role)

	if err != nil {
		log.Printf("API Gateway Login failed for email %s: %v", req.Email, err)
		return h.handleAuthError(c, err)
	}

	log.Printf("Login successful for email: %s, user_uuid: %s", req.Email, resp.UserUUID)

	return c.JSON(resp)
}

// Register обрабатывает регистрацию пользователя
func (h *Handler) Register(c *fiber.Ctx) error {
	var req models.SignUpRequest

	// Парсим тело запроса
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Register failed - body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	log.Printf("Register attempt for email: %s", req.Email)

	// Валидация обязательных полей
	if req.Email == "" || req.Password == "" || req.Role == "" {
		log.Printf("Register failed - missing fields for email: %s", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "MISSING_FIELDS",
			Message: "Email, password and role are required",
		})
	}

	// Проверка минимальной длины пароля
	if len(req.Password) < 6 {
		log.Printf("Register failed - weak password for email: %s", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "WEAK_PASSWORD",
			Message: "Password must be at least 6 characters long",
		})
	}

	log.Printf("Calling API Gateway Register for email: %s, role: %s", req.Email, req.Role)

	// Вызываем сервис (конвертация роли теперь в сервисе)
	resp, err := h.apiService.Auth.Register(c.Context(), req.Email, req.Password, req.Role)

	if err != nil {
		log.Printf("API Gateway Register failed for email %s: %v", req.Email, err)
		return h.handleAuthError(c, err)
	}

	if _, err = h.apiService.User.CreateUser(c.Context(), &usersv1.NewProfileRequest{
		Profile: &usersv1.Profile{
			Id:    resp.UserUUID,
			Email: req.Email,
			Age:   18,
		},
	}); err != nil {
		log.Printf("API Gateway Create failed for email %s: %v", req.Email, err)

		// TODO: Удалить пользователя из auth
		return h.handleAuthError(c, err)
	}

	log.Printf("Register successful for email: %s, user_uuid: %s", req.Email, resp.UserUUID)

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// parseToken проверяет валидность токена
func (h *Handler) parseToken(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		log.Printf("ParseToken failed - missing authorization header")
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "MISSING_TOKEN",
			Message: "Authorization header is required",
		})
	}

	// Убираем "Bearer " префикс если есть
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	log.Printf("ParseToken attempt for token: %s...", token[:min(10, len(token))])

	// Вызываем сервис для проверки токена
	valid, userUUID, role, err := h.apiService.Auth.ValidateToken(context.Background(), token)
	if err != nil {
		log.Printf("ParseToken failed: %v", err)
		return h.handleAuthError(c, err)
	}

	log.Printf("ParseToken result - valid: %t, user_uuid: %s, role: %s", valid, userUUID, role)

	return c.JSON(fiber.Map{
		"valid":     valid,
		"user_uuid": userUUID,
		"role":      role,
	})
}

// handleAuthError обрабатывает ошибки аутентификации
func (h *Handler) handleAuthError(c *fiber.Ctx, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		log.Printf("Non-gRPC error in handleAuthError: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		})
	}

	log.Printf("gRPC error - code: %s, message: %s", st.Code(), st.Message())

	switch st.Code() {
	case codes.Unauthenticated:
		return c.Status(fiber.StatusUnauthorized).JSON(models.Error{
			Code:    "INVALID_CREDENTIALS",
			Message: "Invalid email or password",
		})
	case codes.AlreadyExists:
		return c.Status(fiber.StatusConflict).JSON(models.Error{
			Code:    "USER_EXISTS",
			Message: "User with this email already exists",
		})
	case codes.InvalidArgument:
		return c.Status(fiber.StatusBadRequest).JSON(models.Error{
			Code:    "INVALID_DATA",
			Message: st.Message(),
		})
	case codes.NotFound:
		return c.Status(fiber.StatusNotFound).JSON(models.Error{
			Code:    "NOT_FOUND",
			Message: "Resource not found",
		})
	case codes.PermissionDenied:
		return c.Status(fiber.StatusForbidden).JSON(models.Error{
			Code:    "PERMISSION_DENIED",
			Message: "Insufficient permissions",
		})
	case codes.Unavailable:
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.Error{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Authentication service is temporarily unavailable",
		})
	default:
		log.Printf("Unhandled gRPC error code: %s, message: %s", st.Code(), st.Message())
		return c.Status(fiber.StatusInternalServerError).JSON(models.Error{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		})
	}
}
