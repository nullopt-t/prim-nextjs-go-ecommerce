// Package job routes job messages to designated handler callbacks based on routing keys.
package job

import (
	"context"
	"fmt"
	"sync"
)

// Handler defines a callback function that processes a single background JobMessage with context support.
type Handler func(ctx context.Context, msg *JobMessage) error

// MessageDispatcher holds routing mappings between job types and their execution handlers,
// safely protected for concurrent registration and dispatching.
type MessageDispatcher struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

// NewMessageDispatcher instantiates an empty MessageDispatcher registry.
func NewMessageDispatcher() *MessageDispatcher {
	return &MessageDispatcher{
		handlers: make(map[string]Handler),
	}
}

// Register binds a message type key to a specific handler function thread-safely.
func (d *MessageDispatcher) Register(key string, handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[key] = handler
}

// Dispatch executes the handler registered for the given message key thread-safely.
// Returns an error if no handler has been registered for the key.
func (d *MessageDispatcher) Dispatch(ctx context.Context, key string, msg *JobMessage) error {
	d.mu.RLock()
	handler, ok := d.handlers[key]
	d.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no handler for message :%v", key)
	}

	return handler(ctx, msg)
}

