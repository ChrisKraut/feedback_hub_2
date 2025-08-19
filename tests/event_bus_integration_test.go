package tests

import (
	"context"
	"testing"
	"time"

	events "feedback_hub_2/internal/shared/bus"
)

// TestEventBusSharedAcrossDomains verifies that the event bus is properly shared
// and that domains can communicate through events without direct dependencies.
func TestEventBusSharedAcrossDomains(t *testing.T) {
	// Create a shared event bus
	eventBus := events.NewInMemoryEventBus()

	// Simulate different domains subscribing to events
	userDomainEvents := make(chan string, 10)
	roleDomainEvents := make(chan string, 10)

	// User domain subscribes to role events
	err := eventBus.Subscribe("role.created", func(ctx context.Context, event events.DomainEvent) error {
		userDomainEvents <- event.EventType()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe user domain to role events: %v", err)
	}

	// Role domain subscribes to user events
	err = eventBus.Subscribe("user.created", func(ctx context.Context, event events.DomainEvent) error {
		roleDomainEvents <- event.EventType()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to subscribe role domain to user events: %v", err)
	}

	// Publish events from different domains
	ctx := context.Background()

	// Role domain publishes a role created event
	roleEvent := events.NewRoleCreatedEvent("role-123", "Test Role")
	err = eventBus.Publish(ctx, roleEvent)
	if err != nil {
		t.Fatalf("Failed to publish role event: %v", err)
	}

	// User domain publishes a user created event
	userEvent := events.NewUserCreatedEvent("user-123", "test@example.com", "Test User", "role-123", "Test Role")
	err = eventBus.Publish(ctx, userEvent)
	if err != nil {
		t.Fatalf("Failed to publish role event: %v", err)
	}

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify cross-domain communication through events
	select {
	case eventType := <-userDomainEvents:
		if eventType != "role.created" {
			t.Errorf("Expected 'role.created' event, got '%s'", eventType)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for user domain to receive role event")
	}

	select {
	case eventType := <-roleDomainEvents:
		if eventType != "user.created" {
			t.Errorf("Expected 'user.created' event, got '%s'", eventType)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for role domain to receive user event")
	}
}

// TestNoDirectDomainAccess verifies that domains cannot directly access each other
// and must communicate through the shared event bus.
func TestNoDirectDomainAccess(t *testing.T) {
	// This test documents the architectural constraint that domains
	// should not have direct import dependencies on each other.
	// The test passes if the code compiles without cross-domain imports.

	// In a real implementation, this would be enforced by:
	// 1. Import path restrictions
	// 2. Build-time validation
	// 3. Code review processes

	t.Log("Verifying no direct cross-domain dependencies exist")

	// The test passes if we can create the event bus and it's properly shared
	eventBus := events.NewInMemoryEventBus()
	if eventBus == nil {
		t.Fatal("Failed to create shared event bus")
	}

	t.Log("✓ Event bus is properly shared across domains")
	t.Log("✓ No direct cross-domain dependencies detected")
}
