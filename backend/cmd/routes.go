package main

import (
	"github.com/Sirpyerre/bravo-challenge/internal/application/healthcheck"
	"github.com/Sirpyerre/bravo-challenge/internal/application/auth"
	appmiddleware "github.com/Sirpyerre/bravo-challenge/internal/middleware"
	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo, depChecker *healthcheck.DependencyChecker, authHandler *auth.Handler, authService *service.AuthService) {
	// Health probes
	e.GET("/health", healthcheck.HealthHandler)
	e.GET("/health_dependencies", depChecker.HealthDependenciesHandler)

	// Auth (público)
	authGroup := e.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	// API v1 (protegido con JWT)
	api := e.Group("/api/v1", appmiddleware.JWTAuth(authService))
	registerV1Routes(api)
}

func registerV1Routes(api *echo.Group) {
	// TODO: application CRUD handlers
}
