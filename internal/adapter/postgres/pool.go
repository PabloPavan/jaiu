package postgres

import (
	"context"
	"fmt"

	"github.com/PabloPavan/jaiu/internal/observability"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, fmt.Errorf("database url is required")
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.ConnConfig.Tracer = observability.NewPgxTracer()
	cfg.AfterConnect = composeAfterConnect(cfg.AfterConnect, registerEnumTypes)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func composeAfterConnect(existing, next func(context.Context, *pgx.Conn) error) func(context.Context, *pgx.Conn) error {
	if existing == nil {
		return next
	}
	return func(ctx context.Context, conn *pgx.Conn) error {
		if err := existing(ctx, conn); err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

func registerEnumTypes(ctx context.Context, conn *pgx.Conn) error {
	typeNames := []string{
		"student_status",
		"_student_status",
		"subscription_status",
		"_subscription_status",
		"payment_status",
		"_payment_status",
		"payment_method",
		"_payment_method",
		"payment_kind",
		"_payment_kind",
		"billing_period_status",
		"_billing_period_status",
		"user_role",
		"_user_role",
	}

	for _, typeName := range typeNames {
		dataType, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return fmt.Errorf("load type %s: %w", typeName, err)
		}
		conn.TypeMap().RegisterType(dataType)
	}

	return nil
}
