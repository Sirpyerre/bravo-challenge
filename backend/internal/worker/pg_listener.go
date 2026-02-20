package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// pgNotifyPayload es el JSON que envía el trigger de PostgreSQL.
type pgNotifyPayload struct {
	ApplicationID uuid.UUID `json:"application_id"`
	UserID        uuid.UUID `json:"user_id"`
	Country       string    `json:"country"`
	Status        string    `json:"status"`
}

// PgListener escucha notificaciones de PostgreSQL (LISTEN/NOTIFY) y las
// reenvía a RabbitMQ para que los workers existentes las procesen.
// Es el puente entre el trigger nativo de PostgreSQL y el pipeline async.
type PgListener struct {
	db        *pgxpool.Pool
	publisher *rabbitmq.Publisher
	logger    zerolog.Logger
}

func NewPgListener(db *pgxpool.Pool, publisher *rabbitmq.Publisher, logger zerolog.Logger) *PgListener {
	return &PgListener{
		db:        db,
		publisher: publisher,
		logger:    logger.With().Str("worker", "pg_listener").Logger(),
	}
}

// Start adquiere una conexión dedicada, emite LISTEN y bloquea esperando notificaciones.
// Debe llamarse en una goroutine separada.
func (l *PgListener) Start(ctx context.Context) error {
	// Necesitamos una conexión dedicada (no pooled) para LISTEN/NOTIFY.
	conn, err := l.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "LISTEN \"bravo.application_created\""); err != nil {
		return err
	}

	l.logger.Info().Msg("listening on PostgreSQL channel 'bravo.application_created'")

	for {
		// WaitForNotification bloquea hasta recibir una notificación o que el contexto se cancele.
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				l.logger.Info().Msg("pg_listener stopped")
				return nil
			}
			l.logger.Error().Err(err).Msg("pg_listener: wait error, reconnecting in 5s")
			time.Sleep(5 * time.Second)
			return l.Start(ctx) // reconexión simple
		}

		l.handle(notification.Payload)
	}
}

func (l *PgListener) handle(payload string) {
	var p pgNotifyPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		l.logger.Error().Err(err).Str("payload", payload).Msg("pg_listener: failed to parse notification")
		return
	}

	event := domain.Event{
		ID:            uuid.New().String(),
		Type:          "application.created",
		ApplicationID: p.ApplicationID,
		UserID:        p.UserID,
		Data: map[string]any{
			"country": p.Country,
			"status":  p.Status,
			"source":  "pg_notify",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		l.logger.Error().Err(err).Msg("pg_listener: failed to marshal event")
		return
	}

	l.publisher.Publish("application.created", data)

	l.logger.Info().
		Str("application_id", p.ApplicationID.String()).
		Str("country", p.Country).
		Msg("pg_listener: application.created published to RabbitMQ")
}
