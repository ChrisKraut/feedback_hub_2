package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewUserOrganization tests the creation of a new user-organization relationship.
// AI-hint: These tests ensure that user-organization relationships are properly created
// with validation and business rules enforced.
func TestNewUserOrganization(t *testing.T) {
	tests := []struct {
		name           string
		userID         uuid.UUID
		organizationID uuid.UUID
		roleID         uuid.UUID
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "valid user-organization relationship",
			userID:         uuid.New(),
			organizationID: uuid.New(),
			roleID:         uuid.New(),
			expectError:    false,
		},
		{
			name:           "zero user ID should fail",
			userID:         uuid.Nil,
			organizationID: uuid.New(),
			roleID:         uuid.New(),
			expectError:    true,
			errorMsg:       "user ID cannot be zero",
		},
		{
			name:           "zero organization ID should fail",
			userID:         uuid.New(),
			organizationID: uuid.Nil,
			roleID:         uuid.New(),
			expectError:    true,
			errorMsg:       "organization ID cannot be zero",
		},
		{
			name:           "zero role ID should fail",
			userID:         uuid.New(),
			organizationID: uuid.New(),
			roleID:         uuid.Nil,
			expectError:    true,
			errorMsg:       "role ID cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userOrg, err := NewUserOrganization(tt.userID, tt.organizationID, tt.roleID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, userOrg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, userOrg)
				assert.Equal(t, tt.userID, userOrg.UserID)
				assert.Equal(t, tt.organizationID, userOrg.OrganizationID)
				assert.Equal(t, tt.roleID, userOrg.RoleID)
				assert.True(t, userOrg.IsActive)
				assert.False(t, userOrg.JoinedAt.IsZero())
				assert.False(t, userOrg.CreatedAt.IsZero())
				assert.False(t, userOrg.UpdatedAt.IsZero())
				assert.Equal(t, userOrg.JoinedAt, userOrg.CreatedAt)
				assert.Equal(t, userOrg.CreatedAt, userOrg.UpdatedAt)
			}
		})
	}
}

// TestNewUserOrganizationWithID tests creating a user-organization relationship with a specific ID.
// AI-hint: Factory method for creating relationships with existing IDs, useful for
// reconstruction from database or testing scenarios.
func TestNewUserOrganizationWithID(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	organizationID := uuid.New()
	roleID := uuid.New()
	joinedAt := time.Now()

	userOrg, err := NewUserOrganizationWithID(id, userID, organizationID, roleID, joinedAt, true)

	assert.NoError(t, err)
	assert.NotNil(t, userOrg)
	assert.Equal(t, id, userOrg.ID)
	assert.Equal(t, userID, userOrg.UserID)
	assert.Equal(t, organizationID, userOrg.OrganizationID)
	assert.Equal(t, roleID, userOrg.RoleID)
	assert.Equal(t, joinedAt, userOrg.JoinedAt)
	assert.True(t, userOrg.IsActive)
}

// TestUserOrganization_UpdateRole tests updating the role within a user-organization relationship.
// AI-hint: Business method for updating user roles within organizations.
func TestUserOrganization_UpdateRole(t *testing.T) {
	userOrg, err := NewUserOrganization(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.NotNil(t, userOrg)

	originalUpdatedAt := userOrg.UpdatedAt
	newRoleID := uuid.New()

	// Wait a bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Test valid role update
	err = userOrg.UpdateRole(newRoleID)
	assert.NoError(t, err)
	assert.Equal(t, newRoleID, userOrg.RoleID)
	assert.True(t, userOrg.UpdatedAt.After(originalUpdatedAt))

	// Test invalid role update
	err = userOrg.UpdateRole(uuid.Nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role ID cannot be zero")
}

// TestUserOrganization_SetActive tests setting the active status of a user-organization relationship.
// AI-hint: Business method for activating/deactivating user access to organizations.
func TestUserOrganization_SetActive(t *testing.T) {
	userOrg, err := NewUserOrganization(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.NotNil(t, userOrg)

	// Initially should be active
	assert.True(t, userOrg.IsActive)

	// Test deactivating
	err = userOrg.SetActive(false)
	assert.NoError(t, err)
	assert.False(t, userOrg.IsActive)

	// Test activating
	err = userOrg.SetActive(true)
	assert.NoError(t, err)
	assert.True(t, userOrg.IsActive)
}

// TestUserOrganization_Validate tests user-organization relationship validation.
// AI-hint: Domain validation method that ensures all business rules are satisfied.
func TestUserOrganization_Validate(t *testing.T) {
	tests := []struct {
		name        string
		userOrg     *UserOrganization
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid user-organization relationship",
			userOrg: &UserOrganization{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				OrganizationID: uuid.New(),
				RoleID:         uuid.New(),
				IsActive:       true,
				JoinedAt:       time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			expectError: false,
		},
		{
			name: "zero ID should fail",
			userOrg: &UserOrganization{
				ID:             uuid.Nil,
				UserID:         uuid.New(),
				OrganizationID: uuid.New(),
				RoleID:         uuid.New(),
				IsActive:       true,
				JoinedAt:       time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			expectError: true,
			errorMsg:    "user-organization relationship ID cannot be zero",
		},
		{
			name: "zero user ID should fail",
			userOrg: &UserOrganization{
				ID:             uuid.New(),
				UserID:         uuid.Nil,
				OrganizationID: uuid.New(),
				RoleID:         uuid.New(),
				IsActive:       true,
				JoinedAt:       time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			expectError: true,
			errorMsg:    "user ID cannot be zero",
		},
		{
			name: "zero organization ID should fail",
			userOrg: &UserOrganization{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				OrganizationID: uuid.Nil,
				RoleID:         uuid.New(),
				IsActive:       true,
				JoinedAt:       time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			expectError: true,
			errorMsg:    "organization ID cannot be zero",
		},
		{
			name: "zero role ID should fail",
			userOrg: &UserOrganization{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				OrganizationID: uuid.New(),
				RoleID:         uuid.Nil,
				IsActive:       true,
				JoinedAt:       time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			expectError: true,
			errorMsg:    "role ID cannot be zero",
		},
		{
			name: "zero joined at should fail",
			userOrg: &UserOrganization{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				OrganizationID: uuid.New(),
				RoleID:         uuid.New(),
				IsActive:       true,
				JoinedAt:       time.Time{},
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			expectError: true,
			errorMsg:    "joined at cannot be zero",
		},
		{
			name: "zero created at should fail",
			userOrg: &UserOrganization{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				OrganizationID: uuid.New(),
				RoleID:         uuid.New(),
				IsActive:       true,
				JoinedAt:       time.Now(),
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Now(),
			},
			expectError: true,
			errorMsg:    "created at cannot be zero",
		},
		{
			name: "zero updated at should fail",
			userOrg: &UserOrganization{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				OrganizationID: uuid.New(),
				RoleID:         uuid.New(),
				IsActive:       true,
				JoinedAt:       time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Time{},
			},
			expectError: true,
			errorMsg:    "updated at cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.userOrg.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUserOrganization_IsActiveInOrganization tests the active status check.
// AI-hint: Business method that checks if a user is active in a specific organization.
func TestUserOrganization_IsActiveInOrganization(t *testing.T) {
	userOrg, err := NewUserOrganization(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	// Initially should be active
	assert.True(t, userOrg.IsActiveInOrganization())

	// Test with inactive status
	userOrg.IsActive = false
	assert.False(t, userOrg.IsActiveInOrganization())

	// Test with active status
	userOrg.IsActive = true
	assert.True(t, userOrg.IsActiveInOrganization())
}

// TestUserOrganization_GetTenure tests calculating the user's tenure in the organization.
// AI-hint: Business method that calculates how long a user has been part of an organization.
func TestUserOrganization_GetTenure(t *testing.T) {
	userOrg, err := NewUserOrganization(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	// Test tenure calculation
	tenure := userOrg.GetTenure()
	assert.True(t, tenure >= 0)

	// Test with past join date
	pastTime := time.Now().Add(-24 * time.Hour)
	userOrg.JoinedAt = pastTime
	tenure = userOrg.GetTenure()
	assert.True(t, tenure >= 24*time.Hour)
}

// TestUserOrganization_Clone tests cloning a user-organization relationship.
// AI-hint: Utility method for creating relationship copies, useful for testing
// and creating relationship templates.
func TestUserOrganization_Clone(t *testing.T) {
	original, err := NewUserOrganization(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	// Wait to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	cloned := original.Clone()

	// Verify cloned relationship has same data but different ID and timestamps
	assert.NotEqual(t, original.ID, cloned.ID)
	assert.Equal(t, original.UserID, cloned.UserID)
	assert.Equal(t, original.OrganizationID, cloned.OrganizationID)
	assert.Equal(t, original.RoleID, cloned.RoleID)
	assert.Equal(t, original.IsActive, cloned.IsActive)
	assert.True(t, cloned.CreatedAt.After(original.CreatedAt))
	assert.True(t, cloned.UpdatedAt.After(original.UpdatedAt))

	// Verify modifying cloned relationship doesn't affect original
	cloned.IsActive = false
	assert.True(t, original.IsActive)
	assert.False(t, cloned.IsActive)
}
