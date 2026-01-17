package queue

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client  *redis.Client
	listKey string
}

func NewRedisQueue(client *redis.Client, listKey string) *RedisQueue {
	return &RedisQueue{
		client:  client,
		listKey: listKey,
	}
}

func (q *RedisQueue) Enqueue(ctx context.Context, msg Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, q.listKey, payload).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context) (Message, error) {
	result, err := q.client.BRPop(ctx, 0, q.listKey).Result()
	if err != nil {
		return Message{}, err
	}
	if len(result) != 2 {
		return Message{}, redis.Nil
	}

	var msg Message
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}
