package events

import (
	"context"
	"testing"
)

func TestEventBusPublisher(t *testing.T) {
	t.Run("should create new event bus publisher", func(t *testing.T) {
		eventBus := NewInMemoryEventBus()
		publisher := NewEventBusPublisher(eventBus)

		if publisher == nil {
			t.Error("expected publisher to be created")
		}
		if publisher.eventBus != eventBus {
			t.Error("expected event bus to be set")
		}
	})

	t.Run("should publish event through event bus", func(t *testing.T) {
		eventBus := NewInMemoryEventBus()
		publisher := NewEventBusPublisher(eventBus)

		receivedEvent := make(chan DomainEvent, 1)
		handler := func(ctx context.Context, event DomainEvent) error {
			receivedEvent <- event
			return nil
		}

		eventBus.Subscribe("test.event", handler)

		testEvent := NewBaseDomainEvent("test.event", "test-123", 1)
		err := publisher.PublishEvent(context.Background(), testEvent)
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
}
