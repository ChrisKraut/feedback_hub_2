package events

// OrganizationCreatedEvent represents the event when an organization is created.
// AI-hint: Domain event for organization creation, allowing other domains to react
// to new organization setup and initialize necessary resources.
type OrganizationCreatedEvent struct {
	BaseDomainEvent
	OrganizationID  string `json:"organization_id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Description     string `json:"description"`
	CreatedByUserID string `json:"created_by_user_id"`
}

// NewOrganizationCreatedEvent creates a new organization created event.
// AI-hint: Factory method for organization creation events with proper initialization.
func NewOrganizationCreatedEvent(organizationID, name, slug, description, createdByUserID string) *OrganizationCreatedEvent {
	return &OrganizationCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("organization.created", organizationID, 1),
		OrganizationID:  organizationID,
		Name:            name,
		Slug:            slug,
		Description:     description,
		CreatedByUserID: createdByUserID,
	}
}

// OrganizationUpdatedEvent represents the event when an organization is updated.
// AI-hint: Domain event for organization updates, allowing other domains to react
// to changes in organization information and settings.
type OrganizationUpdatedEvent struct {
	BaseDomainEvent
	OrganizationID  string            `json:"organization_id"`
	Name            string            `json:"name"`
	Slug            string            `json:"slug"`
	Description     string            `json:"description"`
	Settings        map[string]any    `json:"settings"`
	UpdatedByUserID string            `json:"updated_by_user_id"`
	Changes         map[string]Change `json:"changes"`
}

// Change represents a field change in an organization update.
// AI-hint: Helper struct to track what specific fields changed during updates,
// enabling other domains to react only to relevant changes.
type Change struct {
	OldValue any `json:"old_value"`
	NewValue any `json:"new_value"`
}

// NewOrganizationUpdatedEvent creates a new organization updated event.
// AI-hint: Factory method for organization update events with change tracking.
func NewOrganizationUpdatedEvent(organizationID, name, slug, description string, settings map[string]any, updatedByUserID string, changes map[string]Change, version int) *OrganizationUpdatedEvent {
	return &OrganizationUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("organization.updated", organizationID, version),
		OrganizationID:  organizationID,
		Name:            name,
		Slug:            slug,
		Description:     description,
		Settings:        settings,
		UpdatedByUserID: updatedByUserID,
		Changes:         changes,
	}
}

// OrganizationDeletedEvent represents the event when an organization is deleted.
// AI-hint: Critical domain event for organization deletion, allowing other domains
// to clean up organization-related data and resources.
type OrganizationDeletedEvent struct {
	BaseDomainEvent
	OrganizationID  string `json:"organization_id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	DeletedByUserID string `json:"deleted_by_user_id"`
	DeletionReason  string `json:"deletion_reason"`
}

// NewOrganizationDeletedEvent creates a new organization deleted event.
// AI-hint: Factory method for organization deletion events with cleanup context.
func NewOrganizationDeletedEvent(organizationID, name, slug, deletedByUserID, deletionReason string, version int) *OrganizationDeletedEvent {
	return &OrganizationDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent("organization.deleted", organizationID, version),
		OrganizationID:  organizationID,
		Name:            name,
		Slug:            slug,
		DeletedByUserID: deletedByUserID,
		DeletionReason:  deletionReason,
	}
}

// UserJoinedOrganizationEvent represents the event when a user joins an organization.
// AI-hint: Domain event for user-organization relationship creation, allowing
// other domains to set up user-specific resources within the organization.
type UserJoinedOrganizationEvent struct {
	BaseDomainEvent
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	RoleID         string `json:"role_id"`
	RoleName       string `json:"role_name"`
	JoinedByUserID string `json:"joined_by_user_id"`
}

// NewUserJoinedOrganizationEvent creates a new user joined organization event.
// AI-hint: Factory method for user-organization join events.
func NewUserJoinedOrganizationEvent(organizationID, userID, roleID, roleName, joinedByUserID string) *UserJoinedOrganizationEvent {
	return &UserJoinedOrganizationEvent{
		BaseDomainEvent: NewBaseDomainEvent("user.joined_organization", organizationID, 1),
		OrganizationID:  organizationID,
		UserID:          userID,
		RoleID:          roleID,
		RoleName:        roleName,
		JoinedByUserID:  joinedByUserID,
	}
}

// UserLeftOrganizationEvent represents the event when a user leaves an organization.
// AI-hint: Domain event for user-organization relationship termination, allowing
// other domains to clean up user-specific resources within the organization.
type UserLeftOrganizationEvent struct {
	BaseDomainEvent
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	RoleID         string `json:"role_id"`
	LeftByUserID   string `json:"left_by_user_id"`
	LeaveReason    string `json:"leave_reason"`
}

// NewUserLeftOrganizationEvent creates a new user left organization event.
// AI-hint: Factory method for user-organization leave events.
func NewUserLeftOrganizationEvent(organizationID, userID, roleID, leftByUserID, leaveReason string, version int) *UserLeftOrganizationEvent {
	return &UserLeftOrganizationEvent{
		BaseDomainEvent: NewBaseDomainEvent("user.left_organization", organizationID, version),
		OrganizationID:  organizationID,
		UserID:          userID,
		RoleID:          roleID,
		LeftByUserID:    leftByUserID,
		LeaveReason:     leaveReason,
	}
}

// UserRoleChangedInOrganizationEvent represents the event when a user's role changes within an organization.
// AI-hint: Domain event for role changes within organizations, critical for
// permission updates and access control across the system.
type UserRoleChangedInOrganizationEvent struct {
	BaseDomainEvent
	OrganizationID  string `json:"organization_id"`
	UserID          string `json:"user_id"`
	OldRoleID       string `json:"old_role_id"`
	NewRoleID       string `json:"new_role_id"`
	OldRoleName     string `json:"old_role_name"`
	NewRoleName     string `json:"new_role_name"`
	ChangedByUserID string `json:"changed_by_user_id"`
}

// NewUserRoleChangedInOrganizationEvent creates a new user role changed in organization event.
// AI-hint: Factory method for user role change events within organizations.
func NewUserRoleChangedInOrganizationEvent(organizationID, userID, oldRoleID, newRoleID, oldRoleName, newRoleName, changedByUserID string, version int) *UserRoleChangedInOrganizationEvent {
	return &UserRoleChangedInOrganizationEvent{
		BaseDomainEvent: NewBaseDomainEvent("user.role_changed_in_organization", organizationID, version),
		OrganizationID:  organizationID,
		UserID:          userID,
		OldRoleID:       oldRoleID,
		NewRoleID:       newRoleID,
		OldRoleName:     oldRoleName,
		NewRoleName:     newRoleName,
		ChangedByUserID: changedByUserID,
	}
}
