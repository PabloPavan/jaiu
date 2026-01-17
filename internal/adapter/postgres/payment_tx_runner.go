package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxPaymentTxAttempts = 3

type PaymentTxRunner struct {
	pool *pgxpool.Pool
}

func NewPaymentTxRunner(pool *pgxpool.Pool) *PaymentTxRunner {
	return &PaymentTxRunner{pool: pool}
}

func (r *PaymentTxRunner) RunSerializable(ctx context.Context, fn func(context.Context, ports.PaymentDependencies) error) error {
	if r == nil || r.pool == nil {
		return errors.New("transaction pool unavailable")
	}

	for attempt := 0; attempt < maxPaymentTxAttempts; attempt++ {
		tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			return err
		}

		queries := sqlc.New(tx)
		deps := ports.PaymentDependencies{
			Payments:       NewPaymentRepositoryWithQueries(queries),
			Subscriptions:  NewSubscriptionRepositoryWithQueries(queries),
			Plans:          NewPlanRepositoryWithQueries(queries),
			BillingPeriods: NewBillingPeriodRepositoryWithQueries(queries),
			Balances:       NewSubscriptionBalanceRepositoryWithQueries(queries),
			Allocations:    NewPaymentAllocationRepositoryWithQueries(queries),
			Audit:          NewAuditRepositoryWithTx(tx),
		}

		err = fn(ctx, deps)
		if err != nil {
			_ = tx.Rollback(ctx)
			if isSerializationFailure(err) && attempt < maxPaymentTxAttempts-1 {
				backoff(attempt)
				continue
			}
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			if isSerializationFailure(err) && attempt < maxPaymentTxAttempts-1 {
				backoff(attempt)
				continue
			}
			return err
		}

		return nil
	}

	return errors.New("payment transaction retry limit reached")
}

func isSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40001"
	}
	return false
}

func backoff(attempt int) {
	base := 25 * time.Millisecond
	time.Sleep(base * time.Duration(attempt+1))
}
