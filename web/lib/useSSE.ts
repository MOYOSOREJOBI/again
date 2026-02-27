"use client";
import { useEffect, useRef, useState } from "react";

type SSEStatus = "connecting" | "open" | "closed" | "error";
export function useSSE<T>(url: string, onMessage: (data: T) => void) {
  const [status, setStatus] = useState<SSEStatus>("connecting");
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    let closed = false;
    let retryMs = 500;

    const connect = () => {
      if (closed) return;
      setStatus("connecting");
      const es = new EventSource(url, { withCredentials: true });
      esRef.current = es;

      es.onopen = () => { retryMs = 500; setStatus("open"); };
      es.onerror = () => {
        setStatus("error");
        es.close();
        if (closed) return;
        setTimeout(connect, retryMs);
        retryMs = Math.min(retryMs * 2, 8000);
      };
      es.onmessage = (evt) => {
        try { onMessage(JSON.parse(evt.data)); } catch { }
      };
    };

    connect();
    return () => { closed = true; esRef.current?.close(); setStatus("closed"); };
  }, [url, onMessage]);

  return { status };
}
