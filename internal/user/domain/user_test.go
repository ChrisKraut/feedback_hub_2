package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("valid user creation", func(t *testing.T) {
		user, err := NewUser("123", "test@example.com", "Test User", "role-123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "123", user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "role-123", user.RoleID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("invalid email format", func(t *testing.T) {
		user, err := NewUser("123", "invalid-email", "Test User", "role-123")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("empty required fields", func(t *testing.T) {
		testCases := []struct {
			id            string
			email         string
			name          string
			roleID        string
			expectedError string
		}{
			{"", "test@example.com", "Test User", "role-123", "user ID cannot be empty"},
			{"123", "", "Test User", "role-123", "email cannot be empty"},
			{"123", "test@example.com", "", "role-123", "name cannot be empty"},
			{"123", "test@example.com", "Test User", "", "role ID cannot be empty"},
		}

		for _, tc := range testCases {
			user, err := NewUser(tc.id, tc.email, tc.name, tc.roleID)
			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	})
}

func TestUser_UpdateName(t *testing.T) {
	user, _ := NewUser("123", "test@example.com", "Test User", "role-123")
	originalUpdatedAt := user.UpdatedAt

	t.Run("valid name update", func(t *testing.T) {
		err := user.UpdateName("Updated Name")

		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", user.Name)
		assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("empty name", func(t *testing.T) {
		err := user.UpdateName("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})
}

func TestUser_UpdateRole(t *testing.T) {
	user, _ := NewUser("123", "test@example.com", "Test User", "role-123")
	originalUpdatedAt := user.UpdatedAt

	t.Run("valid role update", func(t *testing.T) {
		err := user.UpdateRole("new-role-456")

		assert.NoError(t, err)
		assert.Equal(t, "new-role-456", user.RoleID)
		assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("empty role ID", func(t *testing.T) {
		err := user.UpdateRole("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role ID cannot be empty")
	})
}

func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"", false},
		{"invalid", false},
		{"@example.com", false},
		{"test@", false},
		{"test@domain", false},
	}

	for _, tc := range testCases {
		result := isValidEmail(tc.email)
		assert.Equal(t, tc.expected, result, "Email: %s", tc.email)
	}
}
