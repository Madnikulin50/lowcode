const DEFAULT_AGENT = 'http://localhost:8085/api'

function agentBase (url) {
  return String(url || DEFAULT_AGENT).replace(/\/$/, '')
}

function kv (name, value) {
  if (value == null || value === '') return null
  return { name, value: String(value) }
}

function deviceValues (d) {
  return [
    kv('ip_address', d.ip),
    kv('mac_address', d.mac),
    kv('hostname', d.hostname),
    kv('vendor', d.vendor),
    kv('device_type', normalizeType(d.deviceType)),
    kv('os', d.os),
    kv('domain', d.domain),
    kv('open_ports', d.openPorts ? JSON.stringify(d.openPorts) : '[]'),
    kv('services', d.services ? JSON.stringify(d.services) : '[]'),
    kv('shares', d.shares ? JSON.stringify(d.shares) : '[]'),
    kv('vulnerabilities', d.vulnerabilities ? JSON.stringify(d.vulnerabilities) : '[]'),
    kv('last_seen', d.lastSeen),
    kv('status', d.status || 'online'),
  ].filter(Boolean)
}

function normalizeType (t) {
  const v = String(t || 'unknown').toLowerCase()
  const allowed = ['router', 'switch', 'server', 'workstation', 'printer', 'camera', 'firewall', 'iot', 'unknown']
  if (allowed.includes(v)) return v
  if (v.includes('domain') || v.includes('server')) return 'server'
  return 'unknown'
}

function composeStatus (agentStatus) {
  if (agentStatus === 'done') return 'completed'
  if (agentStatus === 'error') return 'failed'
  if (agentStatus === 'running') return 'running'
  return agentStatus || 'running'
}

async function moduleByHandle (api, namespaceID, handle) {
  const res = await api.moduleList({ namespaceID, handle, limit: 50 })
  const set = res.set || res.response?.set || []
  const mod = set.find(m => m.handle === handle && !m.deletedAt)
  return mod?.moduleID || mod?.ID
}

async function upsertDevice (api, namespaceID, moduleID, device) {
  const ip = device.ip
  if (!ip) return
  const q = `ip_address = '${ip.replace(/'/g, "\\'")}'`
  const found = await api.recordList({ namespaceID, moduleID, query: q, limit: 1 })
  const set = found.set || []
  const values = deviceValues(device)
  if (set.length) {
    await api.recordUpdate({ namespaceID, moduleID, recordID: set[0].recordID, values })
    return
  }
  await api.recordCreate({ namespaceID, moduleID, values })
}

async function updateScanRecord (api, namespaceID, moduleID, recordID, status) {
  if (!recordID || !moduleID) return
  const values = [
    kv('status', composeStatus(status.status)),
    kv('progress', status.progress != null ? Math.round(status.progress) : 0),
    kv('found', status.found != null ? status.found : 0),
    kv('scanning_ip', status.scanningIP || ''),
    kv('error', status.error || status.message || ''),
    kv('target', status.target),
  ].filter(Boolean)
  if (status.startedAt) values.push(kv('started_at', status.startedAt))
  if (status.finishedAt) values.push(kv('finished_at', status.finishedAt))
  await api.recordUpdate({ namespaceID, moduleID, recordID, values })
}

function nodeOutput (nodes, id) {
  const n = (nodes || []).find(x => x.nodeID === id || x.NodeID === id)
  return n?.output || n?.Output || {}
}

export function scanIDsFromTrigger (result) {
  const nodes = result?.nodes || result?.Nodes || []
  const http = nodeOutput(nodes, 'http_scan')
  const crud = nodeOutput(nodes, 'record_scan')
  const body = http.body || http.Body || {}
  const out = result?.output || result?.Output || {}
  return {
    scanID: body.id || body.ID || out.scanID || out.ScanID,
    composeScanRecordID: crud.recordID || crud.RecordID || out.createdRecordID || out.CreatedRecordID,
  }
}

export async function pullScanResultsIntoCompose ({
  $ComposeAPI,
  namespaceID,
  agentUrl,
  scanID,
  composeScanRecordID,
  onProgress,
}) {
  const base = agentBase(agentUrl)
  const devicesMod = await moduleByHandle($ComposeAPI, namespaceID, 'devices')
  const scansMod = await moduleByHandle($ComposeAPI, namespaceID, 'scans')
  const deadline = Date.now() + 15 * 60 * 1000
  let last = null

  while (Date.now() < deadline) {
    const res = await fetch(`${base}/scans/${encodeURIComponent(scanID)}`)
    if (!res.ok) throw new Error(`CMDB API ${res.status}`)
    last = await res.json()
    if (onProgress) onProgress(last)
    if (composeScanRecordID && scansMod) {
      try { await updateScanRecord($ComposeAPI, namespaceID, scansMod, composeScanRecordID, last) } catch (e) { console.warn(e) }
    }
    const st = last.status || last.Status
    if (st === 'done' || st === 'error' || st === 'completed' || st === 'failed') break
    await new Promise(r => setTimeout(r, 2000))
  }

  const st = last?.status || last?.Status
  const stillRunning = st === 'running' || st === 'pending'
  if (stillRunning) {
    const msg = last?.error || last?.message || `scan still running on agent (${last?.scannedIPs || 0}/${last?.totalIPs || '?'} hosts)`
    if (composeScanRecordID && scansMod) {
      try {
        await updateScanRecord($ComposeAPI, namespaceID, scansMod, composeScanRecordID, {
          ...last,
          status: 'running',
          error: msg,
        })
      } catch (e) { console.warn(e) }
    }
    window.dispatchEvent(new CustomEvent('refetch-records', { detail: { stayOnPage: true } }))
    const err = new Error(msg)
    err.stillRunning = true
    err.status = last
    throw err
  }

  const listRes = await fetch(`${base}/devices`)
  if (!listRes.ok) throw new Error(`CMDB devices API ${listRes.status}`)
  const devices = await listRes.json()
  const rows = Array.isArray(devices) ? devices : (devices.devices || [])
  if (devicesMod) {
    for (const d of rows) {
      try { await upsertDevice($ComposeAPI, namespaceID, devicesMod, d) } catch (e) { console.warn('device upsert', d.ip, e) }
    }
  }
  if (composeScanRecordID && scansMod && last) {
    try { await updateScanRecord($ComposeAPI, namespaceID, scansMod, composeScanRecordID, last) } catch (e) { console.warn(e) }
  }
  window.dispatchEvent(new CustomEvent('refetch-records', { detail: { stayOnPage: true } }))
  return { found: rows.length, status: last }
}
