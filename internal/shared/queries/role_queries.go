package queries

import (
	"context"
)

// RoleQueries provides read-only access to role information
// AI-hint: Shared query interface that allows domains to access role data
// without creating direct dependencies on the role domain.
type RoleQueries interface {
	// GetRoleByID retrieves a role by its ID
	GetRoleByID(ctx context.Context, roleID string) (*RoleInfo, error)

	// GetRoleByName retrieves a role by its name
	GetRoleByName(ctx context.Context, name string) (*RoleInfo, error)

	// RoleExists checks if a role with the given name exists
	RoleExists(ctx context.Context, name string) (bool, error)
}

// RoleInfo represents role information for cross-domain queries
// AI-hint: DTO that provides role data without exposing internal role domain structures.
type RoleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewRoleInfo creates a new RoleInfo instance
func NewRoleInfo(id, name string) *RoleInfo {
	return &RoleInfo{
		ID:   id,
		Name: name,
	}
}
