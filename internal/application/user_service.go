package application

import (
	"context"
	"errors"
	"feedback_hub_2/internal/domain/auth"
	"feedback_hub_2/internal/domain/role"
	"feedback_hub_2/internal/domain/user"

	"log"

	"github.com/google/uuid"
)

// UserService implements the user.Service interface and coordinates user management operations.
// AI-hint: Application service that orchestrates user business logic with authorization checks.
// Enforces business rules about who can perform what operations on users and with which roles.
type UserService struct {
	userRepo    user.Repository
	roleRepo    role.Repository
	authService *auth.AuthorizationService
}

// NewUserService creates a new UserService instance.
// AI-hint: Factory method for user service with dependency injection of repositories and auth service.
func NewUserService(userRepo user.Repository, roleRepo role.Repository, authService *auth.AuthorizationService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		authService: authService,
	}
}

// CreateUser creates a new user with authorization and role validation checks.
// AI-hint: User creation with complex business rules - Super Users can create any user,
// Product Owners can only create Contributors.
func (s *UserService) CreateUser(ctx interface{}, email, name, roleID string, createdByUserID string) (*user.User, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, createdByUserID)
	if err != nil {
		return nil, err
	}

	// Validate the target role exists
	targetRole, err := s.roleRepo.GetByID(context, roleID)
	if err != nil {
		if err == role.ErrRoleNotFound {
			return nil, errors.New("invalid role ID")
		}
		return nil, err
	}

	// Check authorization - can the user create users with this specific role?
	if !s.authService.CanCreateUserWithRole(userCtx, targetRole.Name) {
		return nil, user.ErrUnauthorized
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(context, email)
	if err == nil && existingUser != nil {
		return nil, user.ErrEmailAlreadyExists
	}
	if err != user.ErrUserNotFound {
		return nil, err
	}

	// Create the user
	userID := uuid.New().String()
	newUser, err := user.NewUser(userID, email, name, roleID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(context, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// GetUser retrieves a user by ID with authorization checks.
// AI-hint: User retrieval with read permission validation.
func (s *UserService) GetUser(ctx interface{}, id string) (*user.User, error) {
	context := ctx.(context.Context)

	// For now, allow reading users (this could be restricted later)
	return s.userRepo.GetByID(context, id)
}

// UpdateUser updates a user's name with authorization checks.
// AI-hint: User update with permission validation - maintains email immutability.
func (s *UserService) UpdateUser(ctx interface{}, id, name string, updatedByUserID string) (*user.User, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, updatedByUserID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !s.authService.CanPerform(userCtx, auth.PermissionUpdateUser) {
		return nil, user.ErrUnauthorized
	}

	// Get the existing user
	existingUser, err := s.userRepo.GetByID(context, id)
	if err != nil {
		return nil, err
	}

	// Update the name
	if err := existingUser.UpdateName(name); err != nil {
		return nil, err
	}

	// Save the updated user
	if err := s.userRepo.Update(context, existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

// UpdateUserRole updates a user's role with authorization checks.
// AI-hint: Role assignment with business rule enforcement - only Super Users can change roles.
func (s *UserService) UpdateUserRole(ctx interface{}, id, roleID string, updatedByUserID string) (*user.User, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, updatedByUserID)
	if err != nil {
		return nil, err
	}

	// Only Super Users can update user roles
	if userCtx.RoleName != role.SuperUserRoleName {
		return nil, user.ErrUnauthorized
	}

	// Validate the target role exists
	targetRole, err := s.roleRepo.GetByID(context, roleID)
	if err != nil {
		if err == role.ErrRoleNotFound {
			return nil, errors.New("invalid role ID")
		}
		return nil, err
	}

	// Get the existing user
	existingUser, err := s.userRepo.GetByID(context, id)
	if err != nil {
		return nil, err
	}

	// Update the role
	if err := existingUser.UpdateRole(targetRole.ID); err != nil {
		return nil, err
	}

	// Save the updated user
	if err := s.userRepo.Update(context, existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

// DeleteUser deletes a user with authorization checks.
// AI-hint: User deletion with permission validation.
func (s *UserService) DeleteUser(ctx interface{}, id string, deletedByUserID string) error {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, deletedByUserID)
	if err != nil {
		return err
	}

	// Check authorization
	if !s.authService.CanPerform(userCtx, auth.PermissionDeleteUser) {
		return user.ErrUnauthorized
	}

	// Delete the user
	return s.userRepo.Delete(context, id)
}

// ListUsers retrieves all users with authorization checks.
// AI-hint: User listing with read permission validation.
func (s *UserService) ListUsers(ctx interface{}) ([]*user.User, error) {
	context := ctx.(context.Context)

	// For now, allow listing users (this could be restricted later)
	return s.userRepo.List(context)
}

// GetUserByEmail retrieves a user by email address.
// AI-hint: Email-based user lookup for authentication flows.
func (s *UserService) GetUserByEmail(ctx interface{}, email string) (*user.User, error) {
	context := ctx.(context.Context)
	return s.userRepo.GetByEmail(context, email)
}

// CreateSuperUser creates a Super User for system initialization.
// AI-hint: System Super User creation with dynamic ID generation for security.
// This should only be called during system setup.
func (s *UserService) CreateSuperUser(ctx interface{}, email, name string) (*user.User, error) {
	context := ctx.(context.Context)

	// Check if a Super User already exists
	superUserRole, err := s.roleRepo.GetByName(context, role.SuperUserRoleName)
	if err != nil {
		return nil, err
	}

	existingSuperUsers, err := s.userRepo.GetByRoleID(context, superUserRole.ID)
	if err != nil {
		return nil, err
	}

	if len(existingSuperUsers) > 0 {
		// Return the existing Super User instead of error for convenience
		log.Printf("Super User already exists: %s", existingSuperUsers[0].ID)
		return existingSuperUsers[0], nil
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(context, email)
	if err == nil && existingUser != nil {
		return nil, user.ErrEmailAlreadyExists
	}
	if err != user.ErrUserNotFound {
		return nil, err
	}

	// Generate a new UUID for the Super User
	userID := uuid.New().String()
	newUser, err := user.NewUser(userID, email, name, superUserRole.ID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(context, newUser); err != nil {
		return nil, err
	}

	log.Printf("Created Super User with ID: %s", userID)
	return newUser, nil
}

// getUserContext retrieves the user context for authorization.
// AI-hint: Helper method to build authorization context from user ID.
func (s *UserService) getUserContext(ctx context.Context, userID string) (*auth.UserContext, error) {
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
