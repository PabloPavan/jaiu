package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PabloPavan/jaiu/imagekit/queue"
)

const (
	defaultTable       = "imagekit_outbox"
	defaultMaxAttempts = 10
	defaultLockTimeout = 2 * time.Minute
)

type SQLStore struct {
	DB          *sql.DB
	Table       string
	MaxAttempts int
	LockTimeout time.Duration
}

func (s *SQLStore) Insert(ctx context.Context, tx Tx, msg queue.Message) error {
	if tx == nil {
		return errors.New("outbox tx is required")
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`INSERT INTO %s (payload, available_at) VALUES ($1, now())`, s.table())
	_, err = tx.ExecContext(ctx, query, payload)
	return err
}

func (s *SQLStore) Claim(ctx context.Context, limit int) ([]Record, error) {
	if s.DB == nil {
		return nil, errors.New("outbox db is required")
	}
	if limit <= 0 {
		limit = 10
	}
	maxAttempts := s.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	lockTimeout := s.LockTimeout
	if lockTimeout <= 0 {
		lockTimeout = defaultLockTimeout
	}

	query := s.claimQuery(lockTimeout)
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, query, maxAttempts, limit)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	var records []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.Payload, &rec.Attempts); err != nil {
			rows.Close()
			_ = tx.Rollback()
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		_ = tx.Rollback()
		return nil, err
	}
	rows.Close()

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *SQLStore) Delete(ctx context.Context, id int64) error {
	if s.DB == nil {
		return errors.New("outbox db is required")
	}
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, s.table())
	_, err := s.DB.ExecContext(ctx, query, id)
	return err
}

func (s *SQLStore) Reschedule(ctx context.Context, id int64, next time.Time, lastErr string) error {
	if s.DB == nil {
		return errors.New("outbox db is required")
	}
	query := fmt.Sprintf(`UPDATE %s SET available_at = $2, locked_at = NULL, last_error = $3 WHERE id = $1`, s.table())
	_, err := s.DB.ExecContext(ctx, query, id, next, lastErr)
	return err
}

func (s *SQLStore) table() string {
	if s.Table == "" {
		return defaultTable
	}
	return s.Table
}

func (s *SQLStore) claimQuery(lockTimeout time.Duration) string {
	seconds := int(lockTimeout.Seconds())
	if seconds <= 0 {
		seconds = int(defaultLockTimeout.Seconds())
	}

	return fmt.Sprintf(`WITH candidates AS (
	SELECT id
	FROM %s
	WHERE available_at <= now()
		AND (locked_at IS NULL OR locked_at < now() - interval '%d seconds')
		AND attempts < $1
	ORDER BY id
	FOR UPDATE SKIP LOCKED
	LIMIT $2
	)
	UPDATE %s AS o
	SET locked_at = now(),
		attempts = o.attempts + 1
	FROM candidates
	WHERE o.id = candidates.id
	RETURNING o.id, o.payload, o.attempts`, s.table(), seconds, s.table())
}
