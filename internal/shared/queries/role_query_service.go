package queries

import (
	"context"
	roledomain "feedback_hub_2/internal/role/domain"
)

// RoleQueryService implements RoleQueries using the role domain
// AI-hint: Implementation of the shared role query interface that provides
// access to role data without creating cross-domain dependencies.
type RoleQueryService struct {
	roleRepo roledomain.Repository
}

// NewRoleQueryService creates a new RoleQueryService instance
func NewRoleQueryService(roleRepo roledomain.Repository) *RoleQueryService {
	return &RoleQueryService{
		roleRepo: roleRepo,
	}
}

// GetRoleByID retrieves a role by its ID
func (s *RoleQueryService) GetRoleByID(ctx context.Context, roleID string) (*RoleInfo, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return NewRoleInfo(role.ID, role.Name), nil
}

// GetRoleByName retrieves a role by its name
func (s *RoleQueryService) GetRoleByName(ctx context.Context, name string) (*RoleInfo, error) {
	role, err := s.roleRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return NewRoleInfo(role.ID, role.Name), nil
}

// RoleExists checks if a role with the given name exists
func (s *RoleQueryService) RoleExists(ctx context.Context, name string) (bool, error) {
	return s.roleRepo.Exists(ctx, name)
}
