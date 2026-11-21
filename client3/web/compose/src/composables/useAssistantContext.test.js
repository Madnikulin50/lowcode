import { describe, it, expect, beforeEach } from 'vitest'
import { useAssistantContext, registerAssistantDocument, openAssistant, closeAssistant, parseAttachmentURL, assistantURL } from './useAssistantContext'

describe('useAssistantContext', () => {
  beforeEach(() => {
    closeAssistant()
    const { state } = useAssistantContext()
    state.stack.splice(0, state.stack.length)
    state.open = false
    globalThis.window = globalThis.window || globalThis
    window.CortezaAPI = 'https://host'
  })

  it('reads the namespace and attachment IDs from an attachment URL', () => {
    expect(parseAttachmentURL('https://h/compose/namespace/123/attachment/record/456/original/a.pdf')).toEqual({ namespaceID: '123', id: '456' })
    expect(parseAttachmentURL('https://h/other')).toBeNull()
  })

  it('shows the most recently registered document and forgets it when it closes', () => {
    const { current } = useAssistantContext()
    expect(current.value).toBeNull()
    const off1 = registerAssistantDocument({ source: 'document', id: '1', namespaceID: '9', name: 'a' })
    const off2 = registerAssistantDocument({ source: 'attachment', id: '2', namespaceID: '9', name: 'b' })
    expect(current.value.id).toBe('2')
    off2()
    expect(current.value.id).toBe('1')
    off1()
    expect(current.value).toBeNull()
  })

  it('ignores a document without an ID and lets a file link pin the assistant', () => {
    const { current, state } = useAssistantContext()
    registerAssistantDocument({ source: 'document', id: '', namespaceID: '9' })
    expect(current.value).toBeNull()
    openAssistant({ source: 'attachment', id: '7', namespaceID: '9', name: 'f' })
    expect(state.open).toBe(true)
    expect(current.value.id).toBe('7')
    closeAssistant()
    expect(state.open).toBe(false)
    expect(current.value).toBeNull()
  })

  it('builds the agent URL of the document viewer', () => {
    const url = assistantURL({ source: 'attachment', id: '7', namespaceID: '9' }, 3)
    expect(url).toBe('https://host/gateway/agents/stroykontrol/?view=doc&source=attachment&id=7&namespaceID=9&page=3')
  })
})
