package main

import (
	"github.com/Sirpyerre/bravo-challenge/internal/application"
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo, depChecker *application.DependencyChecker) {
	e.GET("/health", application.HealthHandler)
	e.GET("/health_dependencies", depChecker.HealthDependenciesHandler)

	api := e.Group("/api")
	registerV1Routes(api)
}

func registerV1Routes(api *echo.Group) {
	// TODO: register versioned routes once handlers are implemented
	// auth.SetupRoutes(api)
	// applications.SetupRoutes(api)
}
