package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	ApplicationID uuid.UUID `json:"application_id"`
	UserID        uuid.UUID `json:"user_id"`
	Data          any       `json:"data,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type AuditLog struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ApplicationID uuid.UUID `json:"application_id" db:"application_id"`
	EventType     string    `json:"event_type" db:"event_type"`
	Details       any       `json:"details" db:"details"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
