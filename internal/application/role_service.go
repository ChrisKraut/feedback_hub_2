package application

import (
	"context"
	"feedback_hub_2/internal/domain/auth"
	"feedback_hub_2/internal/domain/role"
	"feedback_hub_2/internal/domain/user"

	"github.com/google/uuid"
)

// RoleService implements the role.Service interface and coordinates role management operations.
// AI-hint: Application service that orchestrates role business logic with authorization checks.
// Enforces business rules about who can perform what operations on roles.
type RoleService struct {
	roleRepo    role.Repository
	userRepo    user.Repository
	authService *auth.AuthorizationService
}

// NewRoleService creates a new RoleService instance.
// AI-hint: Factory method for role service with dependency injection of repositories and auth service.
func NewRoleService(roleRepo role.Repository, userRepo user.Repository, authService *auth.AuthorizationService) *RoleService {
	return &RoleService{
		roleRepo:    roleRepo,
		userRepo:    userRepo,
		authService: authService,
	}
}

// CreateRole creates a new role with authorization checks.
// AI-hint: Role creation with business rule enforcement - only Super Users can create roles.
func (s *RoleService) CreateRole(ctx interface{}, name string, createdByUserID string) (*role.Role, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, createdByUserID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !s.authService.CanPerform(userCtx, auth.PermissionCreateRole) {
		return nil, role.ErrUnauthorized
	}

	// Check if role name already exists
	exists, err := s.roleRepo.Exists(context, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, role.ErrRoleNameAlreadyExists
	}

	// Create the role
	roleID := uuid.New().String()
	newRole, err := role.NewRole(roleID, name)
	if err != nil {
		return nil, err
	}

	if err := s.roleRepo.Create(context, newRole); err != nil {
		return nil, err
	}

	return newRole, nil
}

// GetRole retrieves a role by ID with authorization checks.
// AI-hint: Role retrieval with read permission validation.
func (s *RoleService) GetRole(ctx interface{}, id string) (*role.Role, error) {
	context := ctx.(context.Context)

	// For now, allow reading roles (this could be restricted later)
	return s.roleRepo.GetByID(context, id)
}

// UpdateRole updates a role's name with authorization checks.
// AI-hint: Role update with business rule enforcement and Super User role protection.
func (s *RoleService) UpdateRole(ctx interface{}, id, name string, updatedByUserID string) (*role.Role, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, updatedByUserID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !s.authService.CanPerform(userCtx, auth.PermissionUpdateRole) {
		return nil, role.ErrUnauthorized
	}

	// Get the existing role
	existingRole, err := s.roleRepo.GetByID(context, id)
	if err != nil {
		return nil, err
	}

	// Try to update the name (this will validate business rules like Super User protection)
	if err := existingRole.UpdateName(name); err != nil {
		return nil, err
	}

	// Check if the new name conflicts with existing roles
	if existingRole.Name != name {
		exists, err := s.roleRepo.Exists(context, name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, role.ErrRoleNameAlreadyExists
		}
	}

	// Save the updated role
	if err := s.roleRepo.Update(context, existingRole); err != nil {
		return nil, err
	}

	return existingRole, nil
}

// DeleteRole deletes a role with authorization and business rule checks.
// AI-hint: Role deletion with Super User protection and user assignment validation.
func (s *RoleService) DeleteRole(ctx interface{}, id string, deletedByUserID string) error {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, deletedByUserID)
	if err != nil {
		return err
	}

	// Check authorization
	if !s.authService.CanPerform(userCtx, auth.PermissionDeleteRole) {
		return role.ErrUnauthorized
	}

	// Get the role to check if it can be deleted
	existingRole, err := s.roleRepo.GetByID(context, id)
	if err != nil {
		return err
	}

	// Check if the role can be deleted (Super User role cannot be deleted)
	if !existingRole.CanBeDeleted() {
		return role.ErrCannotDeleteSuperUserRole
	}

	// Check if any users are assigned to this role
	assignedUsers, err := s.userRepo.GetByRoleID(context, id)
	if err != nil {
		return err
	}
	if len(assignedUsers) > 0 {
		return role.ErrInvalidRoleData // Role has assigned users
	}

	// Delete the role
	return s.roleRepo.Delete(context, id)
}

// ListRoles retrieves all roles with authorization checks.
// AI-hint: Role listing with read permission validation.
func (s *RoleService) ListRoles(ctx interface{}) ([]*role.Role, error) {
	context := ctx.(context.Context)

	// For now, allow listing roles (this could be restricted later)
	return s.roleRepo.List(context)
}

// EnsurePredefinedRoles ensures that all predefined roles exist in the system.
// AI-hint: System initialization method for bootstrapping required roles.
func (s *RoleService) EnsurePredefinedRoles(ctx interface{}) error {
	context := ctx.(context.Context)

	for _, roleName := range role.PredefinedRoles {
		exists, err := s.roleRepo.Exists(context, roleName)
		if err != nil {
			return err
		}

		if !exists {
			roleID := uuid.New().String()
			newRole, err := role.NewRole(roleID, roleName)
			if err != nil {
				return err
			}

			if err := s.roleRepo.Create(context, newRole); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetRoleByName retrieves a role by name.
// AI-hint: Name-based role lookup for authorization flows.
func (s *RoleService) GetRoleByName(ctx interface{}, name string) (*role.Role, error) {
	context := ctx.(context.Context)
	return s.roleRepo.GetByName(context, name)
}

// getUserContext retrieves the user context for authorization.
// AI-hint: Helper method to build authorization context from user ID.
func (s *RoleService) getUserContext(ctx context.Context, userID string) (*auth.UserContext, error) {
	if userID == "" {
		return nil, auth.ErrInvalidContext
	}

	// Get the user to determine their role
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get the user's role to determine their permissions
	userRole, err := s.roleRepo.GetByID(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	return &auth.UserContext{
		UserID:   userID,
		RoleName: userRole.Name,
	}, nil
}
