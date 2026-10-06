import { expect } from 'chai'
import { parseSSE, postForEvents } from './sse'

describe('sse', () => {
  describe('parseSSE', () => {
    it('reads complete events and keeps the unfinished tail', () => {
      const { events, rest } = parseSSE('data: {"token":"Hel"}\n\ndata: {"token":"lo"}\n\ndata: {"tok')
      expect(events).to.deep.equal([{ token: 'Hel' }, { token: 'lo' }])
      expect(rest).to.equal('data: {"tok')
    })

    it('waits for the blank line before trusting a block', () => {
      expect(parseSSE('data: {"a":1}\n').events).to.deep.equal([])
      expect(parseSSE('data: {"a":1}\n\n').events).to.deep.equal([{ a: 1 }])
    })

    it('ignores lines that are not data', () => {
      expect(parseSSE(': keep-alive\n\nevent: x\ndata: {"a":1}\n\n').events).to.deep.equal([{ a: 1 }])
    })

    it('copes with an empty buffer', () => {
      expect(parseSSE('')).to.deep.equal({ events: [], rest: '' })
    })
  })

  describe('postForEvents', () => {
    const realFetch = globalThis.fetch
    afterEach(() => { globalThis.fetch = realFetch })

    const streamOf = (chunks: string[], init: { status?: number; type?: string } = {}) => {
      const enc = new TextEncoder()
      return new Response(new ReadableStream({
        start (c) { chunks.forEach(ch => c.enqueue(enc.encode(ch))); c.close() },
      }), { status: init.status ?? 200, headers: { 'Content-Type': init.type ?? 'text/event-stream' } })
    }

    it('delivers events as they arrive, even when chunks split an event', async () => {
      let sent: any
      globalThis.fetch = (async (url: any, init: any) => {
        sent = { url, init }
        return streamOf(['data: {"token":"Hel"}\n\nda', 'ta: {"token":"lo"}\n\n', 'data: {"result":{"ok":true},"done":true}\n\n'])
      }) as any

      const seen: any[] = []
      const last = await postForEvents({ url: '/x', accessToken: 'tok', body: { a: 1 } }, ev => seen.push(ev))

      expect(seen.map(e => e.token)).to.deep.equal(['Hel', 'lo', undefined])
      expect(last).to.deep.equal({ result: { ok: true }, done: true })
      expect(sent.init.headers.Authorization).to.equal('Bearer tok')
      expect(sent.init.credentials).to.equal('include')
      expect(JSON.parse(sent.init.body)).to.deep.equal({ a: 1 })
    })

    it('rejects with the server error carried by an event', async () => {
      globalThis.fetch = (async () => streamOf(['data: {"token":"a"}\n\n', 'data: {"error":"model down","done":true}\n\n'])) as any
      let message = ''
      await postForEvents({ url: '/x' }, () => {}).catch(e => { message = e.message })
      expect(message).to.equal('model down')
    })

    it('rejects with the message of an ordinary JSON error', async () => {
      globalThis.fetch = (async () => new Response('{"error":{"message":"not allowed to create workflows"}}', { status: 200, headers: { 'Content-Type': 'application/json' } })) as any
      let message = ''
      await postForEvents({ url: '/x' }, () => {}).catch(e => { message = e.message })
      expect(message).to.equal('not allowed to create workflows')
    })

    it('rejects on an HTTP failure', async () => {
      globalThis.fetch = (async () => new Response('forbidden', { status: 403 })) as any
      let message = ''
      await postForEvents({ url: '/x' }, () => {}).catch(e => { message = e.message })
      expect(message).to.equal('forbidden')
    })

    it('rejects when the stream ends without a result', async () => {
      globalThis.fetch = (async () => streamOf(['data: {"token":"a"}\n\n'])) as any
      let message = ''
      await postForEvents({ url: '/x' }, () => {}).catch(e => { message = e.message })
      expect(message).to.contain('ended before the result')
    })
  })
})
