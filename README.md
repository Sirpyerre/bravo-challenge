# Bravo Challenge - Solicitudes de Crédito Multi-País

Sistema de solicitudes de crédito distribuido para Bravo. Actualmente, el procesamiento de pagos está habilitado para México y Brasil, con una arquitectura extensible para habilitar fácilmente otros países (como España, Portugal, Italia y Colombia).

## Tabla de Contenidos

- [Características](#características)
- [Stack Técnico](#stack-técnico)
- [Instalación](#instalación)
- [Configuración](#configuración)
- [Uso](#uso)
- [API Endpoints](#api-endpoints)
- [Arquitectura](#arquitectura)
- [Testing](#testing)
- [Deploy](#deploy)


## ✨ Características

- **Multi-país (extensible)**: Procesamiento activo para México y Brasil, con validaciones y proveedores preparados para habilitar más países
- **Procesamiento Asincrónico**: Workers paralelos vía RabbitMQ
- **Real-time**: WebSocket para actualizaciones en tiempo real
- **Idempotencia**: 24h TTL con Redis + Database hybrid
- **Escalable**: Diseño para millones de solicitudes
- **Observabilidad**: Logs estructurados, triggers de auditoría
- **Seguro**: JWT authentication, constraints en DB

## Stack Técnico

### Backend
- **Go** 1.22+
- **Echo** - Framework HTTP
- **PostgreSQL** - Database principal
- **Redis** - Cache e idempotencia
- **RabbitMQ** - Message queue
- **JWT** - Autenticación

### Frontend
- **React** 18+ (recomendado)
- **WebSocket** (Socket.io)
- **TanStack Query** (React Query)

### DevOps
- **Docker** / **Docker Compose** - Containerización
- **Kubernetes** - Manifests YAML
- **WireMock** - Simulación de bancos


## 🚀 Instalación

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- Make
- Git

### Clonar y Setup

```bash
git clone <repo>
cd bravo-challenge

# Copiar variables de entorno
cp .env.example .env

# Iniciar servicios (PostgreSQL, Redis, RabbitMQ, WireMock)
make docker-up

# Descargar dependencias
go mod download

# Ejecutar migraciones
make migrate

# Iniciar backend
make run
```

**Backend disponible en:** `http://localhost:8000`


## Configuración

### Variables de Entorno (.env)

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=bravo
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0

# RabbitMQ
RABBIT_HOST=localhost
RABBIT_PORT=5672
RABBIT_USER=guest
RABBIT_PASSWORD=guest

# JWT
JWT_SECRET=your-secret-key-min-32-chars-long
JWT_EXPIRATION_HOURS=24

# Bancos Simulados (WireMock)
ESP_BANK_URL=http://localhost:8080/esp
MX_BANK_URL=http://localhost:8080/mx
COL_BANK_URL=http://localhost:8080/col
BR_BANK_URL=http://localhost:8080/br
PT_BANK_URL=http://localhost:8080/pt
IT_BANK_URL=http://localhost:8080/it

# API
PORT=8000
ENVIRONMENT=development
LOG_LEVEL=debug
```

## Uso

### Iniciar Todo

```bash
# Opción 1: Make (recomendado)
make docker-up    # Servicios
make run          # Backend

# Opción 2: Manual
docker-compose up -d
go run ./cmd/main.go
```

### Tests

```bash
make test           # Unit tests
make test-race      # Detectar race conditions
make test-coverage  # Coverage report
```

### Limpiar

```bash
make clean          # Remove binaries
make docker-down    # Stop services
make docker-down-v  # Stop services + remove volumes
```

## API Endpoints

### Authentication
```
POST /auth/register
  - Registrar nuevo usuario
  - Body: { email, password, country }
  - Response: { user_id, token }

POST /auth/login
  - Login
  - Body: { email, password }
  - Response: { token }
```

### Applications (Autenticado)
```
POST /api/v1/applications
  - Crear solicitud de crédito
  - Headers: Authorization: Bearer <token>, Idempotency-Key: <uuid>
  - Body: { country, full_name, identity_document, monthly_income, requested_amount }
  - Response: { id, status, created_at }

GET /api/v1/applications
  - Listar solicitudes del usuario
  - Headers: Authorization: Bearer <token>
  - Query: ?country=ES&status=PENDING&limit=20&offset=0
  - Response: { applications: [...], total }

GET /api/v1/applications/:id
  - Obtener detalles de solicitud
  - Headers: Authorization: Bearer <token>
  - Response: { id, country, status, risk_level, created_at, updated_at, ... }

PUT /api/v1/applications/:id
  - Actualizar estado (admin)
  - Headers: Authorization: Bearer <token>, Idempotency-Key: <uuid>
  - Body: { status, notes }
  - Response: { id, status, updated_at }
```

### Health Check
```
GET /health
  - Check servidor está activo
  - Response: { status: "healthy", uptime: "..." }
```


## Arquitectura

### Componentes

```
┌─────────────────────────┐
│   Frontend (React)      │
│   WebSocket + REST      │
└────────────┬────────────┘
             │
┌────────────▼────────────┐
│   API (Echo + Go)       │
│   ├─ Controllers        │
│   ├─ Services           │
│   ├─ Adapters           │
│   └─ JWT Middleware     │
└────┬──────────┬─────────┘
     │          │
┌────▼─┐    ┌───▼──────┐
│ PG   │    │ RabbitMQ │
└──────┘    └─────┬────┘
                  │
         ┌────────▼────────┐
         │   Workers       │
         ├─ Risk Eval      │
         ├─ Auditor        │
         └─ Notifier       │
```

### Patrones de Diseño

- **Strategy Pattern**: Validaciones por país
- **Adapter Pattern**: Integraciones con bancos
- **Factory Pattern**: Instanciación de adapters
- **Repository Pattern**: Acceso a datos
- **Middleware Pattern**: JWT, idempotencia

## Database

### Tablas Principales

```sql
applications
├─ id (UUID PK)
├─ user_id (FK)
├─ country (VARCHAR)
├─ full_name (VARCHAR)
├─ identity_document (VARCHAR)
├─ monthly_income (DECIMAL)
├─ requested_amount (DECIMAL)
├─ status (VARCHAR) -- PENDING, VALIDATING, APPROVED, DENIED
├─ risk_level (VARCHAR) -- LOW, MEDIUM, HIGH
├─ created_at (TIMESTAMP)
└─ updated_at (TIMESTAMP)

idempotency_keys
├─ idempotency_key (VARCHAR PK)
├─ response_status_code (INT)
├─ response_body (JSONB)
├─ created_at (TIMESTAMP)
└─ expires_at (TIMESTAMP)

processed_events
├─ event_id (VARCHAR PK)
├─ event_type (VARCHAR)
├─ processed_at (TIMESTAMP)
└─ worker_name (VARCHAR)

audit_logs
├─ id (UUID PK)
├─ application_id (FK)
├─ event_type (VARCHAR)
├─ details (JSONB)
└─ created_at (TIMESTAMP)
```

### Migraciones

Ubicadas en `migrations/` (SQL files):
- `001_init.sql` - Schema inicial
- `002_add_indexes.sql` - Índices para performance

## Testing

### Unit Tests
```bash
go test ./internal/... -v
```

### Integration Tests
```bash
go test ./tests/integration/... -v
```

### Coverage
```bash
go test -cover ./...
```

### Race Conditions
```bash
go test -race ./...
```

## Docker Compose

### Servicios

```yaml
- postgres:16 (puerto 5432)
- redis:7 (puerto 6379)
- rabbitmq:3.13-management (5672, 15672 admin)
- wiremock:latest (8080 - mock bancos)
- backend (8000)
```

### Comandos Útiles

```bash
# Ver logs
docker-compose logs -f

# Ingresar a PostgreSQL
docker-compose exec postgres psql -U postgres -d bravo

# Ver RabbitMQ Admin
# http://localhost:15672 (user: guest, pass: guest)

# Recrear todo
docker-compose down -v && docker-compose up -d
```

## Escalabilidad

### Hoy (MVP)
- Índices en (country, status, created_at)
- Redis para cache
- Workers paralelos

### Mañana (Millones de requests)
- Particionamiento PostgreSQL por país
- Sharding horizontal por país
- Distributed locks si necesario

## 🔐 Seguridad

- JWT para autenticación
- Constraints en DB previenen duplicados
- Idempotencia 24h previene duplicados de API
- Triggers PostgreSQL para auditoría
- Logs estructurados
- SSL en CORS configurado


## Convenciones

### Git Branches
```
main              - Production
develop           - Staging
feature/xyz       - Feature branch
bugfix/xyz        - Bug fix
```

### Commits
```
feat: Add JWT authentication
fix: Handle concurrent requests
docs: Update API documentation
test: Add integration tests for risk evaluation
refactor: Simplify bank adapter logic
```

### Code Style
```bash
go fmt ./...
go vet ./...
golangci-lint run ./...  # Si está instalado
```

## Documentación Adicional

- **CLAUDE.md** - Decisiones arquitectónicas detalladas
- **docs/IDEMPOTENCY.md** - Explicación de idempotencia
- **docs/MULTI_PAIS.md** - Cómo agregar nuevo país
- **docs/API.md** - Spec OpenAPI (futuro)

## Contribuir

1. Fork el repo
2. Crear feature branch (`git checkout -b feature/xyz`)
3. Commit cambios (`git commit -am 'feat: xyz'`)
4. Push branch (`git push origin feature/xyz`)
5. Abrir PR


## Autor

Pedro Rojas Reyes - Backend Engineer
