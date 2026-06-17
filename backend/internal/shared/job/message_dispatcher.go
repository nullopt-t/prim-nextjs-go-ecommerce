package job

import (
	"fmt"
)

type Handler func(msg *JobMessage) error

type MessageDispatcher struct {
	handlers map[string]Handler
}

func NewMessageDispatcher() *MessageDispatcher {
	return &MessageDispatcher{
		handlers: make(map[string]Handler),
	}
}

func (d *MessageDispatcher) Register(key string, handler Handler) {
	d.handlers[key] = handler
}

func (d *MessageDispatcher) Dispatch(key string, msg *JobMessage) error {
	handler, ok := d.handlers[key]
	if !ok {
		return fmt.Errorf("no handler for message :%v", key)
	}

	return handler(msg)
}
