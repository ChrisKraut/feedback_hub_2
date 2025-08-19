package auth

import (
	"errors"
	"feedback_hub_2/internal/role/domain"
)

// Permission represents an action that can be performed in the system.
// AI-hint: Enum-like type for defining specific system permissions.
// Used for role-based access control validation.
type Permission string

const (
	// Role management permissions
	PermissionCreateRole Permission = "role:create"
	PermissionReadRole   Permission = "role:read"
	PermissionUpdateRole Permission = "role:update"
	PermissionDeleteRole Permission = "role:delete"

	// User management permissions
	PermissionCreateUser Permission = "user:create"
	PermissionReadUser   Permission = "user:read"
	PermissionUpdateUser Permission = "user:update"
	PermissionDeleteUser Permission = "user:delete"

	// Special permissions
	PermissionCreateAnyUser     Permission = "user:create_any"         // Can create users with any role
	PermissionCreateContributor Permission = "user:create_contributor" // Can only create contributor users
)

// UserContext represents the current user making a request.
// AI-hint: Security context carrying user identity and role information
// for authorization decisions throughout the system.
type UserContext struct {
	UserID   string
	RoleName string
}

// AuthorizationService handles permission checking based on roles.
// AI-hint: Domain service for authorization logic that encapsulates
// business rules about who can perform what actions.
type AuthorizationService struct{}

// NewAuthorizationService creates a new authorization service.
// AI-hint: Factory method for the authorization service.
func NewAuthorizationService() *AuthorizationService {
	return &AuthorizationService{}
}

// CanPerform checks if a user has permission to perform a specific action.
// AI-hint: Core authorization method implementing role-based access control.
// Returns true if the user's role allows the requested permission.
func (s *AuthorizationService) CanPerform(userCtx *UserContext, permission Permission) bool {
	if userCtx == nil {
		return false
	}

	// Super User can do anything
	if userCtx.RoleName == domain.SuperUserRoleName {
		return true
	}

	switch userCtx.RoleName {
	case "Product Owner":
		return s.canProductOwnerPerform(permission)
	case "Contributor":
		return s.canContributorPerform(permission)
	default:
		return false
	}
}

// CanCreateUserWithRole checks if a user can create another user with a specific role.
// AI-hint: Specialized authorization check for user creation with role assignment.
// Implements business rule that Product Owners can only create Contributors.
func (s *AuthorizationService) CanCreateUserWithRole(userCtx *UserContext, targetRoleName string) bool {
	if userCtx == nil {
		return false
	}

	// Super User can create users with any role
	if userCtx.RoleName == domain.SuperUserRoleName {
		return true
	}

	// Product Owner can only create Contributors
	if userCtx.RoleName == "Product Owner" && targetRoleName == "Contributor" {
		return true
	}

	return false
}

// ValidateUserContext ensures the user context is valid for authorization.
// AI-hint: Input validation for authorization context to prevent security bypasses.
func (s *AuthorizationService) ValidateUserContext(userCtx *UserContext) error {
	if userCtx == nil {
		return errors.New("user context is required")
	}
	if userCtx.UserID == "" {
		return errors.New("user ID is required")
	}
	if userCtx.RoleName == "" {
		return errors.New("role name is required")
	}
	return nil
}

// canProductOwnerPerform defines what Product Owners can do.
// AI-hint: Business rules for Product Owner role permissions.
func (s *AuthorizationService) canProductOwnerPerform(permission Permission) bool {
	switch permission {
	case PermissionReadRole, PermissionReadUser:
		return true
	case PermissionCreateUser, PermissionUpdateUser, PermissionDeleteUser:
		return true
	case PermissionCreateContributor:
		return true
	default:
		return false
	}
}

// canContributorPerform defines what Contributors can do.
// AI-hint: Business rules for Contributor role permissions.
func (s *AuthorizationService) canContributorPerform(permission Permission) bool {
	switch permission {
	case PermissionReadRole, PermissionReadUser:
		return true
	default:
		return false
	}
}

// Error types for the auth domain.
// AI-hint: Domain-specific errors for authorization failures.
var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidContext = errors.New("invalid user context")
)
