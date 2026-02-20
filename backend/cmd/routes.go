package main

import (
	"github.com/Sirpyerre/bravo-challenge/internal/application/auth"
	"github.com/Sirpyerre/bravo-challenge/internal/application/credit"
	"github.com/Sirpyerre/bravo-challenge/internal/application/healthcheck"
	"github.com/Sirpyerre/bravo-challenge/internal/application/webhook"
	"github.com/Sirpyerre/bravo-challenge/internal/idempotency"
	appmiddleware "github.com/Sirpyerre/bravo-challenge/internal/middleware"
	"github.com/Sirpyerre/bravo-challenge/internal/service"
	appwebsocket "github.com/Sirpyerre/bravo-challenge/internal/websocket"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/labstack/echo/v4"

	_ "github.com/Sirpyerre/bravo-challenge/docs"
)

func registerRoutes(e *echo.Echo, depChecker *healthcheck.DependencyChecker, authHandler *auth.Handler, creditHandler *credit.Handler, wsHandler *appwebsocket.Handler, webhookHandler *webhook.Handler, authService *service.AuthService, idempotencySvc *idempotency.Service) {
	// Health probes
	e.GET("/health", healthcheck.HealthHandler)
	e.GET("/health_dependencies", depChecker.HealthDependenciesHandler)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Auth (público)
	authGroup := e.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	// WebSocket (token via query param)
	e.GET("/ws", wsHandler.Connect)

	// Webhooks entrantes (sin JWT, autenticados por X-Webhook-Secret)
	webhooks := e.Group("/webhooks")
	webhooks.POST("/bank-callback", webhookHandler.BankCallback)

	// API v1 (protegido con JWT)
	api := e.Group("/api/v1", appmiddleware.JWTAuth(authService))
	registerV1Routes(api, creditHandler, idempotencySvc)
}

func registerV1Routes(api *echo.Group, creditHandler *credit.Handler, idempotencySvc *idempotency.Service) {
	apps := api.Group("/applications")
	apps.POST("", creditHandler.Create, appmiddleware.Idempotency(idempotencySvc))
	apps.GET("", creditHandler.List)
	apps.GET("/:id", creditHandler.GetByID)
	apps.PUT("/:id", creditHandler.UpdateStatus, appmiddleware.Idempotency(idempotencySvc))
}
