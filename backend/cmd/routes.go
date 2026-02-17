package main

import (
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo) {
	e.GET("/health", healthHandler)

	api := e.Group("/api")
	registerV1Routes(api)
}

func registerV1Routes(api *echo.Group) {
	// TODO: register versioned routes once handlers are implemented
	// auth.SetupRoutes(api)
	// applications.SetupRoutes(api)
}
