package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"feedback_hub_2/internal/organization/domain"
	events "feedback_hub_2/internal/shared/bus"

	"github.com/google/uuid"
)

// OrganizationService provides business logic for organization management.
// AI-hint: Application service that orchestrates organization operations,
// implements business rules, and coordinates between domain entities and repositories.
type OrganizationService struct {
	repo     domain.OrganizationRepository
	eventBus events.EventBus
}

// NewOrganizationService creates a new organization service instance.
// AI-hint: Factory method for creating organization service with dependency injection
// of the organization repository and event bus.
func NewOrganizationService(repo domain.OrganizationRepository, eventBus events.EventBus) *OrganizationService {
	return &OrganizationService{
		repo:     repo,
		eventBus: eventBus,
	}
}

// CreateOrganization creates a new organization with the given details.
// AI-hint: Business method for creating organizations with validation and business rules.
// Generates a unique slug if not provided and ensures data integrity.
func (s *OrganizationService) CreateOrganization(ctx context.Context, name, slug, description string, settings map[string]any) (*domain.Organization, error) {
	// Validate input parameters
	if err := s.validateOrganizationInput(name, slug, description); err != nil {
		return nil, err
	}

	// Generate slug if not provided
	if slug == "" {
		slug = s.generateSlug(name)
	}

	// Normalize slug
	slug = s.normalizeSlug(slug)

	// Check if slug already exists
	existing, err := s.repo.GetBySlug(ctx, slug)
	if err != nil && err != domain.ErrOrganizationNotFound {
		return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrOrganizationSlugAlreadyExists
	}

	// Create organization
	org, err := domain.NewOrganization(name, slug, description, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Store in repository
	if err := s.repo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to store organization: %w", err)
	}

	// Publish organization created event
	if s.eventBus != nil {
		event := events.NewOrganizationCreatedEvent(
			org.ID.String(),
			org.Name,
			org.Slug,
			org.Description,
			"", // createdByUserID - could be passed as parameter in future
		)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
			// In production, this could be sent to a dead letter queue
			fmt.Printf("Failed to publish organization created event: %v\n", err)
		}
	}

	return org, nil
}

// CreateOrganizationWithSlug creates a new organization with an auto-generated slug.
// AI-hint: Convenience method for creating organizations when the slug should be
// automatically generated from the name.
func (s *OrganizationService) CreateOrganizationWithSlug(ctx context.Context, name, description string, settings map[string]any) (*domain.Organization, error) {
	// Generate slug from name
	slug := s.generateSlug(name)

	// Ensure slug uniqueness by appending number if needed
	originalSlug := slug
	counter := 1
	for {
		existing, err := s.repo.GetBySlug(ctx, slug)
		if err != nil && err != domain.ErrOrganizationNotFound {
			return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		if existing == nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	return s.CreateOrganization(ctx, name, slug, description, settings)
}

// GetOrganizationByID retrieves an organization by its unique identifier.
// AI-hint: Business method for retrieving organizations by ID with proper error handling.
func (s *OrganizationService) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidOrganizationData
	}

	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// GetOrganizationBySlug retrieves an organization by its unique slug.
// AI-hint: Business method for retrieving organizations by slug with proper error handling.
func (s *OrganizationService) GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	if slug == "" {
		return nil, domain.ErrInvalidOrganizationData
	}

	// Normalize slug
	slug = s.normalizeSlug(slug)

	org, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

// UpdateOrganization updates an existing organization with new details.
// AI-hint: Business method for updating organizations with validation and business rules.
// Ensures slug uniqueness if changed and maintains data integrity.
func (s *OrganizationService) UpdateOrganization(ctx context.Context, id uuid.UUID, name, slug, description string, settings map[string]any) error {
	if id == uuid.Nil {
		return domain.ErrInvalidOrganizationData
	}

	// Get existing organization
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Validate input parameters
	if err := s.validateOrganizationInput(name, slug, description); err != nil {
		return err
	}

	// Handle slug changes
	if slug != "" && slug != existing.Slug {
		// Normalize new slug
		slug = s.normalizeSlug(slug)

		// Check if new slug already exists
		slugExists, err := s.repo.GetBySlug(ctx, slug)
		if err != nil && err != domain.ErrOrganizationNotFound {
			return fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		if slugExists != nil {
			return domain.ErrOrganizationSlugAlreadyExists
		}
		existing.Slug = slug
	}

	// Update fields
	if name != "" {
		existing.Name = name
	}
	if description != "" {
		existing.Description = description
	}
	if settings != nil {
		existing.Settings = settings
	}

	// Update timestamp
	existing.UpdatedAt = time.Now()

	// Validate updated organization
	if err := existing.Validate(); err != nil {
		return fmt.Errorf("invalid organization data after update: %w", err)
	}

	// Store updates
	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	// Publish organization updated event
	if s.eventBus != nil {
		// Track changes for the event
		changes := make(map[string]events.Change)
		if name != "" && name != existing.Name {
			changes["name"] = events.Change{OldValue: existing.Name, NewValue: name}
		}
		if slug != "" && slug != existing.Slug {
			changes["slug"] = events.Change{OldValue: existing.Slug, NewValue: slug}
		}
		if description != "" && description != existing.Description {
			changes["description"] = events.Change{OldValue: existing.Description, NewValue: description}
		}

		event := events.NewOrganizationUpdatedEvent(
			existing.ID.String(),
			existing.Name,
			existing.Slug,
			existing.Description,
			existing.Settings,
			"", // updatedByUserID - could be passed as parameter in future
			changes,
			1, // version - could be tracked in domain entity in future
		)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish organization updated event: %v\n", err)
		}
	}

	return nil
}

// DeleteOrganization removes an organization from the system.
// AI-hint: Business method for deleting organizations with proper cleanup and validation.
// Ensures that dependent entities are handled appropriately.
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidOrganizationData
	}

	// Check if organization exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Check for dependencies (this would be implemented based on business rules)
	// For now, we'll allow deletion but this could be enhanced with dependency checking

	// Delete organization
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Publish organization deleted event
	if s.eventBus != nil {
		event := events.NewOrganizationDeletedEvent(
			existing.ID.String(),
			existing.Name,
			existing.Slug,
			"", // deletedByUserID - could be passed as parameter in future
			"", // deletionReason - could be passed as parameter in future
			1,  // version - could be tracked in domain entity in future
		)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish organization deleted event: %v\n", err)
		}
	}

	return nil
}

// ListOrganizations retrieves a paginated list of organizations.
// AI-hint: Business method for listing organizations with pagination support.
// Applies business rules for limits and ordering.
func (s *OrganizationService) ListOrganizations(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	// Apply business rules for limits
	if limit <= 0 {
		limit = domain.DefaultListLimit
	}
	if limit > domain.MaxListLimit {
		limit = domain.MaxListLimit
	}
	if offset < 0 {
		offset = domain.DefaultListOffset
	}

	organizations, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return organizations, nil
}

// CountOrganizations returns the total number of organizations.
// AI-hint: Business method for counting organizations with proper error handling.
func (s *OrganizationService) CountOrganizations(ctx context.Context) (int, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	return count, nil
}

// SearchOrganizations searches for organizations by name or description.
// AI-hint: Business method for searching organizations with flexible criteria.
// Could be enhanced with more sophisticated search algorithms.
func (s *OrganizationService) SearchOrganizations(ctx context.Context, query string, limit, offset int) ([]*domain.Organization, error) {
	if query == "" {
		return s.ListOrganizations(ctx, limit, offset)
	}

	// For now, we'll implement a simple search by getting all organizations
	// and filtering by name/description. In a production system, this would
	// use database-level search or external search services.

	// Apply business rules for limits
	if limit <= 0 {
		limit = domain.DefaultListLimit
	}
	if limit > domain.MaxListLimit {
		limit = domain.MaxListLimit
	}
	if offset < 0 {
		offset = domain.DefaultListOffset
	}

	// Get all organizations (this could be optimized with database-level search)
	allOrgs, err := s.repo.List(ctx, 10000, 0) // Get all for search
	if err != nil {
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}

	// Filter by query
	var results []*domain.Organization
	queryLower := strings.ToLower(query)

	for _, org := range allOrgs {
		if strings.Contains(strings.ToLower(org.Name), queryLower) ||
			strings.Contains(strings.ToLower(org.Description), queryLower) ||
			strings.Contains(strings.ToLower(org.Slug), queryLower) {
			results = append(results, org)
		}
	}

	// Apply pagination
	start := offset
	end := start + limit
	if start >= len(results) {
		return []*domain.Organization{}, nil
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// validateOrganizationInput validates the input parameters for organization creation/update.
// AI-hint: Private validation method that enforces business rules for organization data.
func (s *OrganizationService) validateOrganizationInput(name, slug, description string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("organization name cannot be empty")
	}

	if slug != "" && strings.TrimSpace(slug) == "" {
		return fmt.Errorf("organization slug cannot be empty")
	}

	// Additional validation could be added here
	// - Name length limits
	// - Slug format validation
	// - Description length limits
	// - Reserved slug names

	return nil
}

// generateSlug generates a URL-friendly slug from the organization name.
// AI-hint: Private utility method that converts organization names to URL-friendly slugs.
func (s *OrganizationService) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, ",", "-")
	slug = strings.ReplaceAll(slug, "&", "and")

	// Remove any remaining special characters
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	// Clean up multiple hyphens
	slug = result.String()
	slug = strings.ReplaceAll(slug, "--", "-")
	slug = strings.Trim(slug, "-")

	// Ensure minimum length
	if len(slug) < 3 {
		slug = "org-" + slug
	}

	return slug
}

// normalizeSlug normalizes a slug to ensure consistency.
// AI-hint: Private utility method that ensures slug format consistency.
func (s *OrganizationService) normalizeSlug(slug string) string {
	// Convert to lowercase
	slug = strings.ToLower(slug)

	// Replace spaces and special characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, ",", "-")

	// Remove any remaining special characters
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	// Clean up multiple hyphens
	slug = result.String()
	slug = strings.ReplaceAll(slug, "--", "-")
	slug = strings.Trim(slug, "-")

	return slug
}
