package application

import (
	"context"
	events "feedback_hub_2/internal/shared/bus"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserEventHandlers(t *testing.T) {
	t.Run("should_handle_organization_created_event", func(t *testing.T) {
		// Arrange
		eventHandlers := &UserEventHandlers{}
		event := events.NewOrganizationCreatedEvent("org123", "Test Org", "test-org", "Test Description", "user123")

		// Act
		err := eventHandlers.HandleOrganizationCreated(context.Background(), event)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should_handle_organization_updated_event", func(t *testing.T) {
		// Arrange
		eventHandlers := &UserEventHandlers{}
		event := events.NewOrganizationUpdatedEvent("org123", "Test Org", "test-org", "Test Description", nil, "user123", nil, 1)

		// Act
		err := eventHandlers.HandleOrganizationUpdated(context.Background(), event)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should_handle_organization_deleted_event", func(t *testing.T) {
		// Arrange
		eventHandlers := &UserEventHandlers{}
		event := events.NewOrganizationDeletedEvent("org123", "Test Org", "test-org", "user123", "Testing", 1)

		// Act
		err := eventHandlers.HandleOrganizationDeleted(context.Background(), event)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should_handle_user_joined_organization_event", func(t *testing.T) {
		// Arrange
		eventHandlers := &UserEventHandlers{}
		event := events.NewUserJoinedOrganizationEvent("org123", "user123", "role123", "Admin", "admin123")

		// Act
		err := eventHandlers.HandleUserJoinedOrganization(context.Background(), event)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should_handle_user_left_organization_event", func(t *testing.T) {
		// Arrange
		eventHandlers := &UserEventHandlers{}
		event := events.NewUserLeftOrganizationEvent("org123", "user123", "role123", "admin123", "Left voluntarily", 1)

		// Act
		err := eventHandlers.HandleUserLeftOrganization(context.Background(), event)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should_handle_user_role_changed_in_organization_event", func(t *testing.T) {
		// Arrange
		eventHandlers := &UserEventHandlers{}
		event := events.NewUserRoleChangedInOrganizationEvent("org123", "user123", "role123", "role456", "Admin", "User", "admin123", 1)

		// Act
		err := eventHandlers.HandleUserRoleChangedInOrganization(context.Background(), event)

		// Assert
		assert.NoError(t, err)
	})
}
