export async function GET() {
  const stream = new ReadableStream({
    start(controller) {
      const enc = new TextEncoder()
      controller.enqueue(enc.encode('event: ping\ndata: {"ok":true}\n\n'))
      const timer = setInterval(() => {
        controller.enqueue(enc.encode('event: ping\ndata: {"ok":true}\n\n'))
      }, 5000)
      setTimeout(() => { clearInterval(timer); controller.close() }, 55000)
    },
  })
  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    },
  })
}
