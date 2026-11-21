#!/usr/bin/env node
/**
 * Lists local PostgreSQL databases and tables for the DSPM discovery chain.
 * Does not return row contents or the connection secret.
 *
 *   node discover.mjs
 *   GET/POST http://127.0.0.1:8091/discover
 */
import { createServer } from 'node:http'
import { readFileSync } from 'node:fs'
import { spawn } from 'node:child_process'
import { networkInterfaces } from 'node:os'
import net from 'node:net'
import { inspectSmb } from './smb.mjs'
import { inspectMinio } from './minio.mjs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const PORT = Number(process.env.DSPM_DISCOVER_PORT || 8091)
const MAX_DB = Number(process.env.DSPM_DISCOVER_MAX_DB || 30)
const MAX_TABLES = Number(process.env.DSPM_DISCOVER_MAX_TABLES || 12)
const PROBE_MS = Number(process.env.DSPM_DISCOVER_PROBE_MS || 250)
const PROBE_PARALLEL = Number(process.env.DSPM_DISCOVER_PARALLEL || 80)
const MAX_NET = Number(process.env.DSPM_DISCOVER_MAX_HOSTS || 60)
const HERE = dirname(fileURLToPath(import.meta.url))

const SKIP_IFACE = /^(lo|docker|br-|veth|virbr|cni|flannel|tun|tap|wg|lxc|vboxnet|vmnet|kube-|cali|nodelocal)/
const DATA_PORTS = [
  [5432, 'database', 'postgresql'],
  [3306, 'database', 'mysql'],
  [1433, 'database', 'mssql'],
  [1521, 'database', 'oracle'],
  [27017, 'database', 'mongodb'],
  [6379, 'database', 'redis'],
  [9200, 'database', 'elasticsearch'],
  [9042, 'database', 'cassandra'],
  [5984, 'database', 'couchdb'],
  [7474, 'database', 'neo4j'],
  [8086, 'database', 'influxdb'],
  [8123, 'database', 'clickhouse'],
  [9000, 'bucket', 'minio'],
  [445, 'fileshare', 'smb'],
  [2049, 'fileshare', 'nfs'],
  [21, 'fileshare', 'ftp'],
  [993, 'mailbox', 'imap'],
  [143, 'mailbox', 'imap'],
  [11434, 'saas_app', 'ollama'],
]

function dsnFromEnvFile () {
  if (process.env.DSPM_PG_DSN) return process.env.DSPM_PG_DSN
  const path = join(HERE, '../../docker-conf/test9/.env')
  const text = readFileSync(path, 'utf8')
  const line = text.split('\n').find(l => l.startsWith('DB_DSN='))
  if (!line) throw new Error('DB_DSN missing')
  const raw = line.slice('DB_DSN='.length).trim()
  const u = new URL(raw)
  u.hostname = '127.0.0.1'
  u.pathname = '/postgres'
  return u.toString()
}

function psql (dsn, database, sql) {
  const u = new URL(dsn)
  const env = {
    ...process.env,
    PGPASSWORD: decodeURIComponent(u.password || ''),
    PGHOST: u.hostname,
    PGPORT: u.port || '5432',
    PGUSER: decodeURIComponent(u.username || 'postgres'),
    PGDATABASE: database,
  }
  return new Promise((resolve, reject) => {
    const child = spawn('psql', ['-At', '-v', 'ON_ERROR_STOP=1', '-c', sql], { env })
    let out = ''
    let err = ''
    child.stdout.on('data', d => { out += d })
    child.stderr.on('data', d => { err += d })
    child.on('error', reject)
    child.on('close', code => {
      if (code !== 0) reject(new Error(err.trim() || `psql ${code}`))
      else resolve(out.split('\n').map(s => s.trim()).filter(Boolean))
    })
  })
}

function prefixLength (netmask) {
  return netmask.split('.').reduce((bits, octet) => bits + (Number(octet).toString(2).match(/1/g) || []).length, 0)
}

function localCIDRs () {
  if (process.env.DSPM_DISCOVER_CIDRS) {
    return process.env.DSPM_DISCOVER_CIDRS.split(',').map(s => s.trim()).filter(Boolean)
  }
  const out = []
  for (const [name, addrs] of Object.entries(networkInterfaces())) {
    if (SKIP_IFACE.test(name)) continue
    for (const addr of addrs || []) {
      if (addr.family !== 'IPv4' || addr.internal) continue
      const bits = Math.max(prefixLength(addr.netmask), 24)
      const ip = addr.address.split('.').map(Number)
      const mask = bits >= 32 ? 0 : (0xffffffff << (32 - bits)) >>> 0
      const raw = ((ip[0] << 24) | (ip[1] << 16) | (ip[2] << 8) | ip[3]) >>> 0
      const network = raw & mask
      const a = [(network >>> 24) & 255, (network >>> 16) & 255, (network >>> 8) & 255, network & 255]
      out.push(`${a.join('.')}/${bits}`)
    }
  }
  return [...new Set(out)]
}

function hostsInCIDR (cidr) {
  const [base, bitsRaw] = cidr.split('/')
  const bits = Number(bitsRaw)
  if (bits < 24 || bits > 32) return []
  const ip = base.split('.').map(Number)
  const raw = ((ip[0] << 24) | (ip[1] << 16) | (ip[2] << 8) | ip[3]) >>> 0
  const count = 2 ** (32 - bits)
  const hosts = []
  for (let i = bits >= 31 ? 0 : 1; i < (bits >= 31 ? count : count - 1); i++) {
    const n = (raw + i) >>> 0
    hosts.push([(n >>> 24) & 255, (n >>> 16) & 255, (n >>> 8) & 255, n & 255].join('.'))
  }
  return hosts
}

function probe (host, port) {
  return new Promise(resolve => {
    const socket = net.connect({ host, port, family: 4 })
    let settled = false
    const done = open => {
      if (settled) return
      settled = true
      socket.destroy()
      resolve(open)
    }
    socket.setTimeout(PROBE_MS)
    socket.on('connect', () => done(true))
    socket.on('timeout', () => done(false))
    socket.on('error', () => done(false))
  })
}

async function mapPool (items, limit, fn) {
  const out = new Array(items.length)
  let cursor = 0
  async function worker () {
    while (cursor < items.length) {
      const idx = cursor++
      out[idx] = await fn(items[idx])
    }
  }
  await Promise.all(Array.from({ length: Math.min(limit, items.length) }, () => worker()))
  return out
}

async function discoverNetwork (now) {
  const cidrs = localCIDRs()
  const hosts = [...new Set(cidrs.flatMap(hostsInCIDR))]
  const jobs = []
  for (const host of hosts) {
    for (const spec of DATA_PORTS) jobs.push({ host, spec })
  }
  const opened = []
  await mapPool(jobs, PROBE_PARALLEL, async job => {
    if (opened.length >= MAX_NET) return
    const open = await probe(job.host, job.spec[0])
    if (open) opened.push(job)
  })
  const seen = new Set()
  const items = []
  for (const job of opened) {
    if (items.length >= MAX_NET) break
    const [port, kind, service] = job.spec
    const externalId = `${service}:${job.host}:${port}`
    if (seen.has(externalId)) continue
    seen.add(externalId)
    items.push({
      name: `${service} ${job.host}`,
      kind,
      location: `${job.host}:${port}`,
      externalId,
      exposure: 'internal',
      residency: 'network',
      sensitivity: 'internal',
      lastSeen: now,
      datasets: [],
    })
  }
  return { cidrs, hosts: hosts.length, items }
}

function run (cmd, args) {
  return new Promise(resolve => {
    const child = spawn(cmd, args)
    let out = ''
    child.stdout.on('data', d => { out += d })
    child.on('error', () => resolve(''))
    child.on('close', code => resolve(code === 0 ? out : ''))
  })
}

function dockerKind (image) {
  const s = image.toLowerCase()
  if (/postgres|mysql|mariadb|mongo|redis|clickhouse/.test(s)) return 'database'
  if (/minio|s3/.test(s)) return 'bucket'
  return 'saas_app'
}

async function discoverDocker (now) {
  const text = await run('docker', ['ps', '--format', '{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Ports}}'])
  const items = []
  for (const line of text.split('\n').map(s => s.trim()).filter(Boolean)) {
    const [id, name, image, ports] = line.split('\t')
    if (!name) continue
    const published = (ports || '').split(',').map(s => s.trim()).filter(s => s.includes('->'))
    items.push({
      name,
      kind: dockerKind(image || ''),
      location: published[0] || image || id,
      externalId: `docker:${name}`,
      exposure: published.length ? 'internal' : 'private',
      residency: 'docker',
      sensitivity: 'internal',
      lastSeen: now,
      datasets: published.slice(0, 12).map(spec => ({
        name: spec.split('->').pop() || name,
        path: spec,
        status: 'shadow',
        purpose: 'production',
      })),
    })
  }
  return items
}

async function discoverOllama (now) {
  const host = '127.0.0.1'
  const open = await probe(host, 11434)
  if (!open) return null
  let models = []
  try {
    const res = await fetch(`http://${host}:11434/api/tags`, { signal: AbortSignal.timeout(3000) })
    if (res.ok) {
      const body = await res.json()
      models = (body.models || []).map(m => m.name).filter(Boolean)
    }
  } catch {
    models = []
  }
  return {
    name: 'ollama',
    kind: 'saas_app',
    location: `${host}:11434`,
    externalId: `ollama:${host}:11434`,
    exposure: 'internal',
    residency: 'local',
    sensitivity: 'internal',
    lastSeen: now,
    datasets: models.map(name => ({
      name,
      path: `ollama/${name}`,
      status: 'shadow',
      purpose: 'ai',
    })),
  }
}

function ident (value) {
  const s = String(value || '')
  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(s)) return ''
  return s
}

function maskCell (value) {
  const s = String(value ?? '').trim()
  if (!s || s === '\\N') return ''
  if (s.includes('@')) return `${s[0]}***@${s.split('@').pop()}`
  if (/^\d{6,}$/.test(s)) return `${s.slice(0, 2)}***${s.slice(-2)}`
  if (s.length <= 2) return '**'
  if (s.length > 8) return `${s.slice(0, 2)}***${s.slice(-2)}`
  return `${s[0]}***`
}

function parseEndpoint (location, database) {
  const raw = String(location || '')
  const mapped = raw.match(/:(\d+)->/)
  if (mapped) return { host: '127.0.0.1', port: mapped[1], database: database || '' }
  const m = raw.match(/([\w.-]+):(\d+)(?:\/([\w.-]+))?/)
  if (!m) return { host: '', port: '', database: database || '' }
  return {
    host: m[1] === '0.0.0.0' ? '127.0.0.1' : m[1],
    port: m[2],
    database: database || m[3] || '',
  }
}

function psqlAuth (host, port, user, password, database, sql) {
  const env = {
    ...process.env,
    PGPASSWORD: password,
    PGHOST: host,
    PGPORT: port || '5432',
    PGUSER: user,
    PGDATABASE: database || 'postgres',
  }
  return new Promise((resolve, reject) => {
    const child = spawn('psql', ['-At', '-v', 'ON_ERROR_STOP=1', '-c', sql], { env })
    let out = ''
    let err = ''
    child.stdout.on('data', d => { out += d })
    child.stderr.on('data', d => { err += d })
    child.on('error', reject)
    child.on('close', code => {
      if (code !== 0) reject(new Error(err.trim().split('\n').pop() || 'psql failed'))
      else resolve(out.split('\n').map(s => s.trim()).filter(Boolean))
    })
  })
}

async function inspectPostgres (endpoint, username, password) {
  const db = ident(endpoint.database) || 'postgres'
  const tables = await psqlAuth(endpoint.host, endpoint.port, username, password, db, `
    SELECT n.nspname || '.' || c.relname
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE c.relkind = 'r'
      AND n.nspname NOT IN ('pg_catalog', 'information_schema')
      AND n.nspname NOT LIKE 'pg_toast%'
    ORDER BY 1
    LIMIT 8
  `)
  const items = []
  for (const path of tables) {
    const [schema, table] = path.split('.')
    if (!ident(schema) || !ident(table)) continue
    const cols = await psqlAuth(endpoint.host, endpoint.port, username, password, db, `
      SELECT column_name FROM information_schema.columns
      WHERE table_schema = '${schema}' AND table_name = '${table}'
      ORDER BY ordinal_position
      LIMIT 6
    `)
    const safeCols = cols.map(ident).filter(Boolean)
    let sample = ''
    if (safeCols.length) {
      const row = await psqlAuth(endpoint.host, endpoint.port, username, password, db,
        `SELECT ${safeCols.map(c => `"${c}"`).join(', ')} FROM "${schema}"."${table}" LIMIT 1`)
      const cells = (row[0] || '').split('|')
      sample = safeCols.map((col, i) => `${col}=${maskCell(cells[i] || '')}`).join(', ')
    }
    items.push({ name: table, path, sample, purpose: 'production', status: 'shadow' })
  }
  return items
}

function mysqlQuery (host, port, user, password, database, sql) {
  const env = { ...process.env, MYSQL_PWD: password }
  const args = ['-N', '-B', '-h', host, '-P', port || '3306', '-u', user]
  if (database) args.push(database)
  args.push('-e', sql)
  return new Promise((resolve, reject) => {
    const child = spawn('mysql', args, { env })
    let out = ''
    let err = ''
    child.stdout.on('data', d => { out += d })
    child.stderr.on('data', d => { err += d })
    child.on('error', () => reject(new Error('mysql client is not installed')))
    child.on('close', code => {
      if (code !== 0) reject(new Error(err.trim().split('\n').pop() || 'mysql failed'))
      else resolve(out.split('\n').map(s => s.trim()).filter(Boolean))
    })
  })
}

async function inspectMysql (endpoint, username, password) {
  const db = ident(endpoint.database)
  if (!db) throw new Error('укажите имя базы')
  const tables = await mysqlQuery(endpoint.host, endpoint.port, username, password, db, 'SHOW TABLES')
  const items = []
  for (const table of tables.slice(0, 8)) {
    if (!ident(table)) continue
    const desc = await mysqlQuery(endpoint.host, endpoint.port, username, password, db, `SHOW COLUMNS FROM \`${table}\``)
    const cols = desc.map(line => ident(line.split('\t')[0])).filter(Boolean).slice(0, 6)
    let sample = ''
    if (cols.length) {
      const row = await mysqlQuery(endpoint.host, endpoint.port, username, password, db,
        `SELECT ${cols.map(c => `\`${c}\``).join(', ')} FROM \`${table}\` LIMIT 1`)
      const cells = (row[0] || '').split('\t')
      sample = cols.map((col, i) => `${col}=${maskCell(cells[i] || '')}`).join(', ')
    }
    items.push({ name: table, path: `${db}.${table}`, sample, purpose: 'production', status: 'shadow' })
  }
  return items
}

async function inspectOllama (endpoint) {
  const host = endpoint.host || '127.0.0.1'
  const port = endpoint.port || '11434'
  const res = await fetch(`http://${host}:${port}/api/tags`, { signal: AbortSignal.timeout(3000) })
  if (!res.ok) throw new Error(`ollama ${res.status}`)
  const body = await res.json()
  return (body.models || []).slice(0, 12).map(model => ({
    name: model.name,
    path: `ollama/${model.name}`,
    sample: `size=${maskCell(String(model.size || ''))}`,
    purpose: 'ai',
    status: 'shadow',
  }))
}

// A rule-chain step interpolates fields like {{database}} from the record;
// when the record's field is empty (e.g. an empty array), the automation
// engine leaves the placeholder text unresolved instead of substituting an
// empty string. Treat any such literal placeholder as "not provided".
function templated (v) {
  const s = String(v || '').trim()
  return /^\{\{.*\}\}$/.test(s) ? '' : s
}

async function inspect (body) {
  const username = templated(body.username)
  const password = templated(body.password)
  const endpoint = parseEndpoint(body.location, templated(body.database))
  const kind = templated(body.kind)
  const port = endpoint.port
  if ((port === '11434' || kind === 'saas_app') && String(body.location || '').includes('11434')) {
    const items = await inspectOllama(endpoint)
    return { found: items.length, items }
  }
  const guestSmb = port === '445' || kind === 'fileshare'
  if (!guestSmb && (!username || !password)) throw new Error('укажите логин и пароль на карточке актива')
  if (!endpoint.host || !endpoint.port) throw new Error('у актива нет адреса host:port')
  let items = []
  if (port === '3306') items = await inspectMysql(endpoint, username, password)
  else if (port === '5432') items = await inspectPostgres(endpoint, username, password)
  else if (port === '445') items = await inspectSmb(endpoint, username, password)
  else if (port === '9000') items = await inspectMinio(endpoint, username, password)
  else throw new Error(`инспекция порта ${port} пока не поддержана`)
  return { found: items.length, items }
}

async function discover () {
  const dsn = dsnFromEnvFile()
  const u = new URL(dsn)
  const host = `${u.hostname}:${u.port || '5432'}`
  const names = await psql(dsn, 'postgres', `
    SELECT datname FROM pg_database
    WHERE datallowconn AND NOT datistemplate
    ORDER BY datname
    LIMIT ${MAX_DB}
  `)
  const now = new Date().toISOString()
  const items = []
  let datasets = 0
  for (const name of names) {
    let tables = []
    try {
      tables = await psql(dsn, name, `
        SELECT n.nspname || '.' || c.relname
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relkind = 'r'
          AND n.nspname NOT IN ('pg_catalog', 'information_schema')
          AND n.nspname NOT LIKE 'pg_toast%'
          AND n.nspname NOT LIKE 'pg_temp%'
        ORDER BY (n.nspname = 'public') DESC, 1
        LIMIT ${MAX_TABLES}
      `)
    } catch {
      tables = []
    }
    datasets += tables.length
    items.push({
      name,
      kind: 'database',
      location: `${host}/${name}`,
      externalId: `postgres:${host}/${name}`,
      exposure: 'private',
      residency: 'local',
      sensitivity: 'internal',
      lastSeen: now,
      datasets: tables.map(path => ({
        name: path.split('.').pop(),
        path,
        status: 'shadow',
        purpose: 'production',
      })),
    })
  }
  const network = await discoverNetwork(now)
  const extra = [...network.items, ...(await discoverDocker(now))]
  const ollama = await discoverOllama(now)
  if (ollama) extra.push(ollama)
  const known = new Set(items.map(item => item.externalId))
  for (const item of extra) {
    if (!known.has(item.externalId)) {
      items.push(item)
      datasets += (item.datasets || []).length
    }
  }
  return {
    found: items.length,
    datasetCount: datasets,
    networkHosts: network.hosts,
    networkCIDRs: network.cidrs,
    connectorName: 'Сетевое обнаружение',
    provider: 'network',
    scannedAt: now,
    items,
  }
}

const server = createServer(async (req, res) => {
  if (req.url === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' })
    res.end('{"ok":true}')
    return
  }
  if (req.url === '/inspect' && req.method === 'POST') {
    try {
      const chunks = []
      for await (const chunk of req) chunks.push(chunk)
      const payload = JSON.parse(Buffer.concat(chunks).toString() || '{}')
      const body = await inspect(payload)
      res.writeHead(200, { 'Content-Type': 'application/json' })
      res.end(JSON.stringify(body))
    } catch (err) {
      const message = String(err.message || 'inspect failed').replace(/password=\S+/gi, 'password=***')
      res.writeHead(400, { 'Content-Type': 'application/json' })
      res.end(JSON.stringify({ error: message }))
    }
    return
  }
  if (req.url !== '/discover') {
    res.writeHead(404)
    res.end()
    return
  }
  try {
    const body = await discover()
    res.writeHead(200, { 'Content-Type': 'application/json' })
    res.end(JSON.stringify(body))
  } catch (err) {
    res.writeHead(500, { 'Content-Type': 'application/json' })
    res.end(JSON.stringify({ error: err.message || 'discover failed' }))
  }
})

server.listen(PORT, '127.0.0.1', () => {
  console.log(`dspm discover listening on 127.0.0.1:${PORT}`)
})
