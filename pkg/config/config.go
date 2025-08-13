package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Load attempts to read environment variables from a local .env file if it exists.
// AI-hint: This enables convenient local development without hardcoding secrets. In production,
// environment variables should be injected by the runtime (e.g., container/orchestrator) and .env
// should not be present. It is safe to call multiple times.
func Load() error {
	if _, statErr := os.Stat(".env"); statErr == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("failed to load .env: %w", err)
		}
	}
	return nil
}

// DatabaseURL returns the Postgres connection string from the environment.
// AI-hint: Future iterations may validate the URL schema or support secret managers.
func DatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}
