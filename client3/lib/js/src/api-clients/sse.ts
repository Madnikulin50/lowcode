// Reading a Server-Sent Events stream sent by POST.
//
// The browser's EventSource can only GET and cannot carry an Authorization
// header, so streams that need either are read with fetch. The framing is the
// one the server uses everywhere ("data: <json>" blocks separated by a blank
// line); every event is a JSON object, and the stream ends with one whose
// `done` is true.

export type StreamEvent = Record<string, any>

/**
 * Splits streamed text into complete `data:` payloads, keeping an incomplete
 * trailing block for the next chunk. Returns the events found and the rest.
 */
export function parseSSE (buffer: string): { events: StreamEvent[]; rest: string } {
  const blocks = buffer.split('\n\n')
  const rest = blocks.pop() || ''
  const events: StreamEvent[] = []

  for (const block of blocks) {
    for (const line of block.split('\n')) {
      if (!line.startsWith('data: ')) continue
      events.push(JSON.parse(line.slice(6)))
    }
  }

  return { events, rest }
}

export interface StreamOptions {
  url: string;
  headers?: Record<string, string>;
  accessToken?: string;
  body?: unknown;
  signal?: AbortSignal | null;
}

/**
 * POSTs `body` and calls `onEvent` for every event as it arrives. Resolves
 * with the final event (the one with `done: true`); rejects when the server
 * reports an error, answers with a failure status, or the stream ends without
 * a final event.
 */
export async function postForEvents (opt: StreamOptions, onEvent: (ev: StreamEvent) => void): Promise<StreamEvent> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(opt.headers || {}) }
  if (opt.accessToken) {
    headers.Authorization = 'Bearer ' + opt.accessToken
  }

  const response = await fetch(opt.url, {
    method: 'POST',
    headers,
    credentials: 'include',
    signal: opt.signal as AbortSignal | null,
    body: JSON.stringify(opt.body ?? {}),
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(readableError(text) || `HTTP ${response.status}`)
  }

  // the server answers an ordinary JSON error (not a stream) when it refuses
  // the request before starting, e.g. for a missing permission
  const type = response.headers.get('Content-Type') || ''
  if (!type.includes('text/event-stream')) {
    const text = await response.text()
    throw new Error(readableError(text) || 'The server did not start a stream')
  }

  const reader = response.body!.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (value) buffer += decoder.decode(value, { stream: !done })

    const parsed = parseSSE(done ? buffer + '\n\n' : buffer)
    buffer = parsed.rest

    for (const ev of parsed.events) {
      if (ev.error) throw new Error(String(ev.error))
      onEvent(ev)
      if (ev.done) return ev
    }

    if (done) break
  }

  throw new Error('The stream ended before the result arrived')
}

function readableError (text: string): string {
  try {
    const parsed = JSON.parse(text)
    return parsed?.error?.message || parsed?.error || text
  } catch {
    return text
  }
}
