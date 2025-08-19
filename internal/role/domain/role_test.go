package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRole(t *testing.T) {
	t.Run("valid role creation with organization", func(t *testing.T) {
		role, err := NewRole("123", "Product Owner", "org-456")

		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "123", role.ID)
		assert.Equal(t, "Product Owner", role.Name)
		assert.Equal(t, "org-456", role.OrganizationID)
		assert.False(t, role.CreatedAt.IsZero())
		assert.False(t, role.UpdatedAt.IsZero())
	})

	t.Run("valid role creation without organization (legacy support)", func(t *testing.T) {
		role, err := NewRoleWithoutOrganization("123", "Product Owner")

		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "123", role.ID)
		assert.Equal(t, "Product Owner", role.Name)
		assert.Equal(t, "", role.OrganizationID)
		assert.False(t, role.CreatedAt.IsZero())
		assert.False(t, role.UpdatedAt.IsZero())
	})

	t.Run("empty required fields", func(t *testing.T) {
		testCases := []struct {
			id             string
			name           string
			organizationID string
			expectedError  string
		}{
			{"", "Product Owner", "org-456", "role ID cannot be empty"},
			{"123", "", "org-456", "role name cannot be empty"},
			{"123", "   ", "org-456", "role name cannot be empty"},
			{"123", "Product Owner", "", "organization ID cannot be empty"},
		}

		for _, tc := range testCases {
			role, err := NewRole(tc.id, tc.name, tc.organizationID)
			assert.Error(t, err)
			assert.Nil(t, role)
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	})
}

func TestRole_UpdateName(t *testing.T) {
	t.Run("valid name update for regular role", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")
		originalUpdatedAt := role.UpdatedAt

		err := role.UpdateName("Updated Role")

		assert.NoError(t, err)
		assert.Equal(t, "Updated Role", role.Name)
		assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("cannot update Super User role name", func(t *testing.T) {
		role, _ := NewRole("123", SuperUserRoleName, "org-456")

		err := role.UpdateName("Hacker")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot modify Super User role")
		assert.Equal(t, SuperUserRoleName, role.Name) // Name should remain unchanged
	})

	t.Run("empty name", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")

		err := role.UpdateName("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role name cannot be empty")
	})
}

func TestRole_UpdateOrganization(t *testing.T) {
	t.Run("valid organization update", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")
		originalUpdatedAt := role.UpdatedAt

		err := role.UpdateOrganization("new-org-789")

		assert.NoError(t, err)
		assert.Equal(t, "new-org-789", role.OrganizationID)
		assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("empty organization ID", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")

		err := role.UpdateOrganization("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID cannot be empty")
	})
}

func TestRole_IsSuperUser(t *testing.T) {
	t.Run("Super User role", func(t *testing.T) {
		role, _ := NewRole("123", SuperUserRoleName, "org-456")
		assert.True(t, role.IsSuperUser())
	})

	t.Run("regular role", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")
		assert.False(t, role.IsSuperUser())
	})
}

func TestRole_CanBeDeleted(t *testing.T) {
	t.Run("Super User role cannot be deleted", func(t *testing.T) {
		role, _ := NewRole("123", SuperUserRoleName, "org-456")
		assert.False(t, role.CanBeDeleted())
	})

	t.Run("regular role can be deleted", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")
		assert.True(t, role.CanBeDeleted())
	})
}

func TestRole_IsInOrganization(t *testing.T) {
	t.Run("role is in specified organization", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")

		assert.True(t, role.IsInOrganization("org-456"))
		assert.False(t, role.IsInOrganization("org-789"))
	})

	t.Run("role without organization", func(t *testing.T) {
		role, _ := NewRoleWithoutOrganization("123", "Product Owner")

		// Roles without organization should return false for any organization check
		assert.False(t, role.IsInOrganization("org-456"))
		assert.False(t, role.IsInOrganization(""))
	})
}

func TestRole_IsOrganizationScoped(t *testing.T) {
	t.Run("organization scoped role", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner", "org-456")

		assert.True(t, role.IsOrganizationScoped())
	})

	t.Run("non-organization scoped role", func(t *testing.T) {
		role, _ := NewRoleWithoutOrganization("123", "Product Owner")

		assert.False(t, role.IsOrganizationScoped())
	})
}

func TestPredefinedRoles(t *testing.T) {
	t.Run("contains required roles", func(t *testing.T) {
		assert.Contains(t, PredefinedRoles, SuperUserRoleName)
		assert.Contains(t, PredefinedRoles, "Product Owner")
		assert.Contains(t, PredefinedRoles, "Contributor")
		assert.Len(t, PredefinedRoles, 3)
	})
}
