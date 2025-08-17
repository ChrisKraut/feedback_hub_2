# Domain Events System

This package implements a comprehensive domain events system for the Feedback Hub backend, enabling loose coupling between different domains through event-driven communication.

## Overview

The domain events system follows Domain-Driven Design (DDD) principles by:

- **Decoupling domains**: Services communicate through events rather than direct dependencies
- **Enabling extensibility**: New domains can react to events without modifying existing code
- **Supporting business rules**: Events carry business context and can trigger cross-domain workflows
- **Maintaining consistency**: Event publishing is integrated into business operations

## Architecture

### Core Components

1. **DomainEvent Interface**: Base interface for all domain events
2. **BaseDomainEvent**: Concrete implementation with common event metadata
3. **EventBus**: Manages event publishing and subscription
4. **EventPublisher**: Abstraction layer for publishing events
5. **Specific Event Types**: Concrete events for different business operations

### Event Flow

```
Business Operation → Service → EventPublisher → EventBus → EventHandlers
```

## Usage

### Publishing Events

Services publish events after successful business operations:

```go
// In UserService.CreateUser()
userCreatedEvent := events.NewUserCreatedEvent(newUser.ID, newUser.Email, newUser.Name, newUser.RoleID, targetRole.Name)
if err := s.eventPublisher.PublishEvent(context, userCreatedEvent); err != nil {
    log.Printf("Warning: failed to publish user created event: %v", err)
    // Don't fail the operation if event publishing fails
}
```

### Subscribing to Events

Domains can subscribe to events they're interested in:

```go
eventBus := events.NewInMemoryEventBus()
eventBus.Subscribe("user.created", handleUserCreated)
eventBus.Subscribe("role.updated", handleRoleUpdated)
```

### Event Handlers

Event handlers process events and can trigger cross-domain operations:

```go
func handleUserCreated(ctx context.Context, event events.DomainEvent) error {
    userEvent := event.(*events.UserCreatedEvent)
    
    // React to user creation (e.g., send welcome email, create profile, etc.)
    log.Printf("New user created: %s with role: %s", userEvent.UserID, userEvent.RoleName)
    
    return nil
}
```

## Event Types

### User Events

- **UserCreatedEvent**: Triggered when a user is created
- **UserUpdatedEvent**: Triggered when a user's information is updated
- **UserRoleUpdatedEvent**: Triggered when a user's role changes

### Role Events

- **RoleCreatedEvent**: Triggered when a role is created
- **RoleUpdatedEvent**: Triggered when a role is updated
- **RoleDeletedEvent**: Triggered when a role is deleted

## Benefits

### 1. Reduced Coupling

**Before (Direct Dependencies)**:
```go
type UserService struct {
    userRepo    user.Repository
    roleRepo    role.Repository        // Direct dependency
    authService *auth.AuthorizationService
    emailService *email.EmailService   // Direct dependency
    auditService *audit.AuditService   // Direct dependency
}
```

**After (Event-Driven)**:
```go
type UserService struct {
    userRepo        user.Repository
    roleRepo        role.Repository
    authService     *auth.AuthorizationService
    eventPublisher  events.EventPublisher  // Single dependency
}
```

### 2. Easy Extension

New domains can react to events without modifying existing services:

```go
// New notification domain
eventBus.Subscribe("user.created", func(ctx context.Context, event events.DomainEvent) error {
    // Send welcome email
    return nil
})

// New audit domain
eventBus.Subscribe("user.created", func(ctx context.Context, event events.DomainEvent) error {
    // Log user creation
    return nil
})
```

### 3. Business Rule Enforcement

Events carry business context and can trigger complex workflows:

```go
// When a user's role changes, multiple domains can react
eventBus.Subscribe("user.role_updated", func(ctx context.Context, event events.DomainEvent) error {
    roleEvent := event.(*events.UserRoleUpdatedEvent)
    
    // Update permissions
    // Send notification
    // Update audit trail
    // Trigger approval workflows
    
    return nil
})
```

## Testing

The system includes comprehensive tests:

```bash
# Run all event system tests
go test ./internal/domain/events/... -v

# Run specific test categories
go test ./internal/domain/events/... -v -run TestInMemoryEventBus
go test ./internal/domain/events/... -v -run TestDomainEventsIntegration
```

## Future Enhancements

### 1. Event Persistence

Store events for audit trails and event sourcing:

```go
type EventStore interface {
    Save(event DomainEvent) error
    GetEvents(aggregateID string) ([]DomainEvent, error)
}
```

### 2. Event Versioning

Support event schema evolution:

```go
type DomainEvent interface {
    EventID() string
    EventType() string
    AggregateID() string
    OccurredAt() time.Time
    Version() int
    SchemaVersion() int  // New field
}
```

### 3. Event Replay

Replay events for debugging and testing:

```go
func (bus *EventBus) ReplayEvents(events []DomainEvent) error {
    for _, event := range events {
        if err := bus.Publish(context.Background(), event); err != nil {
            return err
        }
    }
    return nil
}
```

### 4. Event Correlation

Track related events across domains:

```go
type DomainEvent interface {
    // ... existing methods
    CorrelationID() string  // New field
    CausationID() string    // New field
}
```

## Best Practices

### 1. Event Naming

Use consistent naming conventions:
- `{domain}.{action}` (e.g., `user.created`, `role.updated`)
- Use past tense for completed actions
- Be specific about what changed

### 2. Event Data

Include only necessary data in events:
- **Include**: IDs, essential business data, timestamps
- **Exclude**: Sensitive information, internal implementation details
- **Consider**: What other domains need to know

### 3. Error Handling

Handle event publishing failures gracefully:
- Log warnings but don't fail business operations
- Consider retry mechanisms for critical events
- Monitor event publishing success rates

### 4. Event Ordering

Ensure events are published in the correct order:
- Publish events after successful business operations
- Use database transactions when possible
- Consider event ordering guarantees

## Example: Complete Workflow

Here's how the system works in practice:

1. **User Creation Request**:
   ```http
   POST /users
   {
     "email": "user@example.com",
     "name": "John Doe",
     "role_id": "role-123"
   }
   ```

2. **Service Processing**:
   ```go
   func (s *UserService) CreateUser(ctx context.Context, email, name, roleID string, createdByUserID string) (*user.User, error) {
       // ... business logic ...
       
       // Create user
       newUser, err := user.NewUser(userID, email, name, roleID)
       if err != nil {
           return nil, err
       }
       
       // Save to database
       if err := s.userRepo.Create(ctx, newUser); err != nil {
           return nil, err
       }
       
       // Publish event
       userCreatedEvent := events.NewUserCreatedEvent(newUser.ID, newUser.Email, newUser.Name, newUser.RoleID, targetRole.Name)
       if err := s.eventPublisher.PublishEvent(ctx, userCreatedEvent); err != nil {
           log.Printf("Warning: failed to publish user created event: %v", err)
       }
       
       return newUser, nil
   }
   ```

3. **Event Handling**:
   ```go
   // Notification domain
   eventBus.Subscribe("user.created", func(ctx context.Context, event events.DomainEvent) error {
       userEvent := event.(*events.UserCreatedEvent)
       return sendWelcomeEmail(userEvent.Email, userEvent.Name)
   })
   
   // Audit domain
   eventBus.Subscribe("user.created", func(ctx context.Context, event events.DomainEvent) error {
       userEvent := event.(*events.UserCreatedEvent)
       return logUserCreation(userEvent.UserID, userEvent.Email, userEvent.RoleName)
   })
   ```

This architecture enables the system to grow and evolve while maintaining clean separation between domains and supporting complex business workflows.
