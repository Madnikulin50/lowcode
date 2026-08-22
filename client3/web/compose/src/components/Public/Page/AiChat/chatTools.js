export function parseModelsPayload (payload = {}) {
  const raw = payload.models || []
  const names = []
  const tools = { ...(payload.tools || {}) }
  for (const m of raw) {
    if (typeof m === 'string') {
      names.push(m)
    } else if (m && typeof m === 'object' && m.name) {
      names.push(m.name)
      if (typeof m.tools === 'boolean') {
        tools[m.name] = m.tools
      }
    }
  }
  return {
    names,
    tools,
    defaultModel: payload.default || '',
  }
}

export function modelToolsEnabled (name, map) {
  if (!name || !map || typeof map !== 'object') return null
  if (Object.prototype.hasOwnProperty.call(map, name)) return !!map[name]
  const leaf = String(name).split('/').pop()
  if (Object.prototype.hasOwnProperty.call(map, leaf)) return !!map[leaf]
  const base = leaf.split(':')[0]
  const latest = `${base}:latest`
  if (Object.prototype.hasOwnProperty.call(map, latest)) return !!map[latest]
  if (Object.prototype.hasOwnProperty.call(map, base)) return !!map[base]
  return null
}

export function modelLabel (id) {
  if (!id) return ''
  const [name, tag] = String(id).split(':')
  const pretty = name.replace(/[-_]/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
  return tag ? `${pretty} (${tag})` : pretty
}

export function modelBase (id) {
  const leaf = String(id || '').split('/').pop()
  return leaf.split(':')[0]
}

export function pickChatModel (models, saved, serverDefault) {
  if (!Array.isArray(models) || !models.length) return ''
  if (saved && models.includes(saved)) return saved
  if (serverDefault && models.includes(serverDefault)) return serverDefault
  const defBase = modelBase(serverDefault)
  if (defBase) {
    const latest = models.find(m => modelBase(m) === defBase && (m === defBase || m.endsWith(':latest')))
    if (latest) return latest
    const same = models.find(m => modelBase(m) === defBase)
    if (same) return same
  }
  return models[0]
}

export function readStoredModel (key = 'aiChat.model') {
  try { return localStorage.getItem(key) || '' } catch (e) { return '' }
}

export function writeStoredModel (value, key = 'aiChat.model') {
  if (!value) return
  try { localStorage.setItem(key, value) } catch (e) {}
}
