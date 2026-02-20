package webhook

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	"github.com/Sirpyerre/bravo-challenge/pkg/logger"
	"github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BankCallbackRequest es el payload que el banco envía al webhook.
type BankCallbackRequest struct {
	ApplicationID string `json:"application_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Approved      bool   `json:"approved"        example:"true"`
	RiskLevel     string `json:"risk_level"      example:"LOW"`
	Reason        string `json:"reason"          example:"Perfil crediticio aprobado"`
}

// Handler recibe callbacks entrantes del banco (webhook inbound).
type Handler struct {
	appRepo       repository.ApplicationRepository
	publisher     *rabbitmq.Publisher
	webhookSecret string
}

func NewHandler(
	appRepo repository.ApplicationRepository,
	publisher *rabbitmq.Publisher,
	webhookSecret string,
) *Handler {
	return &Handler{
		appRepo:       appRepo,
		publisher:     publisher,
		webhookSecret: webhookSecret,
	}
}

// BankCallback godoc
// @Summary     Webhook: callback del banco
// @Description Endpoint que recibe la decisión de crédito enviada por el banco de forma asíncrona.
// @Description Requiere la cabecera X-Webhook-Secret con el secreto compartido.
// @Tags        webhooks
// @Accept      json
// @Produce     json
// @Param       X-Webhook-Secret  header    string               true  "Secreto compartido"
// @Param       body              body      BankCallbackRequest  true  "Decisión del banco"
// @Success     200               {object}  map[string]string
// @Failure     400               {object}  map[string]string
// @Failure     401               {object}  map[string]string
// @Failure     404               {object}  map[string]string
// @Router      /webhooks/bank-callback [post]
func (h *Handler) BankCallback(c echo.Context) error {
	log := logger.FromContext(c)

	// Autenticación por secreto compartido
	if secret := c.Request().Header.Get("X-Webhook-Secret"); secret != h.webhookSecret {
		log.Warn().Msg("webhook: invalid secret")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid webhook secret"})
	}

	var req BankCallbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}

	appID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid application_id"})
	}

	ctx := c.Request().Context()

	app, err := h.appRepo.FindByID(ctx, appID)
	if err != nil || app == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "application not found"})
	}

	status := domain.StatusApproved
	if !req.Approved {
		status = domain.StatusDenied
	}

	riskLevel := parseRiskLevel(req.RiskLevel)

	if err := h.appRepo.UpdateRiskAndStatus(ctx, appID, status, riskLevel, req.Reason); err != nil {
		if errors.Is(err, repository.ErrAlreadyTerminal) {
			log.Warn().
				Str("application_id", req.ApplicationID).
				Msg("webhook: application already in terminal state")
			return c.JSON(http.StatusConflict, map[string]string{"error": "application already processed"})
		}
		log.Error().Err(err).Str("application_id", req.ApplicationID).Msg("webhook: failed to update application")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update application"})
	}

	log.Info().
		Str("application_id", req.ApplicationID).
		Str("status", string(status)).
		Str("risk_level", string(riskLevel)).
		Msg("webhook: bank callback processed")

	// Publicar evento para que el notifier envíe actualización por WebSocket
	event := domain.Event{
		ID:            uuid.New().String(),
		Type:          "application.updated",
		ApplicationID: app.ID,
		UserID:        app.UserID,
		Data: map[string]any{
			"status":     status,
			"risk_level": riskLevel,
			"reason":     req.Reason,
			"source":     "bank_webhook",
		},
	}
	data, _ := json.Marshal(event)
	h.publisher.Publish("application.updated", data)

	return c.JSON(http.StatusOK, map[string]string{"status": "received"})
}

func parseRiskLevel(level string) domain.RiskLevel {
	switch level {
	case "LOW":
		return domain.RiskLow
	case "HIGH":
		return domain.RiskHigh
	default:
		return domain.RiskMedium
	}
}
