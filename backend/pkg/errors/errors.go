package errors

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewNotFound(resource, id string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s no encontrado", resource),
		Detail:  fmt.Sprintf("id: %s", id),
	}
}

func NewValidation(detail string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: "error de validación",
		Detail:  detail,
	}
}

func NewConflict(detail string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: "conflicto",
		Detail:  detail,
	}
}

func NewInternal(detail string) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "error interno del servidor",
		Detail:  detail,
	}
}

// HandleError envía una respuesta JSON de error al cliente.
func HandleError(c echo.Context, err error) error {
	if appErr, ok := err.(*AppError); ok {
		return c.JSON(appErr.Code, appErr)
	}
	return c.JSON(http.StatusInternalServerError, &AppError{
		Message: "error interno del servidor",
	})
}
