package events

// RoleCreatedEvent represents the event when a role is created.
// AI-hint: Domain event for role creation, allowing other domains to react
// to new role availability and set up related permissions.
type RoleCreatedEvent struct {
	BaseDomainEvent
	RoleID string `json:"role_id"`
	Name   string `json:"name"`
}

// NewRoleCreatedEvent creates a new role created event.
// AI-hint: Factory method for role creation events.
func NewRoleCreatedEvent(roleID, name string) *RoleCreatedEvent {
	return &RoleCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("role.created", roleID, 1),
		RoleID:          roleID,
		Name:            name,
	}
}

// RoleUpdatedEvent represents the event when a role is updated.
// AI-hint: Domain event for role updates, critical for notifying
// other domains about role changes that may affect permissions.
type RoleUpdatedEvent struct {
	BaseDomainEvent
	RoleID string `json:"role_id"`
	Name   string `json:"name"`
}

// NewRoleUpdatedEvent creates a new role updated event.
// AI-hint: Factory method for role update events.
func NewRoleUpdatedEvent(roleID, name string, version int) *RoleUpdatedEvent {
	return &RoleUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("role.updated", roleID, version),
		RoleID:          roleID,
		Name:            name,
	}
}

// RoleDeletedEvent represents the event when a role is deleted.
// AI-hint: Domain event for role deletion, allowing other domains to clean up
// related data and handle users who may have had this role.
type RoleDeletedEvent struct {
	BaseDomainEvent
	RoleID string `json:"role_id"`
	Name   string `json:"name"`
}

// NewRoleDeletedEvent creates a new role deleted event.
// AI-hint: Factory method for role deletion events.
func NewRoleDeletedEvent(roleID, name string, version int) *RoleDeletedEvent {
	return &RoleDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent("role.deleted", roleID, version),
		RoleID:          roleID,
		Name:            name,
	}
}
