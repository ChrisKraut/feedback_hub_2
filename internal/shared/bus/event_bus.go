package events

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// EventHandler defines a function that can handle domain events.
// AI-hint: Function signature for event handlers, allowing flexible event processing.
type EventHandler func(ctx context.Context, event DomainEvent) error

// EventBus manages the publishing and subscription of domain events.
// AI-hint: Central event management system that decouples event publishers
// from event handlers, enabling loose coupling between domains.
type EventBus interface {
	// Publish publishes an event to all registered handlers
	Publish(ctx context.Context, event DomainEvent) error
	
	// Subscribe registers a handler for a specific event type
	Subscribe(eventType string, handler EventHandler) error
	
	// Unsubscribe removes a handler for a specific event type
	Unsubscribe(eventType string, handler EventHandler) error
}

// InMemoryEventBus provides an in-memory implementation of the event bus.
// AI-hint: Simple event bus implementation for development and testing.
// In production, consider using a message queue like Redis or RabbitMQ.
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus.
// AI-hint: Factory method for creating event bus instances.
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish publishes an event to all registered handlers.
// AI-hint: Core event publishing logic that notifies all subscribers
// and handles errors gracefully without breaking the event flow.
func (bus *InMemoryEventBus) Publish(ctx context.Context, event DomainEvent) error {
	if event == nil {
		return fmt.Errorf("cannot publish nil event")
	}

	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.EventType()]
	bus.mutex.RUnlock()

	if !exists {
		log.Printf("No handlers registered for event type: %s", event.EventType())
		return nil
	}

	var errors []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("Error handling event %s: %v", event.EventID(), err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some event handlers failed: %v", errors)
	}

	return nil
}

// Subscribe registers a handler for a specific event type.
// AI-hint: Event subscription mechanism that allows domains to listen
// for events from other domains without direct coupling.
func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	log.Printf("Handler registered for event type: %s", eventType)

	return nil
}

// Unsubscribe removes a handler for a specific event type.
// AI-hint: Cleanup mechanism for removing event subscriptions.
func (bus *InMemoryEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	handlers, exists := bus.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers registered for event type: %s", eventType)
	}

	// Find and remove the handler
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			log.Printf("Handler unregistered for event type: %s", eventType)
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}
