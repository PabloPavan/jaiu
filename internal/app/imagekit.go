package app

import (
	"context"

	"github.com/PabloPavan/jaiu/imagekit/queue"
)

type noopEnqueuer struct{}

func (noopEnqueuer) Enqueue(ctx context.Context, msg queue.Message) error {
	_ = ctx
	_ = msg
	return nil
}
