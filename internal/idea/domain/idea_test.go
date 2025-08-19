package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewIdea(t *testing.T) {
	t.Run("valid idea creation with organization", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, err := NewIdea("Test Title", "Test Content", creatorID, organizationID)

		assert.NoError(t, err)
		assert.NotNil(t, idea)
		assert.NotEqual(t, uuid.Nil, idea.ID)
		assert.Equal(t, "Test Title", idea.Title)
		assert.Equal(t, "Test Content", idea.Content)
		assert.Equal(t, creatorID, idea.CreatorUserID)
		assert.Equal(t, organizationID, idea.OrganizationID)
		assert.False(t, idea.CreatedAt.IsZero())
		assert.False(t, idea.UpdatedAt.IsZero())
	})

	t.Run("valid idea creation without organization (legacy support)", func(t *testing.T) {
		creatorID := uuid.New()
		idea, err := NewIdeaWithoutOrganization("Test Title", "Test Content", creatorID)

		assert.NoError(t, err)
		assert.NotNil(t, idea)
		assert.NotEqual(t, uuid.Nil, idea.ID)
		assert.Equal(t, "Test Title", idea.Title)
		assert.Equal(t, "Test Content", idea.Content)
		assert.Equal(t, creatorID, idea.CreatorUserID)
		assert.Equal(t, uuid.Nil, idea.OrganizationID)
		assert.False(t, idea.CreatedAt.IsZero())
		assert.False(t, idea.UpdatedAt.IsZero())
	})

	t.Run("empty required fields", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()

		testCases := []struct {
			title          string
			content        string
			creatorUserID  uuid.UUID
			organizationID uuid.UUID
			expectedError  string
		}{
			{"", "Test Content", creatorID, organizationID, "idea title cannot be empty"},
			{"   ", "Test Content", creatorID, organizationID, "idea title cannot be empty"},
			{"Test Title", "", creatorID, organizationID, "idea content cannot be empty"},
			{"Test Title", "   ", creatorID, organizationID, "idea content cannot be empty"},
			{"Test Title", "Test Content", uuid.Nil, organizationID, "creator user ID cannot be empty"},
			{"Test Title", "Test Content", creatorID, uuid.Nil, "organization ID cannot be empty"},
		}

		for _, tc := range testCases {
			idea, err := NewIdea(tc.title, tc.content, tc.creatorUserID, tc.organizationID)
			assert.Error(t, err)
			assert.Nil(t, idea)
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	})
}

func TestNewIdeaWithID(t *testing.T) {
	t.Run("valid idea creation with specific ID and organization", func(t *testing.T) {
		id := uuid.New()
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, err := NewIdeaWithID(id, "Test Title", "Test Content", creatorID, organizationID)

		assert.NoError(t, err)
		assert.NotNil(t, idea)
		assert.Equal(t, id, idea.ID)
		assert.Equal(t, "Test Title", idea.Title)
		assert.Equal(t, "Test Content", idea.Content)
		assert.Equal(t, creatorID, idea.CreatorUserID)
		assert.Equal(t, organizationID, idea.OrganizationID)
		assert.False(t, idea.CreatedAt.IsZero())
		assert.False(t, idea.UpdatedAt.IsZero())
	})

	t.Run("valid idea creation with specific ID without organization (legacy support)", func(t *testing.T) {
		id := uuid.New()
		creatorID := uuid.New()
		idea, err := NewIdeaWithIDWithoutOrganization(id, "Test Title", "Test Content", creatorID)

		assert.NoError(t, err)
		assert.NotNil(t, idea)
		assert.Equal(t, id, idea.ID)
		assert.Equal(t, "Test Title", idea.Title)
		assert.Equal(t, "Test Content", idea.Content)
		assert.Equal(t, creatorID, idea.CreatorUserID)
		assert.Equal(t, uuid.Nil, idea.OrganizationID)
		assert.False(t, idea.CreatedAt.IsZero())
		assert.False(t, idea.UpdatedAt.IsZero())
	})

	t.Run("empty required fields with ID", func(t *testing.T) {
		id := uuid.New()
		creatorID := uuid.New()
		organizationID := uuid.New()

		testCases := []struct {
			id             uuid.UUID
			title          string
			content        string
			creatorUserID  uuid.UUID
			organizationID uuid.UUID
			expectedError  string
		}{
			{uuid.Nil, "Test Title", "Test Content", creatorID, organizationID, "idea ID cannot be empty"},
			{id, "", "Test Content", creatorID, organizationID, "idea title cannot be empty"},
			{id, "   ", "Test Content", creatorID, organizationID, "idea title cannot be empty"},
			{id, "Test Title", "", creatorID, organizationID, "idea content cannot be empty"},
			{id, "Test Title", "   ", creatorID, organizationID, "idea content cannot be empty"},
			{id, "Test Title", "Test Content", uuid.Nil, organizationID, "creator user ID cannot be empty"},
			{id, "Test Title", "Test Content", creatorID, uuid.Nil, "organization ID cannot be empty"},
		}

		for _, tc := range testCases {
			idea, err := NewIdeaWithID(tc.id, tc.title, tc.content, tc.creatorUserID, tc.organizationID)
			assert.Error(t, err)
			assert.Nil(t, idea)
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	})
}

func TestIdea_UpdateTitle(t *testing.T) {
	t.Run("valid title update", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Original Title", "Original Content", creatorID, organizationID)
		originalUpdatedAt := idea.UpdatedAt

		err := idea.UpdateTitle("Updated Title")

		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", idea.Title)
		assert.True(t, idea.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("empty title", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Original Title", "Original Content", creatorID, organizationID)

		err := idea.UpdateTitle("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "idea title cannot be empty")

		err = idea.UpdateTitle("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "idea title cannot be empty")
	})
}

func TestIdea_UpdateContent(t *testing.T) {
	t.Run("valid content update", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Original Title", "Original Content", creatorID, organizationID)
		originalUpdatedAt := idea.UpdatedAt

		err := idea.UpdateContent("Updated Content")

		assert.NoError(t, err)
		assert.Equal(t, "Updated Content", idea.Content)
		assert.True(t, idea.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("empty content", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Original Title", "Original Content", creatorID, organizationID)

		err := idea.UpdateContent("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "idea content cannot be empty")

		err = idea.UpdateContent("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "idea content cannot be empty")
	})
}

func TestIdea_UpdateOrganization(t *testing.T) {
	t.Run("valid organization update", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Test Title", "Test Content", creatorID, organizationID)
		originalUpdatedAt := idea.UpdatedAt

		newOrganizationID := uuid.New()
		err := idea.UpdateOrganization(newOrganizationID)

		assert.NoError(t, err)
		assert.Equal(t, newOrganizationID, idea.OrganizationID)
		assert.True(t, idea.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("empty organization ID", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Test Title", "Test Content", creatorID, organizationID)

		err := idea.UpdateOrganization(uuid.Nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID cannot be empty")
	})
}

func TestIdea_IsInOrganization(t *testing.T) {
	t.Run("idea is in specified organization", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Test Title", "Test Content", creatorID, organizationID)

		assert.True(t, idea.IsInOrganization(organizationID))

		otherOrganizationID := uuid.New()
		assert.False(t, idea.IsInOrganization(otherOrganizationID))
	})

	t.Run("idea without organization", func(t *testing.T) {
		creatorID := uuid.New()
		idea, _ := NewIdeaWithoutOrganization("Test Title", "Test Content", creatorID)

		// Ideas without organization should return false for any organization check
		organizationID := uuid.New()
		assert.False(t, idea.IsInOrganization(organizationID))
		assert.False(t, idea.IsInOrganization(uuid.Nil))
	})
}

func TestIdea_IsOrganizationScoped(t *testing.T) {
	t.Run("organization scoped idea", func(t *testing.T) {
		creatorID := uuid.New()
		organizationID := uuid.New()
		idea, _ := NewIdea("Test Title", "Test Content", creatorID, organizationID)

		assert.True(t, idea.IsOrganizationScoped())
	})

	t.Run("non-organization scoped idea", func(t *testing.T) {
		creatorID := uuid.New()
		idea, _ := NewIdeaWithoutOrganization("Test Title", "Test Content", creatorID)

		assert.False(t, idea.IsOrganizationScoped())
	})
}
