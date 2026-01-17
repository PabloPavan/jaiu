package outbox

import (
	"context"
	"database/sql"
)

type txKey struct{}

type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func ContextWithTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func TxFromContext(ctx context.Context) (Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(Tx)
	return tx, ok
}
