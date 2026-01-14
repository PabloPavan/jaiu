package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres"
	redisadapter "github.com/PabloPavan/jaiu/internal/adapter/redis"
	"github.com/PabloPavan/jaiu/internal/http/handlers"
	"github.com/PabloPavan/jaiu/internal/http/router"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/PabloPavan/jaiu/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	redis "github.com/redis/go-redis/v9"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	SessionCookieName string
	SessionTTL        time.Duration
	SessionSecure     bool
}

type App struct {
	Router http.Handler
	DB     *pgxpool.Pool
	Redis  *redis.Client
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

	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("init redis: %w", err)
		}
	}

	var authService handlers.AuthService
	var planService handlers.PlanService
	var studentService handlers.StudentService
	var subscriptionService handlers.SubscriptionService
	var sessionStore ports.SessionStore
	sessionConfig := handlers.SessionConfig{
		CookieName: cfg.SessionCookieName,
		TTL:        cfg.SessionTTL,
		Secure:     cfg.SessionSecure,
	}
	if sessionConfig.CookieName == "" {
		sessionConfig.CookieName = "jaiu_session"
	}
	if sessionConfig.TTL == 0 {
		sessionConfig.TTL = 24 * time.Hour
	}
	if sessionConfig.SameSite == 0 {
		sessionConfig.SameSite = http.SameSiteLaxMode
	}

	if pool != nil {
		userRepo := postgres.NewUserRepository(pool)
		authService = service.NewAuthService(userRepo)

		planRepo := postgres.NewPlanRepository(pool)
		planService = service.NewPlanService(planRepo)

		studentRepo := postgres.NewStudentRepository(pool)
		studentService = service.NewStudentService(studentRepo)

		subscriptionRepo := postgres.NewSubscriptionRepository(pool)
		subscriptionService = service.NewSubscriptionService(subscriptionRepo, planRepo, studentRepo)
	}

	if redisClient != nil {
		sessionStore = redisadapter.NewSessionStore(redisClient)
	}

	h := handlers.New(handlers.Services{
		Auth:          authService,
		Plans:         planService,
		Students:      studentService,
		Subscriptions: subscriptionService,
	}, sessionStore, sessionConfig)

	return &App{
		Router: router.New(h, sessionStore, sessionConfig.CookieName),
		DB:     pool,
		Redis:  redisClient,
	}, nil
}

func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
}
