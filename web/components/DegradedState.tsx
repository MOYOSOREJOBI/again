export default function DegradedState({ message = 'Service degraded; showing best available data.' }: { message?: string }) {
  return <div className="empty-state">⚠️ {message}</div>
}
