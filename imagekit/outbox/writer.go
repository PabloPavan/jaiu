package outbox

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/imagekit/queue"
)

type Writer struct {
	Store Store
}

func (w *Writer) Enqueue(ctx context.Context, msg queue.Message) error {
	if w.Store == nil {
		return errors.New("outbox store is required")
	}
	tx, ok := TxFromContext(ctx)
	if !ok {
		return errors.New("outbox tx missing from context")
	}
	return w.Store.Insert(ctx, tx, msg)
}
