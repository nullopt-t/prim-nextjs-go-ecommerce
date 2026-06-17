package eventbus

import "context"

type Event struct {
	Name string
	Data any
}

type Handler func(Event) error

type Bus interface {
	Publish(context.Context, Event) error
	Subscribe(string, Handler) error
}

type InMemoryBus struct {
	handlers map[string][]Handler
}

func (b *InMemoryBus) Publish(ctx context.Context, event Event) error {
	handlers, ok := b.handlers[event.Name]
	if !ok {
		return nil
	}
	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return err
		}
	}
	return nil
}

func (b *InMemoryBus) Subscribe(event string, handler Handler) error {
	if b.handlers == nil {
		b.handlers = make(map[string][]Handler)
	}
	b.handlers[event] = append(b.handlers[event], handler)
	return nil
}
