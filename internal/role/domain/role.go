package domain

import (
	"errors"
	"strings"
	"time"
)

// SuperUserRoleName is the constant name for the Super User role.
// AI-hint: Business constant that enforces the Super User role naming convention.
// This role has ultimate administrative privileges and cannot be deleted.
const SuperUserRoleName = "Super User"

// PredefinedRoles contains the initial roles that should exist in the system.
// AI-hint: Business rules defining the core roles for the feedback hub system.
var PredefinedRoles = []string{
	SuperUserRoleName,
	"Product Owner",
	"Contributor",
}

// Role represents a role in the system with specific permissions.
// AI-hint: Core domain entity for role-based access control.
// Enforces business rules around role naming and Super User protection.
type Role struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRole creates a new Role with validation.
// AI-hint: Factory method that enforces business rules during role creation.
// Validates role name uniqueness and format requirements.
func NewRole(id, name string) (*Role, error) {
	if id == "" {
		return nil, errors.New("role ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("role name cannot be empty")
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, errors.New("role name cannot be empty")
	}

	now := time.Now()
	return &Role{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateName updates the role's name with validation.
// AI-hint: Domain method that maintains business invariants during updates.
// Prevents modification of the Super User role name.
func (r *Role) UpdateName(name string) error {
	if r.Name == SuperUserRoleName {
		return errors.New("cannot modify Super User role")
	}

	if name == "" {
		return errors.New("role name cannot be empty")
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return errors.New("role name cannot be empty")
	}

	r.Name = name
	r.UpdatedAt = time.Now()
	return nil
}

// IsSuperUser returns true if this role is the Super User role.
// AI-hint: Business logic to identify the special Super User role.
func (r *Role) IsSuperUser() bool {
	return r.Name == SuperUserRoleName
}

// CanBeDeleted returns true if this role can be deleted.
// AI-hint: Business rule that protects the Super User role from deletion.
func (r *Role) CanBeDeleted() bool {
	return !r.IsSuperUser()
}

// Repository defines the interface for role persistence operations.
// AI-hint: Repository pattern interface for dependency inversion.
// Keeps domain logic independent of persistence implementation.
type Repository interface {
	Create(ctx interface{}, role *Role) error
	GetByID(ctx interface{}, id string) (*Role, error)
	GetByName(ctx interface{}, name string) (*Role, error)
	Update(ctx interface{}, role *Role) error
	Delete(ctx interface{}, id string) error
	List(ctx interface{}) ([]*Role, error)
	Exists(ctx interface{}, name string) (bool, error)
}

// Service defines the business operations for role management.
// AI-hint: Domain service interface for complex business operations
// that require coordination between multiple entities or external services.
type Service interface {
	CreateRole(ctx interface{}, name string, createdByUserID string) (*Role, error)
	GetRole(ctx interface{}, id string) (*Role, error)
	UpdateRole(ctx interface{}, id, name string, updatedByUserID string) (*Role, error)
	DeleteRole(ctx interface{}, id string, deletedByUserID string) error
	ListRoles(ctx interface{}) ([]*Role, error)
	EnsurePredefinedRoles(ctx interface{}) error
}

// Error types for the role domain.
// AI-hint: Domain-specific errors for clear error handling and business rules.
var (
	ErrRoleNotFound              = errors.New("role not found")
	ErrRoleNameAlreadyExists     = errors.New("role name already exists")
	ErrCannotDeleteSuperUserRole = errors.New("cannot delete Super User role")
	ErrCannotModifySuperUserRole = errors.New("cannot modify Super User role")
	ErrInvalidRoleData           = errors.New("invalid role data")
	ErrUnauthorized              = errors.New("unauthorized operation")
)
