# Feedback Hub Development Guide

## Table of Contents
- [Overview](#overview)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Swagger Documentation](#swagger-documentation)
- [Adding New Features](#adding-new-features)
- [Testing](#testing)
- [Deployment](#deployment)
- [Common Issues & Solutions](#common-issues--solutions)

## Overview

Feedback Hub is a Go-based application built with Domain-Driven Design (DDD) principles, featuring JWT authentication, role-based access control, and a PostgreSQL database. The application follows clean architecture patterns with clear separation of concerns.

## Project Structure

```
feedback_hub_2/
├── cmd/api/                    # Application entry point
├── internal/                   # Private application code
│   ├── application/           # Application services (business logic)
│   ├── domain/               # Domain models and business rules
│   ├── infrastructure/       # External concerns (database, auth)
│   └── interfaces/           # HTTP handlers and API layer
├── docs/                     # Swagger documentation
├── pkg/                      # Public packages
├── scripts/                  # Database migrations
└── tests/                    # Integration tests
```

## Development Workflow

### 1. **Feature Development Process**
```
Domain Model → Repository → Application Service → HTTP Handler → Tests → Swagger Docs
```

### 2. **Code Organization Principles**
- **Domain Layer**: Pure business logic, no external dependencies
- **Application Layer**: Orchestrates domain entities and repositories
- **Infrastructure Layer**: Implements interfaces defined in domain
- **Interface Layer**: HTTP transport, validation, and error handling

### 3. **AI-Friendly Development**
- Add comprehensive comments starting with `// AI-hint:` for future AI iterations
- Document business rules and domain invariants
- Explain architectural decisions and patterns

## Swagger Documentation

### ⚠️ **CRITICAL LEARNING: Swagger Generation**

**The most important lesson learned:** This project uses `swaggo/swag` which **automatically generates Swagger documentation from Go code annotations**, NOT from static files.

#### **What We Learned:**
1. **Static files are ignored** - `swagger.yaml` and `swagger.json` are overwritten during build
2. **Documentation comes from Go code** - Add Swagger annotations to your handlers
3. **Always regenerate docs** after adding new endpoints

#### **How to Add Swagger Documentation:**

1. **Add annotations to your HTTP handler:**
```go
// CreateIdea handles POST /v1/ideas requests.
// AI-hint: Idea creation endpoint with authentication, validation, and proper error handling.
//
// @Summary Create a new idea
// @Description Create a new feedback idea (authentication required)
// @Tags ideas
// @Accept json
// @Produce json
// @Param idea body CreateIdeaRequest true "Idea creation request"
// @Success 201 {object} CreateIdeaResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security JWTAuth
// @Router /v1/ideas [post]
func (h *IdeaHandler) CreateIdea(w http.ResponseWriter, r *http.Request) {
    // ... implementation
}
```

2. **Regenerate Swagger docs:**
```bash
# Install swag tool (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate documentation
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g cmd/api/main.go -o docs
```

#### **Swagger File Management:**
- **`docs/docs.go`**: Auto-generated, DO NOT EDIT
- **`docs/swagger.json`**: Auto-generated, DO NOT EDIT  
- **`docs/swagger.yaml`**: Auto-generated, DO NOT EDIT
- **Static files are for reference only** - they get overwritten

## Adding New Features

### **Step-by-Step Process:**

#### 1. **Create Domain Model** (`internal/domain/feature/feature.go`)
```go
package feature

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

type Feature struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Repository interface
type Repository interface {
    Save(ctx interface{}, feature *Feature) error
    FindByID(ctx interface{}, id uuid.UUID) (*Feature, error)
    // ... other methods
}

// Error types
var (
    ErrFeatureNotFound = errors.New("feature not found")
    ErrInvalidData     = errors.New("invalid feature data")
)
```

#### 2. **Create Repository Implementation** (`internal/infrastructure/persistence/feature_repository.go`)
```go
package persistence

import (
    "context"
    "feedback_hub_2/internal/domain/feature"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

type FeatureRepository struct {
    pool *pgxpool.Pool
}

func NewFeatureRepository(pool *pgxpool.Pool) *FeatureRepository {
    return &FeatureRepository{pool: pool}
}

func (r *FeatureRepository) Save(ctx interface{}, feature *feature.Feature) error {
    // Implementation with proper error handling
}
```

#### 3. **Create Application Service** (`internal/application/feature_service.go`)
```go
package application

import (
    "context"
    "feedback_hub_2/internal/domain/feature"
    "github.com/google/uuid"
)

type CreateFeatureCommand struct {
    Name string `json:"name"`
}

type FeatureApplicationService struct {
    featureRepo feature.Repository
}

func (s *FeatureApplicationService) CreateFeature(ctx context.Context, cmd CreateFeatureCommand) (uuid.UUID, error) {
    // Business logic implementation
}
```

#### 4. **Create HTTP Handler** (`internal/interfaces/http/feature_handler.go`)
```go
package http

import (
    "encoding/json"
    "feedback_hub_2/internal/application"
    "net/http"
)

type FeatureHandler struct {
    featureService *application.FeatureApplicationService
}

// Add Swagger annotations here!
func (h *FeatureHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
    // HTTP handling implementation
}
```

#### 5. **Add Database Migration** (`scripts/migrate_feature.sql`)
```sql
-- Feature table migration
CREATE TABLE IF NOT EXISTS features (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add constraints and indexes
ALTER TABLE features ADD CONSTRAINT chk_features_name_not_empty CHECK (trim(name) != '');
CREATE INDEX IF NOT EXISTS idx_features_name ON features(name);
```

#### 6. **Update Main Schema** (`internal/infrastructure/persistence/schema.sql`)
```sql
-- Add your new table to the main schema file
CREATE TABLE IF NOT EXISTS features (
    -- ... table definition
);
```

#### 7. **Add Routing** (in your main server file)
```go
// Initialize repositories and services
featureRepo := persistence.NewFeatureRepository(dbPool)
featureService := application.NewFeatureApplicationService(featureRepo, userRepo)
featureHandler := http.NewFeatureHandler(featureService)

// Add routes
mux.HandleFunc("/v1/features", featureHandler.CreateFeature)
```

#### 8. **Regenerate Swagger Documentation**
```bash
swag init -g cmd/api/main.go -o docs
```

### **Key Patterns to Follow:**

1. **Error Handling**: Use domain-specific error types
2. **Validation**: Validate at both HTTP and domain layers
3. **Authentication**: Use the existing JWT middleware
4. **Dependency Injection**: Pass dependencies through constructors
5. **Context Usage**: Use context.Context for cancellation and timeouts

## Testing

### **Test Structure:**
- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test database interactions
- **Handler Tests**: Test HTTP endpoints with mocked services

### **Testing Best Practices:**
```go
// Use mocks for external dependencies
type MockFeatureRepository struct {
    mock.Mock
}

// Test both success and failure scenarios
func TestCreateFeature_Success(t *testing.T) { ... }
func TestCreateFeature_ValidationError(t *testing.T) { ... }
func TestCreateFeature_DatabaseError(t *testing.T) { ... }
```

## Deployment

### **Pre-Deployment Checklist:**
1. ✅ All tests pass
2. ✅ Swagger documentation is regenerated
3. ✅ Database migrations are ready
4. ✅ Environment variables are configured
5. ✅ Build succeeds without errors

### **Deployment Commands:**
```bash
# Build the application
go build -o feedback-api cmd/api/main.go

# Run database migrations
psql -d your_database -f scripts/migrate_feature.sql

# Start the application
./feedback-api
```

## Common Issues & Solutions

### **1. Swagger Endpoint Not Visible**
**Problem**: Added endpoint to Go code but not visible in Swagger UI
**Solution**: 
- Ensure Swagger annotations are added to handler
- Run `swag init -g cmd/api/main.go -o docs`
- Check that handler is properly wired in routing

### **2. Compilation Errors**
**Problem**: `undefined: application.FeatureService`
**Solution**: 
- Create the missing service/repository first
- Follow the dependency chain: Domain → Repository → Service → Handler
- Check import paths and package names

### **3. Database Connection Issues**
**Problem**: Cannot connect to PostgreSQL
**Solution**:
- Verify database is running
- Check connection string in environment variables
- Ensure database exists and user has proper permissions

### **4. Authentication Not Working**
**Problem**: JWT tokens not being validated
**Solution**:
- Check that auth middleware is properly configured
- Verify JWT secret is set in environment
- Ensure cookies are being sent with requests

## Best Practices Summary

1. **Always follow the layered architecture** - don't skip layers
2. **Add comprehensive Swagger annotations** to all handlers
3. **Use domain-specific error types** for clear error handling
4. **Write tests for all new functionality**
5. **Regenerate Swagger docs** after adding new endpoints
6. **Follow the established naming conventions** and patterns
7. **Add AI-friendly comments** for future development
8. **Validate input at multiple layers** for security
9. **Use proper HTTP status codes** and error responses
10. **Keep dependencies minimal** and well-defined

## Quick Reference Commands

```bash
# Install swag tool
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate Swagger docs
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g cmd/api/main.go -o docs

# Run tests
go test ./...

# Build application
go build -o feedback-api cmd/api/main.go

# Check for compilation errors
go build ./...
```

---

**Remember**: The key to success with this project is understanding that **Swagger documentation is generated from Go code, not static files**. Always add proper annotations and regenerate docs after changes!
