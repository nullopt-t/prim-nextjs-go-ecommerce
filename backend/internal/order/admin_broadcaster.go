package order

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/redis/go-redis/v9"
)

const (
	// AdminEventsChannel is the Redis Pub/Sub topic used to broadcast live administrative updates.
	AdminEventsChannel = "admin:events:live"
	// Deprecated: AdminOrderEventsChannel alias for backward compatibility.
	AdminOrderEventsChannel = AdminEventsChannel
)

// AdminOrderEvent alias pointing to the canonical notifier.AdminOrderEvent.
type AdminOrderEvent = notifier.AdminOrderEvent

// AdminStockEvent alias pointing to the canonical notifier.AdminStockEvent.
type AdminStockEvent = notifier.AdminStockEvent

// AdminRealtimeMessage alias pointing to the canonical notifier.AdminRealtimeMessage.
type AdminRealtimeMessage = notifier.AdminRealtimeMessage

// Broadcaster defines the contract for real-time administrative event distribution.
type Broadcaster interface {
	Publish(ctx context.Context, ev AdminOrderEvent) error
	PublishStock(ctx context.Context, ev AdminStockEvent) error
	Subscribe() chan AdminRealtimeMessage
	Unsubscribe(ch chan AdminRealtimeMessage)
	Close()
}

// AdminBroadcaster coordinates real-time server-sent events for administrative operators.
type AdminBroadcaster struct {
	redisClient *redis.Client
	logger      log.Logger
	mu          sync.RWMutex
	clients     map[chan AdminRealtimeMessage]struct{}
	stopCh      chan struct{}
}

// NewAdminBroadcaster initializes and starts listening on Redis Pub/Sub if available.
func NewAdminBroadcaster(redisClient *redis.Client, logger log.Logger) *AdminBroadcaster {
	if logger == nil {
		panic("order: logger is required for admin broadcaster")
	}

	b := &AdminBroadcaster{
		redisClient: redisClient,
		logger:      logger,
		clients:     make(map[chan AdminRealtimeMessage]struct{}),
		stopCh:      make(chan struct{}),
	}

	if redisClient != nil {
		go b.listenRedisPubSub()
	}

	return b
}

func (b *AdminBroadcaster) listenRedisPubSub() {
	pubsub := b.redisClient.Subscribe(context.Background(), AdminEventsChannel)
	defer func() {
		_ = pubsub.Close()
	}()

	ch := pubsub.Channel()
	for {
		select {
		case <-b.stopCh:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var ev AdminRealtimeMessage
			if err := json.Unmarshal([]byte(msg.Payload), &ev); err != nil {
				b.logger.Warn("admin broadcaster: failed to decode pubsub message", log.Meta{"error": err})
				continue
			}
			b.dispatchLocal(ev)
		}
	}
}

// Publish broadcasts an order event to Redis pubsub (or dispatches locally if Redis is nil).
func (b *AdminBroadcaster) Publish(ctx context.Context, ev AdminOrderEvent) error {
	msg := AdminRealtimeMessage{
		Type:  "order.created",
		Order: &ev,
	}
	return b.publishMessage(ctx, msg)
}

// PublishStock broadcasts a stock change / low-stock warning event to Redis pubsub.
func (b *AdminBroadcaster) PublishStock(ctx context.Context, ev AdminStockEvent) error {
	msg := AdminRealtimeMessage{
		Type:  "stock.updated",
		Stock: &ev,
	}
	return b.publishMessage(ctx, msg)
}

func (b *AdminBroadcaster) publishMessage(ctx context.Context, msg AdminRealtimeMessage) error {
	b.logger.Info("admin broadcaster: publishing message", log.Meta{"type": msg.Type})
	if b.redisClient != nil {
		payload, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal admin realtime event: %w", err)
		}
		if err := b.redisClient.Publish(ctx, AdminEventsChannel, payload).Err(); err != nil {
			b.logger.Warn("admin broadcaster: redis publish failed, falling back to local dispatch", log.Meta{"error": err})
			b.dispatchLocal(msg)
			return err
		}
		return nil
	}

	b.dispatchLocal(msg)
	return nil
}

// Subscribe registers a new subscriber channel.
func (b *AdminBroadcaster) Subscribe() chan AdminRealtimeMessage {
	ch := make(chan AdminRealtimeMessage, 32)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[ch] = struct{}{}
	return ch
}

// Unsubscribe unregisters and closes a subscriber channel.
func (b *AdminBroadcaster) Unsubscribe(ch chan AdminRealtimeMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
}

func (b *AdminBroadcaster) dispatchLocal(ev AdminRealtimeMessage) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- ev:
		default:
			// Non-blocking drop if channel buffer full
		}
	}
}

// Close gracefully shuts down the broadcaster.
func (b *AdminBroadcaster) Close() {
	close(b.stopCh)
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		delete(b.clients, ch)
		close(ch)
	}
}
