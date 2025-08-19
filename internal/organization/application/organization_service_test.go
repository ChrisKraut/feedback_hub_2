package application

import (
	"context"
	"fmt"
	"testing"

	"feedback_hub_2/internal/organization/domain"
	events "feedback_hub_2/internal/shared/bus"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockOrganizationRepository is a mock implementation for testing the service.
// AI-hint: Mock repository used for testing the application service without
// requiring a real database connection or repository implementation.
type MockOrganizationRepository struct {
	organizations  map[uuid.UUID]*domain.Organization
	slugs          map[string]uuid.UUID
	createError    error
	getByIDError   error
	getBySlugError error
	updateError    error
	deleteError    error
	listError      error
	countError     error
}

// NewMockOrganizationRepository creates a new mock repository for testing.
func NewMockOrganizationRepository() *MockOrganizationRepository {
	return &MockOrganizationRepository{
		organizations: make(map[uuid.UUID]*domain.Organization),
		slugs:         make(map[string]uuid.UUID),
	}
}

// MockEventBus is a mock implementation for testing event publishing.
// AI-hint: Mock event bus used for testing event publishing without
// requiring a real event bus implementation.
type MockEventBus struct {
	publishedEvents []events.DomainEvent
	publishError    error
}

// NewMockEventBus creates a new mock event bus for testing.
func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		publishedEvents: make([]events.DomainEvent, 0),
	}
}

// Publish implements events.EventBus.Publish for the mock.
func (m *MockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	if m.publishError != nil {
		return m.publishError
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

// Subscribe implements events.EventBus.Subscribe for the mock.
func (m *MockEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	return nil
}

// Unsubscribe implements events.EventBus.Unsubscribe for the mock.
func (m *MockEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	return nil
}

// GetPublishedEvents returns all published events for testing.
func (m *MockEventBus) GetPublishedEvents() []events.DomainEvent {
	return m.publishedEvents
}

// ClearPublishedEvents clears the published events list.
func (m *MockEventBus) ClearPublishedEvents() {
	m.publishedEvents = make([]events.DomainEvent, 0)
}

// Create implements domain.OrganizationRepository.Create for the mock.
func (m *MockOrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	if m.createError != nil {
		return m.createError
	}

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

// GetByID implements domain.OrganizationRepository.GetByID for the mock.
func (m *MockOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}

	if id == uuid.Nil {
		return nil, domain.ErrInvalidOrganizationData
	}

	org, exists := m.organizations[id]
	if !exists {
		return nil, domain.ErrOrganizationNotFound
	}

	return org, nil
}

// GetBySlug implements domain.OrganizationRepository.GetBySlug for the mock.
func (m *MockOrganizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	if m.getBySlugError != nil {
		return nil, m.getBySlugError
	}

	if slug == "" {
		return nil, domain.ErrInvalidOrganizationData
	}

	id, exists := m.slugs[slug]
	if !exists {
		return nil, domain.ErrOrganizationNotFound
	}

	return m.GetByID(ctx, id)
}

// Update implements domain.OrganizationRepository.Update for the mock.
func (m *MockOrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	if m.updateError != nil {
		return m.updateError
	}

	if org == nil {
		return domain.ErrInvalidOrganizationData
	}

	if err := org.Validate(); err != nil {
		return domain.ErrInvalidOrganizationData
	}

	existing, exists := m.organizations[org.ID]
	if !exists {
		return domain.ErrOrganizationNotFound
	}

	// Check if slug changed and if new slug already exists
	if existing.Slug != org.Slug {
		if existingID, slugExists := m.slugs[org.Slug]; slugExists && existingID != org.ID {
			return domain.ErrOrganizationSlugAlreadyExists
		}
		// Remove old slug mapping
		delete(m.slugs, existing.Slug)
		// Add new slug mapping
		m.slugs[org.Slug] = org.ID
	}

	m.organizations[org.ID] = org
	return nil
}

// Delete implements domain.OrganizationRepository.Delete for the mock.
func (m *MockOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}

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

// List implements domain.OrganizationRepository.List for the mock.
func (m *MockOrganizationRepository) List(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	if limit < 0 || offset < 0 {
		return nil, domain.ErrInvalidOrganizationData
	}

	orgs := make([]*domain.Organization, 0, len(m.organizations))
	for _, org := range m.organizations {
		orgs = append(orgs, org)
	}

	// Simple pagination (in real implementation, this would be more sophisticated)
	if offset >= len(orgs) {
		return []*domain.Organization{}, nil
	}

	end := offset + limit
	if end > len(orgs) {
		end = len(orgs)
	}

	return orgs[offset:end], nil
}

// Count implements domain.OrganizationRepository.Count for the mock.
func (m *MockOrganizationRepository) Count(ctx context.Context) (int, error) {
	if m.countError != nil {
		return 0, m.countError
	}

	return len(m.organizations), nil
}

// TestOrganizationService tests the organization application service.
// AI-hint: These tests ensure that the service correctly implements business logic
// for organization management and properly delegates to the repository layer.
func TestOrganizationService(t *testing.T) {
	ctx := context.Background()

	t.Run("create organization", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Test creating a valid organization
		org, err := service.CreateOrganization(ctx, "Test Org", "test-org", "Test organization", nil)
		require.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, "Test Org", org.Name)
		assert.Equal(t, "test-org", org.Slug)
		assert.Equal(t, "Test organization", org.Description)
		assert.NotEmpty(t, org.ID)
		assert.False(t, org.CreatedAt.IsZero())
		assert.False(t, org.UpdatedAt.IsZero())

		// Verify event was published
		publishedEvents := eventBus.GetPublishedEvents()
		require.Len(t, publishedEvents, 1)
		assert.Equal(t, "organization.created", publishedEvents[0].EventType())
		assert.Equal(t, org.ID.String(), publishedEvents[0].AggregateID())
	})

	t.Run("create organization with duplicate slug", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create first organization
		org1, err := service.CreateOrganization(ctx, "First Org", "test-org", "First organization", nil)
		require.NoError(t, err)
		assert.NotNil(t, org1)

		// Try to create second organization with same slug
		org2, err := service.CreateOrganization(ctx, "Second Org", "test-org", "Second organization", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationSlugAlreadyExists)
		assert.Nil(t, org2)
	})

	t.Run("create organization with invalid data", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Test with empty name
		org, err := service.CreateOrganization(ctx, "", "test-org", "Test organization", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization name cannot be empty")
		assert.Nil(t, org)

		// Test with empty slug (should auto-generate slug)
		org, err = service.CreateOrganization(ctx, "Test Org", "", "Test organization", nil)
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.NotEmpty(t, org.Slug)
		assert.Equal(t, "test-org", org.Slug)
	})
}

// TestOrganizationServiceGetByID tests organization retrieval by ID.
// AI-hint: These tests ensure that the service correctly retrieves organizations
// by ID and handles various error scenarios.
func TestOrganizationServiceGetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing organization", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create an organization
		created, err := service.CreateOrganization(ctx, "Test Org", "test-org", "Test organization", nil)
		require.NoError(t, err)

		// Retrieve it by ID
		retrieved, err := service.GetOrganizationByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Name, retrieved.Name)
		assert.Equal(t, created.Slug, retrieved.Slug)
	})

	t.Run("get non-existent organization", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Try to get non-existent organization
		org, err := service.GetOrganizationByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationNotFound)
		assert.Nil(t, org)
	})

	t.Run("get organization with zero UUID", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Try to get organization with zero UUID
		org, err := service.GetOrganizationByID(ctx, uuid.Nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidOrganizationData)
		assert.Nil(t, org)
	})
}

// TestOrganizationServiceGetBySlug tests organization retrieval by slug.
// AI-hint: These tests ensure that the service correctly retrieves organizations
// by slug and handles various error scenarios.
func TestOrganizationServiceGetBySlug(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing organization by slug", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create an organization
		created, err := service.CreateOrganization(ctx, "Test Org", "test-org", "Test organization", nil)
		require.NoError(t, err)

		// Retrieve it by slug
		retrieved, err := service.GetOrganizationBySlug(ctx, created.Slug)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Name, retrieved.Name)
		assert.Equal(t, created.Slug, retrieved.Slug)
	})

	t.Run("get non-existent organization by slug", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Try to get non-existent organization by slug
		org, err := service.GetOrganizationBySlug(ctx, "non-existent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationNotFound)
		assert.Nil(t, org)
	})

	t.Run("get organization with empty slug", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Try to get organization with empty slug
		org, err := service.GetOrganizationBySlug(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidOrganizationData)
		assert.Nil(t, org)
	})
}

// TestOrganizationServiceUpdate tests organization updates.
// AI-hint: These tests ensure that the service correctly updates organizations
// while maintaining data integrity and business rules.
func TestOrganizationServiceUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("update existing organization", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create an organization
		org, err := service.CreateOrganization(ctx, "Original Name", "original-slug", "Original description", nil)
		require.NoError(t, err)

		// Update it
		err = service.UpdateOrganization(ctx, org.ID, "Updated Name", "updated-slug", "Updated description", map[string]any{"theme": "dark"})
		assert.NoError(t, err)

		// Retrieve and verify updates
		updated, err := service.GetOrganizationByID(ctx, org.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "updated-slug", updated.Slug)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, map[string]any{"theme": "dark"}, updated.Settings)
		assert.True(t, updated.UpdatedAt.After(org.UpdatedAt) || updated.UpdatedAt.Equal(org.UpdatedAt))
	})

	t.Run("update non-existent organization", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Try to update non-existent organization
		err := service.UpdateOrganization(ctx, uuid.New(), "Updated Name", "updated-slug", "Updated description", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	})

	t.Run("update organization with duplicate slug", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create two organizations
		_, err := service.CreateOrganization(ctx, "First Org", "first-org", "First organization", nil)
		require.NoError(t, err)

		org2, err := service.CreateOrganization(ctx, "Second Org", "second-org", "Second organization", nil)
		require.NoError(t, err)

		// Try to update second organization with first organization's slug
		err = service.UpdateOrganization(ctx, org2.ID, "Second Org", "first-org", "Second organization", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationSlugAlreadyExists)
	})
}

// TestOrganizationServiceDelete tests organization deletion.
// AI-hint: These tests ensure that the service correctly deletes organizations
// and handles various error scenarios.
func TestOrganizationServiceDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("delete existing organization", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create an organization
		org, err := service.CreateOrganization(ctx, "Test Org", "test-org", "Test organization", nil)
		require.NoError(t, err)

		// Delete it
		err = service.DeleteOrganization(ctx, org.ID)
		assert.NoError(t, err)

		// Verify it's deleted
		retrieved, err := service.GetOrganizationByID(ctx, org.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationNotFound)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent organization", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Try to delete non-existent organization
		err := service.DeleteOrganization(ctx, uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	})
}

// TestOrganizationServiceList tests organization listing and pagination.
// AI-hint: These tests ensure that the service correctly lists organizations
// with proper pagination and error handling.
func TestOrganizationServiceList(t *testing.T) {
	ctx := context.Background()

	t.Run("list organizations with pagination", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create multiple organizations
		for i := 0; i < 5; i++ {
			_, err := service.CreateOrganization(ctx,
				fmt.Sprintf("Org %d", i),
				fmt.Sprintf("org-%d", i),
				fmt.Sprintf("Organization %d", i),
				nil)
			require.NoError(t, err)
		}

		// List with pagination
		orgs, err := service.ListOrganizations(ctx, 3, 0)
		assert.NoError(t, err)
		assert.Len(t, orgs, 3)

		// List with offset
		orgs, err = service.ListOrganizations(ctx, 3, 3)
		assert.NoError(t, err)
		assert.Len(t, orgs, 2)
	})

	t.Run("list organizations with invalid pagination", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Test with negative limit (should be handled gracefully)
		orgs, err := service.ListOrganizations(ctx, -1, 0)
		assert.NoError(t, err)
		assert.NotNil(t, orgs)

		// Test with negative offset (should be handled gracefully)
		orgs, err = service.ListOrganizations(ctx, 10, -1)
		assert.NoError(t, err)
		assert.NotNil(t, orgs)
	})
}

// TestOrganizationServiceCount tests organization counting.
// AI-hint: These tests ensure that the service correctly counts organizations
// and handles error scenarios.
func TestOrganizationServiceCount(t *testing.T) {
	ctx := context.Background()

	t.Run("count organizations", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Initially should be 0
		count, err := service.CountOrganizations(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)

		// Create an organization
		_, err = service.CreateOrganization(ctx, "Test Org", "test-org", "Test organization", nil)
		require.NoError(t, err)

		// Should now be 1
		count, err = service.CountOrganizations(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

// TestOrganizationServiceBusinessRules tests business rule enforcement.
// AI-hint: These tests ensure that the service enforces business rules
// beyond simple validation, such as slug uniqueness and organization state.
func TestOrganizationServiceBusinessRules(t *testing.T) {
	ctx := context.Background()

	t.Run("slug uniqueness across organizations", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create organization with auto-generated slug
		org1, err := service.CreateOrganizationWithSlug(ctx, "Acme Corporation", "Acme Corporation description", nil)
		require.NoError(t, err)
		assert.Equal(t, "acme-corporation", org1.Slug)

		// Try to create another with same name (should generate unique slug)
		org2, err := service.CreateOrganizationWithSlug(ctx, "Acme Corporation", "Another Acme description", nil)
		assert.NoError(t, err)
		assert.NotNil(t, org2)
		assert.NotEqual(t, org1.Slug, org2.Slug)
		assert.Contains(t, org2.Slug, "acme-corporation")
	})

	t.Run("organization state management", func(t *testing.T) {
		repo := NewMockOrganizationRepository()
		eventBus := NewMockEventBus()
		service := NewOrganizationService(repo, eventBus)

		// Create organization
		org, err := service.CreateOrganization(ctx, "Test Org", "test-org", "Test organization", nil)
		require.NoError(t, err)

		// Initially should be active
		assert.True(t, org.IsActive())

		// Note: Organization state management would be implemented here
		// For now, we just verify the organization was created correctly
		assert.NotNil(t, org)
		assert.Equal(t, "Test Org", org.Name)
	})
}
