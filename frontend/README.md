# Frontend Bravo Dashboard (Vite + React + Tailwind)

Dashboard con rutas protegidas, Auth persistente (JWT), listado de solicitudes, creación y actualización de estado, y eventos en tiempo real vía WebSocket.

## Setup

```bash
cd frontend
npm install

# variables
echo "VITE_API_BASE_URL=http://localhost:8000" > .env.development

# desarrollo
npm run dev -- --host
```

Abre `http://localhost:5173` (o el puerto que indique Vite). El backend debe estar corriendo en el puerto definido por `VITE_API_BASE_URL`.

## Características
- Login/registro y sesión persistente en localStorage.
- Axios con interceptor `Authorization: Bearer <token>` y auto-logout en 401.
- Layout de dashboard con sidebar, header e indicador de realtime.
- Listado con filtros (país, estado, from_date, to_date, límite, offset) usando React Query.
- Creación con validación básica y Idempotency-Key automática.
- Drawer de detalle con actualización de estado.
- WebSocket para refrescar y resaltar filas; toasts en cambios.
