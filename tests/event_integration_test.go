package tests

import (
	"context"
	"testing"

	"feedback_hub_2/internal/organization/application"
	"feedback_hub_2/internal/organization/domain"
	events "feedback_hub_2/internal/shared/bus"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventIntegration tests the complete event flow from organization service to downstream consumers.
// AI-hint: Integration tests that verify the complete event-driven architecture works correctly,
// ensuring that organization lifecycle events properly trigger responses in other domains.
func TestEventIntegration(t *testing.T) {
	t.Run("should publish events when organization is created", func(t *testing.T) {
		// Arrange
		eventBus := events.NewInMemoryEventBus()
		repo := NewMockOrganizationRepository()
		orgService := application.NewOrganizationService(repo, eventBus)

		// Create event handlers for different domains
		userServiceHandler := &UserServiceEventHandler{}
		roleServiceHandler := &RoleServiceEventHandler{}
		ideaServiceHandler := &IdeaServiceEventHandler{}

		// Subscribe handlers to organization events
		err := eventBus.Subscribe("organization.created", userServiceHandler.HandleOrganizationCreated)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.created", roleServiceHandler.HandleOrganizationCreated)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.created", ideaServiceHandler.HandleOrganizationCreated)
		require.NoError(t, err)

		// Act
		ctx := context.Background()
		org, err := orgService.CreateOrganization(ctx, "Test Event Org", "test-event-org", "Organization for event testing", nil)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, org)

		// Verify all handlers received the event
		assert.True(t, userServiceHandler.OrganizationCreatedHandled)
		assert.True(t, roleServiceHandler.OrganizationCreatedHandled)
		assert.True(t, ideaServiceHandler.OrganizationCreatedHandled)
	})

	t.Run("should publish events when organization is updated", func(t *testing.T) {
		// Arrange
		eventBus := events.NewInMemoryEventBus()
		repo := NewMockOrganizationRepository()
		orgService := application.NewOrganizationService(repo, eventBus)

		// Create event handlers for different domains
		userServiceHandler := &UserServiceEventHandler{}
		roleServiceHandler := &RoleServiceEventHandler{}
		ideaServiceHandler := &IdeaServiceEventHandler{}

		// Subscribe handlers to organization events
		err := eventBus.Subscribe("organization.updated", userServiceHandler.HandleOrganizationUpdated)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.updated", roleServiceHandler.HandleOrganizationUpdated)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.updated", ideaServiceHandler.HandleOrganizationUpdated)
		require.NoError(t, err)

		// Create an organization first
		ctx := context.Background()
		org, err := orgService.CreateOrganization(ctx, "Update Event Org", "update-event-org", "Organization for update event testing", nil)
		require.NoError(t, err)

		// Reset handler state for clean testing
		userServiceHandler.OrganizationUpdatedHandled = false
		roleServiceHandler.OrganizationUpdatedHandled = false
		ideaServiceHandler.OrganizationUpdatedHandled = false

		// Act - Update the organization
		err = orgService.UpdateOrganization(ctx, org.ID, "Updated Event Org", "", "Updated description", nil)

		// Assert
		require.NoError(t, err)

		// Verify all handlers received the event
		assert.True(t, userServiceHandler.OrganizationUpdatedHandled)
		assert.True(t, roleServiceHandler.OrganizationUpdatedHandled)
		assert.True(t, ideaServiceHandler.OrganizationUpdatedHandled)
	})

	t.Run("should publish events when organization is deleted", func(t *testing.T) {
		// Arrange
		eventBus := events.NewInMemoryEventBus()
		repo := NewMockOrganizationRepository()
		orgService := application.NewOrganizationService(repo, eventBus)

		// Create event handlers for different domains
		userServiceHandler := &UserServiceEventHandler{}
		roleServiceHandler := &RoleServiceEventHandler{}
		ideaServiceHandler := &IdeaServiceEventHandler{}

		// Subscribe handlers to organization events
		err := eventBus.Subscribe("organization.deleted", userServiceHandler.HandleOrganizationDeleted)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.deleted", roleServiceHandler.HandleOrganizationDeleted)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.deleted", ideaServiceHandler.HandleOrganizationDeleted)
		require.NoError(t, err)

		// Create an organization first
		ctx := context.Background()
		org, err := orgService.CreateOrganization(ctx, "Delete Event Org", "delete-event-org", "Organization for delete event testing", nil)
		require.NoError(t, err)

		// Reset handler state for clean testing
		userServiceHandler.OrganizationDeletedHandled = false
		roleServiceHandler.OrganizationDeletedHandled = false
		ideaServiceHandler.OrganizationDeletedHandled = false

		// Act - Delete the organization
		err = orgService.DeleteOrganization(ctx, org.ID)

		// Assert
		require.NoError(t, err)

		// Verify all handlers received the event
		assert.True(t, userServiceHandler.OrganizationDeletedHandled)
		assert.True(t, roleServiceHandler.OrganizationDeletedHandled)
		assert.True(t, ideaServiceHandler.OrganizationDeletedHandled)
	})
}

// TestEventOrderingAndConsistency tests that events are published in the correct order and maintain consistency.
// AI-hint: Tests that verify event ordering and consistency, ensuring that the event-driven
// system maintains proper state across domain boundaries.
func TestEventOrderingAndConsistency(t *testing.T) {
	t.Run("should maintain event ordering for organization lifecycle", func(t *testing.T) {
		// Arrange
		eventBus := events.NewInMemoryEventBus()
		repo := NewMockOrganizationRepository()
		orgService := application.NewOrganizationService(repo, eventBus)

		// Create an event recorder to track event order
		eventRecorder := &EventOrderRecorder{}

		// Subscribe to all organization events
		err := eventBus.Subscribe("organization.created", eventRecorder.RecordEvent)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.updated", eventRecorder.RecordEvent)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.deleted", eventRecorder.RecordEvent)
		require.NoError(t, err)

		// Act - Perform organization lifecycle operations
		ctx := context.Background()

		// 1. Create organization
		org, err := orgService.CreateOrganization(ctx, "Order Test Org", "order-test-org", "Organization for order testing", nil)
		require.NoError(t, err)

		// 2. Update organization
		err = orgService.UpdateOrganization(ctx, org.ID, "Updated Order Test Org", "", "Updated description", nil)
		require.NoError(t, err)

		// 3. Delete organization
		err = orgService.DeleteOrganization(ctx, org.ID)
		require.NoError(t, err)

		// Assert - Events should be in correct order
		require.Len(t, eventRecorder.Events, 3)
		assert.Equal(t, "organization.created", eventRecorder.Events[0].EventType())
		assert.Equal(t, "organization.updated", eventRecorder.Events[1].EventType())
		assert.Equal(t, "organization.deleted", eventRecorder.Events[2].EventType())

		// All events should have the same aggregate ID
		for _, event := range eventRecorder.Events {
			assert.Equal(t, org.ID.String(), event.AggregateID())
		}
	})
}

// TestEventErrorHandlingAndRecovery tests how the event system handles errors and recovers.
// AI-hint: Tests that verify the event system's resilience to failures and its ability
// to recover gracefully, ensuring system stability under error conditions.
func TestEventErrorHandlingAndRecovery(t *testing.T) {
	t.Run("should continue processing when some handlers fail", func(t *testing.T) {
		// Arrange
		eventBus := events.NewInMemoryEventBus()
		repo := NewMockOrganizationRepository()
		orgService := application.NewOrganizationService(repo, eventBus)

		// Create handlers - one that fails, one that succeeds
		failingHandler := &ErrorProneEventHandler{}
		successfulHandler := &UserServiceEventHandler{}

		// Subscribe handlers to organization events
		err := eventBus.Subscribe("organization.created", failingHandler.HandleOrganizationCreated)
		require.NoError(t, err)
		err = eventBus.Subscribe("organization.created", successfulHandler.HandleOrganizationCreated)
		require.NoError(t, err)

		// Act
		ctx := context.Background()
		org, err := orgService.CreateOrganization(ctx, "Error Test Org", "error-test-org", "Organization for error testing", nil)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, org)

		// Both handlers should have been called
		assert.True(t, failingHandler.OrganizationCreatedHandled)
		assert.True(t, successfulHandler.OrganizationCreatedHandled)
	})
}

// Mock implementations for testing

// EventOrderRecorder records events in the order they are received.
// AI-hint: Mock implementation used to test event ordering and consistency
// in the event-driven system.
type EventOrderRecorder struct {
	Events []events.DomainEvent
}

func (r *EventOrderRecorder) RecordEvent(ctx context.Context, event events.DomainEvent) error {
	r.Events = append(r.Events, event)
	return nil
}

// MockOrganizationRepository is a mock implementation for testing.
// AI-hint: Mock repository that simulates organization persistence without
// requiring a real database connection.
type MockOrganizationRepository struct {
	organizations map[uuid.UUID]*domain.Organization
	slugs         map[string]uuid.UUID
}

func NewMockOrganizationRepository() *MockOrganizationRepository {
	return &MockOrganizationRepository{
		organizations: make(map[uuid.UUID]*domain.Organization),
		slugs:         make(map[string]uuid.UUID),
	}
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	if org == nil {
		return domain.ErrInvalidOrganizationData
	}

	if err := org.Validate(); err != nil {
		return domain.ErrInvalidOrganizationData
	}

	// Check if slug already exists
	if existingID, exists := m.slugs[org.Slug]; exists {
		if existingID != org.ID {
			return domain.ErrOrganizationSlugAlreadyExists
		}
	}

	m.organizations[org.ID] = org
	m.slugs[org.Slug] = org.ID
	return nil
}

func (m *MockOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidOrganizationData
	}

	org, exists := m.organizations[id]
	if !exists {
		return nil, domain.ErrOrganizationNotFound
	}

	return org, nil
}

func (m *MockOrganizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	if slug == "" {
		return nil, domain.ErrInvalidOrganizationData
	}

	id, exists := m.slugs[slug]
	if !exists {
		return nil, domain.ErrOrganizationNotFound
	}

	return m.GetByID(ctx, id)
}

func (m *MockOrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	if org == nil {
		return domain.ErrInvalidOrganizationData
	}

	if err := org.Validate(); err != nil {
		return domain.ErrInvalidOrganizationData
	}

	// Check if organization exists
	if _, exists := m.organizations[org.ID]; !exists {
		return domain.ErrOrganizationNotFound
	}

	// Check if new slug conflicts with existing organization
	if existingID, exists := m.slugs[org.Slug]; exists {
		if existingID != org.ID {
			return domain.ErrOrganizationSlugAlreadyExists
		}
	}

	m.organizations[org.ID] = org
	m.slugs[org.Slug] = org.ID
	return nil
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidOrganizationData
	}

	org, exists := m.organizations[id]
	if !exists {
		return domain.ErrOrganizationNotFound
	}

	// Remove from both maps
	delete(m.organizations, id)
	delete(m.slugs, org.Slug)
	return nil
}

func (m *MockOrganizationRepository) List(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	if limit < 0 || offset < 0 {
		return nil, domain.ErrInvalidOrganizationData
	}

	orgs := make([]*domain.Organization, 0, len(m.organizations))
	for _, org := range m.organizations {
		orgs = append(orgs, org)
	}

	// Simple pagination
	if offset >= len(orgs) {
		return []*domain.Organization{}, nil
	}

	end := offset + limit
	if end > len(orgs) {
		end = len(orgs)
	}

	return orgs[offset:end], nil
}

func (m *MockOrganizationRepository) Count(ctx context.Context) (int, error) {
	return len(m.organizations), nil
}
