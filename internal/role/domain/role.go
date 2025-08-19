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
// Enforces business rules around role naming, organization scoping, and Super User protection.
type Role struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	OrganizationID string    `json:"organization_id"` // Organization scoping for multi-tenant support
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewRole creates a new Role with validation.
// AI-hint: Factory method that enforces business rules during role creation.
// Validates role name uniqueness, format requirements, and organization scoping.
func NewRole(id, name, organizationID string) (*Role, error) {
	if id == "" {
		return nil, errors.New("role ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("role name cannot be empty")
	}
	if organizationID == "" {
		return nil, errors.New("organization ID cannot be empty")
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, errors.New("role name cannot be empty")
	}

	now := time.Now()
	return &Role{
		ID:             id,
		Name:           name,
		OrganizationID: organizationID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// NewRoleWithoutOrganization creates a new Role without organization scoping (legacy support).
// AI-hint: Factory method for backward compatibility with existing roles.
// Validates role name uniqueness and format requirements.
func NewRoleWithoutOrganization(id, name string) (*Role, error) {
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
		ID:             id,
		Name:           name,
		OrganizationID: "", // No organization scoping for legacy roles
		CreatedAt:      now,
		UpdatedAt:      now,
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

// UpdateOrganization updates the role's organization.
// AI-hint: Domain method for organization assignment with validation.
func (r *Role) UpdateOrganization(organizationID string) error {
	if organizationID == "" {
		return errors.New("organization ID cannot be empty")
	}
	r.OrganizationID = organizationID
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

// IsInOrganization checks if the role belongs to the specified organization.
// AI-hint: Business logic method for organization membership validation.
// Returns false if role has no organization or if the specified organization doesn't match.
func (r *Role) IsInOrganization(organizationID string) bool {
	// Roles without organization cannot be in any organization
	if !r.IsOrganizationScoped() {
		return false
	}
	return r.OrganizationID == organizationID
}

// IsOrganizationScoped checks if the role is scoped to an organization.
// AI-hint: Business logic method to determine if role has organization context.
func (r *Role) IsOrganizationScoped() bool {
	return r.OrganizationID != ""
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
	GetByOrganizationID(ctx interface{}, organizationID string) ([]*Role, error)
	GetByNameAndOrganization(ctx interface{}, name, organizationID string) (*Role, error)
}

// Service defines the business operations for role management.
// AI-hint: Domain service interface for complex business operations
// that require coordination between multiple entities or external services.
type Service interface {
	CreateRole(ctx interface{}, name, organizationID string, createdByUserID string) (*Role, error)
	GetRole(ctx interface{}, id string) (*Role, error)
	UpdateRole(ctx interface{}, id, name string, updatedByUserID string) (*Role, error)
	UpdateRoleOrganization(ctx interface{}, id, organizationID string, updatedByUserID string) (*Role, error)
	DeleteRole(ctx interface{}, id string, deletedByUserID string) error
	ListRoles(ctx interface{}) ([]*Role, error)
	ListRolesByOrganization(ctx interface{}, organizationID string) ([]*Role, error)
	EnsurePredefinedRoles(ctx interface{}, organizationID string) error
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
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrRoleNotInOrganization     = errors.New("role not in specified organization")
)
