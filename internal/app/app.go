package app

import (
	"fmt"
	"net/http"

	"github.com/PabloPavan/jaiu/internal/http/handlers"
	"github.com/PabloPavan/jaiu/internal/http/router"
	"github.com/PabloPavan/jaiu/internal/view"
)

type Config struct {
	Addr string
}

type App struct {
	Router http.Handler
}

func New(cfg Config) (*App, error) {
	renderer, err := view.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("init renderer: %w", err)
	}

	h := handlers.New(renderer)

	return &App{
		Router: router.New(h),
	}, nil
}
