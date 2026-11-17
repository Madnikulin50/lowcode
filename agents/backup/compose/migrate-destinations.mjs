#!/usr/bin/env node
/**
 * Одноразовая миграция: policies.dest_bucket (legacy free-text) → записи
 * destinations + связь policies.destination.
 *
 *   node migrate-destinations.mjs
 *
 * Идемпотентна: политики со заполненным destination пропускаются.
 * Endpoint назначений берётся из MINIO_ENDPOINT (дефолт агента) — туда
 * legacy-значения и резолвились.
 */
import { mintToken, detectBase, apiFactory, setOf, createRecord } from './helpers.mjs'

function recMap (rec) {
  const m = {}
  for (const v of rec.values || []) {
    if (m[v.name] === undefined) m[v.name] = v.value
    else if (v.value !== '') m[v.name] += '\n' + v.value
  }
  return m
}

async function updateRecord (api, nsID, moduleID, recordID, values) {
  const payload = Object.entries(values).map(([name, value]) => ({ name, value: value == null ? '' : String(value) }))
  return api('POST', `/namespace/${nsID}/module/${moduleID}/record/${recordID}`, { values: payload })
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const endpoint = process.env.MINIO_ENDPOINT || '127.0.0.1:9000'

  console.log('API', base, 'endpoint', endpoint)

  const nsList = setOf(await api('GET', '/namespace/?slug=backup&limit=50'))
  const ns = nsList.find(n => n.slug === 'backup')
  if (!ns) throw new Error('namespace backup not found; run apply.mjs first')
  const nsID = ns.namespaceID

  const modList = setOf(await api('GET', `/namespace/${nsID}/module/?limit=200`))
  const mod = h => modList.find(m => m.handle === h)?.moduleID
  const policiesID = mod('policies')
  const destinationsID = mod('destinations')
  if (!policiesID || !destinationsID) throw new Error('modules policies/destinations not found; run apply.mjs first')

  const policies = setOf(await api('GET', `/namespace/${nsID}/module/${policiesID}/record/?limit=500`))
  const destinations = setOf(await api('GET', `/namespace/${nsID}/module/${destinationsID}/record/?limit=500`))
  const destByBucket = new Map()
  for (const d of destinations) {
    const m = recMap(d)
    if (m.bucket) destByBucket.set(m.bucket, d)
  }

  let linked = 0
  let created = 0
  let skipped = 0
  for (const rec of policies) {
    const m = recMap(rec)
    const bucket = (m.dest_bucket || '').trim()
    if (!bucket) {
      skipped++
      continue
    }
    if ((m.destination || '').trim()) {
      console.log(`policy ${rec.recordID} "${m.name}": already has destination, skip`)
      skipped++
      continue
    }
    let dest = destByBucket.get(bucket)
    if (!dest) {
      dest = await createRecord(api, nsID, destinationsID, {
        name: `legacy-${bucket}`,
        handle: `legacy-${bucket}`,
        type: 'minio',
        endpoint,
        bucket,
        enabled: '1',
        notes: 'Создано миграцией из policies.dest_bucket',
      })
      destByBucket.set(bucket, dest)
      created++
      console.log(`destination legacy-${bucket} created (${dest.recordID || dest.ID})`)
    }
    const destID = String(dest.recordID || dest.ID)
    await updateRecord(api, nsID, policiesID, rec.recordID, {
      destination: destID,
      dest_bucket: '',
    })
    linked++
    console.log(`policy ${rec.recordID} "${m.name}" → destination ${destID} (bucket ${bucket})`)
  }
  console.log(`\ndone: ${linked} linked, ${created} destinations created, ${skipped} skipped`)
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
