package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository interface {
	Save(ctx context.Context, log *domain.AuditLog) error
}

type auditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Save(ctx context.Context, entry *domain.AuditLog) error {
	details, err := json.Marshal(entry.Details)
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}

	query := `
		INSERT INTO audit_logs (application_id, event_type, details)
		VALUES ($1, $2, $3)`

	_, err = r.db.Exec(ctx, query, entry.ApplicationID, entry.EventType, details)
	if err != nil {
		return fmt.Errorf("save audit log: %w", err)
	}
	return nil
}
