package middleware

import (
	"bytes"
	"net/http"

	"github.com/Sirpyerre/bravo-challenge/internal/idempotency"
	"github.com/labstack/echo/v4"
)

// bodyCapture captura el response body para almacenarlo en idempotencia.
type bodyCapture struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (bc *bodyCapture) Write(b []byte) (int, error) {
	bc.body.Write(b)
	return bc.ResponseWriter.Write(b)
}

func (bc *bodyCapture) WriteHeader(code int) {
	bc.statusCode = code
	bc.ResponseWriter.WriteHeader(code)
}

// Idempotency middleware intercepta POST y PUT que incluyan el header Idempotency-Key.
// Si la clave ya fue procesada, retorna la respuesta cacheada sin ejecutar el handler.
func Idempotency(svc *idempotency.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			if method != http.MethodPost && method != http.MethodPut {
				return next(c)
			}

			key := c.Request().Header.Get("Idempotency-Key")
			if key == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"message": "Idempotency-Key header es requerido para POST/PUT",
				})
			}

			ctx := c.Request().Context()

			// Verificar si ya existe respuesta
			statusCode, body, found, err := svc.Check(ctx, key)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"message": "error verificando idempotencia",
				})
			}
			if found {
				return c.JSONBlob(statusCode, body)
			}

			// Capturar la respuesta del handler
			capture := &bodyCapture{
				ResponseWriter: c.Response().Writer,
				body:           new(bytes.Buffer),
				statusCode:     http.StatusOK,
			}
			c.Response().Writer = capture

			// Ejecutar handler
			if err := next(c); err != nil {
				c.Error(err)
			}

			// Almacenar respuesta para futuras solicitudes con la misma clave
			svc.Store(ctx, key, capture.statusCode, capture.body.Bytes())

			return nil
		}
	}
}
