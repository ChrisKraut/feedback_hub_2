package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("valid user creation with organization", func(t *testing.T) {
		user, err := NewUser("123", "test@example.com", "Test User", "role-123", "org-456")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "123", user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "role-123", user.RoleID)
		assert.Equal(t, "org-456", user.OrganizationID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("valid user creation without organization (legacy support)", func(t *testing.T) {
		user, err := NewUserWithoutOrganization("123", "test@example.com", "Test User", "role-123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "123", user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "role-123", user.RoleID)
		assert.Equal(t, "", user.OrganizationID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("invalid email format", func(t *testing.T) {
		user, err := NewUser("123", "invalid-email", "Test User", "role-123", "org-456")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("empty required fields", func(t *testing.T) {
		testCases := []struct {
			id             string
			email          string
			name           string
			roleID         string
			organizationID string
			expectedError  string
		}{
			{"", "test@example.com", "Test User", "role-123", "org-456", "user ID cannot be empty"},
			{"123", "", "Test User", "role-123", "org-456", "email cannot be empty"},
			{"123", "test@example.com", "", "role-123", "org-456", "name cannot be empty"},
			{"123", "test@example.com", "Test User", "", "org-456", "role ID cannot be empty"},
			{"123", "test@example.com", "Test User", "role-123", "", "organization ID cannot be empty"},
		}

		for _, tc := range testCases {
			user, err := NewUser(tc.id, tc.email, tc.name, tc.roleID, tc.organizationID)
			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	})
}

func TestNewUserWithPassword(t *testing.T) {
	t.Run("valid user creation with password and organization", func(t *testing.T) {
		user, err := NewUserWithPassword("123", "test@example.com", "Test User", "password-hash", "role-123", "org-456")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "123", user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "password-hash", user.PasswordHash)
		assert.Equal(t, "role-123", user.RoleID)
		assert.Equal(t, "org-456", user.OrganizationID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("valid user creation with password without organization (legacy support)", func(t *testing.T) {
		user, err := NewUserWithPasswordWithoutOrganization("123", "test@example.com", "Test User", "password-hash", "role-123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "123", user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "password-hash", user.PasswordHash)
		assert.Equal(t, "role-123", user.RoleID)
		assert.Equal(t, "", user.OrganizationID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("empty required fields with password", func(t *testing.T) {
		testCases := []struct {
			id             string
			email          string
			name           string
			passwordHash   string
			roleID         string
			organizationID string
			expectedError  string
		}{
			{"", "test@example.com", "Test User", "password-hash", "role-123", "org-456", "user ID cannot be empty"},
			{"123", "", "Test User", "password-hash", "role-123", "org-456", "email cannot be empty"},
			{"123", "test@example.com", "", "password-hash", "role-123", "org-456", "name cannot be empty"},
			{"123", "test@example.com", "Test User", "", "role-123", "org-456", "password hash cannot be empty"},
			{"123", "test@example.com", "Test User", "password-hash", "", "org-456", "role ID cannot be empty"},
			{"123", "test@example.com", "Test User", "password-hash", "role-123", "", "organization ID cannot be empty"},
		}

		for _, tc := range testCases {
			user, err := NewUserWithPassword(tc.id, tc.email, tc.name, tc.passwordHash, tc.roleID, tc.organizationID)
			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	})
}

func TestUser_UpdateName(t *testing.T) {
	user, _ := NewUser("123", "test@example.com", "Test User", "role-123", "org-456")
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
	user, _ := NewUser("123", "test@example.com", "Test User", "role-123", "org-456")
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

func TestUser_UpdateOrganization(t *testing.T) {
	user, _ := NewUser("123", "test@example.com", "Test User", "role-123", "org-456")
	originalUpdatedAt := user.UpdatedAt

	t.Run("valid organization update", func(t *testing.T) {
		err := user.UpdateOrganization("new-org-789")

		assert.NoError(t, err)
		assert.Equal(t, "new-org-789", user.OrganizationID)
		assert.True(t, user.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("empty organization ID", func(t *testing.T) {
		err := user.UpdateOrganization("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID cannot be empty")
	})
}

func TestUser_IsInOrganization(t *testing.T) {
	t.Run("user is in specified organization", func(t *testing.T) {
		user, _ := NewUser("123", "test@example.com", "Test User", "role-123", "org-456")

		assert.True(t, user.IsInOrganization("org-456"))
		assert.False(t, user.IsInOrganization("org-789"))
	})

	t.Run("user without organization", func(t *testing.T) {
		user, _ := NewUserWithoutOrganization("123", "test@example.com", "Test User", "role-123")

		// Users without organization should return false for any organization check
		assert.False(t, user.IsInOrganization("org-456"))
		assert.False(t, user.IsInOrganization(""))
	})
}

func TestUser_IsOrganizationScoped(t *testing.T) {
	t.Run("organization scoped user", func(t *testing.T) {
		user, _ := NewUser("123", "test@example.com", "Test User", "role-123", "org-456")

		assert.True(t, user.IsOrganizationScoped())
	})

	t.Run("non-organization scoped user", func(t *testing.T) {
		user, _ := NewUserWithoutOrganization("123", "test@example.com", "Test User", "role-123")

		assert.False(t, user.IsOrganizationScoped())
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
