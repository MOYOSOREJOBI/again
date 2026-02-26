export type QueuePatch = { type: "upsert" | "delete"; incident: any };

function sub(path: string, event: string, onPatch: (p: any) => void, onError?: (err: any) => void) {
  const es = new EventSource(path, { withCredentials: true });
  es.addEventListener(event, (evt) => {
    try { onPatch(JSON.parse((evt as MessageEvent).data)); } catch (err) { onError?.(err); }
  });
  es.onerror = (err) => onError?.(err);
  return () => es.close();
}

export function subscribeQueuePatches(onPatch: (patch: QueuePatch) => void, onError?: (err: any) => void) {
  return sub("/api/stream/queue", "queue_patch", onPatch, onError);
}
export function subscribeCommandCenterPatches(onPatch: (patch: any) => void, onError?: (err: any) => void) {
  return sub("/api/stream/command-center", "command_center_patch", onPatch, onError);
}
export function subscribeTrustPatches(onPatch: (patch: any) => void, onError?: (err: any) => void) {
  return sub("/api/stream/trust", "trust_patch", onPatch, onError);
}
