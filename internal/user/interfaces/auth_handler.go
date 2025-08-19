package interfaces

import (
	"encoding/json"
	"net/http"
	"os"

	roleapp "feedback_hub_2/internal/role/application"
	"feedback_hub_2/internal/shared/web"
	userapp "feedback_hub_2/internal/user/application"
	"feedback_hub_2/internal/user/infrastructure/auth"

	"github.com/google/uuid"
)

// AuthHandler handles HTTP requests for authentication operations.
// AI-hint: HTTP transport layer for authentication with JWT and password-based login.
// Provides secure login/logout with HTTP-only cookie token storage.
type AuthHandler struct {
	userService     *userapp.UserService
	roleService     *roleapp.RoleService
	jwtService      *auth.JWTService
	passwordService *auth.PasswordService
}

// NewAuthHandler creates a new AuthHandler instance.
// AI-hint: Factory method for auth handler with dependency injection of required services.
func NewAuthHandler(userService *userapp.UserService, roleService *roleapp.RoleService, jwtService *auth.JWTService, passwordService *auth.PasswordService) *AuthHandler {
	return &AuthHandler{
		userService:     userService,
		roleService:     roleService,
		jwtService:      jwtService,
		passwordService: passwordService,
	}
}

// LoginRequest represents the request body for user login.
// AI-hint: DTO for login API with email and password fields.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents the request body for user registration.
// AI-hint: User registration with email, name, password, and organization scoping.
type RegisterRequest struct {
	Email          string `json:"email"`
	Name           string `json:"name"`
	Password       string `json:"password"`
	OrganizationID string `json:"organization_id"` // Organization scoping for multi-tenant support
}

// AuthResponse represents the response body for authentication operations.
// AI-hint: DTO for auth responses with user info (token stored in HTTP-only cookie).
type AuthResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	RoleName string `json:"role_name"`
	Message  string `json:"message"`
}

// Login handles POST /auth/login requests.
// AI-hint: User authentication endpoint with password verification and JWT token generation.
//
// @Summary User login
// @Description Authenticate user with email and password, returns JWT in HTTP-only cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if user.PasswordHash == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Password login not available for this account")
		return
	}

	if err := h.passwordService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Get user's role for JWT claims
	role, err := h.roleService.GetRole(r.Context(), user.RoleID)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user role")
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user.ID, user.Email, role.Name)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Set HTTP-only cookie with environment-based security
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Secure:   isProduction, // HTTPS only in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
		Path:     "/",
	})

	// Return user info (not the token)
	response := AuthResponse{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleName: role.Name,
		Message:  "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Register handles POST /auth/register requests.
// AI-hint: User registration endpoint with password hashing and role assignment.
//
// @Summary User registration
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Registration details"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Name == "" || req.Password == "" || req.OrganizationID == "" {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Email, name, password, and organization ID are required")
		return
	}

	// Validate password strength
	if !h.passwordService.IsValidPassword(req.Password) {
		web.WriteErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	// Hash password
	hashedPassword, err := h.passwordService.HashPassword(req.Password)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Get default "Contributor" role for new users
	contributorRole, err := h.roleService.GetRoleByName(r.Context(), "Contributor")
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get default role")
		return
	}

	// Create user with hashed password
	userID := uuid.New().String()
	user, err := h.userService.CreateUserWithPassword(r.Context(), userID, req.Email, req.Name, hashedPassword, contributorRole.ID, req.OrganizationID)
	if err != nil {
		if err.Error() == "email already exists" {
			web.WriteErrorResponse(w, http.StatusConflict, "Email already registered")
			return
		}
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token for immediate login
	token, err := h.jwtService.GenerateToken(user.ID, user.Email, contributorRole.Name)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Set HTTP-only cookie with environment-based security
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Secure:   isProduction, // HTTPS only in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
		Path:     "/",
	})

	// Return user info
	response := AuthResponse{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleName: contributorRole.Name,
		Message:  "Registration successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Logout handles POST /auth/logout requests.
// AI-hint: User logout endpoint that clears the authentication cookie.
//
// @Summary User logout
// @Description Logout user by clearing authentication cookie
// @Tags auth
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie with environment-based security
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		Secure:   isProduction, // HTTPS only in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete cookie
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

// Me handles GET /auth/me requests.
// AI-hint: Current user info endpoint that returns authenticated user details.
//
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Security JWTAuth
// @Produce json
// @Success 200 {object} AuthResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// User ID is already set in context by auth middleware
	userID := web.GetUserIDFromContext(r.Context())
	if userID == "" {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Get user details
	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Get user's role
	role, err := h.roleService.GetRole(r.Context(), user.RoleID)
	if err != nil {
		web.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user role")
		return
	}

	// Return user info
	response := AuthResponse{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleName: role.Name,
		Message:  "Authenticated",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
