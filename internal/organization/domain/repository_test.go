package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrganizationRepositoryInterface tests the organization repository interface contracts.
// AI-hint: These tests ensure that any implementation of OrganizationRepository
// follows the expected contract and handles errors correctly.
func TestOrganizationRepositoryInterface(t *testing.T) {
	// This is a compile-time test to ensure the interface is properly defined
	var _ OrganizationRepository = (*MockOrganizationRepository)(nil)
}

// TestOrganizationRepositoryErrors tests the organization repository error types.
// AI-hint: These tests validate that repository errors are properly typed and
// provide meaningful error messages for different failure scenarios.
func TestOrganizationRepositoryErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedErr error
		checkType   func(error) bool
	}{
		{
			name:        "ErrOrganizationNotFound should be properly typed",
			err:         ErrOrganizationNotFound,
			expectedErr: ErrOrganizationNotFound,
			checkType:   func(err error) bool { return errors.Is(err, ErrOrganizationNotFound) },
		},
		{
			name:        "ErrOrganizationAlreadyExists should be properly typed",
			err:         ErrOrganizationAlreadyExists,
			expectedErr: ErrOrganizationAlreadyExists,
			checkType:   func(err error) bool { return errors.Is(err, ErrOrganizationAlreadyExists) },
		},
		{
			name:        "ErrOrganizationSlugAlreadyExists should be properly typed",
			err:         ErrOrganizationSlugAlreadyExists,
			expectedErr: ErrOrganizationSlugAlreadyExists,
			checkType:   func(err error) bool { return errors.Is(err, ErrOrganizationSlugAlreadyExists) },
		},
		{
			name:        "ErrInvalidOrganizationData should be properly typed",
			err:         ErrInvalidOrganizationData,
			expectedErr: ErrInvalidOrganizationData,
			checkType:   func(err error) bool { return errors.Is(err, ErrInvalidOrganizationData) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.True(t, tt.checkType(tt.err))
			assert.Equal(t, tt.expectedErr, tt.err)
		})
	}
}

// TestOrganizationRepositoryErrorMessages tests that repository errors have meaningful messages.
// AI-hint: These tests ensure that error messages provide clear information about
// what went wrong, helping with debugging and user experience.
func TestOrganizationRepositoryErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "ErrOrganizationNotFound should have meaningful message",
			err:         ErrOrganizationNotFound,
			expectedMsg: "organization not found",
		},
		{
			name:        "ErrOrganizationAlreadyExists should have meaningful message",
			err:         ErrOrganizationAlreadyExists,
			expectedMsg: "organization already exists",
		},
		{
			name:        "ErrOrganizationSlugAlreadyExists should have meaningful message",
			err:         ErrOrganizationSlugAlreadyExists,
			expectedMsg: "organization slug already exists",
		},
		{
			name:        "ErrInvalidOrganizationData should have meaningful message",
			err:         ErrInvalidOrganizationData,
			expectedMsg: "invalid organization data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.err.Error(), tt.expectedMsg)
		})
	}
}

// MockOrganizationRepository is a mock implementation for testing the interface.
// AI-hint: Mock implementation used for testing repository interface contracts
// and ensuring proper error handling in tests.
type MockOrganizationRepository struct {
	organizations map[uuid.UUID]*Organization
	slugs         map[string]uuid.UUID
}

// NewMockOrganizationRepository creates a new mock repository for testing.
func NewMockOrganizationRepository() *MockOrganizationRepository {
	return &MockOrganizationRepository{
		organizations: make(map[uuid.UUID]*Organization),
		slugs:         make(map[string]uuid.UUID),
	}
}

// Create implements OrganizationRepository.Create for the mock.
func (m *MockOrganizationRepository) Create(ctx context.Context, org *Organization) error {
	if org == nil {
		return ErrInvalidOrganizationData
	}

	if err := org.Validate(); err != nil {
		return ErrInvalidOrganizationData
	}

	// Check if slug already exists
	if existingID, exists := m.slugs[org.Slug]; exists {
		if existingID != org.ID {
			return ErrOrganizationSlugAlreadyExists
		}
	}

	m.organizations[org.ID] = org
	m.slugs[org.Slug] = org.ID
	return nil
}

// GetByID implements OrganizationRepository.GetByID for the mock.
func (m *MockOrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidOrganizationData
	}

	org, exists := m.organizations[id]
	if !exists {
		return nil, ErrOrganizationNotFound
	}

	return org, nil
}

// GetBySlug implements OrganizationRepository.GetBySlug for the mock.
func (m *MockOrganizationRepository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	if slug == "" {
		return nil, ErrInvalidOrganizationData
	}

	id, exists := m.slugs[slug]
	if !exists {
		return nil, ErrOrganizationNotFound
	}

	return m.GetByID(ctx, id)
}

// Update implements OrganizationRepository.Update for the mock.
func (m *MockOrganizationRepository) Update(ctx context.Context, org *Organization) error {
	if org == nil {
		return ErrInvalidOrganizationData
	}

	if err := org.Validate(); err != nil {
		return ErrInvalidOrganizationData
	}

	existing, exists := m.organizations[org.ID]
	if !exists {
		return ErrOrganizationNotFound
	}

	// Check if slug changed and if new slug already exists
	if existing.Slug != org.Slug {
		if existingID, slugExists := m.slugs[org.Slug]; slugExists && existingID != org.ID {
			return ErrOrganizationSlugAlreadyExists
		}
		// Remove old slug mapping
		delete(m.slugs, existing.Slug)
		// Add new slug mapping
		m.slugs[org.Slug] = org.ID
	}

	m.organizations[org.ID] = org
	return nil
}

// Delete implements OrganizationRepository.Delete for the mock.
func (m *MockOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidOrganizationData
	}

	org, exists := m.organizations[id]
	if !exists {
		return ErrOrganizationNotFound
	}

	// Remove from both maps
	delete(m.organizations, id)
	delete(m.slugs, org.Slug)
	return nil
}

// List implements OrganizationRepository.List for the mock.
func (m *MockOrganizationRepository) List(ctx context.Context, limit, offset int) ([]*Organization, error) {
	if limit < 0 || offset < 0 {
		return nil, ErrInvalidOrganizationData
	}

	orgs := make([]*Organization, 0, len(m.organizations))
	for _, org := range m.organizations {
		orgs = append(orgs, org)
	}

	// Simple pagination (in real implementation, this would be more sophisticated)
	if offset >= len(orgs) {
		return []*Organization{}, nil
	}

	end := offset + limit
	if end > len(orgs) {
		end = len(orgs)
	}

	return orgs[offset:end], nil
}

// Count implements OrganizationRepository.Count for the mock.
func (m *MockOrganizationRepository) Count(ctx context.Context) (int, error) {
	return len(m.organizations), nil
}

// TestMockOrganizationRepository tests the mock repository implementation.
// AI-hint: These tests ensure the mock repository behaves correctly and can be
// used reliably in other tests that depend on the repository interface.
func TestMockOrganizationRepository(t *testing.T) {
	repo := NewMockOrganizationRepository()
	ctx := context.Background()

	// Test creating an organization
	org, err := NewOrganization("Test Org", "test-org", "Test organization", nil)
	require.NoError(t, err)

	err = repo.Create(ctx, org)
	assert.NoError(t, err)

	// Test getting by ID
	retrieved, err := repo.GetByID(ctx, org.ID)
	assert.NoError(t, err)
	assert.Equal(t, org.ID, retrieved.ID)

	// Test getting by slug
	retrieved, err = repo.GetBySlug(ctx, org.Slug)
	assert.NoError(t, err)
	assert.Equal(t, org.Slug, retrieved.Slug)

	// Test updating
	err = org.Update("Updated Org", "updated-org", "Updated organization", nil)
	require.NoError(t, err)

	err = repo.Update(ctx, org)
	assert.NoError(t, err)

	// Test counting
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test listing
	orgs, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, orgs, 1)

	// Test deleting
	err = repo.Delete(ctx, org.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, org.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOrganizationNotFound)
}
