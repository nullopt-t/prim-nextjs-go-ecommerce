package job

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const (
	EmailQueue = "email"
)

type JobQueue interface {
	Enqueue(ctx context.Context, msg *JobMessage) error
	Dequeue(ctx context.Context) (*JobMessage, error)
}

type redisJobQueue struct {
	client *redis.Client
	name   string
}

func NewJobQueue(client *redis.Client, name string) *redisJobQueue {
	return &redisJobQueue{client: client, name: name}
}

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

func (q *redisJobQueue) Dequeue(ctx context.Context) (*JobMessage, error) {
	var msg JobMessage
	res, err := q.client.LPop(ctx, q.name).Result()
	if err != nil {
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
