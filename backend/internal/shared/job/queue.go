package job

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
)

const (
	// EmailQueue is the default Redis list key for transactional email jobs.
	EmailQueue = "email"
)

// JobQueue defines operations for pushing and popping background jobs.
type JobQueue interface {
	Enqueue(ctx context.Context, msg *JobMessage) error
	Dequeue(ctx context.Context) (*JobMessage, error)
}

// redisJobQueue is an implementation of JobQueue backed by Redis lists (LPUSH / LPOP).
type redisJobQueue struct {
	client *redis.Client
	name   string
}

// NewJobQueue initializes a redisJobQueue with a specific Redis list key.
func NewJobQueue(client *redis.Client, name string) *redisJobQueue {
	return &redisJobQueue{client: client, name: name}
}

// Enqueue serializes a JobMessage to JSON and prepends it to the head of the Redis list.
func (q *redisJobQueue) Enqueue(ctx context.Context, msg *JobMessage) error {
	mm, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = q.client.LPush(ctx, q.name, mm).Err()
	if err != nil {
		return err
	}
	return nil
}

// Dequeue pops the next job from the head of the Redis list.
// Returns (nil, nil) when the queue is empty (redis.Nil).
func (q *redisJobQueue) Dequeue(ctx context.Context) (*JobMessage, error) {
	var msg JobMessage
	res, err := q.client.LPop(ctx, q.name).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if res != "" {
		err = json.Unmarshal([]byte(res), &msg)
		if err != nil {
			return nil, err
		}
		return &msg, nil
	}
	return nil, nil
}

