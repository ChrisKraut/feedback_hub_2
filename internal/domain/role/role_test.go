package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRole(t *testing.T) {
	t.Run("valid role creation", func(t *testing.T) {
		role, err := NewRole("123", "Product Owner")

		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "123", role.ID)
		assert.Equal(t, "Product Owner", role.Name)
		assert.False(t, role.CreatedAt.IsZero())
		assert.False(t, role.UpdatedAt.IsZero())
	})

	t.Run("empty required fields", func(t *testing.T) {
		testCases := []struct {
			id            string
			name          string
			expectedError string
		}{
			{"", "Product Owner", "role ID cannot be empty"},
			{"123", "", "role name cannot be empty"},
			{"123", "   ", "role name cannot be empty"},
		}

		for _, tc := range testCases {
			role, err := NewRole(tc.id, tc.name)
			assert.Error(t, err)
			assert.Nil(t, role)
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	})
}

func TestRole_UpdateName(t *testing.T) {
	t.Run("valid name update for regular role", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner")
		originalUpdatedAt := role.UpdatedAt

		err := role.UpdateName("Updated Role")

		assert.NoError(t, err)
		assert.Equal(t, "Updated Role", role.Name)
		assert.True(t, role.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("cannot update Super User role name", func(t *testing.T) {
		role, _ := NewRole("123", SuperUserRoleName)

		err := role.UpdateName("Hacker")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot modify Super User role")
		assert.Equal(t, SuperUserRoleName, role.Name) // Name should remain unchanged
	})

	t.Run("empty name", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner")

		err := role.UpdateName("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role name cannot be empty")
	})
}

func TestRole_IsSuperUser(t *testing.T) {
	t.Run("Super User role", func(t *testing.T) {
		role, _ := NewRole("123", SuperUserRoleName)
		assert.True(t, role.IsSuperUser())
	})

	t.Run("regular role", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner")
		assert.False(t, role.IsSuperUser())
	})
}

func TestRole_CanBeDeleted(t *testing.T) {
	t.Run("Super User role cannot be deleted", func(t *testing.T) {
		role, _ := NewRole("123", SuperUserRoleName)
		assert.False(t, role.CanBeDeleted())
	})

	t.Run("regular role can be deleted", func(t *testing.T) {
		role, _ := NewRole("123", "Product Owner")
		assert.True(t, role.CanBeDeleted())
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
