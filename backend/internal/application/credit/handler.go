package credit

import (
	"net/http"
	"strconv"

	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	appService *service.ApplicationService
}

func NewHandler(appService *service.ApplicationService) *Handler {
	return &Handler{appService: appService}
}

func (h *Handler) Create(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "usuario no autenticado"})
	}

	var req service.CreateApplicationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "request inválido"})
	}

	app, err := h.appService.Create(c.Request().Context(), userID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusCreated, app)
}

func (h *Handler) List(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "usuario no autenticado"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	req := service.ListApplicationsRequest{
		Country: c.QueryParam("country"),
		Status:  c.QueryParam("status"),
		Limit:   limit,
		Offset:  offset,
	}

	resp, err := h.appService.List(c.Request().Context(), userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "id inválido"})
	}

	app, err := h.appService.GetByID(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "solicitud no encontrada" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, app)
}

func (h *Handler) UpdateStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "id inválido"})
	}

	var req service.UpdateApplicationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "request inválido"})
	}

	if err := h.appService.UpdateStatus(c.Request().Context(), id, req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "solicitud actualizada"})
}
