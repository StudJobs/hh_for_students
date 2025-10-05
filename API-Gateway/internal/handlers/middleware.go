package handlers

import (
	apigatewayV1 "github.com/StudJobs/proto_srtucture/gen/go/proto/apigateway/v1"
	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

// Константы для ключей контекста
type contextKey string
type Role string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
	TokenKey  contextKey = "token"

	// Roles
	ROLE_DEVELOPER Role = "ROLE_DEVELOPER"
	ROLE_STUDENT   Role = "ROLE_STUDENT"
	ROLE_HR        Role = "ROLE_EMPLOYER"

	ID string = "id"
)

// AuthMiddleware проверяет JWT токен через ApiGateway
func AuthMiddleware(client apigatewayV1.ApiGatewayServiceClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/api/v1/auth/login" || c.Path() == "/api/v1/auth/register" {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Bearer token required",
			})
		}

		validation, err := client.ParseToken(c.Context(), &authv1.ParseTokenRequest{
			AccessToken: token,
		})
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication service error",
			})
		}

		if !validation.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		c.Locals(string(UserIDKey), validation.UserUuid)
		c.Locals(string(RoleKey), validation.Role.String())
		c.Locals(string(TokenKey), token)
		return c.Next()
	}
}

// RoleMiddleware проверяет роль пользователя
func RoleMiddleware(allowedRoles ...Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := getRoleFromContext(c)
		if userRole == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		// ROLE_DEVELOPER имеет доступ ко всему
		if userRole == ROLE_DEVELOPER {
			return c.Next()
		}

		// Проверяем разрешенные роли
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

// OwnerOrRoleMiddleware проверяет владельца или роль
func OwnerOrRoleMiddleware(paramName string, allowedRoles ...Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := getUserIDFromContext(c)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		targetID := c.Params(paramName)
		userRole := getRoleFromContext(c)

		// ROLE_DEVELOPER имеет доступ ко всему
		if userRole == ROLE_DEVELOPER {
			return c.Next()
		}

		// Владелец имеет доступ к своим данным
		if userID == targetID {
			return c.Next()
		}

		// Проверяем разрешенные роли
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

// getUserIDFromContext возвращает user_id из контекста
func getUserIDFromContext(c *fiber.Ctx) string {
	if userID, ok := c.Locals(string(UserIDKey)).(string); ok {
		return userID
	}
	return ""
}

// getRoleFromContext возвращает роль из контекста
func getRoleFromContext(c *fiber.Ctx) Role {
	if role, ok := c.Locals(string(RoleKey)).(Role); ok { // Type Assertion НУЖНО ПОНАБЛЮДАТЬ как он парсит строку в роль
		return role
	}
	return ""
}

// getTokenFromContext возвращает токен из контекста
func getTokenFromContext(c *fiber.Ctx) string {
	if token, ok := c.Locals(string(TokenKey)).(string); ok {
		return token
	}
	return ""
}
