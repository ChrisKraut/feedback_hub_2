package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewOrganization tests the creation of a new organization with valid data
func TestNewOrganization(t *testing.T) {
	tests := []struct {
		name        string
		nameValue   string
		slugValue   string
		description string
		settings    map[string]any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid organization with all fields",
			nameValue:   "Acme Corporation",
			slugValue:   "acme-corp",
			description: "A leading technology company",
			settings:    map[string]any{"theme": "dark", "timezone": "UTC"},
			expectError: false,
		},
		{
			name:        "valid organization with minimal fields",
			nameValue:   "Startup Inc",
			slugValue:   "startup",
			description: "",
			settings:    map[string]any{},
			expectError: false,
		},
		{
			name:        "empty name should fail",
			nameValue:   "",
			slugValue:   "empty-name",
			description: "Organization with empty name",
			expectError: true,
			errorMsg:    "organization name cannot be empty",
		},
		{
			name:        "empty slug should fail",
			nameValue:   "Valid Name",
			slugValue:   "",
			description: "Organization with empty slug",
			expectError: true,
			errorMsg:    "organization slug cannot be empty",
		},
		{
			name:        "name too long should fail",
			nameValue:   "This is a very long organization name that exceeds the maximum allowed length of two hundred and fifty five characters which should cause a validation error to be returned and this is even longer to make sure we exceed the limit and we need to add more text to actually go over the character limit so let me add some more words here to make sure we have enough characters to trigger the validation error that we are testing for in this particular test case",
			slugValue:   "long-name",
			description: "Organization with very long name",
			expectError: true,
			errorMsg:    "organization name cannot exceed 255 characters",
		},
		{
			name:        "slug too long should fail",
			nameValue:   "Valid Name",
			slugValue:   "this-is-a-very-long-slug-that-exceeds-the-maximum-allowed-length-of-one-hundred-characters-which-should-cause-a-validation-error",
			description: "Organization with very long slug",
			expectError: true,
			errorMsg:    "organization slug cannot exceed 100 characters",
		},
		{
			name:        "invalid slug format should fail",
			nameValue:   "Valid Name",
			slugValue:   "invalid slug with spaces",
			description: "Organization with invalid slug format",
			expectError: true,
			errorMsg:    "organization slug can only contain lowercase letters, numbers, and hyphens",
		},
		{
			name:        "slug starting with hyphen should fail",
			nameValue:   "Valid Name",
			slugValue:   "-invalid-slug",
			description: "Organization with slug starting with hyphen",
			expectError: true,
			errorMsg:    "organization slug can only contain lowercase letters, numbers, and hyphens",
		},
		{
			name:        "slug ending with hyphen should fail",
			nameValue:   "Valid Name",
			slugValue:   "invalid-slug-",
			description: "Organization with slug ending with hyphen",
			expectError: true,
			errorMsg:    "organization slug can only contain lowercase letters, numbers, and hyphens",
		},
		{
			name:        "consecutive hyphens in slug should fail",
			nameValue:   "Valid Name",
			slugValue:   "invalid--slug",
			description: "Organization with consecutive hyphens in slug",
			expectError: true,
			errorMsg:    "organization slug can only contain lowercase letters, numbers, and hyphens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := NewOrganization(tt.nameValue, tt.slugValue, tt.description, tt.settings)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, org)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, org)
				assert.NotEmpty(t, org.ID)
				assert.Equal(t, tt.nameValue, org.Name)
				assert.Equal(t, tt.slugValue, org.Slug)
				assert.Equal(t, tt.description, org.Description)
				if tt.settings == nil {
					assert.Empty(t, org.Settings)
				} else {
					assert.Equal(t, tt.settings, org.Settings)
				}
				assert.False(t, org.CreatedAt.IsZero())
				assert.False(t, org.UpdatedAt.IsZero())
				assert.Equal(t, org.CreatedAt, org.UpdatedAt)
			}
		})
	}
}

// TestNewOrganizationWithID tests creating an organization with a specific ID
func TestNewOrganizationWithID(t *testing.T) {
	id := uuid.New()
	name := "Test Organization"
	slug := "test-org"
	description := "A test organization"
	settings := map[string]any{"test": true}

	org, err := NewOrganizationWithID(id, name, slug, description, settings)

	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, id, org.ID)
	assert.Equal(t, name, org.Name)
	assert.Equal(t, slug, org.Slug)
	assert.Equal(t, description, org.Description)
	assert.Equal(t, settings, org.Settings)
}

// TestOrganization_Update tests updating organization fields
func TestOrganization_Update(t *testing.T) {
	org, err := NewOrganization("Original Name", "original-slug", "Original description", nil)
	require.NoError(t, err)
	require.NotNil(t, org)

	originalCreatedAt := org.CreatedAt
	originalUpdatedAt := org.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Test valid update
	err = org.Update("Updated Name", "updated-slug", "Updated description", map[string]any{"new": "setting"})
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", org.Name)
	assert.Equal(t, "updated-slug", org.Slug)
	assert.Equal(t, "Updated description", org.Description)
	assert.Equal(t, map[string]any{"new": "setting"}, org.Settings)
	assert.Equal(t, originalCreatedAt, org.CreatedAt)
	assert.True(t, org.UpdatedAt.After(originalUpdatedAt))

	// Test invalid update
	err = org.Update("", "valid-slug", "Valid description", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization name cannot be empty")
}

// TestOrganization_Validate tests organization validation
func TestOrganization_Validate(t *testing.T) {
	tests := []struct {
		name        string
		org         *Organization
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid organization",
			org: &Organization{
				ID:          uuid.New(),
				Name:        "Valid Organization",
				Slug:        "valid-org",
				Description: "A valid organization",
				Settings:    map[string]any{"valid": true},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectError: false,
		},
		{
			name: "zero UUID should fail",
			org: &Organization{
				ID:          uuid.Nil,
				Name:        "Valid Name",
				Slug:        "valid-slug",
				Description: "Valid description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "organization ID cannot be zero",
		},
		{
			name: "empty name should fail",
			org: &Organization{
				ID:          uuid.New(),
				Name:        "",
				Slug:        "valid-slug",
				Description: "Valid description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "organization name cannot be empty",
		},
		{
			name: "empty slug should fail",
			org: &Organization{
				ID:          uuid.New(),
				Name:        "Valid Name",
				Slug:        "",
				Description: "Valid description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "organization slug cannot be empty",
		},
		{
			name: "zero created at should fail",
			org: &Organization{
				ID:          uuid.New(),
				Name:        "Valid Name",
				Slug:        "valid-slug",
				Description: "Valid description",
				CreatedAt:   time.Time{},
				UpdatedAt:   time.Now(),
			},
			expectError: true,
			errorMsg:    "organization created at cannot be zero",
		},
		{
			name: "zero updated at should fail",
			org: &Organization{
				ID:          uuid.New(),
				Name:        "Valid Name",
				Slug:        "valid-slug",
				Description: "Valid description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Time{},
			},
			expectError: true,
			errorMsg:    "organization updated at cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.org.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGenerateSlug tests automatic slug generation from organization name
func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"Acme Corporation", "acme-corporation"},
		{"Startup Inc.", "startup-inc"},
		{"Tech Company 2024", "tech-company-2024"},
		{"Company & Associates", "company-associates"},
		{"Company@Tech", "company-tech"},
		{"Company (LLC)", "company-llc"},
		{"Company - The Best", "company-the-best"},
		{"Company...", "company"},
		{"   Company   ", "company"},
		{"C", "c"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganization_IsActive tests organization active status
func TestOrganization_IsActive(t *testing.T) {
	org, err := NewOrganization("Test Org", "test-org", "Test organization", nil)
	require.NoError(t, err)

	// By default, organization should be active
	assert.True(t, org.IsActive())

	// Test with deactivated setting
	org.Settings["active"] = false
	assert.False(t, org.IsActive())

	// Test with activated setting
	org.Settings["active"] = true
	assert.True(t, org.IsActive())
}

// TestOrganization_GetSetting tests getting organization settings
func TestOrganization_GetSetting(t *testing.T) {
	org, err := NewOrganization("Test Org", "test-org", "Test organization", map[string]any{
		"theme":     "dark",
		"timezone":  "UTC",
		"max_users": 100,
	})
	require.NoError(t, err)

	// Test getting existing settings
	theme, exists := org.GetSetting("theme")
	assert.True(t, exists)
	assert.Equal(t, "dark", theme)

	timezone, exists := org.GetSetting("timezone")
	assert.True(t, exists)
	assert.Equal(t, "UTC", timezone)

	maxUsers, exists := org.GetSetting("max_users")
	assert.True(t, exists)
	assert.Equal(t, 100, maxUsers)

	// Test getting non-existent setting
	_, exists = org.GetSetting("non_existent")
	assert.False(t, exists)
}

// TestOrganization_SetSetting tests setting organization settings
func TestOrganization_SetSetting(t *testing.T) {
	org, err := NewOrganization("Test Org", "test-org", "Test organization", nil)
	require.NoError(t, err)

	// Test setting new setting
	org.SetSetting("theme", "light")
	theme, exists := org.GetSetting("theme")
	assert.True(t, exists)
	assert.Equal(t, "light", theme)

	// Test updating existing setting
	org.SetSetting("theme", "dark")
	theme, exists = org.GetSetting("theme")
	assert.True(t, exists)
	assert.Equal(t, "dark", theme)

	// Test setting multiple settings
	org.SetSetting("timezone", "EST")
	org.SetSetting("max_users", 50)

	timezone, exists := org.GetSetting("timezone")
	assert.True(t, exists)
	assert.Equal(t, "EST", timezone)

	maxUsers, exists := org.GetSetting("max_users")
	assert.True(t, exists)
	assert.Equal(t, 50, maxUsers)
}

// TestOrganization_RemoveSetting tests removing organization settings
func TestOrganization_RemoveSetting(t *testing.T) {
	org, err := NewOrganization("Test Org", "test-org", "Test organization", map[string]any{
		"theme":    "dark",
		"timezone": "UTC",
	})
	require.NoError(t, err)

	// Test removing existing setting
	org.RemoveSetting("theme")
	_, exists := org.GetSetting("theme")
	assert.False(t, exists)

	// Test removing non-existent setting (should not error)
	org.RemoveSetting("non_existent")

	// Verify other settings remain
	timezone, exists := org.GetSetting("timezone")
	assert.True(t, exists)
	assert.Equal(t, "UTC", timezone)
}

// TestOrganization_Clone tests cloning an organization
func TestOrganization_Clone(t *testing.T) {
	original, err := NewOrganization("Original Org", "original-org", "Original organization", map[string]any{
		"theme": "dark",
	})
	require.NoError(t, err)

	// Wait to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	cloned := original.Clone()

	// Verify cloned organization has same data but different ID and timestamps
	assert.NotEqual(t, original.ID, cloned.ID)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Slug, cloned.Slug)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, original.Settings, cloned.Settings)
	assert.True(t, cloned.CreatedAt.After(original.CreatedAt))
	assert.True(t, cloned.UpdatedAt.After(original.UpdatedAt))

	// Verify modifying cloned organization doesn't affect original
	cloned.Name = "Modified Name"
	assert.Equal(t, "Original Org", original.Name)
	assert.Equal(t, "Modified Name", cloned.Name)
}
