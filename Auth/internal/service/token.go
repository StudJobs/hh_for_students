package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	authv1 "github.com/StudJobs/proto_srtucture/gen/go/proto/auth/v1"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type Claims struct {
	UserUUID string      `json:"user_uuid"`
	Email    string      `json:"email"`
	Role     authv1.Role `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(cfg JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey:     cfg.SecretKey,
		tokenDuration: cfg.TokenDuration,
	}
}

func (m *JWTManager) GenerateToken(userUUID, email string, role authv1.Role) (string, error) {
	log.Printf("Generating JWT token for user: %s, role: %v", userUUID, role)

	claims := Claims{
		UserUUID: userUUID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		log.Printf("JWT token generation failed: %v", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Printf("JWT token generated successfully for user: %s", userUUID)
	return tokenString, nil
}

func (m *JWTManager) ValidateToken(tokenString string) (string, authv1.Role, error) {
	log.Printf("Validating JWT token: %s...", tokenString[:10])

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		log.Printf("JWT token parsing failed: %v", err)
		return "", authv1.Role_ROLE_UNSPECIFIED, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		log.Printf("JWT token validated - user_uuid: %s, role: %v", claims.UserUUID, claims.Role)
		return claims.UserUUID, claims.Role, nil
	}

	log.Printf("JWT token validation failed - invalid token")
	return "", authv1.Role_ROLE_UNSPECIFIED, ErrInvalidToken
}
