#!/usr/bin/env node
/**
 * Provision the Backup Compose namespace.
 *
 *   node apply.mjs
 *   TOKEN=... node apply.mjs   # optional override
 */
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { writeFileSync } from 'node:fs'

import {
  mintToken, detectBase, apiFactory, ensureNamespace, ensureModule, ensureChart,
  ensurePage, ensureRuleChain, parentPages, doughnutChart, createRecord, setOf,
} from './helpers.mjs'
import {
  agentFields, credentialFields, sourceFields, destinationFields, policyFields, jobFields,
  snapshotFields, restoreFields, verifyReportFields, backupRequestFields,
  auditScopeFields, auditFindingFields, destructionActFields,
} from './fields.mjs'
import { buildPages } from './pages.mjs'
import { buildRuleChains } from './chains.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const NS_META = {
  subtitle: 'Резервное копирование во внутренний MinIO',
  description: 'SMB, локальные папки, базы данных и внешний S3. Каталог в Compose, байты в MinIO, исполнение — backup-agent.',
  prompt: 'Это пространство резервного копирования. Модули: agents, credentials, sources, destinations, policies, jobs, snapshots, restores. Секреты не хранятся в записях — только handle (BACKUP_SECRET_<handle> на агенте). Куда сохраняется бэкап, задаётся записью destinations (связь policies.destination; пусто = default-назначение или env агента). Запуски идут через rule chains на backup-agent :8087.',
}

async function seedIfEmpty (api, nsID, modules) {
  const existing = setOf(await api('GET', `/namespace/${nsID}/module/${modules.sources}/record/?limit=5`))
  if (existing.length) {
    console.log('sources already have records, skip seed')
    return
  }
  const agent = await createRecord(api, nsID, modules.agents, {
    name: 'local-agent',
    url: process.env.BACKUP_AGENT_URL || 'http://localhost:8087/api',
    hostname: 'localhost',
    capabilities: 'fs,smb,database,s3',
    status: 'offline',
  })
  const src = await createRecord(api, nsID, modules.sources, {
    name: 'demo-fs',
    type: 'fs',
    path: '/tmp/backup-demo',
    agent: String(agent.recordID || agent.ID),
    enabled: '1',
    notes: 'Демо-источник. Создайте /tmp/backup-demo и нажмите «Запустить бэкап».',
  })
  await createRecord(api, nsID, modules.policies, {
    name: 'nightly-demo',
    source: String(src.recordID || src.ID),
    cron: '0 2 * * *',
    retention_days: '14',
    incremental: '0',
    compression: 'gzip',
    enabled: '1',
  })
  console.log('seeded demo-fs + nightly-demo')
}

async function ensureDefaultDestination (api, nsID, modules) {
  const existing = setOf(await api('GET', `/namespace/${nsID}/module/${modules.destinations}/record/?limit=50`))
  if (existing.length) {
    console.log('destinations already exist, skip default seed')
    return
  }
  await createRecord(api, nsID, modules.destinations, {
    name: 'minio-default',
    handle: 'minio-default',
    type: 'minio',
    endpoint: process.env.MINIO_ENDPOINT || '127.0.0.1:9000',
    bucket: process.env.MINIO_BUCKET || 'backups',
    secure: process.env.MINIO_SECURE === 'true' ? '1' : '0',
    default: '1',
    enabled: '1',
    notes: 'Дефолтное назначение, созданное apply.mjs из env агента (MINIO_*)',
  })
  console.log('seeded default destination from MINIO_* env')
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const agentUrl = (process.env.BACKUP_AGENT_URL || 'http://localhost:8087/api').replace(/\/$/, '')

  console.log('API', base)

  const nsID = await ensureNamespace(api, {
    name: 'Backup',
    slug: 'backup',
    meta: NS_META,
  })

  const agents = await ensureModule(api, nsID, { name: 'Агенты', handle: 'agents', fields: agentFields() })
  const credentials = await ensureModule(api, nsID, { name: 'Секреты', handle: 'credentials', fields: credentialFields() })
  const sources = await ensureModule(api, nsID, {
    name: 'Источники',
    handle: 'sources',
    fields: sourceFields(agents, credentials),
    config: { records: { readOnly: false, noCreate: false, noDelete: false } },
  })
  const destinations = await ensureModule(api, nsID, { name: 'Назначения', handle: 'destinations', fields: destinationFields(credentials) })
  const policies = await ensureModule(api, nsID, { name: 'Политики', handle: 'policies', fields: policyFields(sources, destinations) })
  const jobs = await ensureModule(api, nsID, { name: 'Джобы', handle: 'jobs', fields: jobFields(policies, sources) })
  const snapshots = await ensureModule(api, nsID, { name: 'Снапшоты', handle: 'snapshots', fields: snapshotFields(jobs, sources, policies, destinations) })
  const restores = await ensureModule(api, nsID, { name: 'Восстановления', handle: 'restores', fields: restoreFields(snapshots) })
  const verifyReports = await ensureModule(api, nsID, { name: 'Верификации', handle: 'verify_reports', fields: verifyReportFields(snapshots, sources) })
  const backupRequests = await ensureModule(api, nsID, { name: 'Заявки на резервирование', handle: 'backup_requests', fields: backupRequestFields() })
  const auditScopes = await ensureModule(api, nsID, { name: 'Аудит: области', handle: 'audit_scopes', fields: auditScopeFields(credentials, agents) })
  const auditFindings = await ensureModule(api, nsID, { name: 'Аудит: находки', handle: 'audit_findings', fields: auditFindingFields(auditScopes) })
  const destructionActs = await ensureModule(api, nsID, { name: 'Акты уничтожения', handle: 'destruction_acts', fields: destructionActFields(snapshots) })

  const modules = {
    agents, credentials, sources, destinations, policies, jobs, snapshots, restores,
    verify_reports: verifyReports, backup_requests: backupRequests,
    audit_scopes: auditScopes, audit_findings: auditFindings, destruction_acts: destructionActs,
  }

  const jobsByStatus = await ensureChart(api, nsID, doughnutChart('Джобы по статусу', 'jobs-by-status', jobs, 'status'))
  const sourcesByType = await ensureChart(api, nsID, doughnutChart('Источники по типу', 'sources-by-type', sources, 'type'))
  const snapshotsByEngine = await ensureChart(api, nsID, doughnutChart('Снапшоты по движку', 'snapshots-by-engine', snapshots, 'engine'))
  const charts = { jobsByStatus, sourcesByType, snapshotsByEngine }

  const pageIDs = {}
  for (const page of buildPages({ modules, charts })) {
    pageIDs[page.handle] = await ensurePage(api, nsID, page)
  }

  await parentPages(api, nsID, pageIDs, {
    source: 'sources',
    policy: 'policies',
    job: 'jobs',
    snapshot: 'snapshots',
    restore: 'restores',
    agent: 'agents',
    credential: 'credentials',
    destination: 'destinations',
  })

  for (const chain of buildRuleChains({ nsID, modules, agentUrl })) {
    await ensureRuleChain(api, chain)
  }

  try {
    await ensureDefaultDestination(api, nsID, modules)
  } catch (e) {
    console.warn('default destination seed skipped:', e.message)
  }

  try {
    await seedIfEmpty(api, nsID, modules)
  } catch (e) {
    console.warn('seed skipped:', e.message)
  }

  const summary = {
    namespaceID: String(nsID),
    slug: 'backup',
    api: base,
    modules: Object.fromEntries(Object.entries(modules).map(([k, v]) => [k, String(v)])),
    charts: Object.fromEntries(Object.entries(charts).map(([k, v]) => [k, String(v)])),
    urls: {
      namespace: '/ns/backup',
      dashboard: '/ns/backup/pages',
    },
    agent: {
      flags: `--api=${base.replace(/\/compose$/, '').replace(/\/api$/, '')} --namespace=${nsID}`,
      listen: ':8087',
    },
  }
  writeFileSync(join(HERE, 'applied.json'), JSON.stringify(summary, null, 2))
  console.log('\nBackup ready')
  console.log(JSON.stringify(summary, null, 2))
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
