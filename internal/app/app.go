package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PabloPavan/jaiu/imagekit"
	kitconfig "github.com/PabloPavan/jaiu/imagekit/config"
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
	ImageUploadDir    string
	SessionCookieName string
	SessionTTL        time.Duration
	SessionSecure     bool
}

type App struct {
	Router   http.Handler
	DB       *pgxpool.Pool
	Redis    *redis.Client
	ImageKit *imagekit.Kit
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
	var paymentService handlers.PaymentService
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

	imageCfg := kitconfig.Config{
		StorageType: kitconfig.StorageLocal,
		LocalDir:    cfg.ImageUploadDir,
		QueueName:   "imagekit:queue",
		OriginalKey: "original.jpg",
		Sizes: []kitconfig.ImageSize{
			{Name: "preview", Width: 640, Height: 640},
			{Name: "list", Width: 96, Height: 96},
		},
		Redis: kitconfig.RedisConfig{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
	}
	if imageCfg.LocalDir == "" {
		imageCfg.LocalDir = "tmp/uploads/images"
	}

	imageCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imageKit, err := imagekit.New(imageCtx, imageCfg)
	if err != nil {
		return nil, fmt.Errorf("init imagekit: %w", err)
	}

	if pool != nil {
		imageKit.EnableOutbox(pool)
		userRepo := postgres.NewUserRepository(pool)
		auditRepo := postgres.NewAuditRepository(pool)
		authService = service.NewAuthService(userRepo, auditRepo)

		planRepo := postgres.NewPlanRepository(pool)
		studentRepo := postgres.NewStudentRepository(pool)
		subscriptionRepo := postgres.NewSubscriptionRepository(pool)
		planService = service.NewPlanService(planRepo, subscriptionRepo, auditRepo)
		studentService = service.NewStudentService(studentRepo, subscriptionRepo, auditRepo)
		subscriptionService = service.NewSubscriptionService(subscriptionRepo, planRepo, studentRepo, auditRepo)

		paymentRepo := postgres.NewPaymentRepository(pool)
		periodRepo := postgres.NewBillingPeriodRepository(pool)
		balanceRepo := postgres.NewSubscriptionBalanceRepository(pool)
		allocationRepo := postgres.NewPaymentAllocationRepository(pool)
		paymentTx := postgres.NewPaymentTxRunner(pool)
		paymentService = service.NewPaymentService(paymentRepo, subscriptionRepo, planRepo, periodRepo, balanceRepo, allocationRepo, auditRepo, paymentTx)
	}

	if redisClient != nil {
		sessionStore = redisadapter.NewSessionStore(redisClient)
	}

	h := handlers.New(handlers.Services{
		Auth:          authService,
		Plans:         planService,
		Students:      studentService,
		Subscriptions: subscriptionService,
		Payments:      paymentService,
	}, sessionStore, sessionConfig)
	h.SetImageConfig(handlers.ImageConfig{
		ImageService: imageKit,
		BaseURL:      "/images",
		OriginalKey:  imageCfg.OriginalKey,
	})

	return &App{
		Router:   router.New(h, sessionStore, sessionConfig.CookieName),
		DB:       pool,
		Redis:    redisClient,
		ImageKit: imageKit,
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
