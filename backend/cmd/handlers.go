package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
