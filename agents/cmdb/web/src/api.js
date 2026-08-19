const BASE = '/api'

async function req(path, opts = {}) {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  return res.json()
}

export function startScan(cidr, namespaceID) {
  return req('/scan', { method: 'POST', body: { cidr, namespaceID } })
}

export function listScans() {
  return req('/scans')
}

export function getScan(scanID) {
  return req(`/scans/${scanID}`)
}

export async function listDevices(moduleID) {
  const q = moduleID ? `?moduleID=${moduleID}` : ''
  const data = await req(`/devices${q}`)
  return Array.isArray(data) ? data : []
}

export function getDevice(recordID, moduleID) {
  const q = moduleID ? `?moduleID=${moduleID}` : ''
  return req(`/devices/${recordID}${q}`)
}

export function deleteDevice(recordID, moduleID) {
  const q = moduleID ? `?moduleID=${moduleID}` : ''
  return req(`/devices/${recordID}${q}`, { method: 'DELETE' })
}

export function ensureModule() {
  return req('/modules/ensure', { method: 'POST' })
}
