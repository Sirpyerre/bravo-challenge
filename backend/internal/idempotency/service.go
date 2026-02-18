package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	"github.com/redis/go-redis/v9"
)

// cachedResponse es la estructura almacenada en Redis y DB.
type cachedResponse struct {
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
}

type Service struct {
	redis *redis.Client
	repo  repository.IdempotencyRepository
	ttl   time.Duration
}

func NewService(redisClient *redis.Client, repo repository.IdempotencyRepository, ttl time.Duration) *Service {
	return &Service{
		redis: redisClient,
		repo:  repo,
		ttl:   ttl,
	}
}

// Check verifica si ya existe una respuesta para la clave dada.
// Primero consulta Redis (~1ms), luego DB como fallback.
// Retorna (statusCode, body, found, error).
func (s *Service) Check(ctx context.Context, key string) (int, []byte, bool, error) {
	// 1. Verificar Redis
	val, err := s.redis.Get(ctx, s.redisKey(key)).Bytes()
	if err == nil {
		var cached cachedResponse
		if err := json.Unmarshal(val, &cached); err == nil {
			return cached.StatusCode, cached.Body, true, nil
		}
	}

	// 2. Fallback a DB
	entry, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return 0, nil, false, fmt.Errorf("check idempotency DB: %w", err)
	}
	if entry != nil {
		// Re-cachear en Redis para futuras consultas
		s.cacheInRedis(ctx, key, entry.ResponseStatusCode, entry.ResponseBody)
		return entry.ResponseStatusCode, entry.ResponseBody, true, nil
	}

	return 0, nil, false, nil
}

// Store almacena la respuesta en Redis y DB.
func (s *Service) Store(ctx context.Context, key string, statusCode int, body []byte) error {
	// Guardar en Redis
	s.cacheInRedis(ctx, key, statusCode, body)

	// Guardar en DB
	entry := &repository.IdempotencyEntry{
		Key:                key,
		ResponseStatusCode: statusCode,
		ResponseBody:       body,
		ExpiresAt:          time.Now().Add(s.ttl),
	}
	if err := s.repo.Save(ctx, entry); err != nil {
		return fmt.Errorf("store idempotency: %w", err)
	}

	return nil
}

func (s *Service) cacheInRedis(ctx context.Context, key string, statusCode int, body []byte) {
	cached := cachedResponse{StatusCode: statusCode, Body: body}
	data, err := json.Marshal(cached)
	if err != nil {
		return
	}
	s.redis.Set(ctx, s.redisKey(key), data, s.ttl)
}

func (s *Service) redisKey(key string) string {
	return "idempotency:" + key
}
