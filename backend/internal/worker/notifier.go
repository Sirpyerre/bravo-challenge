package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	ws "github.com/Sirpyerre/bravo-challenge/internal/websocket"
	"github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	"github.com/rs/zerolog"
)

const notifierName = "notifier"

type Notifier struct {
	consumer      *rabbitmq.Consumer
	processedRepo repository.ProcessedEventRepository
	hub           *ws.Hub
	logger        zerolog.Logger
}

func NewNotifier(
	consumer *rabbitmq.Consumer,
	processedRepo repository.ProcessedEventRepository,
	hub *ws.Hub,
	logger zerolog.Logger,
) *Notifier {
	return &Notifier{
		consumer:      consumer,
		processedRepo: processedRepo,
		hub:           hub,
		logger:        logger.With().Str("worker", notifierName).Logger(),
	}
}

// Start escucha eventos de actualización para enviar notificaciones.
func (w *Notifier) Start(ctx context.Context) error {
	return w.consumer.Consume("notifications", "application.updated", func(body []byte) error {
		return w.handle(ctx, body)
	})
}

func (w *Notifier) handle(ctx context.Context, body []byte) error {
	var event domain.Event
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	processed, err := w.processedRepo.IsProcessed(ctx, event.ID, notifierName)
	if err != nil {
		return err
	}
	if processed {
		return nil
	}

	w.logger.Info().
		Str("application_id", event.ApplicationID.String()).
		Str("event_type", event.Type).
		Interface("data", event.Data).
		Msg("notificación enviada (placeholder)")

	// Broadcast en tiempo real al usuario via WebSocket
	if payload, err := json.Marshal(event); err == nil {
		w.hub.BroadcastToUser(event.UserID.String(), payload)
	}

	return w.processedRepo.MarkProcessed(ctx, event.ID, event.Type, notifierName)
}
