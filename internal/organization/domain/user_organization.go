package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserOrganization represents the relationship between a user and an organization.
// AI-hint: Junction entity for many-to-many relationship between users and organizations.
// Each relationship includes the user's role within that organization and active status.
type UserOrganization struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	RoleID         uuid.UUID `json:"role_id"`
	IsActive       bool      `json:"is_active"`
	JoinedAt       time.Time `json:"joined_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewUserOrganization creates a new user-organization relationship.
// AI-hint: Factory method for creating new user-organization relationships with
// automatic ID generation and timestamp creation. Validates all input parameters.
func NewUserOrganization(userID, organizationID, roleID uuid.UUID) (*UserOrganization, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be zero")
	}

	if organizationID == uuid.Nil {
		return nil, fmt.Errorf("organization ID cannot be zero")
	}

	if roleID == uuid.Nil {
		return nil, fmt.Errorf("role ID cannot be zero")
	}

	now := time.Now()
	userOrg := &UserOrganization{
		ID:             uuid.New(),
		UserID:         userID,
		OrganizationID: organizationID,
		RoleID:         roleID,
		IsActive:       true, // Default to active
		JoinedAt:       now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return userOrg, nil
}

// NewUserOrganizationWithID creates a new user-organization relationship with a specific ID.
// AI-hint: Factory method for creating relationships with existing IDs, useful for
// reconstruction from database or testing scenarios.
func NewUserOrganizationWithID(id, userID, organizationID, roleID uuid.UUID, joinedAt time.Time, isActive bool) (*UserOrganization, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("user-organization relationship ID cannot be zero")
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be zero")
	}

	if organizationID == uuid.Nil {
		return nil, fmt.Errorf("organization ID cannot be zero")
	}

	if roleID == uuid.Nil {
		return nil, fmt.Errorf("role ID cannot be zero")
	}

	if joinedAt.IsZero() {
		return nil, fmt.Errorf("joined at cannot be zero")
	}

	now := time.Now()
	userOrg := &UserOrganization{
		ID:             id,
		UserID:         userID,
		OrganizationID: organizationID,
		RoleID:         roleID,
		IsActive:       isActive,
		JoinedAt:       joinedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return userOrg, nil
}

// UpdateRole updates the user's role within the organization.
// AI-hint: Business method for updating user roles within organizations.
// Automatically updates the UpdatedAt timestamp and validates the new role ID.
func (uo *UserOrganization) UpdateRole(roleID uuid.UUID) error {
	if roleID == uuid.Nil {
		return fmt.Errorf("role ID cannot be zero")
	}

	uo.RoleID = roleID
	uo.UpdatedAt = time.Now()

	return nil
}

// SetActive sets the active status of the user-organization relationship.
// AI-hint: Business method for activating/deactivating user access to organizations.
// Automatically updates the UpdatedAt timestamp.
func (uo *UserOrganization) SetActive(isActive bool) error {
	uo.IsActive = isActive
	uo.UpdatedAt = time.Now()

	return nil
}

// Validate ensures the user-organization relationship is in a valid state.
// AI-hint: Domain validation method that ensures all business rules are satisfied.
// Called before persistence operations to maintain data integrity.
func (uo *UserOrganization) Validate() error {
	if uo.ID == uuid.Nil {
		return fmt.Errorf("user-organization relationship ID cannot be zero")
	}

	if uo.UserID == uuid.Nil {
		return fmt.Errorf("user ID cannot be zero")
	}

	if uo.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization ID cannot be zero")
	}

	if uo.RoleID == uuid.Nil {
		return fmt.Errorf("role ID cannot be zero")
	}

	if uo.JoinedAt.IsZero() {
		return fmt.Errorf("joined at cannot be zero")
	}

	if uo.CreatedAt.IsZero() {
		return fmt.Errorf("created at cannot be zero")
	}

	if uo.UpdatedAt.IsZero() {
		return fmt.Errorf("updated at cannot be zero")
	}

	return nil
}

// IsActiveInOrganization checks if the user is currently active in the organization.
// AI-hint: Business method that checks the active status of the relationship.
// Returns true if the user is active, false otherwise.
func (uo *UserOrganization) IsActiveInOrganization() bool {
	return uo.IsActive
}

// GetTenure calculates how long the user has been part of the organization.
// AI-hint: Business method that calculates the user's tenure in the organization.
// Useful for analytics and user management features.
func (uo *UserOrganization) GetTenure() time.Duration {
	return time.Since(uo.JoinedAt)
}

// Clone creates a deep copy of the user-organization relationship with a new ID and timestamps.
// AI-hint: Utility method for creating relationship copies, useful for testing
// and creating relationship templates. Ensures complete isolation between copies.
func (uo *UserOrganization) Clone() *UserOrganization {
	now := time.Now()
	return &UserOrganization{
		ID:             uuid.New(),
		UserID:         uo.UserID,
		OrganizationID: uo.OrganizationID,
		RoleID:         uo.RoleID,
		IsActive:       uo.IsActive,
		JoinedAt:       uo.JoinedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// IsSameUser checks if this relationship belongs to the same user.
// AI-hint: Utility method for comparing user-organization relationships.
// Useful for validation and business logic that needs to check user identity.
func (uo *UserOrganization) IsSameUser(other *UserOrganization) bool {
	if other == nil {
		return false
	}
	return uo.UserID == other.UserID
}

// IsSameOrganization checks if this relationship belongs to the same organization.
// AI-hint: Utility method for comparing user-organization relationships.
// Useful for validation and business logic that needs to check organization identity.
func (uo *UserOrganization) IsSameOrganization(other *UserOrganization) bool {
	if other == nil {
		return false
	}
	return uo.OrganizationID == other.OrganizationID
}

// HasRole checks if the user has a specific role in the organization.
// AI-hint: Business method for role-based access control.
// Useful for permission checking and authorization logic.
func (uo *UserOrganization) HasRole(roleID uuid.UUID) bool {
	return uo.RoleID == roleID
}

// IsRecentlyJoined checks if the user joined the organization recently.
// AI-hint: Business method for identifying new users in organizations.
// Useful for onboarding flows and user experience features.
func (uo *UserOrganization) IsRecentlyJoined(threshold time.Duration) bool {
	return time.Since(uo.JoinedAt) <= threshold
}
