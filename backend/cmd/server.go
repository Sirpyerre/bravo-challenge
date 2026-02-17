package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Sirpyerre/bravo-challenge/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type server struct {
	echo *echo.Echo
	cfg  *config.Config
}

func newServer(cfg *config.Config) *server {
	e := echo.New()
	e.HideBanner = true

	registerMiddleware(e, cfg)
	registerRoutes(e)

	return &server{echo: e, cfg: cfg}
}

func registerMiddleware(e *echo.Echo, cfg *config.Config) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization", "Idempotency-Key"},
	}))
}

func (s *server) start() {
	addr := ":" + s.cfg.Port
	fmt.Printf("Server starting on %s\n", addr)
	if err := s.echo.Start(addr); err != nil {
		log.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
