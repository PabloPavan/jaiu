package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/imagekit/queue"
)

const (
	defaultBatchSize    = 25
	defaultPollInterval = 2 * time.Second
	defaultRetryDelay   = 10 * time.Second
)

// Dispatcher publishes outbox entries into the queue with retries.
type Dispatcher struct {
	Store        Store
	Queue        queue.Queue
	BatchSize    int
	PollInterval time.Duration
	RetryDelay   time.Duration
}

// Run continuously publishes outbox entries until the context is canceled.
func (d *Dispatcher) Run(ctx context.Context) error {
	if d.Store == nil {
		return errors.New("outbox store is required")
	}
	if d.Queue == nil {
		return errors.New("queue is required")
	}

	batchSize := d.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	pollInterval := d.PollInterval
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}
	retryDelay := d.RetryDelay
	if retryDelay <= 0 {
		retryDelay = defaultRetryDelay
	}

	for {
		records, err := d.Store.Claim(ctx, batchSize)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return ctx.Err()
			}
			return err
		}

		if len(records) == 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(pollInterval):
				continue
			}
		}

		for _, record := range records {
			var msg queue.Message
			if err := json.Unmarshal(record.Payload, &msg); err != nil {
				if err := d.Store.Reschedule(ctx, record.ID, time.Now().Add(retryDelay), "invalid payload: "+err.Error()); err != nil {
					return err
				}
				continue
			}

			if err := d.Queue.Enqueue(ctx, msg); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return ctx.Err()
				}
				if err := d.Store.Reschedule(ctx, record.ID, time.Now().Add(retryDelay), err.Error()); err != nil {
					return err
				}
				continue
			}

			if err := d.Store.Delete(ctx, record.ID); err != nil {
				return err
			}
		}
	}
}
