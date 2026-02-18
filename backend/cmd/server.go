package main

import (
	"context"
	"os"

	"github.com/Sirpyerre/bravo-challenge/internal/application"
	"github.com/Sirpyerre/bravo-challenge/internal/config"
	"github.com/Sirpyerre/bravo-challenge/pkg/database"
	"github.com/Sirpyerre/bravo-challenge/pkg/logger"
	pkgrabbit "github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	pkgredis "github.com/Sirpyerre/bravo-challenge/pkg/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type server struct {
	echo     *echo.Echo
	cfg      *config.Config
	logger   zerolog.Logger
	db       *pgxpool.Pool
	redis    *redis.Client
	rabbitmq *amqp.Connection
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

	depChecker := &application.DependencyChecker{
		DB:       db,
		Redis:    rdb,
		RabbitMQ: rmq,
	}

	e := echo.New()
	e.HideBanner = true

	registerMiddleware(e, cfg, log)
	registerRoutes(e, depChecker)

	return &server{
		echo:     e,
		cfg:      cfg,
		logger:   log,
		db:       db,
		redis:    rdb,
		rabbitmq: rmq,
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
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			evt := log.Info()
			if v.Error != nil {
				evt = log.Error().Err(v.Error)
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

func (s *server) start() {
	addr := ":" + s.cfg.Port
	s.logger.Info().Str("addr", addr).Msg("server starting")
	if err := s.echo.Start(addr); err != nil {
		s.logger.Fatal().Err(err).Msg("server failed to start")
		os.Exit(1)
	}
}
