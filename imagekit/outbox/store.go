package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PabloPavan/jaiu/imagekit/queue"
)

type Record struct {
	ID       int64
	Payload  json.RawMessage
	Attempts int
}

type Store interface {
	Insert(ctx context.Context, tx Tx, msg queue.Message) error
	Claim(ctx context.Context, limit int) ([]Record, error)
	Delete(ctx context.Context, id int64) error
	Reschedule(ctx context.Context, id int64, next time.Time, lastErr string) error
}
