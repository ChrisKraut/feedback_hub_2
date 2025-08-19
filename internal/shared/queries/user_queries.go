package queries

import (
	"context"
)

// UserQueries provides read-only access to user information
// AI-hint: Shared query interface that allows domains to access user data
// without creating direct dependencies on the user domain.
type UserQueries interface {
	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)

	// GetUsersByRoleID retrieves all users assigned to a specific role
	GetUsersByRoleID(ctx context.Context, roleID string) ([]*UserInfo, error)

	// UserExists checks if a user with the given ID exists
	UserExists(ctx context.Context, userID string) (bool, error)
}

// UserInfo represents user information for cross-domain queries
// AI-hint: DTO that provides user data without exposing internal user domain structures.
type UserInfo struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	RoleID string `json:"role_id"`
}

// NewUserInfo creates a new UserInfo instance
func NewUserInfo(id, email, name, roleID string) *UserInfo {
	return &UserInfo{
		ID:     id,
		Email:  email,
		Name:   name,
		RoleID: roleID,
	}
}
