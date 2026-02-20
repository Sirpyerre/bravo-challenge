package auth

import (
	"net/http"

	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	authService *service.AuthService
}

func NewHandler(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Register(c echo.Context) error {
	var req service.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "request inválido"})
	}

	if req.Email == "" || req.Password == "" || req.Country == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "email, password y country son requeridos"})
	}

	resp, err := h.authService.Register(c.Request().Context(), req)
	if err != nil {
		if err.Error() == "email already registered" {
			return c.JSON(http.StatusConflict, map[string]string{"message": "email ya registrado"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "error interno").SetInternal(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Login(c echo.Context) error {
	var req service.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "request inválido"})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "email y password son requeridos"})
	}

	resp, err := h.authService.Login(c.Request().Context(), req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "credenciales inválidas"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "error interno").SetInternal(err)
	}

	return c.JSON(http.StatusOK, resp)
}
