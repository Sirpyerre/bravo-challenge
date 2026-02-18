# Frontend Bravo Realtime

Frontend ligero (HTML + JS) para crear, listar, ver detalle y actualizar solicitudes con actualizaciones en tiempo real vía WebSocket.

## Ejecutar

```bash
# desde la raíz del repo
cd frontend/src
python -m http.server 5173
# abre http://localhost:5173
```

- Ajusta `Base URL` en la UI si tu backend no corre en `http://localhost:8080`.
- Regístrate o haz login para obtener el token JWT; el WebSocket se conecta con ese token.
- Usa los filtros `country`, `status`, `from_date`, `to_date`, `limit`, `offset` para listar.

## Funcionalidad
- Registro y login.
- Crear solicitud (usa Idempotency-Key aleatoria).
- Listar con filtros incluyendo rango de fechas.
- Ver detalle y actualizar estado.
- Panel de eventos en tiempo real usando `/ws`.
