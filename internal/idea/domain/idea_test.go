package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdea_UpdateTitle(t *testing.T) {
	// Create a test idea
	creatorID := uuid.New()
	idea, err := NewIdea("Original Title", "Original Content", creatorID)
	if err != nil {
		t.Fatalf("Failed to create idea: %v", err)
	}

	// Test successful title update
	newTitle := "Updated Title"
	err = idea.UpdateTitle(newTitle)
	if err != nil {
		t.Errorf("UpdateTitle failed: %v", err)
	}

	if idea.Title != newTitle {
		t.Errorf("Expected title to be %s, got %s", newTitle, idea.Title)
	}

	if idea.UpdatedAt.Equal(idea.CreatedAt) {
		t.Error("Expected UpdatedAt to be different from CreatedAt after update")
	}

	// Test empty title validation
	err = idea.UpdateTitle("")
	if err == nil {
		t.Error("Expected error when updating with empty title")
	}

	err = idea.UpdateTitle("   ")
	if err == nil {
		t.Error("Expected error when updating with whitespace-only title")
	}
}

func TestIdea_UpdateContent(t *testing.T) {
	// Create a test idea
	creatorID := uuid.New()
	idea, err := NewIdea("Original Title", "Original Content", creatorID)
	if err != nil {
		t.Fatalf("Failed to create idea: %v", err)
	}

	// Test successful content update
	newContent := "Updated Content"
	err = idea.UpdateContent(newContent)
	if err != nil {
		t.Errorf("UpdateContent failed: %v", err)
	}

	if idea.Content != newContent {
		t.Errorf("Expected content to be %s, got %s", newContent, idea.Content)
	}

	if idea.UpdatedAt.Equal(idea.CreatedAt) {
		t.Error("Expected UpdatedAt to be different from CreatedAt after update")
	}

	// Test empty content validation
	err = idea.UpdateContent("")
	if err == nil {
		t.Error("Expected error when updating with empty content")
	}

	err = idea.UpdateContent("   ")
	if err == nil {
		t.Error("Expected error when updating with whitespace-only content")
	}
}

func TestIdea_UpdateTitleAndContent(t *testing.T) {
	// Create a test idea
	creatorID := uuid.New()
	idea, err := NewIdea("Original Title", "Original Content", creatorID)
	if err != nil {
		t.Fatalf("Failed to create idea: %v", err)
	}

	originalUpdatedAt := idea.UpdatedAt

	// Update both title and content
	newTitle := "Updated Title"
	newContent := "Updated Content"

	err = idea.UpdateTitle(newTitle)
	if err != nil {
		t.Errorf("UpdateTitle failed: %v", err)
	}

	err = idea.UpdateContent(newContent)
	if err != nil {
		t.Errorf("UpdateContent failed: %v", err)
	}

	// Verify both fields were updated
	if idea.Title != newTitle {
		t.Errorf("Expected title to be %s, got %s", newTitle, idea.Title)
	}

	if idea.Content != newContent {
		t.Errorf("Expected content to be %s, got %s", newContent, idea.Content)
	}

	// Verify UpdatedAt was updated
	if idea.UpdatedAt.Equal(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after both changes")
	}
}
