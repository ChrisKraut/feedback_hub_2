package events

import (
	"testing"
	"time"
)

func TestBaseDomainEvent(t *testing.T) {
	t.Run("should create base domain event with correct values", func(t *testing.T) {
		eventType := "test.event"
		aggregateID := "test-123"
		version := 1

		event := NewBaseDomainEvent(eventType, aggregateID, version)

		if event.EventType() != eventType {
			t.Errorf("expected event type %s, got %s", eventType, event.EventType())
		}

		if event.AggregateID() != aggregateID {
			t.Errorf("expected aggregate ID %s, got %s", aggregateID, event.AggregateID())
		}

		if event.Version() != version {
			t.Errorf("expected version %d, got %d", version, event.Version())
		}

		if event.EventID() == "" {
			t.Error("expected event ID to be generated")
		}

		if event.OccurredAt().IsZero() {
			t.Error("expected occurred at to be set")
		}
	})

	t.Run("should generate unique event IDs", func(t *testing.T) {
		event1 := NewBaseDomainEvent("test.event1", "test-1", 1)
		event2 := NewBaseDomainEvent("test.event2", "test-2", 1)

		if event1.EventID() == event2.EventID() {
			t.Error("expected different event IDs for different events")
		}
	})

	t.Run("should set occurred at to current time", func(t *testing.T) {
		before := time.Now()
		event := NewBaseDomainEvent("test.event", "test-123", 1)
		after := time.Now()

		if event.OccurredAt().Before(before) || event.OccurredAt().After(after) {
			t.Errorf("expected occurred at to be between %v and %v, got %v", before, after, event.OccurredAt())
		}
	})
}

func TestUserCreatedEvent(t *testing.T) {
	t.Run("should create user created event with correct values", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		name := "Test User"
		roleID := "role-456"
		roleName := "Contributor"

		event := NewUserCreatedEvent(userID, email, name, roleID, roleName)

		if event.EventType() != "user.created" {
			t.Errorf("expected event type 'user.created', got %s", event.EventType())
		}

		if event.AggregateID() != userID {
			t.Errorf("expected aggregate ID %s, got %s", userID, event.AggregateID())
		}

		if event.UserID != userID {
			t.Errorf("expected user ID %s, got %s", userID, event.UserID)
		}

		if event.Email != email {
			t.Errorf("expected email %s, got %s", email, event.Email)
		}

		if event.Name != name {
			t.Errorf("expected name %s, got %s", name, event.Name)
		}

		if event.RoleID != roleID {
			t.Errorf("expected role ID %s, got %s", roleID, event.RoleID)
		}

		if event.RoleName != roleName {
			t.Errorf("expected role name %s, got %s", roleName, event.RoleName)
		}

		if event.Version() != 1 {
			t.Errorf("expected version 1, got %d", event.Version())
		}
	})
}

func TestUserUpdatedEvent(t *testing.T) {
	t.Run("should create user updated event with correct values", func(t *testing.T) {
		userID := "user-123"
		name := "Updated Name"
		version := 2

		event := NewUserUpdatedEvent(userID, name, version)

		if event.EventType() != "user.updated" {
			t.Errorf("expected event type 'user.updated', got %s", event.EventType())
		}

		if event.AggregateID() != userID {
			t.Errorf("expected aggregate ID %s, got %s", userID, event.AggregateID())
		}

		if event.UserID != userID {
			t.Errorf("expected user ID %s, got %s", userID, event.UserID)
		}

		if event.Name != name {
			t.Errorf("expected name %s, got %s", name, event.Name)
		}

		if event.Version() != version {
			t.Errorf("expected version %d, got %d", version, event.Version())
		}
	})
}

func TestUserRoleUpdatedEvent(t *testing.T) {
	t.Run("should create user role updated event with correct values", func(t *testing.T) {
		userID := "user-123"
		oldRoleID := "role-456"
		newRoleID := "role-789"
		oldRole := "Contributor"
		newRole := "Product Owner"
		version := 3

		event := NewUserRoleUpdatedEvent(userID, oldRoleID, newRoleID, oldRole, newRole, version)

		if event.EventType() != "user.role_updated" {
			t.Errorf("expected event type 'user.role_updated', got %s", event.EventType())
		}

		if event.AggregateID() != userID {
			t.Errorf("expected aggregate ID %s, got %s", userID, event.AggregateID())
		}

		if event.UserID != userID {
			t.Errorf("expected user ID %s, got %s", userID, event.UserID)
		}

		if event.OldRoleID != oldRoleID {
			t.Errorf("expected old role ID %s, got %s", oldRoleID, event.OldRoleID)
		}

		if event.NewRoleID != newRoleID {
			t.Errorf("expected new role ID %s, got %s", newRoleID, event.NewRoleID)
		}

		if event.OldRole != oldRole {
			t.Errorf("expected old role %s, got %s", oldRole, event.OldRole)
		}

		if event.NewRole != newRole {
			t.Errorf("expected new role %s, got %s", newRole, event.NewRole)
		}

		if event.Version() != version {
			t.Errorf("expected version %d, got %d", version, event.Version())
		}
	})
}

func TestRoleCreatedEvent(t *testing.T) {
	t.Run("should create role created event with correct values", func(t *testing.T) {
		roleID := "role-123"
		name := "New Role"

		event := NewRoleCreatedEvent(roleID, name)

		if event.EventType() != "role.created" {
			t.Errorf("expected event type 'role.created', got %s", event.EventType())
		}

		if event.AggregateID() != roleID {
			t.Errorf("expected aggregate ID %s, got %s", roleID, event.AggregateID())
		}

		if event.RoleID != roleID {
			t.Errorf("expected role ID %s, got %s", roleID, event.RoleID)
		}

		if event.Name != name {
			t.Errorf("expected name %s, got %s", name, event.Name)
		}

		if event.Version() != 1 {
			t.Errorf("expected version 1, got %d", event.Version())
		}
	})
}

func TestRoleUpdatedEvent(t *testing.T) {
	t.Run("should create role updated event with correct values", func(t *testing.T) {
		roleID := "role-123"
		name := "Updated Role"
		version := 2

		event := NewRoleUpdatedEvent(roleID, name, version)

		if event.EventType() != "role.updated" {
			t.Errorf("expected event type 'role.updated', got %s", event.EventType())
		}

		if event.AggregateID() != roleID {
			t.Errorf("expected aggregate ID %s, got %s", roleID, event.AggregateID())
		}

		if event.RoleID != roleID {
			t.Errorf("expected role ID %s, got %s", roleID, event.RoleID)
		}

		if event.Name != name {
			t.Errorf("expected name %s, got %s", name, event.Name)
		}

		if event.Version() != version {
			t.Errorf("expected version %d, got %d", version, event.Version())
		}
	})
}

func TestRoleDeletedEvent(t *testing.T) {
	t.Run("should create role deleted event with correct values", func(t *testing.T) {
		roleID := "role-123"
		name := "Deleted Role"
		version := 3

		event := NewRoleDeletedEvent(roleID, name, version)

		if event.EventType() != "role.deleted" {
			t.Errorf("expected event type 'role.deleted', got %s", event.EventType())
		}

		if event.AggregateID() != roleID {
			t.Errorf("expected aggregate ID %s, got %s", roleID, event.AggregateID())
		}

		if event.RoleID != roleID {
			t.Errorf("expected role ID %s, got %s", roleID, event.RoleID)
		}

		if event.Name != name {
			t.Errorf("expected name %s, got %s", name, event.Name)
		}

		if event.Version() != version {
			t.Errorf("expected version %d, got %d", version, event.Version())
		}
	})
}
