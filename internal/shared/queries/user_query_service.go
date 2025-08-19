package queries

import (
	"context"
	userdomain "feedback_hub_2/internal/user/domain"
)

// UserQueryService implements UserQueries using the user domain
// AI-hint: Implementation of the shared user query interface that provides
// access to user data without creating cross-domain dependencies.
type UserQueryService struct {
	userRepo userdomain.Repository
}

// NewUserQueryService creates a new UserQueryService instance
func NewUserQueryService(userRepo userdomain.Repository) *UserQueryService {
	return &UserQueryService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by their ID
func (s *UserQueryService) GetUserByID(ctx context.Context, userID string) (*UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return NewUserInfo(user.ID, user.Email, user.Name, user.RoleID), nil
}

// GetUsersByRoleID retrieves all users assigned to a specific role
func (s *UserQueryService) GetUsersByRoleID(ctx context.Context, roleID string) ([]*UserInfo, error) {
	users, err := s.userRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	var userInfos []*UserInfo
	for _, user := range users {
		userInfos = append(userInfos, NewUserInfo(user.ID, user.Email, user.Name, user.RoleID))
	}

	return userInfos, nil
}

// UserExists checks if a user with the given ID exists
func (s *UserQueryService) UserExists(ctx context.Context, userID string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == userdomain.ErrUserNotFound {
			return false, nil
		}
		return false, err
	}

	return user != nil, nil
}
