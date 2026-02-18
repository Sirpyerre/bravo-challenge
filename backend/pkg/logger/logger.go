package logger

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// New crea un zerolog.Logger configurado según el nivel y formato indicados.
// level: "debug", "info", "warn", "error" (default: "info")
// format: "console" para salida legible, cualquier otro valor usa JSON.
func New(level, format string) zerolog.Logger {
	var logger zerolog.Logger

	if format == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		logger = zerolog.New(os.Stdout)
	}

	logger = logger.With().
		Timestamp().
		Str("service", "bravo-api").
		Logger().
		Level(parseLevel(level))

	return logger
}

// FromContext extrae el zerolog.Logger almacenado en el contexto de Echo.
func FromContext(c echo.Context) zerolog.Logger {
	if l, ok := c.Get("logger").(zerolog.Logger); ok {
		return l
	}
	return New("info", "json")
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
