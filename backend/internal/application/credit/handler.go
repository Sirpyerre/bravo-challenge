package credit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
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

// Create crea una nueva solicitud de crédito.
//
//	@Summary		Crear solicitud de crédito
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Idempotency-Key	header		string								true	"Clave de idempotencia (UUID)"
//	@Param			body			body		service.CreateApplicationRequest	true	"Datos de la solicitud"
//	@Success		201				{object}	domain.Application
//	@Failure		400				{object}	map[string]string
//	@Failure		401				{object}	map[string]string
//	@Router			/api/v1/applications [post]
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

// List lista solicitudes de crédito. AGENT y ADMIN ven todas; USER solo las propias.
//
//	@Summary		Listar solicitudes
//	@Tags			applications
//	@Produce		json
//	@Security		BearerAuth
//	@Param			country		query		string	false	"Filtro por país (MX, BR, ES…)"
//	@Param			status		query		string	false	"Filtro por estado (PENDING, APPROVED, DENIED…)"
//	@Param			from_date	query		string	false	"Fecha inicio YYYY-MM-DD"
//	@Param			to_date		query		string	false	"Fecha fin YYYY-MM-DD"
//	@Param			limit		query		int		false	"Límite (default 20, max 100)"
//	@Param			offset		query		int		false	"Offset"
//	@Success		200			{object}	service.ListApplicationsResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Router			/api/v1/applications [get]
func (h *Handler) List(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "usuario no autenticado"})
	}
	role, _ := c.Get("role").(domain.Role)

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	var fromDate *time.Time
	if v := c.QueryParam("from_date"); v != "" {
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "from_date debe tener formato YYYY-MM-DD"})
		}
		fromDate = &parsed
	}

	var toDate *time.Time
	if v := c.QueryParam("to_date"); v != "" {
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"message": "to_date debe tener formato YYYY-MM-DD"})
		}
		toDate = &parsed
	}

	req := service.ListApplicationsRequest{
		Country:  c.QueryParam("country"),
		Status:   c.QueryParam("status"),
		FromDate: fromDate,
		ToDate:   toDate,
		Limit:    limit,
		Offset:   offset,
	}

	resp, err := h.appService.List(c.Request().Context(), userID, role, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetByID obtiene una solicitud por ID. AGENT y ADMIN pueden ver cualquiera; USER solo la propia.
//
//	@Summary		Obtener solicitud por ID
//	@Tags			applications
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"UUID de la solicitud"
//	@Success		200	{object}	domain.Application
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/api/v1/applications/{id} [get]
func (h *Handler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "id inválido"})
	}
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "usuario no autenticado"})
	}
	role, _ := c.Get("role").(domain.Role)

	app, err := h.appService.GetByID(c.Request().Context(), userID, role, id)
	if err != nil {
		if err.Error() == "solicitud no encontrada" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
		}
		if err.Error() == "acceso no permitido" {
			return c.JSON(http.StatusForbidden, map[string]string{"message": err.Error()})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "error interno").SetInternal(err)
	}

	return c.JSON(http.StatusOK, app)
}

// UpdateStatus actualiza el estado de una solicitud. Requiere rol AGENT o ADMIN.
//
//	@Summary		Actualizar estado de solicitud
//	@Tags			applications
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Idempotency-Key	header		string								true	"Clave de idempotencia (UUID)"
//	@Param			id				path		string								true	"UUID de la solicitud"
//	@Param			body			body		service.UpdateApplicationRequest	true	"Nuevo estado"
//	@Success		200				{object}	map[string]string
//	@Failure		400				{object}	map[string]string
//	@Failure		401				{object}	map[string]string
//	@Failure		403				{object}	map[string]string
//	@Router			/api/v1/applications/{id} [put]
func (h *Handler) UpdateStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "id inválido"})
	}
	role, _ := c.Get("role").(domain.Role)
	if !role.IsPrivileged() {
		return c.JSON(http.StatusForbidden, map[string]string{"message": "acceso no permitido"})
	}

	var req service.UpdateApplicationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "request inválido"})
	}

	if err := h.appService.UpdateStatus(c.Request().Context(), role, id, req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "solicitud actualizada"})
}
