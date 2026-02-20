package main

import (
	"context"

	"github.com/Sirpyerre/bravo-challenge/internal/application/auth"
	"github.com/Sirpyerre/bravo-challenge/internal/application/credit"
	"github.com/Sirpyerre/bravo-challenge/internal/application/healthcheck"
	"github.com/Sirpyerre/bravo-challenge/internal/config"
	"github.com/Sirpyerre/bravo-challenge/internal/idempotency"
	_ "github.com/Sirpyerre/bravo-challenge/internal/metrics" // registra métricas custom
	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	"github.com/Sirpyerre/bravo-challenge/internal/service"
	appwebsocket "github.com/Sirpyerre/bravo-challenge/internal/websocket"
	"github.com/Sirpyerre/bravo-challenge/internal/worker"
	"github.com/Sirpyerre/bravo-challenge/pkg/database"
	"github.com/Sirpyerre/bravo-challenge/pkg/logger"
	pkgrabbit "github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	pkgredis "github.com/Sirpyerre/bravo-challenge/pkg/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type server struct {
	echo         *echo.Echo
	cfg          *config.Config
	logger       zerolog.Logger
	db           *pgxpool.Pool
	redis        *redis.Client
	rabbitmq     *amqp.Connection
	riskWorker   *worker.RiskEvaluator
	auditWorker  *worker.Auditor
	notifyWorker *worker.Notifier
	wsHub        *appwebsocket.Hub
}

func newServer(cfg *config.Config) *server {
	ctx := context.Background()
	log := logger.New(cfg.LogLevel, cfg.LogFormat)

	// PostgreSQL
	db, err := database.New(ctx, database.Config{
		Server:          cfg.DB.Server,
		Port:            cfg.DB.Port,
		Database:        cfg.DB.Database,
		User:            cfg.DB.User,
		Password:        cfg.DB.Password,
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		ConnectTimeout:  cfg.DB.ConnectTimeOut,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("connected to postgresql")

	// Redis
	rdb, err := pkgredis.New(ctx, pkgredis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	log.Info().Msg("connected to redis")

	// RabbitMQ
	rmq, err := pkgrabbit.New(pkgrabbit.Config{
		Host:     cfg.RabbitMQ.Host,
		Port:     cfg.RabbitMQ.Port,
		User:     cfg.RabbitMQ.User,
		Password: cfg.RabbitMQ.Password,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	log.Info().Msg("connected to rabbitmq")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	appRepo := repository.NewApplicationRepository(db)
	idempotencyRepo := repository.NewIdempotencyRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	processedRepo := repository.NewProcessedEventRepository(db)

	// RabbitMQ publisher/consumers
	publisher := pkgrabbit.NewPublisher(rmq)
	consumer := pkgrabbit.NewConsumer(rmq)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	appService := service.NewApplicationService(appRepo, publisher)
	idempotencySvc := idempotency.NewService(rdb, idempotencyRepo, cfg.IdempotencyTTL)

	// WebSocket hub
	wsHub := appwebsocket.NewHub()

	// Workers
	riskWorker := worker.NewRiskEvaluator(consumer, publisher, appRepo, processedRepo, cfg.BankURLs, log)
	auditWorker := worker.NewAuditor(pkgrabbit.NewConsumer(rmq), auditRepo, processedRepo, log)
	notifyWorker := worker.NewNotifier(pkgrabbit.NewConsumer(rmq), processedRepo, wsHub, log)

	// Handlers
	authHandler := auth.NewHandler(authService)
	creditHandler := credit.NewHandler(appService)
	wsHandler := appwebsocket.NewHandler(wsHub, authService)
	depChecker := &healthcheck.DependencyChecker{
		DB:       db,
		Redis:    rdb,
		RabbitMQ: rmq,
	}

	e := echo.New()
	e.HideBanner = true

	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Namespace: "credit",
		Skipper: func(c echo.Context) bool {
			p := c.Path()
			return p == "/metrics" || p == "/health" || p == "/health_dependencies"
		},
	}))
	e.GET("/metrics", echoprometheus.NewHandler())

	registerMiddleware(e, cfg, log)
	registerRoutes(e, depChecker, authHandler, creditHandler, wsHandler, authService, idempotencySvc)

	return &server{
		echo:         e,
		cfg:          cfg,
		logger:       log,
		db:           db,
		redis:        rdb,
		rabbitmq:     rmq,
		riskWorker:   riskWorker,
		auditWorker:  auditWorker,
		notifyWorker: notifyWorker,
		wsHub:        wsHub,
	}
}

func registerMiddleware(e *echo.Echo, cfg *config.Config, log zerolog.Logger) {
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", log)
			return next(c)
		}
	})

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			p := c.Path()
			if p == "" {
				p = c.Request().URL.Path
			}
			return p == "/metrics" || p == "/health" || p == "/health_dependencies"
		},
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			evt := log.Info()
			if v.Error != nil {
				cause := v.Error
				if he, ok := v.Error.(*echo.HTTPError); ok && he.Internal != nil {
					cause = he.Internal
				}
				evt = log.Error().Err(cause)
			}
			evt.
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Str("remote_ip", v.RemoteIP).
				Msg("request")
			return nil
		},
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization", "Idempotency-Key"},
	}))
}

func (s *server) startWorkers(ctx context.Context) {
	if err := s.riskWorker.Start(ctx); err != nil {
		s.logger.Fatal().Err(err).Msg("failed to start risk evaluator worker")
	}
	if err := s.auditWorker.Start(ctx); err != nil {
		s.logger.Fatal().Err(err).Msg("failed to start auditor worker")
	}
	if err := s.notifyWorker.Start(ctx); err != nil {
		s.logger.Fatal().Err(err).Msg("failed to start notifier worker")
	}
	s.logger.Info().Msg("workers started")
}

func (s *server) start() {
	addr := ":" + s.cfg.Port
	s.logger.Info().Str("addr", addr).Msg("server starting")
	if err := s.echo.Start(addr); err != nil {
		s.logger.Info().Msg("server stopped")
	}
}

func (s *server) close() {
	if s.db != nil {
		s.db.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
	if s.rabbitmq != nil {
		s.rabbitmq.Close()
	}
}
