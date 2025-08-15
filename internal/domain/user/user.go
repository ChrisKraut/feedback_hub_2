package user

import (
	"errors"
	"strings"
	"time"
)

// User represents a user in the system with role-based access control.
// AI-hint: Core domain entity containing user business logic and invariants.
// Enforces email uniqueness, role assignment, and Super User protection rules.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUser creates a new User with validation.
// AI-hint: Factory method that enforces business rules during user creation.
// Validates email format and ensures required fields are present.
func NewUser(id, email, name, roleID string) (*User, error) {
	if id == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if !isValidEmail(email) {
		return nil, errors.New("invalid email format")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if roleID == "" {
		return nil, errors.New("role ID cannot be empty")
	}

	now := time.Now()
	return &User{
		ID:        id,
		Email:     strings.ToLower(strings.TrimSpace(email)),
		Name:      strings.TrimSpace(name),
		RoleID:    roleID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateName updates the user's name with validation.
// AI-hint: Domain method that maintains business invariants during updates.
func (u *User) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	u.Name = strings.TrimSpace(name)
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateRole updates the user's role.
// AI-hint: Domain method for role assignment with validation.
func (u *User) UpdateRole(roleID string) error {
	if roleID == "" {
		return errors.New("role ID cannot be empty")
	}
	u.RoleID = roleID
	u.UpdatedAt = time.Now()
	return nil
}

// isValidEmail performs basic email validation.
// AI-hint: Simple email validation for domain integrity.
// More sophisticated validation can be added in future iterations.
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) == 0 {
		return false
	}
	atIndex := strings.Index(email, "@")
	if atIndex <= 0 || atIndex == len(email)-1 {
		return false
	}
	dotIndex := strings.LastIndex(email, ".")
	return dotIndex > atIndex && dotIndex < len(email)-1
}

// Repository defines the interface for user persistence operations.
// AI-hint: Repository pattern interface for dependency inversion.
// Keeps domain logic independent of persistence implementation.
type Repository interface {
	Create(ctx interface{}, user *User) error
	GetByID(ctx interface{}, id string) (*User, error)
	GetByEmail(ctx interface{}, email string) (*User, error)
	Update(ctx interface{}, user *User) error
	Delete(ctx interface{}, id string) error
	List(ctx interface{}) ([]*User, error)
	GetByRoleID(ctx interface{}, roleID string) ([]*User, error)
}

// Service defines the business operations for user management.
// AI-hint: Domain service interface for complex business operations
// that don't naturally belong to a single entity.
type Service interface {
	CreateUser(ctx interface{}, email, name, roleID string, createdByUserID string) (*User, error)
	GetUser(ctx interface{}, id string) (*User, error)
	UpdateUser(ctx interface{}, id, name string, updatedByUserID string) (*User, error)
	UpdateUserRole(ctx interface{}, id, roleID string, updatedByUserID string) (*User, error)
	DeleteUser(ctx interface{}, id string, deletedByUserID string) error
	ListUsers(ctx interface{}) ([]*User, error)
}

// Error types for the user domain.
// AI-hint: Domain-specific errors for clear error handling and business rules.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidUserData    = errors.New("invalid user data")
	ErrUnauthorized       = errors.New("unauthorized operation")
)
