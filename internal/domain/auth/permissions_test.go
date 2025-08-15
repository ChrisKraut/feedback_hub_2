package auth

import (
	"feedback_hub_2/internal/domain/role"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizationService_CanPerform(t *testing.T) {
	service := NewAuthorizationService()

	t.Run("Super User can perform any action", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "super-user-id",
			RoleName: role.SuperUserRoleName,
		}

		permissions := []Permission{
			PermissionCreateRole, PermissionReadRole, PermissionUpdateRole, PermissionDeleteRole,
			PermissionCreateUser, PermissionReadUser, PermissionUpdateUser, PermissionDeleteUser,
			PermissionCreateAnyUser, PermissionCreateContributor,
		}

		for _, permission := range permissions {
			result := service.CanPerform(userCtx, permission)
			assert.True(t, result, "Super User should be able to perform %s", permission)
		}
	})

	t.Run("Product Owner permissions", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "po-user-id",
			RoleName: "Product Owner",
		}

		// Can perform these actions
		allowedPermissions := []Permission{
			PermissionReadRole, PermissionReadUser,
			PermissionCreateUser, PermissionUpdateUser, PermissionDeleteUser,
			PermissionCreateContributor,
		}

		for _, permission := range allowedPermissions {
			result := service.CanPerform(userCtx, permission)
			assert.True(t, result, "Product Owner should be able to perform %s", permission)
		}

		// Cannot perform these actions
		deniedPermissions := []Permission{
			PermissionCreateRole, PermissionUpdateRole, PermissionDeleteRole,
			PermissionCreateAnyUser,
		}

		for _, permission := range deniedPermissions {
			result := service.CanPerform(userCtx, permission)
			assert.False(t, result, "Product Owner should NOT be able to perform %s", permission)
		}
	})

	t.Run("Contributor permissions", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "contributor-user-id",
			RoleName: "Contributor",
		}

		// Can perform these actions
		allowedPermissions := []Permission{
			PermissionReadRole, PermissionReadUser,
		}

		for _, permission := range allowedPermissions {
			result := service.CanPerform(userCtx, permission)
			assert.True(t, result, "Contributor should be able to perform %s", permission)
		}

		// Cannot perform these actions
		deniedPermissions := []Permission{
			PermissionCreateRole, PermissionUpdateRole, PermissionDeleteRole,
			PermissionCreateUser, PermissionUpdateUser, PermissionDeleteUser,
			PermissionCreateAnyUser, PermissionCreateContributor,
		}

		for _, permission := range deniedPermissions {
			result := service.CanPerform(userCtx, permission)
			assert.False(t, result, "Contributor should NOT be able to perform %s", permission)
		}
	})

	t.Run("nil user context", func(t *testing.T) {
		result := service.CanPerform(nil, PermissionReadUser)
		assert.False(t, result)
	})

	t.Run("unknown role", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "unknown-user-id",
			RoleName: "Unknown Role",
		}

		result := service.CanPerform(userCtx, PermissionReadUser)
		assert.False(t, result)
	})
}

func TestAuthorizationService_CanCreateUserWithRole(t *testing.T) {
	service := NewAuthorizationService()

	t.Run("Super User can create users with any role", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "super-user-id",
			RoleName: role.SuperUserRoleName,
		}

		roles := []string{role.SuperUserRoleName, "Product Owner", "Contributor", "Custom Role"}

		for _, roleName := range roles {
			result := service.CanCreateUserWithRole(userCtx, roleName)
			assert.True(t, result, "Super User should be able to create user with role %s", roleName)
		}
	})

	t.Run("Product Owner can only create Contributors", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "po-user-id",
			RoleName: "Product Owner",
		}

		// Can create Contributors
		result := service.CanCreateUserWithRole(userCtx, "Contributor")
		assert.True(t, result)

		// Cannot create other roles
		deniedRoles := []string{role.SuperUserRoleName, "Product Owner", "Custom Role"}

		for _, roleName := range deniedRoles {
			result := service.CanCreateUserWithRole(userCtx, roleName)
			assert.False(t, result, "Product Owner should NOT be able to create user with role %s", roleName)
		}
	})

	t.Run("Contributor cannot create any users", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "contributor-user-id",
			RoleName: "Contributor",
		}

		roles := []string{role.SuperUserRoleName, "Product Owner", "Contributor"}

		for _, roleName := range roles {
			result := service.CanCreateUserWithRole(userCtx, roleName)
			assert.False(t, result, "Contributor should NOT be able to create user with role %s", roleName)
		}
	})

	t.Run("nil user context", func(t *testing.T) {
		result := service.CanCreateUserWithRole(nil, "Contributor")
		assert.False(t, result)
	})
}

func TestAuthorizationService_ValidateUserContext(t *testing.T) {
	service := NewAuthorizationService()

	t.Run("valid user context", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "user-123",
			RoleName: "Product Owner",
		}

		err := service.ValidateUserContext(userCtx)
		assert.NoError(t, err)
	})

	t.Run("nil user context", func(t *testing.T) {
		err := service.ValidateUserContext(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user context is required")
	})

	t.Run("empty user ID", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "",
			RoleName: "Product Owner",
		}

		err := service.ValidateUserContext(userCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("empty role name", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:   "user-123",
			RoleName: "",
		}

		err := service.ValidateUserContext(userCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role name is required")
	})
}
