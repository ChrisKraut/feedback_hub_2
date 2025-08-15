package events

import (
	"time"
)

// DomainEvent represents a domain event that occurred in the system.
// AI-hint: Base interface for all domain events, providing common metadata
// and ensuring consistency across the event system.
type DomainEvent interface {
	// EventID returns the unique identifier for this event
	EventID() string
	
	// EventType returns the type/category of this event
	EventType() string
	
	// AggregateID returns the ID of the aggregate that generated this event
	AggregateID() string
	
	// OccurredAt returns when this event occurred
	OccurredAt() time.Time
	
	// Version returns the version of the aggregate when this event was created
	Version() int
}

// BaseDomainEvent provides common implementation for domain events.
// AI-hint: Concrete base class that implements common event functionality,
// reducing boilerplate in specific event implementations.
type BaseDomainEvent struct {
	eventID     string
	eventType   string
	aggregateID string
	occurredAt  time.Time
	version     int
}

// NewBaseDomainEvent creates a new base domain event.
// AI-hint: Factory method for base events with proper initialization.
func NewBaseDomainEvent(eventType, aggregateID string, version int) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:     generateEventID(),
		eventType:   eventType,
		aggregateID: aggregateID,
		occurredAt:  time.Now(),
		version:     version,
	}
}

// EventID returns the unique identifier for this event.
func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

// EventType returns the type/category of this event.
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// AggregateID returns the ID of the aggregate that generated this event.
func (e BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

// OccurredAt returns when this event occurred.
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// Version returns the version of the aggregate when this event was created.
func (e BaseDomainEvent) Version() int {
	return e.version
}

// generateEventID generates a unique event ID.
// AI-hint: Simple UUID generation for event identification.
func generateEventID() string {
	// For now, use timestamp-based ID. In production, consider using UUID.
	return time.Now().Format("20060102150405.000000000")
}
