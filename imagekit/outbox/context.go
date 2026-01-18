package outbox

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

type txKey struct{}

type Tx interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

func ContextWithTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func TxFromContext(ctx context.Context) (Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(Tx)
	return tx, ok
}
