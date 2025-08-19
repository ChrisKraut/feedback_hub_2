package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"

	orgapp "feedback_hub_2/internal/organization/application"
	roleapp "feedback_hub_2/internal/role/application"
	userapp "feedback_hub_2/internal/user/application"
)

// BootstrapService handles system initialization and setup.
// AI-hint: System bootstrap service for initializing required data and first user.
// Ensures predefined roles exist and creates initial Super User from environment variables.
type BootstrapService struct {
	roleService         *roleapp.RoleService
	userService         *userapp.UserService
	organizationService *orgapp.OrganizationService
}

// NewBootstrapService creates a new BootstrapService instance.
// AI-hint: Factory method for bootstrap service with dependency injection.
func NewBootstrapService(roleService *roleapp.RoleService, userService *userapp.UserService, organizationService *orgapp.OrganizationService) *BootstrapService {
	return &BootstrapService{
		roleService:         roleService,
		userService:         userService,
		organizationService: organizationService,
	}
}

// Initialize performs system initialization tasks.
// AI-hint: System setup method that ensures required data exists and creates initial Super User.
// Safe to run multiple times - will not duplicate data.
func (s *BootstrapService) Initialize(ctx context.Context) error {
	log.Println("Starting system initialization...")

	// Ensure default organization exists
	if err := s.ensureDefaultOrganization(ctx); err != nil {
		return fmt.Errorf("failed to ensure default organization: %w", err)
	}
	log.Println("Default organization ensured")

	// Ensure predefined roles exist
	if err := s.roleService.EnsurePredefinedRoles(ctx); err != nil {
		return fmt.Errorf("failed to ensure predefined roles: %w", err)
	}
	log.Println("Predefined roles ensured")

	// Create initial Super User if specified in environment variables
	if err := s.createInitialSuperUser(ctx); err != nil {
		return fmt.Errorf("failed to create initial Super User: %w", err)
	}

	log.Println("System initialization completed successfully")
	return nil
}

// createInitialSuperUser creates the initial Super User from environment variables.
// AI-hint: Initial Super User creation for system bootstrapping.
// Uses SUPER_USER_EMAIL and SUPER_USER_NAME environment variables.
func (s *BootstrapService) createInitialSuperUser(ctx context.Context) error {
	email := os.Getenv("SUPER_USER_EMAIL")
	name := os.Getenv("SUPER_USER_NAME")

	// If no Super User environment variables are set, skip creation
	if email == "" || name == "" {
		log.Println("No Super User environment variables set (SUPER_USER_EMAIL, SUPER_USER_NAME), skipping initial Super User creation")
		return nil
	}

	// Check if a Super User already exists
	superUserRole, err := s.roleService.GetRoleByName(ctx, "Super User")
	if err != nil {
		return fmt.Errorf("failed to get Super User role: %w", err)
	}

	// Try to create the Super User
	superUser, err := s.userService.CreateSuperUser(ctx, email, name)
	if err != nil {
		// If Super User already exists, that's fine
		if err.Error() == "Super User already exists" {
			log.Println("Super User already exists, skipping creation")
			return nil
		}
		if err.Error() == "email already exists" {
			log.Printf("User with email %s already exists, skipping Super User creation", email)
			return nil
		}
		return fmt.Errorf("failed to create Super User: %w", err)
	}

	log.Printf("Created initial Super User: %s (%s) with role %s", superUser.Name, superUser.Email, superUserRole.Name)
	return nil
}

// ensureDefaultOrganization ensures that a default organization exists.
// AI-hint: Creates a default organization if none exists, ensuring the system
// has at least one organization for new users and existing data.
func (s *BootstrapService) ensureDefaultOrganization(ctx context.Context) error {
	// Try to get the default organization by slug
	defaultOrg, err := s.organizationService.GetOrganizationBySlug(ctx, "default")
	if err == nil {
		log.Println("Default organization already exists")
		return nil
	}

	// If not found, create the default organization
	log.Println("Creating default organization...")
	defaultOrg, err = s.organizationService.CreateOrganization(ctx, "Default Organization", "default", "Default organization for new installations", nil)
	if err != nil {
		return fmt.Errorf("failed to create default organization: %w", err)
	}

	log.Printf("Created default organization: %s (%s)", defaultOrg.Name, defaultOrg.Slug)
	return nil
}
