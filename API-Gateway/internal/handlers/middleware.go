package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/services"
	"log"
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

// AuthMiddleware проверяет JWT токен через APIGatewayService
func AuthMiddleware(apiService *services.ApiGateway) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Пропускаем auth endpoints и health check
		if c.Path() == "/api/v1/auth/login" ||
			c.Path() == "/api/v1/auth/register" ||
			c.Path() == "/health" {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Printf("AuthMiddleware: Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			log.Printf("AuthMiddleware: Bearer token prefix missing")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Bearer token required",
			})
		}

		log.Printf("AuthMiddleware: Validating token: %s...", token[:min(10, len(token))])

		// Вызываем сервис для проверки токена
		valid, userUUID, role, err := apiService.Auth.ValidateToken(context.Background(), token)
		if err != nil {
			log.Printf("AuthMiddleware: Token validation error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication service error",
			})
		}

		if !valid {
			log.Printf("AuthMiddleware: Invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		log.Printf("AuthMiddleware: Token validated - user_uuid: %s, role: %s", userUUID, role)

		// Конвертируем строку роли в тип Role
		var userRole Role
		switch role {
		case "ROLE_DEVELOPER":
			userRole = ROLE_DEVELOPER
		case "ROLE_STUDENT":
			userRole = ROLE_STUDENT
		case "ROLE_EMPLOYER":
			userRole = ROLE_HR
		default:
			log.Printf("AuthMiddleware: Unknown role: %s", role)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user role",
			})
		}

		// Сохраняем данные в контекст
		c.Locals(string(UserIDKey), userUUID)
		c.Locals(string(RoleKey), userRole)
		c.Locals(string(TokenKey), token)

		return c.Next()
	}
}

// RoleMiddleware проверяет роль пользователя
func RoleMiddleware(allowedRoles ...Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := getRoleFromContext(c)
		if userRole == "" {
			log.Printf("RoleMiddleware: User not authenticated")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		log.Printf("RoleMiddleware: Checking role %s against allowed roles: %v", userRole, allowedRoles)

		// ROLE_DEVELOPER имеет доступ ко всему
		if userRole == ROLE_DEVELOPER {
			log.Printf("RoleMiddleware: Developer role - access granted")
			return c.Next()
		}

		// Проверяем разрешенные роли
		for _, role := range allowedRoles {
			if userRole == role {
				log.Printf("RoleMiddleware: Role %s allowed - access granted", userRole)
				return c.Next()
			}
		}

		log.Printf("RoleMiddleware: Role %s not in allowed roles - access denied", userRole)
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
			log.Printf("OwnerOrRoleMiddleware: User not authenticated")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		targetID := c.Params(paramName)
		userRole := getRoleFromContext(c)

		log.Printf("OwnerOrRoleMiddleware: User %s (role: %s) accessing resource %s", userID, userRole, targetID)

		// ROLE_DEVELOPER имеет доступ ко всему
		if userRole == ROLE_DEVELOPER {
			log.Printf("OwnerOrRoleMiddleware: Developer role - access granted")
			return c.Next()
		}

		// Владелец имеет доступ к своим данным
		if userID == targetID {
			log.Printf("OwnerOrRoleMiddleware: Owner access - access granted")
			return c.Next()
		}

		// Проверяем разрешенные роли
		for _, role := range allowedRoles {
			if userRole == role {
				log.Printf("OwnerOrRoleMiddleware: Role %s allowed - access granted", userRole)
				return c.Next()
			}
		}

		log.Printf("OwnerOrRoleMiddleware: Access denied for user %s to resource %s", userID, targetID)
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
	if role, ok := c.Locals(string(RoleKey)).(Role); ok {
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
