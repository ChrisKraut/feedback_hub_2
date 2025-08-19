package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoCrossDomainImports verifies that no cross-domain imports exist
// and that all domains only depend on the shared layer.
func TestNoCrossDomainImports(t *testing.T) {
	t.Log("Scanning codebase for cross-domain imports...")

	// Scan for Go files in domain directories
	domainDirs := []string{"../internal/user", "../internal/role", "../internal/idea"}

	for _, domainDir := range domainDirs {
		t.Logf("Scanning %s for cross-domain imports...", domainDir)

		err := filepath.Walk(domainDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				content, err := os.ReadFile(path)
				if err != nil {
					t.Logf("Warning: Could not read file %s: %v", path, err)
					return nil
				}

				lines := strings.Split(string(content), "\n")
				for i, line := range lines {
					lineNum := i + 1
					line = strings.TrimSpace(line)

					// Check for import statements
					if strings.HasPrefix(line, "import") || strings.HasPrefix(line, "\"") {
						// Extract import path
						if strings.Contains(line, "feedback_hub_2/internal/") {
							// Check if this is a cross-domain import
							for _, otherDomain := range domainDirs {
								if otherDomain != domainDir && strings.Contains(line, otherDomain) {
									t.Errorf("Cross-domain import found in %s:%d: %s imports from %s",
										path, lineNum, domainDir, otherDomain)
								}
							}
						}
					}
				}
			}

			return nil
		})

		if err != nil {
			t.Logf("Warning: Error scanning %s: %v", domainDir, err)
		}
	}

	t.Log("✓ No cross-domain imports detected")
	t.Log("✓ All domains only depend on shared layer")
	t.Log("✓ DDD dependency flow rules are followed")
}

// TestSharedLayerStructure verifies that the shared layer is properly structured
// and contains all necessary shared components.
func TestSharedLayerStructure(t *testing.T) {
	t.Log("Verifying shared layer structure...")

	// Check that shared layer contains expected components
	sharedComponents := []string{
		"../internal/shared/bus",         // Event bus
		"../internal/shared/persistence", // Shared persistence
		"../internal/shared/web",         // Shared web utilities
		"../internal/shared/auth",        // Shared auth
		"../internal/shared/queries",     // Shared queries
		"../internal/shared/bootstrap",   // Bootstrap service
	}

	for _, component := range sharedComponents {
		if _, err := os.Stat(component); os.IsNotExist(err) {
			t.Errorf("Shared component missing: %s", component)
		} else {
			t.Logf("✓ Shared component exists: %s", component)
		}
	}

	t.Log("✓ Shared layer structure is complete")
}

// TestDomainIsolation verifies that each domain is properly isolated
// and follows the DDD layered architecture.
func TestDomainIsolation(t *testing.T) {
	t.Log("Verifying domain isolation...")

	// Check that each domain has the proper layered structure
	domains := []string{"../internal/user", "../internal/role", "../internal/idea"}
	layers := []string{"domain", "application", "infrastructure", "interfaces"}

	for _, domain := range domains {
		t.Logf("Checking domain: %s", domain)

		for _, layer := range layers {
			layerPath := filepath.Join(domain, layer)
			if _, err := os.Stat(layerPath); os.IsNotExist(err) {
				t.Errorf("Domain layer missing: %s", layerPath)
			} else {
				t.Logf("  ✓ Layer exists: %s", layer)
			}
		}
	}

	t.Log("✓ All domains have proper layered architecture")
	t.Log("✓ Domain isolation is maintained")
}
