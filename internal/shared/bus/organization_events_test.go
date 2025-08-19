package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrganizationCreatedEvent tests the creation and properties of organization created events.
// AI-hint: Comprehensive testing of organization creation events to ensure proper
// initialization and data integrity for event-driven communication.
func TestOrganizationCreatedEvent(t *testing.T) {
	t.Run("should create organization created event with correct properties", func(t *testing.T) {
		// Arrange
		orgID := "org-123"
		name := "Test Organization"
		slug := "test-org"
		description := "A test organization"
		createdByUserID := "user-456"

		// Act
		event := NewOrganizationCreatedEvent(orgID, name, slug, description, createdByUserID)

		// Assert
		require.NotNil(t, event)
		assert.Equal(t, "organization.created", event.EventType())
		assert.Equal(t, orgID, event.AggregateID())
		assert.Equal(t, orgID, event.OrganizationID)
		assert.Equal(t, name, event.Name)
		assert.Equal(t, slug, event.Slug)
		assert.Equal(t, description, event.Description)
		assert.Equal(t, createdByUserID, event.CreatedByUserID)
		assert.Equal(t, 1, event.Version())
		assert.NotEmpty(t, event.EventID())
		assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
	})

	t.Run("should generate unique event IDs for different events", func(t *testing.T) {
		// Arrange & Act
		event1 := NewOrganizationCreatedEvent("org-1", "Org 1", "org-1", "First org", "user-1")
		// Add a small delay to ensure different timestamps
		time.Sleep(time.Microsecond)
		event2 := NewOrganizationCreatedEvent("org-2", "Org 2", "org-2", "Second org", "user-2")

		// Assert
		assert.NotEqual(t, event1.EventID(), event2.EventID())
		assert.NotEqual(t, event1.OccurredAt(), event2.OccurredAt())
	})

	t.Run("should handle empty description and created by user ID", func(t *testing.T) {
		// Arrange & Act
		event := NewOrganizationCreatedEvent("org-123", "Test Org", "test-org", "", "")

		// Assert
		assert.Equal(t, "", event.Description)
		assert.Equal(t, "", event.CreatedByUserID)
	})
}

// TestOrganizationUpdatedEvent tests the creation and properties of organization updated events.
// AI-hint: Testing organization update events with change tracking to ensure proper
// event data for downstream consumers that need to react to specific changes.
func TestOrganizationUpdatedEvent(t *testing.T) {
	t.Run("should create organization updated event with correct properties", func(t *testing.T) {
		// Arrange
		orgID := "org-123"
		name := "Updated Organization"
		slug := "updated-org"
		description := "An updated organization"
		settings := map[string]any{"theme": "dark", "timezone": "UTC"}
		updatedByUserID := "user-789"
		changes := map[string]Change{
			"name": {OldValue: "Old Name", NewValue: name},
			"slug": {OldValue: "old-org", NewValue: slug},
		}
		version := 2

		// Act
		event := NewOrganizationUpdatedEvent(orgID, name, slug, description, settings, updatedByUserID, changes, version)

		// Assert
		require.NotNil(t, event)
		assert.Equal(t, "organization.updated", event.EventType())
		assert.Equal(t, orgID, event.AggregateID())
		assert.Equal(t, orgID, event.OrganizationID)
		assert.Equal(t, name, event.Name)
		assert.Equal(t, slug, event.Slug)
		assert.Equal(t, description, event.Description)
		assert.Equal(t, settings, event.Settings)
		assert.Equal(t, updatedByUserID, event.UpdatedByUserID)
		assert.Equal(t, changes, event.Changes)
		assert.Equal(t, version, event.Version())
	})

	t.Run("should handle empty changes map", func(t *testing.T) {
		// Arrange & Act
		event := NewOrganizationUpdatedEvent("org-123", "Test Org", "test-org", "Description", nil, "user-1", nil, 1)

		// Assert
		assert.Nil(t, event.Changes)
		assert.Nil(t, event.Settings)
	})

	t.Run("should handle empty settings map", func(t *testing.T) {
		// Arrange
		settings := map[string]any{}

		// Act
		event := NewOrganizationUpdatedEvent("org-123", "Test Org", "test-org", "Description", settings, "user-1", nil, 1)

		// Assert
		assert.Equal(t, settings, event.Settings)
	})
}

// TestOrganizationDeletedEvent tests the creation and properties of organization deleted events.
// AI-hint: Testing organization deletion events to ensure proper cleanup context
// is provided for downstream consumers that need to remove organization-related data.
func TestOrganizationDeletedEvent(t *testing.T) {
	t.Run("should create organization deleted event with correct properties", func(t *testing.T) {
		// Arrange
		orgID := "org-123"
		name := "Deleted Organization"
		slug := "deleted-org"
		deletedByUserID := "user-999"
		deletionReason := "Company closure"
		version := 3

		// Act
		event := NewOrganizationDeletedEvent(orgID, name, slug, deletedByUserID, deletionReason, version)

		// Assert
		require.NotNil(t, event)
		assert.Equal(t, "organization.deleted", event.EventType())
		assert.Equal(t, orgID, event.AggregateID())
		assert.Equal(t, orgID, event.OrganizationID)
		assert.Equal(t, name, event.Name)
		assert.Equal(t, slug, event.Slug)
		assert.Equal(t, deletedByUserID, event.DeletedByUserID)
		assert.Equal(t, deletionReason, event.DeletionReason)
		assert.Equal(t, version, event.Version())
	})

	t.Run("should handle empty deletion reason", func(t *testing.T) {
		// Arrange & Act
		event := NewOrganizationDeletedEvent("org-123", "Test Org", "test-org", "user-1", "", 1)

		// Assert
		assert.Equal(t, "", event.DeletionReason)
	})
}

// TestUserJoinedOrganizationEvent tests the creation and properties of user joined organization events.
// AI-hint: Testing user-organization join events to ensure proper relationship
// creation context for downstream consumers that need to set up user resources.
func TestUserJoinedOrganizationEvent(t *testing.T) {
	t.Run("should create user joined organization event with correct properties", func(t *testing.T) {
		// Arrange
		orgID := "org-123"
		userID := "user-456"
		roleID := "role-789"
		roleName := "Member"
		joinedByUserID := "user-999"

		// Act
		event := NewUserJoinedOrganizationEvent(orgID, userID, roleID, roleName, joinedByUserID)

		// Assert
		require.NotNil(t, event)
		assert.Equal(t, "user.joined_organization", event.EventType())
		assert.Equal(t, orgID, event.AggregateID())
		assert.Equal(t, orgID, event.OrganizationID)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, roleID, event.RoleID)
		assert.Equal(t, roleName, event.RoleName)
		assert.Equal(t, joinedByUserID, event.JoinedByUserID)
		assert.Equal(t, 1, event.Version())
	})

	t.Run("should handle empty joined by user ID", func(t *testing.T) {
		// Arrange & Act
		event := NewUserJoinedOrganizationEvent("org-123", "user-456", "role-789", "Member", "")

		// Assert
		assert.Equal(t, "", event.JoinedByUserID)
	})
}

// TestUserLeftOrganizationEvent tests the creation and properties of user left organization events.
// AI-hint: Testing user-organization leave events to ensure proper relationship
// termination context for downstream consumers that need to clean up user resources.
func TestUserLeftOrganizationEvent(t *testing.T) {
	t.Run("should create user left organization event with correct properties", func(t *testing.T) {
		// Arrange
		orgID := "org-123"
		userID := "user-456"
		roleID := "role-789"
		leftByUserID := "user-999"
		leaveReason := "User request"
		version := 2

		// Act
		event := NewUserLeftOrganizationEvent(orgID, userID, roleID, leftByUserID, leaveReason, version)

		// Assert
		require.NotNil(t, event)
		assert.Equal(t, "user.left_organization", event.EventType())
		assert.Equal(t, orgID, event.AggregateID())
		assert.Equal(t, orgID, event.OrganizationID)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, roleID, event.RoleID)
		assert.Equal(t, leftByUserID, event.LeftByUserID)
		assert.Equal(t, leaveReason, event.LeaveReason)
		assert.Equal(t, version, event.Version())
	})

	t.Run("should handle empty leave reason", func(t *testing.T) {
		// Arrange & Act
		event := NewUserLeftOrganizationEvent("org-123", "user-456", "role-789", "user-999", "", 1)

		// Assert
		assert.Equal(t, "", event.LeaveReason)
	})
}

// TestUserRoleChangedInOrganizationEvent tests the creation and properties of user role change events.
// AI-hint: Testing user role change events within organizations to ensure proper
// permission update context for downstream consumers that need to adjust access control.
func TestUserRoleChangedInOrganizationEvent(t *testing.T) {
	t.Run("should create user role changed in organization event with correct properties", func(t *testing.T) {
		// Arrange
		orgID := "org-123"
		userID := "user-456"
		oldRoleID := "role-789"
		newRoleID := "role-999"
		oldRoleName := "Member"
		newRoleName := "Admin"
		changedByUserID := "user-888"
		version := 3

		// Act
		event := NewUserRoleChangedInOrganizationEvent(orgID, userID, oldRoleID, newRoleID, oldRoleName, newRoleName, changedByUserID, version)

		// Assert
		require.NotNil(t, event)
		assert.Equal(t, "user.role_changed_in_organization", event.EventType())
		assert.Equal(t, orgID, event.AggregateID())
		assert.Equal(t, orgID, event.OrganizationID)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, oldRoleID, event.OldRoleID)
		assert.Equal(t, newRoleID, event.NewRoleID)
		assert.Equal(t, oldRoleName, event.OldRoleName)
		assert.Equal(t, newRoleName, event.NewRoleName)
		assert.Equal(t, changedByUserID, event.ChangedByUserID)
		assert.Equal(t, version, event.Version())
	})

	t.Run("should handle empty changed by user ID", func(t *testing.T) {
		// Arrange & Act
		event := NewUserRoleChangedInOrganizationEvent("org-123", "user-456", "role-789", "role-999", "Member", "Admin", "", 1)

		// Assert
		assert.Equal(t, "", event.ChangedByUserID)
	})
}

// TestChangeStruct tests the Change helper struct for tracking field changes.
// AI-hint: Testing the Change struct to ensure proper tracking of field modifications
// for downstream consumers that need to react to specific changes.
func TestChangeStruct(t *testing.T) {
	t.Run("should create change with old and new values", func(t *testing.T) {
		// Arrange
		oldValue := "old value"
		newValue := "new value"

		// Act
		change := Change{
			OldValue: oldValue,
			NewValue: newValue,
		}

		// Assert
		assert.Equal(t, oldValue, change.OldValue)
		assert.Equal(t, newValue, change.NewValue)
	})

	t.Run("should handle nil values", func(t *testing.T) {
		// Arrange & Act
		change := Change{
			OldValue: nil,
			NewValue: "new value",
		}

		// Assert
		assert.Nil(t, change.OldValue)
		assert.Equal(t, "new value", change.NewValue)
	})
}

// TestOrganizationEventsIntegration tests the integration of organization events with the event system.
// AI-hint: Integration testing to ensure organization events work properly
// with the event bus and can be published and consumed correctly.
func TestOrganizationEventsIntegration(t *testing.T) {
	t.Run("should publish and handle organization created event", func(t *testing.T) {
		// Arrange
		eventBus := NewInMemoryEventBus()
		var receivedEvent *OrganizationCreatedEvent

		handler := func(ctx context.Context, event DomainEvent) error {
			if orgEvent, ok := event.(*OrganizationCreatedEvent); ok {
				receivedEvent = orgEvent
			}
			return nil
		}

		// Subscribe to organization events
		err := eventBus.Subscribe("organization.created", handler)
		require.NoError(t, err)

		// Create and publish event
		event := NewOrganizationCreatedEvent("org-123", "Test Org", "test-org", "Description", "user-1")
		err = eventBus.Publish(context.Background(), event)
		require.NoError(t, err)

		// Assert
		require.NotNil(t, receivedEvent)
		assert.Equal(t, event.OrganizationID, receivedEvent.OrganizationID)
		assert.Equal(t, event.Name, receivedEvent.Name)
		assert.Equal(t, event.Slug, receivedEvent.Slug)
	})

	t.Run("should handle multiple organization event types", func(t *testing.T) {
		// Arrange
		eventBus := NewInMemoryEventBus()
		var receivedEvents []DomainEvent

		handler := func(ctx context.Context, event DomainEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		// Subscribe to multiple event types
		err := eventBus.Subscribe("organization.created", handler)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.updated", handler)
		require.NoError(t, err)

		// Create and publish multiple events
		createdEvent := NewOrganizationCreatedEvent("org-1", "Org 1", "org-1", "First org", "user-1")
		updatedEvent := NewOrganizationUpdatedEvent("org-1", "Org 1 Updated", "org-1", "Updated org", nil, "user-2", nil, 2)

		err = eventBus.Publish(context.Background(), createdEvent)
		require.NoError(t, err)
		err = eventBus.Publish(context.Background(), updatedEvent)
		require.NoError(t, err)

		// Assert
		assert.Len(t, receivedEvents, 2)
		assert.Equal(t, "organization.created", receivedEvents[0].EventType())
		assert.Equal(t, "organization.updated", receivedEvents[1].EventType())
	})
}
