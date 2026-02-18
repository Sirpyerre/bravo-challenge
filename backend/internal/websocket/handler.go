package websocket

import (
	"net/http"

	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Permitir cualquier origen (ajustar en producción via CheckOrigin)
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	hub         *Hub
	authService *service.AuthService
}

func NewHandler(hub *Hub, authService *service.AuthService) *Handler {
	return &Handler{hub: hub, authService: authService}
}

// Connect maneja GET /ws?token=<jwt>
func (h *Handler) Connect(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "token requerido")
	}

	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "token inválido")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	userID := claims.UserID.String()
	client := NewClient(h.hub, userID, conn)
	h.hub.Register(userID, client)

	go client.WritePump()
	go client.ReadPump()

	return nil
}
