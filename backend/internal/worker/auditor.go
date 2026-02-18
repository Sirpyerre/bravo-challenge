package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	"github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	"github.com/rs/zerolog"
)

const auditorName = "auditor"

type Auditor struct {
	consumer      *rabbitmq.Consumer
	auditRepo     repository.AuditRepository
	processedRepo repository.ProcessedEventRepository
	logger        zerolog.Logger
}

func NewAuditor(
	consumer *rabbitmq.Consumer,
	auditRepo repository.AuditRepository,
	processedRepo repository.ProcessedEventRepository,
	logger zerolog.Logger,
) *Auditor {
	return &Auditor{
		consumer:      consumer,
		auditRepo:     auditRepo,
		processedRepo: processedRepo,
		logger:        logger.With().Str("worker", auditorName).Logger(),
	}
}

// Start escucha todos los eventos de aplicación para registrar auditoría.
func (w *Auditor) Start(ctx context.Context) error {
	return w.consumer.Consume("audit.logs", "application.*", func(body []byte) error {
		return w.handle(ctx, body)
	})
}

func (w *Auditor) handle(ctx context.Context, body []byte) error {
	var event domain.Event
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	processed, err := w.processedRepo.IsProcessed(ctx, event.ID, auditorName)
	if err != nil {
		return err
	}
	if processed {
		return nil
	}

	entry := &domain.AuditLog{
		ApplicationID: event.ApplicationID,
		EventType:     event.Type,
		Details:       event.Data,
	}

	if err := w.auditRepo.Save(ctx, entry); err != nil {
		return fmt.Errorf("save audit log: %w", err)
	}

	w.logger.Info().
		Str("application_id", event.ApplicationID.String()).
		Str("event_type", event.Type).
		Msg("audit log saved")

	return w.processedRepo.MarkProcessed(ctx, event.ID, event.Type, auditorName)
}
