# User and Role Management API

This API provides comprehensive user and role management functionality for the Feedback Hub application. It implements role-based access control with three predefined roles: Super User, Product Owner, and Contributor.

## Features

- **Role-based Access Control**: Three predefined roles with specific permissions
- **Super User Management**: Ultimate administrative privileges
- **User Management**: Create, read, update, and delete users
- **Role Management**: Create, read, update, and delete roles
- **Authorization**: Fine-grained permission checks
- **RESTful API**: Standard HTTP methods and status codes

## Architecture

The implementation follows Domain-Driven Design (DDD) principles:

- **Domain Layer**: Core business logic and entities
- **Application Layer**: Use cases and orchestration
- **Infrastructure Layer**: Database persistence
- **Interface Layer**: HTTP handlers and middleware

## API Endpoints

### Authentication

All endpoints require authentication via the `X-User-ID` header (simplified for demo purposes).

```
X-User-ID: <user-uuid>
```

### Roles

#### List Roles
```
GET /roles
```
Returns all roles in the system.

#### Get Role
```
GET /roles/{id}
```
Returns a specific role by ID.

#### Create Role
```
POST /roles
```
Creates a new role (Super User only).

**Request Body:**
```json
{
  "name": "Custom Role"
}
```

#### Update Role
```
PUT /roles/{id}
```
Updates a role's name (Super User only, cannot modify Super User role).

**Request Body:**
```json
{
  "name": "Updated Role Name"
}
```

#### Delete Role
```
DELETE /roles/{id}
```
Deletes a role (Super User only, cannot delete Super User role or roles with assigned users).

### Users

#### List Users
```
GET /users
```
Returns all users in the system.

#### Get User
```
GET /users/{id}
```
Returns a specific user by ID.

#### Create User
```
POST /users
```
Creates a new user with role assignment.

**Request Body:**
```json
{
  "email": "user@example.com",
  "name": "User Name",
  "role_id": "role-uuid"
}
```

**Authorization Rules:**
- Super User: Can create users with any role
- Product Owner: Can only create users with Contributor role
- Contributor: Cannot create users

#### Update User
```
PUT /users/{id}
```
Updates a user's name (email is immutable).

**Request Body:**
```json
{
  "name": "Updated Name"
}
```

#### Update User Role
```
PUT /users/{id}/role
```
Updates a user's role (Super User only).

**Request Body:**
```json
{
  "role_id": "new-role-uuid"
}
```

#### Delete User
```
DELETE /users/{id}
```
Deletes a user from the system.

## Business Rules

### Super User
- Has ultimate administrative privileges
- Can perform all operations
- Can create, update, and delete any role except modifying/deleting the Super User role itself
- Can create users with any role
- Cannot be deleted (the role itself)

### Product Owner
- Can create, read, update, and delete users
- Can only create users with the Contributor role
- Can read roles but cannot create, update, or delete them
- Cannot create users with Product Owner or Super User roles

### Contributor
- Can read users and roles
- Cannot create, update, or delete users or roles

## Error Handling

The API returns standard HTTP status codes:

- `200` - OK
- `201` - Created
- `204` - No Content
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

Error responses follow this format:
```json
{
  "error": "Bad Request",
  "message": "Detailed error message"
}
```

## Database Schema

### Roles Table
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

## Setup Instructions

1. **Environment Variables:**
   ```
   DATABASE_URL=postgresql://user:password@localhost:5432/feedback_hub
   PORT=8080
   SUPER_USER_EMAIL=admin@example.com
   SUPER_USER_NAME=System Administrator
   ```

2. **Database Setup:**
   ```bash
   # Run the migration script
   psql $DATABASE_URL -f scripts/migrate.sql
   ```

3. **Build and Run:**
   ```bash
   go build -o feedback-api ./cmd/api
   ./feedback-api
   ```

4. **The server will:**
   - Create predefined roles (Super User, Product Owner, Contributor)
   - Create an initial Super User if environment variables are set
   - Start the HTTP server on the specified port

## Testing

Run all tests:
```bash
go test ./...
```

Run domain tests only:
```bash
go test ./internal/domain/...
```

## API Documentation

Interactive API documentation is available at:
```
http://localhost:8080/swagger/
```

## Example Usage

1. **Create a Product Owner (as Super User):**
   ```bash
   curl -X POST http://localhost:8080/users \
     -H "Content-Type: application/json" \
     -H "X-User-ID: super-user-uuid" \
     -d '{
       "email": "po@example.com",
       "name": "Product Owner",
       "role_id": "product-owner-role-uuid"
     }'
   ```

2. **Create a Contributor (as Product Owner):**
   ```bash
   curl -X POST http://localhost:8080/users \
     -H "Content-Type: application/json" \
     -H "X-User-ID: product-owner-uuid" \
     -d '{
       "email": "contributor@example.com",
       "name": "Contributor User",
       "role_id": "contributor-role-uuid"
     }'
   ```

3. **List all roles:**
   ```bash
   curl -H "X-User-ID: any-user-uuid" http://localhost:8080/roles
   ```
