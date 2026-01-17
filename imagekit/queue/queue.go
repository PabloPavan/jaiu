package queue

import "context"

type Message struct {
	ObjectKey string `json:"object_key"`
}

type Queue interface {
	Enqueue(ctx context.Context, msg Message) error
	Dequeue(ctx context.Context) (Message, error)
}
