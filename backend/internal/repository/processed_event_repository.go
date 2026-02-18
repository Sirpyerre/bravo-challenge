package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessedEventRepository interface {
	IsProcessed(ctx context.Context, eventID, workerName string) (bool, error)
	MarkProcessed(ctx context.Context, eventID, eventType, workerName string) error
}

type processedEventRepository struct {
	db *pgxpool.Pool
}

func NewProcessedEventRepository(db *pgxpool.Pool) ProcessedEventRepository {
	return &processedEventRepository{db: db}
}

func (r *processedEventRepository) IsProcessed(ctx context.Context, eventID, workerName string) (bool, error) {
	query := `SELECT event_id FROM processed_events WHERE event_id = $1 AND worker_name = $2`

	var id string
	err := r.db.QueryRow(ctx, query, eventID, workerName).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check processed event: %w", err)
	}
	return true, nil
}

func (r *processedEventRepository) MarkProcessed(ctx context.Context, eventID, eventType, workerName string) error {
	query := `
		INSERT INTO processed_events (event_id, event_type, worker_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id) DO NOTHING`

	_, err := r.db.Exec(ctx, query, eventID, eventType, workerName)
	if err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}
	return nil
}
