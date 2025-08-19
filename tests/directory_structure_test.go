package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDirectoryStructure verifies that the new DDD directory structure has been created correctly.
// AI-hint: This test ensures the foundation for DDD refactoring is in place.
func TestDirectoryStructure(t *testing.T) {
	// Define the expected directory structure
	expectedDirs := []string{
		"internal/shared/bus",
		"internal/shared/persistence",
		"internal/shared/web",
		"internal/shared/auth",
		"internal/shared/bootstrap",
		"internal/user/domain",
		"internal/user/application",
		"internal/user/infrastructure",
		"internal/user/interfaces",
		"internal/idea/domain",
		"internal/idea/application",
		"internal/idea/infrastructure",
		"internal/idea/interfaces",
		"internal/role/domain",
		"internal/role/application",
		"internal/role/infrastructure",
		"internal/role/interfaces",
	}

	// Check each expected directory
	for _, dir := range expectedDirs {
		// Use relative path from tests directory
		relativePath := "../" + dir
		info, err := os.Stat(relativePath)
		if err != nil {
			t.Errorf("Directory %s does not exist: %v", relativePath, err)
			continue
		}

		if !info.IsDir() {
			t.Errorf("Path %s exists but is not a directory", relativePath)
		}
	}

	// Verify the structure follows DDD principles
	t.Run("DDD Structure Validation", func(t *testing.T) {
		// Check that shared directories exist
		sharedDirs := []string{"bus", "persistence", "web", "auth", "bootstrap"}
		for _, subdir := range sharedDirs {
			path := filepath.Join("..", "internal", "shared", subdir)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("Shared directory %s missing: %v", path, err)
			}
		}

		// Check that each domain has the four required subdirectories
		domains := []string{"user", "idea", "role"}
		requiredSubdirs := []string{"domain", "application", "infrastructure", "interfaces"}

		for _, domain := range domains {
			for _, subdir := range requiredSubdirs {
				path := filepath.Join("..", "internal", domain, subdir)
				if _, err := os.Stat(path); err != nil {
					t.Errorf("Domain directory %s missing: %v", path, err)
				}
			}
		}
	})
}
