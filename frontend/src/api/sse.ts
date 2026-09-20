import type { PipelineEvent } from '../types'

/**
 * The pipeline endpoints (create session, run, evolve) are POST requests
 * that stream text/event-stream frames, so the browser's EventSource (GET
 * only) can't be used — we read the fetch body stream manually instead.
 */
export async function streamPipeline(
  url: string,
  body: unknown,
  onEvent: (event: PipelineEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal,
  })

  if (!res.ok || !res.body) {
    const text = await res.text().catch(() => '')
    throw new Error(text || `request failed with status ${res.status}`)
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })

    let sepIndex: number
    while ((sepIndex = buffer.indexOf('\n\n')) !== -1) {
      const frame = buffer.slice(0, sepIndex)
      buffer = buffer.slice(sepIndex + 2)

      const line = frame.split('\n').find((l) => l.startsWith('data: '))
      if (!line) continue
      const json = line.slice('data: '.length)
      try {
        onEvent(JSON.parse(json) as PipelineEvent)
      } catch {
        // ignore malformed frames rather than crashing the whole stream
      }
    }
  }
}
