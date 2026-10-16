// Package job provides background asynchronous workers for processing queued notification tasks.
package job

import (
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/log"

	"context"
	"fmt"
	"time"
)

// Worker continuously polls a JobQueue and dispatches messages to registered handlers.
type Worker struct {
	queue      JobQueue
	dispatcher *MessageDispatcher
	logger     log.Logger
	clientCfg  config.ClientConfig
}

// NewNotifierWorker constructs a new background worker instance.
func NewNotifierWorker(
	queue JobQueue,
	dispatcher *MessageDispatcher,
	logger log.Logger,
	clientCfg config.ClientConfig,
) *Worker {
	return &Worker{
		queue:      queue,
		dispatcher: dispatcher,
		logger:     logger,
		clientCfg:  clientCfg,
	}
}

// Start initiates the worker event loop. It dequeues jobs, executes their handlers, and handles retries.
func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("context done, stopping worker")
			return
		default:
		}

		msg, err := w.queue.Dequeue(ctx)
		if err != nil {
			time.Sleep(500 * time.Millisecond) // basic backoff
			continue
		}

		// Queue was empty, sleep briefly before next polling iteration
		if msg == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Check message status
		if msg.Status != MessageStatusPending {
			continue
		}

		w.logger.Info(
			"dequeued message:",
			log.Meta{
				"id":          msg.ID,
				"type":        msg.Type,
				"status":      msg.Status,
				"attempts":    msg.Attempts,
				"enqueued_at": msg.EnqueuedAt,
			},
		)

		err = w.dispatcher.Dispatch(ctx, string(msg.Type), msg)
		if err != nil {
			msg.Attempts++
			if msg.Attempts >= 3 {
				// mark as failed and skip re-enqueue
				msg.Status = MessageStatusFailure
				w.logger.Error("message reached max attempts, discarding", log.Meta{"id": msg.ID, "error": err})
				continue
			}

			// Exponential backoff before re-enqueuing
			time.Sleep(time.Duration(msg.Attempts*200) * time.Millisecond)
			err = w.queue.Enqueue(ctx, msg)
			if err != nil {
				w.logger.Error("failed to re-enqueue message", log.Meta{"error": err})
			}
			continue
		}

		// Message processed successfully; do not re-enqueue completed message
		msg.Status = MessageStatusSuccess
		w.logger.Info("message processed successfully", log.Meta{"id": msg.ID})
	}
}

