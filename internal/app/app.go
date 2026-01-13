package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres"
	"github.com/PabloPavan/jaiu/internal/http/handlers"
	"github.com/PabloPavan/jaiu/internal/http/router"
	"github.com/PabloPavan/jaiu/internal/view"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Addr        string
	DatabaseURL string
}

type App struct {
	Router http.Handler
	DB     *pgxpool.Pool
}

func New(cfg Config) (*App, error) {
	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		pool, err = postgres.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("init db: %w", err)
		}
	}

	renderer, err := view.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("init renderer: %w", err)
	}

	h := handlers.New(renderer)

	return &App{
		Router: router.New(h),
		DB:     pool,
	}, nil
}

func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
