package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password hashing and verification.
// AI-hint: Secure password service using bcrypt for hashing with salt and proper cost factor.
type PasswordService struct{}

// NewPasswordService creates a new password service instance.
// AI-hint: Factory method for password service with no dependencies.
func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

// HashPassword creates a bcrypt hash of the given password.
// AI-hint: Password hashing with bcrypt default cost (10) for security without performance impact.
func (s *PasswordService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword checks if the provided password matches the hash.
// AI-hint: Password verification using bcrypt's constant-time comparison to prevent timing attacks.
func (s *PasswordService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// IsValidPassword checks if a password meets basic requirements.
// AI-hint: Basic password validation for minimum length - extend with complexity rules as needed.
func (s *PasswordService) IsValidPassword(password string) bool {
	// Basic validation - extend as needed
	return len(password) >= 8
}
