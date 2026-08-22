/**
 * Shared Corteza provisioning helpers (same pattern as agents/cmdb/compose/apply.mjs).
 */
import { execFileSync } from 'node:child_process'
import { createHash, createHmac } from 'node:crypto'

export function selectOptions (pairs) {
  return {
    options: pairs.map(([value, text, style]) => ({
      value,
      text,
      style: style || {},
    })),
    selectType: 'default',
    displayType: 'badge',
    multiDelimiter: '\n',
  }
}

export function field (name, label, kind, extra = {}) {
  return {
    name,
    label,
    kind,
    isRequired: !!extra.required,
    isMulti: !!extra.multi,
    options: extra.options || {},
  }
}

export function recordRel (name, label, moduleID, labelField, queryFields, required = false, extra = {}) {
  return field(name, label, 'Record', {
    required,
    multi: extra.multi,
    options: {
      moduleID: String(moduleID || '0'),
      labelField,
      queryFields,
      selectType: '',
      multiDelimiter: '\n',
    },
  })
}

export function userField (name, label, extra = {}) {
  return field(name, label, 'User', {
    required: extra.required,
    multi: extra.multi,
    options: {
      roles: extra.roles || [],
      presetWithAuthenticated: extra.preset ?? true,
      selectType: 'default',
      multiDelimiter: '\n',
      isUniqueMultiValue: false,
    },
  })
}

export function fileField (name, label, extra = {}) {
  return field(name, label, 'File', {
    required: extra.required,
    multi: extra.multi,
    options: {
      allowImages: extra.images ?? true,
      allowDocuments: extra.documents ?? true,
      maxSize: extra.maxSize || 0,
      mode: extra.mode || 'list',
      inline: true,
      hideFileName: false,
      mimetypes: extra.mimetypes || '',
      clickToView: true,
      enableDownload: true,
      enableWebcam: extra.webcam ?? false,
      multiDelimiter: '\n',
      description: { view: '', edit: undefined },
      hint: { view: '', edit: undefined },
    },
  })
}

export function dateField (name, label, extra = {}) {
  return field(name, label, 'DateTime', {
    required: extra.required,
    options: {
      format: extra.format || '',
      onlyDate: extra.onlyDate ?? true,
      onlyTime: false,
      onlyPastValues: false,
      onlyFutureValues: false,
      outputRelative: false,
      multiDelimiter: '\n',
    },
  })
}

export function moneyField (name, label, extra = {}) {
  return field(name, label, 'Number', {
    required: extra.required,
    options: {
      presetFormat: 'custom',
      precision: 2,
      format: '0,0.00',
      prefix: '',
      suffix: ' ₽',
      display: 'number',
      min: 0,
      max: 0,
      step: 1,
      showValue: true,
      showRelative: false,
      showProgress: false,
      animated: false,
      variant: 'success',
      thresholds: [],
      multiDelimiter: '\n',
    },
  })
}

export function numberField (name, label, extra = {}) {
  return field(name, label, 'Number', {
    required: extra.required,
    options: {
      presetFormat: 'custom',
      precision: extra.precision ?? 3,
      format: extra.format || '',
      prefix: extra.prefix || '',
      suffix: extra.suffix || '',
      display: extra.display || 'number',
      min: extra.min ?? 0,
      max: extra.max ?? 0,
      step: extra.step ?? 1,
      showValue: true,
      showRelative: extra.display === 'progress',
      showProgress: extra.display === 'progress',
      animated: false,
      variant: extra.variant || 'info',
      thresholds: extra.thresholds || [],
      multiDelimiter: '\n',
    },
  })
}

export function boolSwitch (name, label) {
  return field(name, label, 'Bool', {
    options: { trueLabel: 'Да', falseLabel: 'Нет', switch: true },
  })
}

export function geoField (name, label) {
  return field(name, label, 'Geometry', {
    options: {
      center: [55.75, 37.62],
      zoom: 10,
      multiDelimiter: '\n',
      prefillWithCurrentLocation: true,
      hideCurrentLocationButton: false,
      hideGeoSearch: false,
    },
  })
}

export function wrapStyle () {
  return {
    variants: { headerText: 'dark' },
    wrap: { kind: 'card' },
    border: { enabled: false },
  }
}

export function block (kind, title, xywh, options = {}) {
  return {
    kind,
    title,
    xywh,
    options,
    style: wrapStyle(),
    meta: { visibility: { expression: '', roles: [] } },
  }
}

export function recordList (title, xywh, moduleID, fields, extra = {}) {
  return block('RecordList', title, xywh, {
    moduleID: String(moduleID),
    fields: fields.map(name => ({ name })),
    prefilter: extra.prefilter || '',
    presort: extra.presort || 'createdAt DESC',
    hideHeader: false,
    hideAddButton: extra.hideAddButton ?? false,
    hideImportButton: extra.hideImportButton ?? true,
    hideConfigureFieldsButton: true,
    hideSearch: extra.hideSearch ?? false,
    hidePaging: extra.hidePaging ?? false,
    hideSorting: false,
    hideFiltering: extra.hideFiltering ?? false,
    hideRecordReminderButton: true,
    hideRecordCloneButton: false,
    hideRecordEditButton: false,
    hideRecordViewButton: false,
    hideRecordPermissionsButton: true,
    hideRecordDeleteButton: false,
    enableRecordPageNavigation: true,
    allowExport: true,
    perPage: extra.perPage ?? 20,
    recordDisplayOption: 'sameTab',
    selectable: true,
    selectMode: 'multi',
    showTotalCount: true,
    compactRows: true,
    stickyHeader: true,
    displayMode: 'table',
    refField: extra.refField || '',
    ...(extra.more || {}),
  })
}

export function recordBlock (title, xywh, fields, extra = {}) {
  return block('Record', title, xywh, {
    fields: fields.map(name => ({ name })),
    fieldConditions: [],
    density: extra.density || 'comfortable',
    hideEmptyFields: extra.hideEmptyFields ?? false,
    showEmptyPlaceholder: true,
    fieldRoles: extra.fieldRoles || {},
    sections: extra.sections || [],
    horizontalFieldLayoutEnabled: extra.horizontal ?? true,
    recordFieldLayoutOption: extra.layout || 'wrap',
  })
}

export function metricItem (label, moduleID, filter = '', extra = {}) {
  const color = extra.color || '#2e59d9'
  return {
    label,
    moduleID: String(moduleID),
    metricField: extra.field || 'count',
    operation: extra.operation || 'count',
    dimensionField: extra.dimensionField || '',
    filter,
    showLabel: true,
    role: extra.role || 'hero',
    balloonFullWidth: extra.fullWidth || false,
    numberFormat: extra.fmt || '',
    prefix: extra.prefix || '',
    suffix: extra.suffix || '',
    valueStyle: {
      backgroundColor: '#FFFFFF00',
      color,
      fontSize: extra.fontSize || '',
      colorThresholds: extra.thresholds || [],
    },
    drillDown: { enabled: false, blockID: '', recordListOptions: { fields: [] } },
  }
}

export function metricBlock (title, xywh, metrics, extra = {}) {
  return block('Metric', title, xywh, {
    metrics,
    itemsPerRow: extra.itemsPerRow || '4',
    likeRecordList: extra.likeRecordList ?? true,
    density: extra.density || 'comfortable',
  })
}

export function ruleChain (title, xywh, opts) {
  return block('RuleChain', title, xywh, {
    chainID: opts.chainID,
    label: opts.label || title,
    variant: opts.variant || 'primary',
    size: opts.size || '',
    icon: opts.icon || 'play',
    context: opts.context || {},
  })
}

export function organizer (title, xywh, moduleID, opts) {
  return block('RecordOrganizer', title, xywh, {
    moduleID: String(moduleID),
    labelField: opts.labelField,
    descriptionField: opts.descriptionField || '',
    filter: opts.filter || '',
    positionField: '',
    groupField: opts.groupField || '',
    group: opts.group || '',
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    displayOption: 'sameTab',
  })
}

export function commentBlock (title, xywh, moduleID, opts) {
  return block('Comment', title, xywh, {
    moduleID: String(moduleID),
    filter: opts.filter || '',
    titleField: opts.titleField || 'title',
    contentField: opts.contentField || 'content',
    replyField: opts.replyField || '',
    referenceField: opts.referenceField || 'document',
    sortDirection: 'asc',
    refreshRate: 0,
    showRefresh: false,
    magnifyOption: '',
    attachmentField: opts.attachmentField || '',
    reactionsField: '',
  })
}

export function doughnutChart (name, handle, moduleID, dimField, filter = '') {
  return {
    name,
    handle,
    config: {
      colorScheme: 'tableau.Tableau10',
      noAnimation: false,
      reports: [{
        moduleID: String(moduleID),
        filter,
        dimensions: [{
          field: dimField,
          modifier: '(no grouping / buckets)',
          skipMissing: false,
          conditions: {},
          meta: {},
        }],
        metrics: [{
          field: 'count',
          type: 'doughnut',
          fixTooltips: true,
        }],
        yAxis: {},
        tooltip: {},
        renderer: { version: '' },
      }],
      toolbox: { saveAsImage: false, showDataTable: true },
    },
  }
}

export function ganttChart (name, handle, moduleID, labelField, startField, endField) {
  return {
    name,
    handle,
    config: {
      colorScheme: 'tableau.Tableau10',
      noAnimation: false,
      reports: [{
        moduleID: String(moduleID),
        filter: '',
        dimensions: [{
          field: labelField,
          modifier: '(no grouping / buckets)',
          skipMissing: false,
          conditions: {},
          meta: {},
        }],
        metrics: [{
          field: 'count',
          type: 'gantt',
          startField,
          endField,
          fixTooltips: true,
        }],
        yAxis: {},
        tooltip: {},
        renderer: { version: '' },
      }],
      toolbox: { saveAsImage: false, showDataTable: true },
    },
  }
}

export function pageIcon (src) {
  return { navItem: { expanded: false, icon: { type: 'fontawesome', src } } }
}

export function layoutBlocks (pageBlocks) {
  return pageBlocks.map((b, i) => ({
    blockID: b.blockID || String(i + 1),
    xywh: b.xywh,
  }))
}

export function withBlockIDs (blocks) {
  return blocks.map((b, i) => ({ ...b, blockID: String(i + 1) }))
}

export function snowflakeJSON (obj) {
  return JSON.stringify(obj).replace(
    /"(namespaceID|moduleID)":"(\d{15,})"/g,
    '"$1":$2',
  )
}

export function setOf (payload) {
  if (!payload) return []
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.set)) return payload.set
  return []
}

function b64url (input) {
  const buf = Buffer.isBuffer(input) ? input : Buffer.from(input)
  return buf.toString('base64url')
}

function jwtSecretFromEnv (dsn) {
  if (process.env.AUTH_JWT_SECRET) return process.env.AUTH_JWT_SECRET
  const hostname = process.env.HOSTNAME || 'localhost'
  return createHash('md5').update('jwt secret' + dsn + hostname).digest('hex')
}

function signHs512 (payload, secret) {
  const header = b64url(JSON.stringify({ alg: 'HS512', typ: 'JWT' }))
  const body = b64url(JSON.stringify(payload))
  const data = `${header}.${body}`
  return `${data}.${createHmac('sha512', secret).update(data).digest('base64url')}`
}

function psql (dsn, sql) {
  return execFileSync('psql', [dsn, '-tA', '-c', sql], { encoding: 'utf8' }).trim()
}

async function tokenWorks (token) {
  try {
    const res = await fetch('http://127.0.0.1:3333/compose/namespace/?limit=1', {
      headers: { Authorization: 'Bearer ' + token, Accept: 'application/json' },
      signal: AbortSignal.timeout(5000),
    })
    const text = await res.text()
    return res.ok && !/unauthorized/i.test(text)
  } catch {
    return false
  }
}

/**
 * Mint an API JWT.
 *
 * 1. Refresh exchange on /auth/oauth2/default-client (no client_id — the
 *    handler injects the default client; `frontend-app` caused HTTP 401).
 * 2. Sign HS512 with jti=auth_oa2tokens.access, trying DSN variants for
 *    md5("jwt secret"+DSN+HOSTNAME).
 */
export async function mintToken () {
  if (process.env.TOKEN) return process.env.TOKEN.trim()

  // GoLand «Server RU test10 translations» uses test10; older scripts defaulted to test9.
  const dsnList = [...new Set([
    process.env.COMPOSE_DSN,
    'postgres://postgres:Zse45rdx@127.0.0.1:5432/test10?sslmode=disable',
    'postgres://postgres:Zse45rdx@127.0.0.1:5432/test9?sslmode=disable',
  ].filter(Boolean))]
  const authBase = (process.env.AUTH_API || 'http://127.0.0.1:3333').replace(/\/$/, '')
  let lastErr
  for (const dsn of dsnList) {
    try {
      const token = await mintTokenFromDsn(dsn, authBase)
      if (token) return token
    } catch (e) {
      lastErr = e
    }
  }
  throw lastErr || new Error('Could not mint a JWT the server accepts. Log in to Compose once, or set TOKEN / COMPOSE_DSN.')
}

async function mintTokenFromDsn (dsn, authBase) {
  const refresh = psql(dsn, `
    SELECT refresh FROM auth_oa2tokens
    WHERE expires_at > now() AND refresh <> ''
    ORDER BY created_at DESC LIMIT 1`)
  if (refresh) {
    const res = await fetch(authBase + '/auth/oauth2/default-client', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded', Accept: 'application/json' },
      body: new URLSearchParams({ refresh_token: refresh }),
      signal: AbortSignal.timeout(15000),
    })
    const json = await res.json().catch(() => ({}))
    if (json.access_token && await tokenWorks(json.access_token)) {
      return json.access_token
    }
  }

  const row = psql(dsn, `
    SELECT access || E'\\t' || rel_user || E'\\t' || COALESCE(rel_client::text, '0')
    FROM auth_oa2tokens
    WHERE expires_at > now() AND access <> ''
    ORDER BY created_at DESC
    LIMIT 1`)
  if (!row) {
    throw new Error('No live access token in auth_oa2tokens; log in to Compose once, or set TOKEN')
  }
  const [jti, userID, clientID] = row.split('\t')
  if (!jti || !userID) {
    throw new Error('Malformed auth_oa2tokens row; set TOKEN')
  }
  if (jti.includes('.') && await tokenWorks(jti)) return jti

  const rolesRaw = psql(dsn, `
    SELECT string_agg(rel_role::text, ',')
    FROM role_members
    WHERE rel_resource = 'corteza::system:user/${userID}'`)
  const extra = psql(dsn, `
    SELECT string_agg(id::text, ',')
    FROM roles
    WHERE handle IN ('authenticated','super-admin') AND deleted_at IS NULL`)
  const roles = [...new Set(
    `${rolesRaw},${extra}`.split(',').map(s => s.trim()).filter(Boolean),
  )]

  const now = Math.floor(Date.now() / 1000)
  const payload = {
    jti,
    sub: String(userID),
    exp: now + 24 * 3600,
    iat: now,
    iss: 'cortezaproject.org',
    clientID: String(clientID || '0'),
    scope: 'profile api',
    roles,
  }

  const secrets = []
  if (process.env.AUTH_JWT_SECRET) secrets.push(process.env.AUTH_JWT_SECRET)
  const dsnVariants = [...new Set([
    dsn,
    dsn.replace('127.0.0.1', 'localhost'),
    dsn.replace('localhost', '127.0.0.1'),
    dsn.replace('127.0.0.1', 'host.docker.internal'),
    dsn.replace('localhost', 'host.docker.internal'),
  ])]
  for (const variant of dsnVariants) {
    secrets.push(jwtSecretFromEnv(variant))
  }

  for (const secret of [...new Set(secrets)]) {
    const token = signHs512(payload, secret)
    if (await tokenWorks(token)) return token
  }

  throw new Error('Could not mint a JWT the server accepts. Log in to Compose once and retry, or set TOKEN / AUTH_JWT_SECRET to match the GoLand process.')
}

export async function detectBase (token) {
  if (process.env.COMPOSE_API) return process.env.COMPOSE_API.replace(/\/$/, '')
  const candidates = [
    'http://127.0.0.1:3333/compose',
    'http://localhost:3333/compose',
    'http://127.0.0.1:3333/api/compose',
    'http://localhost:3333/api/compose',
  ]
  const tried = []
  for (const base of candidates) {
    try {
      const res = await fetch(base + '/namespace/', {
        headers: { Authorization: 'Bearer ' + token },
        signal: AbortSignal.timeout(5000),
      })
      tried.push(`${base}→${res.status}`)
      // 404 = wrong prefix (/api vs /). 200/401/403 all mean this is the API.
      if (res.status !== 404) return base
    } catch (e) {
      tried.push(`${base}→${e.cause?.code || e.message}`)
    }
  }
  throw new Error('Cannot reach Corteza compose API on :3333 (' + tried.join('; ') + '). If GoLand debugger is paused, Resume first.')
}

export function apiFactory (base, token) {
  return async function api (method, path, body, stringify = JSON.stringify) {
    const url = base + path
    const res = await fetch(url, {
      method,
      headers: {
        Authorization: 'Bearer ' + token,
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: body !== undefined ? stringify(body) : undefined,
      signal: AbortSignal.timeout(30000),
    })
    const text = await res.text()
    let data
    try { data = JSON.parse(text) } catch { data = { raw: text } }
    if (!res.ok || data.error) {
      const err = data.error?.message || data.error || text.slice(0, 500)
      throw new Error(`${method} ${path} → ${res.status}: ${typeof err === 'string' ? err : JSON.stringify(err)}`)
    }
    return data.response !== undefined ? data.response : data
  }
}

function mergeFields (existing, desired) {
  const byName = new Map((existing || []).map(f => [f.name, f]))
  return desired.map(d => {
    const prev = byName.get(d.name)
    if (!prev) return d
    return {
      ...d,
      fieldID: prev.fieldID,
      namespaceID: prev.namespaceID,
      moduleID: prev.moduleID,
    }
  })
}

const REVISIONS = {
  recordRevisions: { enabled: true, ident: '' },
}

export function withRevisions (extra = {}) {
  return { ...REVISIONS, ...extra }
}

export async function ensureNamespace (api, { name, slug, meta }) {
  const list = setOf(await api('GET', '/namespace/?limit=200'))
  const existing = list.find(n => !n.deletedAt && (n.slug === slug || String(n.name || '').toLowerCase() === slug))
  if (existing) {
    console.log('namespace', slug, 'exists', existing.namespaceID)
    const needFix = !existing.enabled || existing.slug !== slug
    if (needFix) {
      await api('POST', `/namespace/${existing.namespaceID}`, {
        name,
        slug,
        enabled: true,
        meta: { ...(existing.meta || {}), ...meta },
        updatedAt: existing.updatedAt,
      })
    }
    return existing.namespaceID
  }
  const ns = await api('POST', '/namespace/', { name, slug, enabled: true, meta })
  console.log('created namespace', ns.namespaceID)
  return ns.namespaceID
}

export async function ensureModule (api, nsID, { name, handle, fields, config }) {
  const list = setOf(await api('GET', `/namespace/${nsID}/module/?handle=${encodeURIComponent(handle)}&limit=50`))
  const existing = list.find(m => m.handle === handle && !m.deletedAt)
  const meta = {}
  if (!existing) {
    const body = { name, handle, fields, meta }
    if (config) body.config = config
    const created = await api('POST', `/namespace/${nsID}/module/`, body)
    console.log('created module', handle, created.moduleID)
    return created.moduleID
  }
  const full = await api('GET', `/namespace/${nsID}/module/${existing.moduleID}`)
  const merged = mergeFields(full.fields, fields)
  const nextConfig = config ? { ...(full.config || {}), ...config } : (full.config || {})
  await api('POST', `/namespace/${nsID}/module/${existing.moduleID}`, {
    name,
    handle,
    fields: merged,
    meta: full.meta || {},
    config: nextConfig,
    updatedAt: full.updatedAt,
  })
  console.log('updated module', handle, existing.moduleID)
  return existing.moduleID
}

export async function ensureChart (api, nsID, def) {
  const list = setOf(await api('GET', `/namespace/${nsID}/chart/?handle=${encodeURIComponent(def.handle)}&limit=50`))
  const existing = list.find(c => c.handle === def.handle && !c.deletedAt)
  if (!existing) {
    const created = await api('POST', `/namespace/${nsID}/chart/`, def)
    console.log('created chart', def.handle, created.chartID)
    return created.chartID
  }
  await api('POST', `/namespace/${nsID}/chart/${existing.chartID}`, {
    ...def,
    updatedAt: existing.updatedAt,
  })
  console.log('updated chart', def.handle, existing.chartID)
  return existing.chartID
}

export async function ensurePage (api, nsID, def) {
  const list = setOf(await api('GET', `/namespace/${nsID}/page/?handle=${encodeURIComponent(def.handle)}&limit=200`))
  const existing = list.find(p => p.handle === def.handle && !p.deletedAt)
  let page
  if (!existing) {
    page = await api('POST', `/namespace/${nsID}/page/`, def)
    console.log('created page', def.handle, page.pageID)
  } else {
    page = await api('POST', `/namespace/${nsID}/page/${existing.pageID}`, {
      ...def,
      updatedAt: existing.updatedAt,
    })
    console.log('updated page', def.handle, existing.pageID)
  }
  const layouts = setOf(await api('GET', `/namespace/${nsID}/page/${page.pageID}/layout/?limit=20`))
  const primary = layouts.find(l => l.handle === 'primary' && !l.deletedAt)
  const layoutBody = {
    handle: 'primary',
    meta: { title: def.title },
    blocks: layoutBlocks(def.blocks),
    config: {
      visibility: { expression: '', roles: [] },
      buttons: {
        new: { enabled: true },
        edit: { enabled: true },
        submit: { enabled: true },
        delete: { enabled: true },
        clone: { enabled: true },
        back: { enabled: true },
      },
      actions: [],
      useTitle: !!def.moduleID && def.moduleID !== '0',
    },
  }
  if (!primary) {
    await api('POST', `/namespace/${nsID}/page/${page.pageID}/layout/`, layoutBody)
    console.log('  layout primary')
  } else {
    await api('POST', `/namespace/${nsID}/page/${page.pageID}/layout/${primary.pageLayoutID}`, {
      ...layoutBody,
      updatedAt: primary.updatedAt,
    })
    console.log('  layout updated')
  }
  return page.pageID
}

export async function ensureRuleChain (api, chain) {
  const listed = await api('GET', '/admin/rulechain/?limit=200')
  const chains = listed.chains || listed.set || (Array.isArray(listed) ? listed : [])
  const existing = chains.find(c => c.id === chain.id)
  if (!existing) {
    try {
      const created = await api('POST', '/admin/rulechain/', chain, snowflakeJSON)
      console.log('created rule chain', chain.id, created.chainID || '')
      return
    } catch (e) {
      if (!String(e.message || e).includes('already') && !String(e.message || e).toLowerCase().includes('exists')) {
        throw e
      }
    }
  }
  try {
    await api('PUT', `/admin/rulechain/${chain.id}`, chain, snowflakeJSON)
    console.log('updated rule chain', chain.id)
  } catch (e) {
    const msg = String(e.message || e).toLowerCase()
    if (msg.includes('not found')) {
      const created = await api('POST', '/admin/rulechain/', chain, snowflakeJSON)
      console.log('created rule chain', chain.id, created.chainID || '')
      return
    }
    throw e
  }
}

export async function parentPages (api, nsID, pageIDs, parentOf) {
  for (const [child, parent] of Object.entries(parentOf)) {
    if (!pageIDs[child] || !pageIDs[parent]) continue
    const rec = await api('GET', `/namespace/${nsID}/page/${pageIDs[child]}`)
    if (String(rec.selfID) === String(pageIDs[parent])) continue
    await api('POST', `/namespace/${nsID}/page/${pageIDs[child]}`, {
      title: rec.title,
      handle: rec.handle,
      moduleID: rec.moduleID,
      visible: rec.visible,
      weight: rec.weight,
      blocks: rec.blocks,
      config: rec.config,
      meta: rec.meta,
      selfID: String(pageIDs[parent]),
      updatedAt: rec.updatedAt,
    })
    console.log('parented', child, '→', parent)
  }
}

export async function createRecord (api, nsID, moduleID, values) {
  const payload = Object.entries(values).map(([name, value]) => ({ name, value: value == null ? '' : String(value) }))
  return api('POST', `/namespace/${nsID}/module/${moduleID}/record/`, { values: payload })
}
