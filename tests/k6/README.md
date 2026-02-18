# k6 Integration Tests

## Requisitos

```bash
brew install k6          # macOS
# o: https://k6.io/docs/get-started/installation/
```

El servidor debe estar corriendo en `localhost:8000` (o sobreescribir con `BASE_URL`).

## Tests disponibles

| Archivo | Descripción |
|---------|-------------|
| `auth.js` | Register, login, token validation |
| `applications.js` | CRUD completo + idempotency + worker status update |
| `load_test.js` | Prueba de carga con múltiples VUs concurrentes |

## Cómo correr

```bash
# Smoke test de auth
k6 run tests/k6/auth.js

# CRUD de aplicaciones (incluye espera al worker: ~3s)
k6 run tests/k6/applications.js

# Load test (ramp-up 5→10 VUs, 40s total)
k6 run tests/k6/load_test.js

# Load test personalizado
k6 run --vus 20 --duration 30s tests/k6/load_test.js

# Contra otro ambiente
BASE_URL=http://staging.example.com k6 run tests/k6/applications.js
```

## Resultado esperado

```
✓ register status 201
✓ login status 200
✓ create status 201
✓ idempotency replay returns same id
...
checks.........................: 100.00%
http_req_duration..............: p(95)=120ms
```
