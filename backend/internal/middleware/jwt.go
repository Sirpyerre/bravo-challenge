package middleware

import (
	"net/http"
	"strings"

	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/labstack/echo/v4"
)

func JWTAuth(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "token requerido"})
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "formato de token inválido"})
			}

			claims, err := authService.ValidateToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "token inválido o expirado"})
			}

			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("country", claims.Country)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}
