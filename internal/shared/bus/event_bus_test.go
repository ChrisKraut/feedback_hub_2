package events

import (
	"context"
	"errors"
	"testing"
)

func TestInMemoryEventBus(t *testing.T) {
	t.Run("should create new event bus", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		if bus == nil {
			t.Error("expected event bus to be created")
		}
		if bus.handlers == nil {
			t.Error("expected handlers map to be initialized")
		}
	})

	t.Run("should subscribe to event type", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		handler := func(ctx context.Context, event DomainEvent) error { return nil }

		err := bus.Subscribe("test.event", handler)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		handlers, exists := bus.handlers["test.event"]
		if !exists {
			t.Error("expected handlers to be registered")
		}
		if len(handlers) != 1 {
			t.Errorf("expected 1 handler, got %d", len(handlers))
		}
	})

	t.Run("should not subscribe with empty event type", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		handler := func(ctx context.Context, event DomainEvent) error { return nil }

		err := bus.Subscribe("", handler)
		if err == nil {
			t.Error("expected error for empty event type")
		}
	})

	t.Run("should not subscribe with nil handler", func(t *testing.T) {
		bus := NewInMemoryEventBus()

		err := bus.Subscribe("test.event", nil)
		if err == nil {
			t.Error("expected error for nil handler")
		}
	})

	t.Run("should publish event to subscribed handlers", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		receivedEvent := make(chan DomainEvent, 1)
		handler := func(ctx context.Context, event DomainEvent) error {
			receivedEvent <- event
			return nil
		}

		err := bus.Subscribe("test.event", handler)
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}

		testEvent := NewBaseDomainEvent("test.event", "test-123", 1)
		err = bus.Publish(context.Background(), testEvent)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		select {
		case event := <-receivedEvent:
			if event.EventID() != testEvent.EventID() {
				t.Errorf("expected event ID %s, got %s", testEvent.EventID(), event.EventID())
			}
		default:
			t.Error("expected event to be received by handler")
		}
	})

	t.Run("should handle multiple handlers for same event type", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		receivedCount := 0
		handler1 := func(ctx context.Context, event DomainEvent) error {
			receivedCount++
			return nil
		}
		handler2 := func(ctx context.Context, event DomainEvent) error {
			receivedCount++
			return nil
		}

		bus.Subscribe("test.event", handler1)
		bus.Subscribe("test.event", handler2)

		testEvent := NewBaseDomainEvent("test.event", "test-123", 1)
		err := bus.Publish(context.Background(), testEvent)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if receivedCount != 2 {
			t.Errorf("expected 2 handlers to be called, got %d", receivedCount)
		}
	})

	t.Run("should handle handler errors gracefully", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		handler := func(ctx context.Context, event DomainEvent) error {
			return errors.New("handler error")
		}

		bus.Subscribe("test.event", handler)

		testEvent := NewBaseDomainEvent("test.event", "test-123", 1)
		err := bus.Publish(context.Background(), testEvent)
		if err == nil {
			t.Error("expected error when handler fails")
		}
	})

	t.Run("should not publish nil event", func(t *testing.T) {
		bus := NewInMemoryEventBus()

		err := bus.Publish(context.Background(), nil)
		if err == nil {
			t.Error("expected error for nil event")
		}
	})

	t.Run("should handle no handlers gracefully", func(t *testing.T) {
		bus := NewInMemoryEventBus()

		testEvent := NewBaseDomainEvent("test.event", "test-123", 1)
		err := bus.Publish(context.Background(), testEvent)
		if err != nil {
			t.Errorf("expected no error when no handlers, got %v", err)
		}
	})

	t.Run("should unsubscribe handler", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		handler := func(ctx context.Context, event DomainEvent) error { return nil }

		bus.Subscribe("test.event", handler)
		if len(bus.handlers["test.event"]) != 1 {
			t.Error("expected handler to be subscribed")
		}

		err := bus.Unsubscribe("test.event", handler)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(bus.handlers["test.event"]) != 0 {
			t.Error("expected handler to be unsubscribed")
		}
	})

	t.Run("should not unsubscribe with empty event type", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		handler := func(ctx context.Context, event DomainEvent) error { return nil }

		err := bus.Unsubscribe("", handler)
		if err == nil {
			t.Error("expected error for empty event type")
		}
	})

	t.Run("should not unsubscribe with nil handler", func(t *testing.T) {
		bus := NewInMemoryEventBus()

		err := bus.Unsubscribe("test.event", nil)
		if err == nil {
			t.Error("expected error for nil handler")
		}
	})

	t.Run("should handle unsubscribe for non-existent event type", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		handler := func(ctx context.Context, event DomainEvent) error { return nil }

		err := bus.Unsubscribe("nonexistent.event", handler)
		if err == nil {
			t.Error("expected error for non-existent event type")
		}
	})

	t.Run("should handle unsubscribe for non-existent handler", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		handler1 := func(ctx context.Context, event DomainEvent) error { return nil }
		handler2 := func(ctx context.Context, event DomainEvent) error { return nil }

		bus.Subscribe("test.event", handler1)

		err := bus.Unsubscribe("test.event", handler2)
		if err == nil {
			t.Error("expected error for non-existent handler")
		}
	})
}
