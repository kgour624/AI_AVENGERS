import { useEffect, useRef, useState } from 'react';
import { useAuthStore } from '@/stores/authStore';

export function useReceptionistSSE(sessionId: string | null) {
  const [data, setData] = useState<any>(null);
  const [connected, setConnected] = useState(false);
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!sessionId) return;
    let token = '';
    try { token = useAuthStore.getState().accessToken || ''; } catch {}
    if (!token) token = localStorage.getItem('token') || localStorage.getItem('access_token') || localStorage.getItem('accessToken') || localStorage.getItem('auth_token') || '';
    const base = `/api/receptionist/sessions/${sessionId}/events`;
    const url = token ? `${base}?token=${encodeURIComponent(token)}` : base;
    const es = new EventSource(url, { withCredentials: true } as any);
    esRef.current = es;
    let retryTimer: any = null;
    let closed = false;
    es.addEventListener('snapshot', (e: MessageEvent) => {
      try { setData(JSON.parse(e.data)); } catch {}
    });
    es.onopen = () => setConnected(true);
    es.onerror = () => {
      setConnected(false);
      // backoff retry na karo tightly - bar bar reconnect = api call increase
      if (!closed) {
        es.close();
        clearTimeout(retryTimer);
        retryTimer = setTimeout(() => {
          if (!closed) setData((d:any)=>d); // trigger re-effect via sessionId stable, outer will reconnect on next render if needed
        }, 5000);
      }
    };
    return () => { closed = true; clearTimeout(retryTimer); es.close(); setConnected(false); };
  }, [sessionId]);

  return { data, connected };
}
