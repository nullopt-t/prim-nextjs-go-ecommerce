package eventbus

import (
	"context"
	"testing"
)

func TestEventBus(t *testing.T) {
	bus := &InMemoryBus{}
	bus.Subscribe("event1", func(event Event) error {
		if event.Name != "event1" {
			t.Errorf("Expected event name to be 'event1', got '%s'", event.Name)
		}
		return nil
	})

	bus.Subscribe("event1", func(event Event) error {
		if event.Name != "event1" {
			t.Errorf("Expected event name to be 'event1', got '%s'", event.Name)
		}
		return nil
	})

	event := Event{Name: "event1", Data: nil}
	if err := bus.Publish(context.Background(), event); err != nil {
		t.Errorf("Expected Publish to return nil, got %v", err)
	}
}
