package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Organization represents a business entity that can contain users, roles, and ideas.
// AI-hint: Core domain entity for multi-tenant functionality. Each organization operates
// in complete isolation with its own users, roles, and ideas. This entity enforces
// business rules and validation for organization management.
type Organization struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Description string          `json:"description"`
	Settings    map[string]any  `json:"settings"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Organization constants for validation
const (
	MaxNameLength = 255
	MaxSlugLength = 100
)

// slugRegex defines the valid format for organization slugs
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NewOrganization creates a new organization with the given details.
// AI-hint: Factory method for creating new organizations with automatic ID generation
// and timestamp creation. Validates all input parameters before creation.
func NewOrganization(name, slug, description string, settings map[string]any) (*Organization, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateSlug(slug); err != nil {
		return nil, err
	}

	// Initialize settings map if nil
	if settings == nil {
		settings = make(map[string]any)
	}

	now := time.Now()
	org := &Organization{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: description,
		Settings:    settings,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return org, nil
}

// NewOrganizationWithID creates a new organization with a specific ID.
// AI-hint: Factory method for creating organizations with existing IDs, useful for
// reconstruction from database or testing scenarios.
func NewOrganizationWithID(id uuid.UUID, name, slug, description string, settings map[string]any) (*Organization, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("organization ID cannot be zero")
	}

	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateSlug(slug); err != nil {
		return nil, err
	}

	// Initialize settings map if nil
	if settings == nil {
		settings = make(map[string]any)
	}

	now := time.Now()
	org := &Organization{
		ID:          id,
		Name:        name,
		Slug:        slug,
		Description: description,
		Settings:    settings,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return org, nil
}

// Update modifies the organization's basic information.
// AI-hint: Business method for updating organization details. Automatically updates
// the UpdatedAt timestamp and validates all input parameters.
func (o *Organization) Update(name, slug, description string, settings map[string]any) error {
	if err := validateName(name); err != nil {
		return err
	}

	if err := validateSlug(slug); err != nil {
		return err
	}

	o.Name = name
	o.Slug = slug
	o.Description = description
	o.Settings = settings
	o.UpdatedAt = time.Now()

	return nil
}

// Validate ensures the organization is in a valid state.
// AI-hint: Domain validation method that ensures all business rules are satisfied.
// Called before persistence operations to maintain data integrity.
func (o *Organization) Validate() error {
	if o.ID == uuid.Nil {
		return fmt.Errorf("organization ID cannot be zero")
	}

	if err := validateName(o.Name); err != nil {
		return err
	}

	if err := validateSlug(o.Slug); err != nil {
		return err
	}

	if o.CreatedAt.IsZero() {
		return fmt.Errorf("organization created at cannot be zero")
	}

	if o.UpdatedAt.IsZero() {
		return fmt.Errorf("organization updated at cannot be zero")
	}

	return nil
}

// IsActive checks if the organization is currently active.
// AI-hint: Business method that checks organization status from settings.
// Defaults to active if no explicit setting is provided.
func (o *Organization) IsActive() bool {
	if active, exists := o.Settings["active"]; exists {
		if activeBool, ok := active.(bool); ok {
			return activeBool
		}
	}
	return true // Default to active
}

// GetSetting retrieves a setting value from the organization.
// AI-hint: Utility method for accessing organization configuration settings.
// Returns the value and a boolean indicating if the setting exists.
func (o *Organization) GetSetting(key string) (any, bool) {
	value, exists := o.Settings[key]
	return value, exists
}

// SetSetting sets a configuration setting for the organization.
// AI-hint: Utility method for updating organization configuration settings.
// Automatically updates the UpdatedAt timestamp.
func (o *Organization) SetSetting(key string, value any) {
	if o.Settings == nil {
		o.Settings = make(map[string]any)
	}
	o.Settings[key] = value
	o.UpdatedAt = time.Now()
}

// RemoveSetting removes a configuration setting from the organization.
// AI-hint: Utility method for removing organization configuration settings.
// Automatically updates the UpdatedAt timestamp.
func (o *Organization) RemoveSetting(key string) {
	if o.Settings != nil {
		delete(o.Settings, key)
		o.UpdatedAt = time.Now()
	}
}

// Clone creates a deep copy of the organization with a new ID and timestamps.
// AI-hint: Utility method for creating organization copies, useful for testing
// and creating organization templates. Ensures complete isolation between copies.
func (o *Organization) Clone() *Organization {
	// Deep copy settings map
	settingsCopy := make(map[string]any)
	for k, v := range o.Settings {
		settingsCopy[k] = v
	}

	now := time.Now()
	return &Organization{
		ID:          uuid.New(),
		Name:        o.Name,
		Slug:        o.Slug,
		Description: o.Description,
		Settings:    settingsCopy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// validateName validates the organization name according to business rules.
// AI-hint: Private validation method for organization names. Enforces length limits
// and ensures names are not empty.
func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("organization name cannot be empty")
	}

	if len(name) > MaxNameLength {
		return fmt.Errorf("organization name cannot exceed %d characters", MaxNameLength)
	}

	return nil
}

// validateSlug validates the organization slug according to business rules.
// AI-hint: Private validation method for organization slugs. Enforces format rules,
// length limits, and ensures slugs are URL-safe and unique-friendly.
func validateSlug(slug string) error {
	if strings.TrimSpace(slug) == "" {
		return fmt.Errorf("organization slug cannot be empty")
	}

	if len(slug) > MaxSlugLength {
		return fmt.Errorf("organization slug cannot exceed %d characters", MaxSlugLength)
	}

	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("organization slug can only contain lowercase letters, numbers, and hyphens")
	}

	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return fmt.Errorf("organization slug cannot start or end with hyphen")
	}

	if strings.Contains(slug, "--") {
		return fmt.Errorf("organization slug cannot contain consecutive hyphens")
	}

	return nil
}

// GenerateSlug automatically generates a URL-friendly slug from an organization name.
// AI-hint: Utility function for automatic slug generation. Converts organization names
// to URL-safe slugs by removing special characters and converting spaces to hyphens.
func GenerateSlug(name string) string {
	if name == "" {
		return ""
	}

	// Convert to lowercase
	slug := strings.ToLower(name)

	// Remove special characters and replace with spaces
	slug = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(slug, " ")

	// Replace multiple spaces with single space
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, " ")

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}
