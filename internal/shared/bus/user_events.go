package events

// UserCreatedEvent represents the event when a user is created.
// AI-hint: Specific domain event for user creation, carrying relevant
// user data for other domains that need to react to user creation.
type UserCreatedEvent struct {
	BaseDomainEvent
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	RoleID   string `json:"role_id"`
	RoleName string `json:"role_name"`
}

// NewUserCreatedEvent creates a new user created event.
// AI-hint: Factory method for user creation events with proper initialization.
func NewUserCreatedEvent(userID, email, name, roleID, roleName string) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("user.created", userID, 1),
		UserID:          userID,
		Email:           email,
		Name:            name,
		RoleID:          roleID,
		RoleName:        roleName,
	}
}

// UserUpdatedEvent represents the event when a user is updated.
// AI-hint: Domain event for user updates, allowing other domains to react
// to changes in user information.
type UserUpdatedEvent struct {
	BaseDomainEvent
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// NewUserUpdatedEvent creates a new user updated event.
// AI-hint: Factory method for user update events.
func NewUserUpdatedEvent(userID, name string, version int) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("user.updated", userID, version),
		UserID:          userID,
		Name:            name,
	}
}

// UserRoleUpdatedEvent represents the event when a user's role is updated.
// AI-hint: Domain event for role changes, critical for authorization
// and permission updates across the system.
type UserRoleUpdatedEvent struct {
	BaseDomainEvent
	UserID    string `json:"user_id"`
	OldRoleID string `json:"old_role_id"`
	NewRoleID string `json:"new_role_id"`
	OldRole   string `json:"old_role"`
	NewRole   string `json:"new_role"`
}

// NewUserRoleUpdatedEvent creates a new user role updated event.
// AI-hint: Factory method for user role update events.
func NewUserRoleUpdatedEvent(userID, oldRoleID, newRoleID, oldRole, newRole string, version int) *UserRoleUpdatedEvent {
	return &UserRoleUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("user.role_updated", userID, version),
		UserID:          userID,
		OldRoleID:       oldRoleID,
		NewRoleID:       newRoleID,
		OldRole:         oldRole,
		NewRole:         newRole,
	}
}
