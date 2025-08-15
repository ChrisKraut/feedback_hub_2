package events

import (
	"context"
	"testing"
)

func TestTestEventHandlers(t *testing.T) {
	t.Run("should create new test event handlers", func(t *testing.T) {
		handlers := NewTestEventHandlers()
		if handlers == nil {
			t.Error("expected handlers to be created")
		}
	})

	t.Run("should handle user created events", func(t *testing.T) {
		handlers := NewTestEventHandlers()
		event := NewUserCreatedEvent("user-123", "test@example.com", "Test User", "role-456", "Contributor")

		err := handlers.HandleUserCreated(context.Background(), event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		counts := handlers.GetCounts()
		if counts["user_created"] != 1 {
			t.Errorf("expected user created count to be 1, got %d", counts["user_created"])
		}
	})

	t.Run("should handle user updated events", func(t *testing.T) {
		handlers := NewTestEventHandlers()
		event := NewUserUpdatedEvent("user-123", "Updated Name", 2)

		err := handlers.HandleUserUpdated(context.Background(), event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		counts := handlers.GetCounts()
		if counts["user_updated"] != 1 {
			t.Errorf("expected user updated count to be 1, got %d", counts["user_updated"])
		}
	})

	t.Run("should handle user role updated events", func(t *testing.T) {
		handlers := NewTestEventHandlers()
		event := NewUserRoleUpdatedEvent("user-123", "role-456", "role-789", "Contributor", "Product Owner", 3)

		err := handlers.HandleUserRoleUpdated(context.Background(), event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		counts := handlers.GetCounts()
		if counts["user_role_updated"] != 1 {
			t.Errorf("expected user role updated count to be 1, got %d", counts["user_role_updated"])
		}
	})

	t.Run("should handle role created events", func(t *testing.T) {
		handlers := NewTestEventHandlers()
		event := NewRoleCreatedEvent("role-123", "New Role")

		err := handlers.HandleRoleCreated(context.Background(), event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		counts := handlers.GetCounts()
		if counts["role_created"] != 1 {
			t.Errorf("expected role created count to be 1, got %d", counts["role_created"])
		}
	})

	t.Run("should handle role updated events", func(t *testing.T) {
		handlers := NewTestEventHandlers()
		event := NewRoleUpdatedEvent("role-123", "Updated Role", 2)

		err := handlers.HandleRoleUpdated(context.Background(), event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		counts := handlers.GetCounts()
		if counts["role_updated"] != 1 {
			t.Errorf("expected role updated count to be 1, got %d", counts["role_updated"])
		}
	})

	t.Run("should handle role deleted events", func(t *testing.T) {
		handlers := NewTestEventHandlers()
		event := NewRoleDeletedEvent("role-123", "Deleted Role", 3)

		err := handlers.HandleRoleDeleted(context.Background(), event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		counts := handlers.GetCounts()
		if counts["role_deleted"] != 1 {
			t.Errorf("expected role deleted count to be 1, got %d", counts["role_deleted"])
		}
	})

	t.Run("should accumulate multiple events", func(t *testing.T) {
		handlers := NewTestEventHandlers()

		// Handle multiple events
		handlers.HandleUserCreated(context.Background(), NewUserCreatedEvent("user-1", "test1@example.com", "User 1", "role-1", "Contributor"))
		handlers.HandleUserCreated(context.Background(), NewUserCreatedEvent("user-2", "test2@example.com", "User 2", "role-2", "Product Owner"))
		handlers.HandleRoleCreated(context.Background(), NewRoleCreatedEvent("role-1", "Role 1"))

		counts := handlers.GetCounts()
		if counts["user_created"] != 2 {
			t.Errorf("expected user created count to be 2, got %d", counts["user_created"])
		}
		if counts["role_created"] != 1 {
			t.Errorf("expected role created count to be 1, got %d", counts["role_created"])
		}
	})

	t.Run("should reset counts", func(t *testing.T) {
		handlers := NewTestEventHandlers()

		// Handle some events
		handlers.HandleUserCreated(context.Background(), NewUserCreatedEvent("user-123", "test@example.com", "Test User", "role-456", "Contributor"))
		handlers.HandleRoleCreated(context.Background(), NewRoleCreatedEvent("role-123", "Test Role"))

		// Verify counts are not zero
		counts := handlers.GetCounts()
		if counts["user_created"] == 0 || counts["role_created"] == 0 {
			t.Error("expected counts to be non-zero before reset")
		}

		// Reset counts
		handlers.ResetCounts()

		// Verify counts are zero
		counts = handlers.GetCounts()
		if counts["user_created"] != 0 {
			t.Errorf("expected user created count to be 0 after reset, got %d", counts["user_created"])
		}
		if counts["role_created"] != 0 {
			t.Errorf("expected role created count to be 0 after reset, got %d", counts["role_created"])
		}
	})
}
