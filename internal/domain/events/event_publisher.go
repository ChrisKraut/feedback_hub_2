package events

import (
	"context"
)

// EventPublisher defines the interface for publishing domain events.
// AI-hint: Interface for publishing events, allowing services to publish
// events without directly depending on the event bus implementation.
type EventPublisher interface {
	// PublishEvent publishes a domain event
	PublishEvent(ctx context.Context, event DomainEvent) error
}

// EventBusPublisher implements EventPublisher using the event bus.
// AI-hint: Concrete implementation that wraps the event bus for publishing.
type EventBusPublisher struct {
	eventBus EventBus
}

// NewEventBusPublisher creates a new event bus publisher.
// AI-hint: Factory method for creating event publishers.
func NewEventBusPublisher(eventBus EventBus) *EventBusPublisher {
	return &EventBusPublisher{
		eventBus: eventBus,
	}
}

// PublishEvent publishes a domain event through the event bus.
// AI-hint: Delegates event publishing to the underlying event bus.
func (p *EventBusPublisher) PublishEvent(ctx context.Context, event DomainEvent) error {
	return p.eventBus.Publish(ctx, event)
}
