import { useEffect, useRef, useState } from "react";
import { baseURL } from "../api/client";

interface EventMessage {
  type?: string;
  application_id?: string;
  status?: string;
  [key: string]: any;
}

const RECONNECT_DELAYS = [1000, 2000, 5000, 10000, 30000]; // ms, capped at 30s

export function useRealtime(token: string | null, onEvent?: (msg: EventMessage) => void) {
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const onEventRef = useRef(onEvent);
  const attemptRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const destroyedRef = useRef(false);

  // Mantener la ref actualizada sin recrear el WebSocket
  useEffect(() => {
    onEventRef.current = onEvent;
  });

  useEffect(() => {
    if (!token) {
      setConnected(false);
      wsRef.current?.close();
      return;
    }

    destroyedRef.current = false;
    attemptRef.current = 0;

    function connect() {
      if (destroyedRef.current) return;

      const wsUrl = baseURL.replace(/^http/, "ws") + `/ws?token=${token}`;
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        attemptRef.current = 0; // reset backoff on successful connection
      };

      ws.onclose = () => {
        setConnected(false);
        if (destroyedRef.current) return;
        // Reconectar con backoff exponencial
        const delay = RECONNECT_DELAYS[Math.min(attemptRef.current, RECONNECT_DELAYS.length - 1)];
        attemptRef.current += 1;
        timerRef.current = setTimeout(connect, delay);
      };

      ws.onerror = () => {
        setConnected(false);
      };

      ws.onmessage = (evt) => {
        try {
          const msg = JSON.parse(evt.data);
          onEventRef.current?.(msg);
        } catch (err) {
          console.error("WS parse error", err);
        }
      };
    }

    connect();

    return () => {
      destroyedRef.current = true;
      if (timerRef.current) clearTimeout(timerRef.current);
      wsRef.current?.close();
    };
  }, [token]); // Solo reconecta cuando cambia el token

  return connected;
}
