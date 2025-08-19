package events

import (
	"context"
	"log"
)

// TestEventHandlers provides test handlers for domain events to demonstrate the system working.
// AI-hint: These handlers demonstrate how different domains can react to events
// without direct coupling to the source domain.
type TestEventHandlers struct {
	userCreatedCount    int
	userUpdatedCount    int
	userRoleUpdatedCount int
	roleCreatedCount    int
	roleUpdatedCount    int
	roleDeletedCount    int
}

// NewTestEventHandlers creates new test event handlers.
// AI-hint: Factory method for test event handlers.
func NewTestEventHandlers() *TestEventHandlers {
	return &TestEventHandlers{}
}

// HandleUserCreated handles user creation events.
// AI-hint: Example of how a different domain could react to user creation.
func (h *TestEventHandlers) HandleUserCreated(ctx context.Context, event DomainEvent) error {
	h.userCreatedCount++
	log.Printf("Test handler: User created event received - UserID: %s, EventID: %s", event.AggregateID(), event.EventID())
	return nil
}

// HandleUserUpdated handles user update events.
// AI-hint: Example of how a different domain could react to user updates.
func (h *TestEventHandlers) HandleUserUpdated(ctx context.Context, event DomainEvent) error {
	h.userUpdatedCount++
	log.Printf("Test handler: User updated event received - UserID: %s, EventID: %s", event.AggregateID(), event.EventID())
	return nil
}

// HandleUserRoleUpdated handles user role update events.
// AI-hint: Example of how a different domain could react to role changes.
func (h *TestEventHandlers) HandleUserRoleUpdated(ctx context.Context, event DomainEvent) error {
	h.userRoleUpdatedCount++
	log.Printf("Test handler: User role updated event received - UserID: %s, EventID: %s", event.AggregateID(), event.EventID())
	return nil
}

// HandleRoleCreated handles role creation events.
// AI-hint: Example of how a different domain could react to role creation.
func (h *TestEventHandlers) HandleRoleCreated(ctx context.Context, event DomainEvent) error {
	h.roleCreatedCount++
	log.Printf("Test handler: Role created event received - RoleID: %s, EventID: %s", event.AggregateID(), event.EventID())
	return nil
}

// HandleRoleUpdated handles role update events.
// AI-hint: Example of how a different domain could react to role updates.
func (h *TestEventHandlers) HandleRoleUpdated(ctx context.Context, event DomainEvent) error {
	h.roleUpdatedCount++
	log.Printf("Test handler: Role updated event received - RoleID: %s, EventID: %s", event.AggregateID(), event.EventID())
	return nil
}

// HandleRoleDeleted handles role deletion events.
// AI-hint: Example of how a different domain could react to role deletion.
func (h *TestEventHandlers) HandleRoleDeleted(ctx context.Context, event DomainEvent) error {
	h.roleDeletedCount++
	log.Printf("Test handler: Role deleted event received - RoleID: %s, EventID: %s", event.AggregateID(), event.EventID())
	return nil
}

// GetCounts returns the current event counts for testing.
// AI-hint: Helper method for testing event handling.
func (h *TestEventHandlers) GetCounts() map[string]int {
	return map[string]int{
		"user_created":     h.userCreatedCount,
		"user_updated":     h.userUpdatedCount,
		"user_role_updated": h.userRoleUpdatedCount,
		"role_created":     h.roleCreatedCount,
		"role_updated":     h.roleUpdatedCount,
		"role_deleted":     h.roleDeletedCount,
	}
}

// ResetCounts resets all event counts to zero.
// AI-hint: Helper method for testing event handling.
func (h *TestEventHandlers) ResetCounts() {
	h.userCreatedCount = 0
	h.userUpdatedCount = 0
	h.userRoleUpdatedCount = 0
	h.roleCreatedCount = 0
	h.roleUpdatedCount = 0
	h.roleDeletedCount = 0
}
