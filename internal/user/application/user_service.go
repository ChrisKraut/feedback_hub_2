package application

import (
	"context"
	"errors"
	"feedback_hub_2/internal/shared/auth"
	events "feedback_hub_2/internal/shared/bus"
	"feedback_hub_2/internal/shared/queries"
	"feedback_hub_2/internal/user/domain"

	"log"

	"github.com/google/uuid"
)

// UserService implements the user.Service interface and coordinates user management operations.
// AI-hint: Application service that orchestrates user business logic with authorization checks.
// Enforces business rules about who can perform what operations on users and with which roles.
// Uses domain events for cross-domain communication instead of direct dependencies.
type UserService struct {
	userRepo       domain.Repository
	roleQueries    queries.RoleQueries
	authService    *auth.AuthorizationService
	eventPublisher events.EventPublisher
}

// NewUserService creates a new UserService instance.
// AI-hint: Factory method for user service with dependency injection of repositories, auth service, and event publisher.
func NewUserService(userRepo domain.Repository, roleQueries queries.RoleQueries, authService *auth.AuthorizationService, eventPublisher events.EventPublisher) *UserService {
	return &UserService{
		userRepo:       userRepo,
		roleQueries:    roleQueries,
		authService:    authService,
		eventPublisher: eventPublisher,
	}
}

// CreateUser creates a new user with authorization and role validation checks.
// AI-hint: User creation with complex business rules - Super Users can create any user,
// Product Owners can only create Contributors, with organization scoping.
func (s *UserService) CreateUser(ctx interface{}, email, name, roleID, organizationID string, createdByUserID string) (*domain.User, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, createdByUserID)
	if err != nil {
		return nil, err
	}

	// Validate the target role exists using shared queries
	targetRole, err := s.roleQueries.GetRoleByID(context, roleID)
	if err != nil {
		return nil, errors.New("invalid role ID")
	}

	// Check authorization - can the user create users with this specific role?
	if !s.authService.CanCreateUserWithRole(userCtx, targetRole.Name) {
		return nil, domain.ErrUnauthorized
	}

	// Check if email already exists in the organization
	existingUser, err := s.userRepo.GetByEmailAndOrganization(context, email, organizationID)
	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	if err != domain.ErrUserNotFound {
		return nil, err
	}

	// Create the user with organization scoping
	userID := uuid.New().String()
	newUser, err := domain.NewUser(userID, email, name, roleID, organizationID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(context, newUser); err != nil {
		return nil, err
	}

	// Publish domain event for user creation
	userCreatedEvent := events.NewUserCreatedEvent(newUser.ID, newUser.Email, newUser.Name, newUser.RoleID, targetRole.Name)
	if err := s.eventPublisher.PublishEvent(context, userCreatedEvent); err != nil {
		log.Printf("Warning: failed to publish user created event: %v", err)
		// Don't fail the operation if event publishing fails
	}

	return newUser, nil
}

// GetUser retrieves a user by ID with authorization checks.
// AI-hint: User retrieval with read permission validation.
func (s *UserService) GetUser(ctx interface{}, id string) (*domain.User, error) {
	context := ctx.(context.Context)

	// For now, allow reading users (this could be restricted later)
	return s.userRepo.GetByID(context, id)
}

// UpdateUser updates a user's name with authorization checks.
// AI-hint: User update with permission validation - maintains email immutability.
func (s *UserService) UpdateUser(ctx interface{}, id, name string, updatedByUserID string) (*domain.User, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, updatedByUserID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !s.authService.CanPerform(userCtx, auth.PermissionUpdateUser) {
		return nil, domain.ErrUnauthorized
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

	// Publish domain event for user update
	userUpdatedEvent := events.NewUserUpdatedEvent(existingUser.ID, existingUser.Name, 2) // Assuming version 2 for now
	if err := s.eventPublisher.PublishEvent(context, userUpdatedEvent); err != nil {
		log.Printf("Warning: failed to publish user updated event: %v", err)
		// Don't fail the operation if event publishing fails
	}

	return existingUser, nil
}

// UpdateUserRole updates a user's role with authorization checks.
// AI-hint: Role assignment with business rule enforcement - only Super Users can change roles.
func (s *UserService) UpdateUserRole(ctx interface{}, id, roleID string, updatedByUserID string) (*domain.User, error) {
	context := ctx.(context.Context)

	// Get the user context for authorization
	userCtx, err := s.getUserContext(context, updatedByUserID)
	if err != nil {
		return nil, err
	}

	// Only Super Users can update user roles
	if userCtx.RoleName != "Super User" {
		return nil, domain.ErrUnauthorized
	}

	// Validate the target role exists using shared queries
	targetRole, err := s.roleQueries.GetRoleByID(context, roleID)
	if err != nil {
		return nil, errors.New("invalid role ID")
	}

	// Get the existing user
	existingUser, err := s.userRepo.GetByID(context, id)
	if err != nil {
		return nil, err
	}

	// Store old role information for event
	oldRoleID := existingUser.RoleID
	oldRole, err := s.roleQueries.GetRoleByID(context, oldRoleID)
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

	// Publish domain event for user role update
	userRoleUpdatedEvent := events.NewUserRoleUpdatedEvent(existingUser.ID, oldRoleID, targetRole.ID, oldRole.Name, targetRole.Name, 3) // Assuming version 3 for now
	if err := s.eventPublisher.PublishEvent(context, userRoleUpdatedEvent); err != nil {
		log.Printf("Warning: failed to publish user role updated event: %v", err)
		// Don't fail the operation if event publishing fails
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
		return domain.ErrUnauthorized
	}

	// Delete the user
	return s.userRepo.Delete(context, id)
}

// ListUsers retrieves all users with authorization checks.
// AI-hint: User listing with read permission validation.
func (s *UserService) ListUsers(ctx interface{}) ([]*domain.User, error) {
	context := ctx.(context.Context)

	// For now, allow listing users (this could be restricted later)
	return s.userRepo.List(context)
}

// GetUserByEmail retrieves a user by email address.
// AI-hint: Email-based user lookup for authentication flows.
func (s *UserService) GetUserByEmail(ctx interface{}, email string) (*domain.User, error) {
	context := ctx.(context.Context)
	return s.userRepo.GetByEmail(context, email)
}

// CreateSuperUser creates a Super User for system initialization.
// AI-hint: System Super User creation with dynamic ID generation for security.
// This should only be called during system setup.
func (s *UserService) CreateSuperUser(ctx interface{}, email, name string) (*domain.User, error) {
	context := ctx.(context.Context)

	// Check if a Super User already exists
	superUserRole, err := s.roleQueries.GetRoleByName(context, "Super User")
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
		return nil, domain.ErrEmailAlreadyExists
	}
	if err != domain.ErrUserNotFound {
		return nil, err
	}

	// Generate a new UUID for the Super User
	userID := uuid.New().String()
	newUser, err := domain.NewUserWithoutOrganization(userID, email, name, superUserRole.ID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(context, newUser); err != nil {
		return nil, err
	}

	log.Printf("Created Super User with ID: %s", userID)
	return newUser, nil
}

// CreateUserWithPassword creates a user with password authentication.
// AI-hint: User creation method for password-based authentication with proper validation and organization scoping.
func (s *UserService) CreateUserWithPassword(ctx interface{}, userID, email, name, passwordHash, roleID, organizationID string) (*domain.User, error) {
	context := ctx.(context.Context)

	// Check if email already exists in the organization
	existingUser, err := s.userRepo.GetByEmailAndOrganization(context, email, organizationID)
	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	if err != domain.ErrUserNotFound {
		return nil, err
	}

	// Create user with password and organization scoping
	newUser, err := domain.NewUserWithPassword(userID, email, name, passwordHash, roleID, organizationID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(context, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// getUserContext retrieves the user context for authorization.
// AI-hint: Helper method to build authorization context from user ID.
func (s *UserService) getUserContext(ctx context.Context, userID string) (*auth.UserContext, error) {
	if userID == "" {
		return nil, auth.ErrInvalidContext
	}

	// Get the user to determine their role
	userEntity, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get the user's role to determine their permissions
	userRole, err := s.roleQueries.GetRoleByID(ctx, userEntity.RoleID)
	if err != nil {
		return nil, err
	}

	return &auth.UserContext{
		UserID:   userID,
		RoleName: userRole.Name,
	}, nil
}
