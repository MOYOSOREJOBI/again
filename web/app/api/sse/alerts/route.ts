export async function GET() {
  const stream = new ReadableStream({
    start(controller) {
      const enc = new TextEncoder()
      let seq = 0
      const emit = () => {
        seq += 1
        controller.enqueue(enc.encode(`data: ${JSON.stringify({ type: 'activity', seq, message: `activity-${seq}` })}\n\n`))
      }
      emit()
      const timer = setInterval(emit, 3000)
      setTimeout(() => {
        clearInterval(timer)
        controller.close()
      }, 55000)
    },
  })
  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache, no-transform',
      Connection: 'keep-alive',
    },
  })
}
