import { useEffect, useRef, useState } from "react";
import { baseURL } from "../api/client";

interface EventMessage {
  type?: string;
  application_id?: string;
  status?: string;
  [key: string]: any;
}

export function useRealtime(token: string | null, onEvent?: (msg: EventMessage) => void) {
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!token) {
      setConnected(false);
      wsRef.current?.close();
      return;
    }
    const wsUrl = baseURL.replace(/^http/, "ws") + `/ws?token=${token}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    ws.onerror = () => setConnected(false);
    ws.onmessage = (evt) => {
      try {
        const msg = JSON.parse(evt.data);
        onEvent?.(msg);
      } catch (err) {
        console.error("WS parse error", err);
      }
    };

    return () => {
      ws.close();
    };
  }, [token, onEvent]);

  return connected;
}
