const STORAGE_KEY = 'compose.pendingRecordCreate'
const MAX_AGE_MS = 60 * 1000

function writeBoth (raw) {
  try { sessionStorage.setItem(STORAGE_KEY, raw) } catch {}
  try { localStorage.setItem(STORAGE_KEY, raw) } catch {}
}

function readAndClear () {
  let raw
  try { raw = sessionStorage.getItem(STORAGE_KEY) } catch {}
  if (!raw) {
    try { raw = localStorage.getItem(STORAGE_KEY) } catch {}
  }
  try { sessionStorage.removeItem(STORAGE_KEY) } catch {}
  try { localStorage.removeItem(STORAGE_KEY) } catch {}
  return raw
}

/**
 * Vue Router 4 discards params that are not in the path (refRecord, values).
 * Stash them before push / window.open and consume on the create view.
 */
export function stashRecordCreate ({ refRecord, values } = {}) {
  const payload = {
    ts: Date.now(),
    values: values && typeof values === 'object' ? values : {},
    refRecord: refRecord?.recordID
      ? { recordID: refRecord.recordID, moduleID: refRecord.moduleID }
      : null,
  }
  writeBoth(JSON.stringify(payload))
  return payload
}

export function takeRecordCreate () {
  const raw = readAndClear()
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw)
    if (!parsed || (parsed.ts && Date.now() - parsed.ts > MAX_AGE_MS)) return {}
    return {
      values: parsed.values && typeof parsed.values === 'object' ? parsed.values : {},
      refRecord: parsed.refRecord || null,
    }
  } catch {
    return {}
  }
}

export function recordCreateLocation ({ name = 'page.record.create', pageID, moduleID, refRecord, values } = {}) {
  stashRecordCreate({ refRecord, values })
  const params = {}
  if (pageID) params.pageID = pageID
  if (moduleID) params.moduleID = moduleID
  return { name, params }
}
