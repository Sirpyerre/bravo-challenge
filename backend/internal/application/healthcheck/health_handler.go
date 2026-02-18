package healthcheck

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var startTime = time.Now()

// HealthHandler — Liveness probe.
// Responde OK si el proceso está vivo, sin verificar dependencias.
func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
		"uptime": time.Since(startTime).String(),
	})
}

// DependencyChecker contiene las conexiones a verificar en el readiness probe.
type DependencyChecker struct {
	DB       *pgxpool.Pool
	Redis    *redis.Client
	RabbitMQ *amqp.Connection
}

type depStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthDependenciesHandler — Readiness probe.
// Verifica conectividad a PostgreSQL, Redis y RabbitMQ.
func (d *DependencyChecker) HealthDependenciesHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	checks := map[string]depStatus{}
	allHealthy := true

	// PostgreSQL
	if d.DB != nil {
		if err := d.DB.Ping(ctx); err != nil {
			checks["postgresql"] = depStatus{Status: "unhealthy", Error: err.Error()}
			allHealthy = false
		} else {
			checks["postgresql"] = depStatus{Status: "healthy"}
		}
	} else {
		checks["postgresql"] = depStatus{Status: "not_configured"}
		allHealthy = false
	}

	// Redis
	if d.Redis != nil {
		if err := d.Redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = depStatus{Status: "unhealthy", Error: err.Error()}
			allHealthy = false
		} else {
			checks["redis"] = depStatus{Status: "healthy"}
		}
	} else {
		checks["redis"] = depStatus{Status: "not_configured"}
		allHealthy = false
	}

	// RabbitMQ
	if d.RabbitMQ != nil && !d.RabbitMQ.IsClosed() {
		checks["rabbitmq"] = depStatus{Status: "healthy"}
	} else {
		checks["rabbitmq"] = depStatus{Status: "unhealthy", Error: "connection closed or nil"}
		allHealthy = false
	}

	status := http.StatusOK
	overall := "ready"
	if !allHealthy {
		status = http.StatusServiceUnavailable
		overall = "not_ready"
	}

	return c.JSON(status, map[string]any{
		"status":       overall,
		"uptime":       time.Since(startTime).String(),
		"dependencies": checks,
	})
}
