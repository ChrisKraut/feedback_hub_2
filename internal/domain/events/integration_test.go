package events

import (
	"context"
	"testing"
)

// TestDomainEventsIntegration demonstrates the complete domain events system working end-to-end.
// AI-hint: This test shows how the event system enables cross-domain communication
// without direct coupling between services.
func TestDomainEventsIntegration(t *testing.T) {
	t.Run("should demonstrate complete event flow", func(t *testing.T) {
		// Create the event system
		eventBus := NewInMemoryEventBus()
		eventPublisher := NewEventBusPublisher(eventBus)

		// Create test event handlers
		testHandlers := NewTestEventHandlers()

		// Subscribe handlers to events
		eventBus.Subscribe("user.created", testHandlers.HandleUserCreated)
		eventBus.Subscribe("user.updated", testHandlers.HandleUserUpdated)
		eventBus.Subscribe("user.role_updated", testHandlers.HandleUserRoleUpdated)
		eventBus.Subscribe("role.created", testHandlers.HandleRoleCreated)
		eventBus.Subscribe("role.updated", testHandlers.HandleRoleUpdated)
		eventBus.Subscribe("role.deleted", testHandlers.HandleRoleDeleted)

		// Simulate business operations that would trigger events
		ctx := context.Background()

		// Simulate creating a role
		roleCreatedEvent := NewRoleCreatedEvent("role-123", "Contributor")
		err := eventPublisher.PublishEvent(ctx, roleCreatedEvent)
		if err != nil {
			t.Errorf("failed to publish role created event: %v", err)
		}

		// Simulate creating a user
		userCreatedEvent := NewUserCreatedEvent("user-123", "test@example.com", "Test User", "role-123", "Contributor")
		err = eventPublisher.PublishEvent(ctx, userCreatedEvent)
		if err != nil {
			t.Errorf("failed to publish user created event: %v", err)
		}

		// Simulate updating a user
		userUpdatedEvent := NewUserUpdatedEvent("user-123", "Updated Test User", 2)
		err = eventPublisher.PublishEvent(ctx, userUpdatedEvent)
		if err != nil {
			t.Errorf("failed to publish user updated event: %v", err)
		}

		// Simulate updating a user's role
		userRoleUpdatedEvent := NewUserRoleUpdatedEvent("user-123", "role-123", "role-456", "Contributor", "Product Owner", 3)
		err = eventPublisher.PublishEvent(ctx, userRoleUpdatedEvent)
		if err != nil {
			t.Errorf("failed to publish user role updated event: %v", err)
		}

		// Simulate updating a role
		roleUpdatedEvent := NewRoleUpdatedEvent("role-123", "Updated Contributor", 2)
		err = eventPublisher.PublishEvent(ctx, roleUpdatedEvent)
		if err != nil {
			t.Errorf("failed to publish role updated event: %v", err)
		}

		// Simulate deleting a role
		roleDeletedEvent := NewRoleDeletedEvent("role-456", "Product Owner", 3)
		err = eventPublisher.PublishEvent(ctx, roleDeletedEvent)
		if err != nil {
			t.Errorf("failed to publish role deleted event: %v", err)
		}

		// Verify all events were handled
		counts := testHandlers.GetCounts()
		expectedCounts := map[string]int{
			"user_created":     1,
			"user_updated":     1,
			"user_role_updated": 1,
			"role_created":     1,
			"role_updated":     1,
			"role_deleted":     1,
		}

		for eventType, expectedCount := range expectedCounts {
			if counts[eventType] != expectedCount {
				t.Errorf("expected %s count to be %d, got %d", eventType, expectedCount, counts[eventType])
			}
		}
	})

	t.Run("should handle multiple handlers for same event type", func(t *testing.T) {
		eventBus := NewInMemoryEventBus()
		eventPublisher := NewEventBusPublisher(eventBus)

		// Create multiple handlers for the same event type
		handler1 := NewTestEventHandlers()
		handler2 := NewTestEventHandlers()

		// Subscribe both handlers to user created events
		eventBus.Subscribe("user.created", handler1.HandleUserCreated)
		eventBus.Subscribe("user.created", handler2.HandleUserCreated)

		// Publish a user created event
		ctx := context.Background()
		userCreatedEvent := NewUserCreatedEvent("user-123", "test@example.com", "Test User", "role-123", "Contributor")
		err := eventPublisher.PublishEvent(ctx, userCreatedEvent)
		if err != nil {
			t.Errorf("failed to publish user created event: %v", err)
		}

		// Verify both handlers received the event
		counts1 := handler1.GetCounts()
		counts2 := handler2.GetCounts()

		if counts1["user_created"] != 1 {
			t.Errorf("expected handler1 user created count to be 1, got %d", counts1["user_created"])
		}
		if counts2["user_created"] != 1 {
			t.Errorf("expected handler2 user created count to be 1, got %d", counts2["user_created"])
		}
	})

	t.Run("should handle handler errors gracefully", func(t *testing.T) {
		eventBus := NewInMemoryEventBus()
		eventPublisher := NewEventBusPublisher(eventBus)

		// Create a handler that always returns an error
		errorHandler := func(ctx context.Context, event DomainEvent) error {
			return context.DeadlineExceeded
		}

		// Subscribe the error handler
		eventBus.Subscribe("user.created", errorHandler)

		// Publish an event - should not fail the test
		ctx := context.Background()
		userCreatedEvent := NewUserCreatedEvent("user-123", "test@example.com", "Test User", "role-123", "Contributor")
		err := eventPublisher.PublishEvent(ctx, userCreatedEvent)
		if err == nil {
			t.Error("expected error when handler fails, but got none")
		}
	})

	t.Run("should handle events with no subscribers gracefully", func(t *testing.T) {
		eventBus := NewInMemoryEventBus()
		eventPublisher := NewEventBusPublisher(eventBus)

		// Publish an event with no subscribers - should not fail
		ctx := context.Background()
		userCreatedEvent := NewUserCreatedEvent("user-123", "test@example.com", "Test User", "role-123", "Contributor")
		err := eventPublisher.PublishEvent(ctx, userCreatedEvent)
		if err != nil {
			t.Errorf("expected no error when no subscribers, got %v", err)
		}
	})
}
