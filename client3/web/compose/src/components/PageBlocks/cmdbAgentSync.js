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
