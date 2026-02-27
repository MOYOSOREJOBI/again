export type QueuePatch = { type: "upsert" | "delete"; incident: any };

function sub(path: string, onPatch: (p: any) => void, onError?: (err: any) => void) {
  const es = new EventSource(path, { withCredentials: true });
  es.onmessage = (evt) => {
    try { onPatch(JSON.parse((evt as MessageEvent).data)); } catch (err) { onError?.(err); }
  };
  es.onerror = (err) => onError?.(err);
  return () => es.close();
}

export function subscribeQueuePatches(onPatch: (patch: QueuePatch) => void, onError?: (err: any) => void) {
  return sub("/api/sse/alerts", onPatch, onError);
}
export function subscribeCommandCenterPatches(onPatch: (patch: any) => void, onError?: (err: any) => void) {
  return sub("/api/sse/alerts", onPatch, onError);
}
export function subscribeTrustPatches(onPatch: (patch: any) => void, onError?: (err: any) => void) {
  return sub("/api/sse/alerts", onPatch, onError);
}
