# Ideas Domain

The Ideas domain provides functionality for users to submit and manage feedback ideas with markdown-formatted text support.

## Overview

The Ideas domain allows authenticated users to:
- Submit new ideas with markdown-formatted text
- View all submitted ideas (ordered by creation date)
- View ideas by specific user
- Update and delete their own ideas

## API Endpoints

### Create Idea
- **POST** `/ideas`
- **Authentication**: Required
- **Request Body**:
  ```json
  {
    "idea_text": "A brilliant new feature suggestion with **bold** and *italic* text."
  }
  ```
- **Response**: 201 Created with the created idea object
- **Notes**: The `idea_text` field accepts markdown formatting and is automatically sanitized for XSS protection

### List All Ideas
- **GET** `/ideas`
- **Authentication**: Required
- **Response**: 200 OK with array of idea objects
- **Notes**: Ideas are ordered by creation date (most recent first)

### Get Idea by ID
- **GET** `/ideas/{id}`
- **Authentication**: Required
- **Response**: 200 OK with idea object or 404 Not Found

### Get Ideas by User
- **GET** `/users/{id}/ideas`
- **Authentication**: Required
- **Response**: 200 OK with array of idea objects for the specified user

## Data Model

### Idea Entity
```go
type Idea struct {
    ID        string    `json:"id"`
    IdeaText  string    `json:"idea_text"`
    UserID    string    `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Database Schema
```sql
CREATE TABLE ideas (
    id UUID PRIMARY KEY,
    idea_text TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_ideas_user_id ON ideas(user_id);
CREATE INDEX idx_ideas_created_at ON ideas(created_at DESC);

-- Constraints
ALTER TABLE ideas ADD CONSTRAINT chk_ideas_text_not_empty CHECK (trim(idea_text) != '');
```

## Security Features

### XSS Protection
- All idea text is automatically sanitized before storage
- Script tags and dangerous HTML elements are removed
- Markdown formatting is preserved
- Input validation ensures non-empty content

### Authorization
- All endpoints require authentication
- Users can only update/delete their own ideas
- User existence is validated before idea creation

## Business Rules

1. **Idea Creation**:
   - Idea text cannot be empty
   - User ID must reference an existing user
   - Automatic UUID generation for idea ID
   - Timestamp automatically set to creation time

2. **Idea Updates**:
   - Only the creator can update their idea
   - Text validation and sanitization on updates
   - User ID cannot be changed

3. **Idea Deletion**:
   - Only the creator can delete their idea
   - Cascading deletion when user is deleted

4. **Ordering**:
   - Ideas are always ordered by creation date (newest first)
   - Consistent ordering across all listing endpoints

## Error Handling

The API returns appropriate HTTP status codes:
- **400 Bad Request**: Invalid input data
- **401 Unauthorized**: Authentication required
- **404 Not Found**: Idea or user not found
- **500 Internal Server Error**: Server-side errors

## Usage Examples

### Creating an Idea
```bash
curl -X POST http://localhost:8080/ideas \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=your_jwt_token" \
  -d '{
    "idea_text": "Add dark mode support with **bold** and *italic* text formatting"
  }'
```

### Listing All Ideas
```bash
curl -X GET http://localhost:8080/ideas \
  -H "Cookie: auth_token=your_jwt_token"
```

### Getting Ideas by User
```bash
curl -X GET http://localhost:8080/users/user-uuid/ideas \
  -H "Cookie: auth_token=your_jwt_token"
```

## Implementation Details

### Domain Layer
- **Location**: `internal/domain/idea/`
- **Core Entity**: `Idea` with business logic and validation
- **Interfaces**: `Repository` and `Service` for dependency inversion

### Application Layer
- **Location**: `internal/application/idea_service.go`
- **Business Logic**: Coordinates between domain and infrastructure
- **Authorization**: Enforces business rules and permissions

### Infrastructure Layer
- **Location**: `internal/infrastructure/persistence/idea_repository.go`
- **Database**: PostgreSQL implementation with proper indexing
- **Schema**: Automatic table creation and migration

### HTTP Layer
- **Location**: `internal/interfaces/http/idea_handler.go`
- **REST API**: Standard HTTP endpoints with proper status codes
- **Validation**: Request validation and error handling
- **Documentation**: Swagger annotations for API docs

## Future Enhancements

Potential improvements for future iterations:
- Idea categories/tags
- Idea voting/rating system
- Idea status tracking (draft, submitted, in-review, implemented)
- Rich text editor with markdown preview
- Idea search and filtering
- Idea comments and discussions
- Email notifications for idea updates
