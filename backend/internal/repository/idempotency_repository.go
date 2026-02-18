package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyEntry struct {
	Key                string    `json:"idempotency_key"`
	ResponseStatusCode int       `json:"response_status_code"`
	ResponseBody       []byte    `json:"response_body"`
	CreatedAt          time.Time `json:"created_at"`
	ExpiresAt          time.Time `json:"expires_at"`
}

type IdempotencyRepository interface {
	Save(ctx context.Context, entry *IdempotencyEntry) error
	FindByKey(ctx context.Context, key string) (*IdempotencyEntry, error)
}

type idempotencyRepository struct {
	db *pgxpool.Pool
}

func NewIdempotencyRepository(db *pgxpool.Pool) IdempotencyRepository {
	return &idempotencyRepository{db: db}
}

func (r *idempotencyRepository) Save(ctx context.Context, entry *IdempotencyEntry) error {
	query := `
		INSERT INTO idempotency_keys (idempotency_key, response_status_code, response_body, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (idempotency_key) DO NOTHING`

	_, err := r.db.Exec(ctx, query, entry.Key, entry.ResponseStatusCode, entry.ResponseBody, entry.ExpiresAt)
	if err != nil {
		return fmt.Errorf("save idempotency key: %w", err)
	}
	return nil
}

func (r *idempotencyRepository) FindByKey(ctx context.Context, key string) (*IdempotencyEntry, error) {
	query := `
		SELECT idempotency_key, response_status_code, response_body, created_at, expires_at
		FROM idempotency_keys
		WHERE idempotency_key = $1 AND expires_at > NOW()`

	var entry IdempotencyEntry
	err := r.db.QueryRow(ctx, query, key).Scan(
		&entry.Key, &entry.ResponseStatusCode, &entry.ResponseBody,
		&entry.CreatedAt, &entry.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find idempotency key: %w", err)
	}
	return &entry, nil
}
