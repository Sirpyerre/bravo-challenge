package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sirpyerre/bravo-challenge/internal/bank"
	"github.com/Sirpyerre/bravo-challenge/internal/config"
	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	"github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const riskEvaluatorName = "risk_evaluator"

type RiskEvaluator struct {
	consumer       *rabbitmq.Consumer
	publisher      *rabbitmq.Publisher
	appRepo        repository.ApplicationRepository
	processedRepo  repository.ProcessedEventRepository
	bankURLs       config.BankURLsConfig
	logger         zerolog.Logger
}

func NewRiskEvaluator(
	consumer *rabbitmq.Consumer,
	publisher *rabbitmq.Publisher,
	appRepo repository.ApplicationRepository,
	processedRepo repository.ProcessedEventRepository,
	bankURLs config.BankURLsConfig,
	logger zerolog.Logger,
) *RiskEvaluator {
	return &RiskEvaluator{
		consumer:      consumer,
		publisher:     publisher,
		appRepo:       appRepo,
		processedRepo: processedRepo,
		bankURLs:      bankURLs,
		logger:        logger.With().Str("worker", riskEvaluatorName).Logger(),
	}
}

func (w *RiskEvaluator) Start(ctx context.Context) error {
	return w.consumer.Consume("risk.evaluation", "application.created", func(body []byte) error {
		return w.handle(ctx, body)
	})
}

func (w *RiskEvaluator) handle(ctx context.Context, body []byte) error {
	var event domain.Event
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	// Deduplicación
	processed, err := w.processedRepo.IsProcessed(ctx, event.ID, riskEvaluatorName)
	if err != nil {
		return err
	}
	if processed {
		w.logger.Debug().Str("event_id", event.ID).Msg("event already processed, skipping")
		return nil
	}

	app, err := w.appRepo.FindByID(ctx, event.ApplicationID)
	if err != nil || app == nil {
		return fmt.Errorf("find application %s: %w", event.ApplicationID, err)
	}

	w.logger.Info().
		Str("application_id", app.ID.String()).
		Str("country", app.Country).
		Msg("evaluating risk")

	provider, err := bank.NewProvider(app.Country, w.bankURLs)
	if err != nil {
		return fmt.Errorf("get bank provider: %w", err)
	}

	bankResp, err := provider.Evaluate(ctx, bank.EvaluationRequest{
		ApplicationID:    app.ID.String(),
		IdentityDocument: app.IdentityDocument,
		MonthlyIncome:    app.MonthlyIncome,
		RequestedAmount:  app.RequestedAmount,
	})
	if err != nil {
		w.logger.Error().Err(err).Msg("bank evaluation failed")
		// Marcar como DENIED si el banco no responde
		w.appRepo.UpdateRiskAndStatus(ctx, app.ID, domain.StatusDenied, domain.RiskHigh)
		return err
	}

	status := domain.StatusApproved
	if !bankResp.Approved {
		status = domain.StatusDenied
	}

	riskLevel := parseRiskLevel(bankResp.RiskLevel)
	if err := w.appRepo.UpdateRiskAndStatus(ctx, app.ID, status, riskLevel); err != nil {
		return fmt.Errorf("update application: %w", err)
	}

	w.logger.Info().
		Str("application_id", app.ID.String()).
		Str("status", string(status)).
		Str("risk_level", string(riskLevel)).
		Msg("risk evaluation completed")

	// Publicar evento de actualización
	updatedEvent := domain.Event{
		ID:            uuid.New().String(),
		Type:          "application.updated",
		ApplicationID: app.ID,
		UserID:        app.UserID,
		Data: map[string]any{
			"status":     status,
			"risk_level": riskLevel,
			"reason":     bankResp.Reason,
		},
	}
	data, _ := json.Marshal(updatedEvent)
	w.publisher.Publish("application.updated", data)

	return w.processedRepo.MarkProcessed(ctx, event.ID, event.Type, riskEvaluatorName)
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
