package job

import (
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"

	"context"
	"fmt"
	"time"
)

type Worker struct {
	queue      JobQueue
	dispatcher *MessageDispatcher
	logger     log.Logger
	clientCfg  *config.ClientConfig
}

func NewNotifierWorker(
	queue JobQueue,
	dispatcher *MessageDispatcher,
	logger log.Logger,
	clientCfg *config.ClientConfig,
) *Worker {
	return &Worker{
		queue:      queue,
		dispatcher: dispatcher,
		logger:     logger,
		clientCfg:  clientCfg,
	}
}

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

		if msg.Status != "pending" {
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

		err = w.dispatcher.Dispatch(string(msg.Type), msg)
		if err != nil {
			msg.Attempts++
			if msg.Attempts >= 3 {
				// mark as failed and skip re-enqueue
				continue
			}

			err = w.queue.Enqueue(ctx, msg)
			if err != nil {
				w.logger.Error("failed to re-enqueue message", log.Meta{"error": err})
			}
			continue
		}
		msg.Status = "completed"
		err = w.queue.Enqueue(ctx, msg)
		if err != nil {
			w.logger.Error("failed to re-enqueue message", log.Meta{"error": err})
		}
	}
}
