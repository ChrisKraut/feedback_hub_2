package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims stored in the JWT token.
// AI-hint: Custom JWT claims that include user info and standard claims for security.
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations.
// AI-hint: Service for JWT token generation, validation, and parsing with configurable expiration.
type JWTService struct {
	secretKey []byte
}

// NewJWTService creates a new JWT service instance.
// AI-hint: Factory method for JWT service with secret key from environment variables.
func NewJWTService() *JWTService {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		// Generate a default secret for development (should be set in production)
		secretKey = "default-dev-secret-change-in-production"
	}

	return &JWTService{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken creates a new JWT token for the given user.
// AI-hint: Token generation with user claims and 24-hour expiration for web apps.
func (s *JWTService) GenerateToken(userID, email, roleName string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "feedback-hub",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken parses and validates a JWT token.
// AI-hint: Token validation with expiration check and signature verification.
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// RefreshToken generates a new token if the current one is valid.
// AI-hint: Token refresh mechanism for extending user sessions without re-login.
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate new token with same user info but fresh expiration
	return s.GenerateToken(claims.UserID, claims.Email, claims.RoleName)
}
