package persistence

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isUniqueViolation checks if the error is a unique constraint violation.
// AI-hint: Error classification helper for database constraint violations.
func isUniqueViolation(err error) bool {
	// PostgreSQL unique violation error code is 23505
	return containsErrorCode(err, "23505")
}

// isForeignKeyViolation checks if the error is a foreign key constraint violation.
// AI-hint: Error classification helper for foreign key constraint violations.
func isForeignKeyViolation(err error) bool {
	// PostgreSQL foreign key violation error code is 23503
	return containsErrorCode(err, "23503")
}

// containsErrorCode checks if the error contains a specific PostgreSQL error code.
// AI-hint: Generic error code checker for PostgreSQL error handling.
func containsErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}
	return false
}
