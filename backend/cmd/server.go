package main

import (
	"os"

	"github.com/Sirpyerre/bravo-challenge/internal/config"
	"github.com/Sirpyerre/bravo-challenge/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type server struct {
	echo   *echo.Echo
	cfg    *config.Config
	logger zerolog.Logger
}

func newServer(cfg *config.Config) *server {
	log := logger.New(cfg.LogLevel, cfg.LogFormat)

	e := echo.New()
	e.HideBanner = true

	registerMiddleware(e, cfg, log)
	registerRoutes(e)

	return &server{echo: e, cfg: cfg, logger: log}
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
