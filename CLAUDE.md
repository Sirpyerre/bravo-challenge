# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Bravo Challenge is a multi-country credit application system. It processes loan requests across 6 countries (Mexico and Brazil active; Spain, Portugal, Italy, Colombia prepared) with country-specific validation rules and bank provider integrations. The backend is in Go, using Echo framework, PostgreSQL, Redis, and RabbitMQ.

## Build & Run Commands

```bash
# Start infrastructure (Postgres, Redis, RabbitMQ, WireMock)
docker-compose up -d

# Run the backend (from repo root; binary is in backend/cmd/)
cd backend && go run ./cmd/

# Run tests
go test ./internal/... -v          # unit tests
go test ./tests/integration/... -v # integration tests
go test -race ./...                # race condition detection
go test -cover ./...               # coverage

# Lint
go fmt ./...
go vet ./...
```

The backend serves on `:8080` (configurable via PORT env var). Config is loaded via `github.com/sethvargo/go-envconfig` with env vars (see `.env.example`).

## Architecture

**Monorepo layout:** `backend/` (Go API + workers), `frontend/` (React, not yet implemented), `wiremock/` (bank mocks).

**Backend structure** (`backend/`):
- `cmd/` — Entry point (`main.go`), server setup (`server.go`), route registration (`routes.go`)
- `internal/config/` — Config structs with env tags, loaded via `envconfig.Process()`
- `internal/api/handler/` — HTTP handlers (auth, applications, health)
- `internal/api/middleware/` — JWT, idempotency, logging
- `internal/domain/` — Domain models (Application, User, Events)
- `internal/service/` — Business logic layer
- `internal/repository/` — Data access (PostgreSQL)
- `internal/bank/` — Bank provider adapters per country (Strategy + Adapter pattern)
- `internal/validation/` — Country-specific validation rules (Strategy pattern)
- `internal/worker/` — Async event consumers (risk evaluation, audit, notifications)
- `internal/cache/` — Redis client
- `internal/queue/` — RabbitMQ client
- `internal/websocket/` — WebSocket server for real-time updates

**Key design patterns:**
- **Strategy + Factory** for country-specific logic: `bank/factory.go` and `validation/factory.go` return the correct implementation based on country code. Adding a new country means creating a new adapter/validator struct and adding a case to the factory switch.
- **Repository pattern** decouples data access from services.
- **Idempotency middleware** checks Redis first (~1ms), falls back to DB. 24h TTL. Requires `Idempotency-Key` header on POST/PUT.
- **Async processing**: API publishes events to RabbitMQ; workers (risk eval, audit, notification) consume in parallel via goroutines.

## Conventions

- **Language**: Code in Go, documentation in Spanish
- **Git commits**: conventional commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`)
- **Git branches**: `main`, `develop`, `feature/xyz`, `bugfix/xyz`
- **Config**: All config via environment variables with `env:` struct tags (see `internal/config/config.go`)
- **DB env vars** use `POSTGRES_` prefix (e.g., `POSTGRES_SERVER`, `POSTGRES_DATABASE`)

## Infrastructure Services (docker-compose)

| Service    | Port  | Purpose                    |
|------------|-------|----------------------------|
| PostgreSQL | 5432  | Primary database           |
| Redis      | 6379  | Idempotency cache          |
| RabbitMQ   | 5672  | Message queue              |
| RabbitMQ   | 15672 | Management UI (guest/guest)|
| WireMock   | 8080  | Bank API simulators        |