#!/usr/bin/env node
/**
 * Provision the CMDB Lowcode namespace (modules, charts, pages, layouts, rule chains).
 *
 *   COMPOSE_API=http://localhost:3333/api/compose \
 *   TOKEN=$(node /tmp/opencode/mint-token.mjs | head -1) \
 *   node apply.mjs
 *
 * Idempotent: existing resources are reused / updated by handle.
 * Pages: existing Builder geometry (xywh), extra blocks, titles, styles and
 * layout buttons are kept. New template blocks are appended. To restore the
 * YAML layout: APPLY_RESET_PAGES=1 or --reset-pages.
 */
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { writeFileSync } from 'node:fs'

const HERE = dirname(fileURLToPath(import.meta.url))

const DEVICE_TYPES = [
  ['router', 'Router'],
  ['switch', 'Switch'],
  ['server', 'Server'],
  ['workstation', 'Workstation'],
  ['printer', 'Printer'],
  ['camera', 'Camera'],
  ['firewall', 'Firewall'],
  ['iot', 'IoT'],
  ['unknown', 'Unknown'],
]

const DEVICE_STATUS = [
  ['online', 'Online', { backgroundColor: 'success', textColor: 'white' }],
  ['offline', 'Offline', { backgroundColor: 'secondary', textColor: 'white' }],
  ['unknown', 'Unknown', { backgroundColor: 'light', textColor: 'dark' }],
]

const SEVERITY = [
  ['CRITICAL', 'Critical', { backgroundColor: 'danger', textColor: 'white' }],
  ['HIGH', 'High', { backgroundColor: 'warning', textColor: 'dark' }],
  ['MEDIUM', 'Medium', { backgroundColor: 'info', textColor: 'white' }],
  ['LOW', 'Low', { backgroundColor: 'secondary', textColor: 'white' }],
  ['INFO', 'Info', { backgroundColor: 'light', textColor: 'dark' }],
]

const VULN_STATUS = [
  ['open', 'Open', { backgroundColor: 'danger', textColor: 'white' }],
  ['acknowledged', 'Acknowledged', { backgroundColor: 'warning', textColor: 'dark' }],
  ['fixed', 'Fixed', { backgroundColor: 'success', textColor: 'white' }],
  ['false_positive', 'False positive', { backgroundColor: 'light', textColor: 'dark' }],
]

const SCAN_STATUS = [
  ['pending', 'Pending', { backgroundColor: 'light', textColor: 'dark' }],
  ['running', 'Running', { backgroundColor: 'info', textColor: 'white' }],
  ['completed', 'Completed', { backgroundColor: 'success', textColor: 'white' }],
  ['failed', 'Failed', { backgroundColor: 'danger', textColor: 'white' }],
]

const PROTO = [
  ['tcp', 'TCP'],
  ['udp', 'UDP'],
]

const CRITICALITY = [
  ['low', 'Low'],
  ['medium', 'Medium'],
  ['high', 'High'],
  ['critical', 'Critical'],
]

function selectOptions (pairs) {
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

function field (name, label, kind, extra = {}) {
  return {
    name,
    label,
    kind,
    isRequired: !!extra.required,
    isMulti: !!extra.multi,
    options: extra.options || {},
  }
}

function recordRel (name, label, moduleID, labelField, queryFields, required = false) {
  return field(name, label, 'Record', {
    required,
    options: {
      moduleID: String(moduleID),
      labelField,
      queryFields,
      selectType: '',
      multiDelimiter: '\n',
    },
  })
}

function wrapStyle () {
  return {
    variants: { headerText: 'dark' },
    wrap: { kind: 'card' },
    border: { enabled: false },
  }
}

function block (kind, title, xywh, options = {}) {
  return {
    kind,
    title,
    xywh,
    options,
    style: wrapStyle(),
    meta: { visibility: { expression: '', roles: [] } },
  }
}

function recordList (title, xywh, moduleID, fields, extra = {}) {
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

function recordBlock (title, xywh, fields, extra = {}) {
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

function metricItem (label, moduleID, filter = '', extra = {}) {
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

function metricBlock (title, xywh, metrics, extra = {}) {
  return block('Metric', title, xywh, {
    metrics,
    itemsPerRow: extra.itemsPerRow || '4',
    likeRecordList: extra.likeRecordList ?? true,
    density: extra.density || 'comfortable',
  })
}

function ruleChain (title, xywh, opts) {
  return block('RuleChain', title, xywh, {
    chainID: opts.chainID,
    label: opts.label || title,
    variant: opts.variant || 'primary',
    size: opts.size || '',
    icon: opts.icon || 'play',
    context: opts.context || {},
  })
}

function organizer (title, xywh, moduleID, opts) {
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

function progressBlock (title, xywh, opts) {
  const empty = { default: 0, moduleID: '', filter: '', field: '', operation: '' }
  return block('Progress', title, xywh, {
    value: opts.value,
    minValue: opts.minValue || { ...empty, default: 0 },
    maxValue: opts.maxValue || { ...empty, default: 100 },
    display: {
      showValue: true,
      showRelative: true,
      showProgress: true,
      animated: opts.animated ?? true,
      variant: opts.variant || 'info',
      thresholds: opts.thresholds || [],
    },
    refreshRate: opts.refreshRate ?? 15,
    showRefresh: true,
    magnifyOption: '',
  })
}

function doughnutChart (name, handle, moduleID, dimField) {
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

function pageIcon (src) {
  return { navItem: { expanded: false, icon: { type: 'fontawesome', src } } }
}

function layoutBlocks (pageBlocks, existingLayoutBlocks = [], { reset = false } = {}) {
  const prev = new Map((existingLayoutBlocks || []).map(b => [String(b.blockID), b]))
  return pageBlocks.map((b, i) => {
    const id = String(b.blockID || i + 1)
    const old = prev.get(id)
    return {
      blockID: id,
      xywh: (!reset && Array.isArray(old?.xywh) && old.xywh.length === 4) ? old.xywh : b.xywh,
      ...(old?.meta ? { meta: old.meta } : {}),
    }
  })
}

function withBlockIDs (blocks) {
  return blocks.map((b, i) => ({ ...b, blockID: String(i + 1) }))
}

function applyResetPages () {
  return process.env.APPLY_RESET_PAGES === '1' || process.argv.includes('--reset-pages')
}

function blockIdentity (b) {
  const o = b?.options || {}
  const title = String(b?.title || '').trim().toLowerCase()
  const kind = String(b?.kind || '')
  if (kind === 'Chart' && o.chartID) return `chart:${o.chartID}`
  if (kind === 'RuleChain' && o.chainID) return `rulechain:${o.chainID}:${title}`
  if (kind === 'RecordList' && o.moduleID) return `recordlist:${o.moduleID}:${title}`
  if (kind === 'RecordOrganizer') return `organizer:${o.moduleID || ''}:${o.group || title}`
  if (kind === 'Progress') return `progress:${title}`
  if (kind === 'Metric') return `metric:${title}`
  if (kind === 'Record') return `record:${title}`
  return `${kind}:${title}`
}

function mergeBlockOptions (existing = {}, desired = {}) {
  const out = { ...desired, ...existing }
  for (const key of ['moduleID', 'chartID', 'chainID', 'refField']) {
    if (desired[key] !== undefined && desired[key] !== '') out[key] = desired[key]
  }
  return out
}

function mergeBlockMeta (existing = {}, desired = {}) {
  return {
    ...desired,
    ...existing,
    hidden: existing.hidden ?? desired.hidden,
    invisible: existing.invisible ?? desired.invisible,
    customID: existing.customID || desired.customID,
    customCSSClass: existing.customCSSClass || desired.customCSSClass,
    visibility: existing.visibility || desired.visibility || { expression: '', roles: [] },
    tempID: existing.tempID,
  }
}

function cloneXYWH (xywh, fallback) {
  const src = Array.isArray(xywh) && xywh.length === 4 ? xywh : fallback
  return Array.isArray(src) ? src.map(v => Number(v)) : [0, 0, 20, 15]
}

function mergePageBlocks (existingBlocks, desiredBlocks, { reset = false } = {}) {
  const existing = [...(existingBlocks || [])]
  const used = new Set()
  const sameLink = (have, want) => {
    const a = have?.options || {}
    const b = want?.options || {}
    if (want.kind === 'Chart') return String(a.chartID || '') === String(b.chartID || '')
    if (want.kind === 'RuleChain') return a.chainID === b.chainID
    if (want.kind === 'RecordList' || want.kind === 'RecordOrganizer') {
      return String(a.moduleID || '') === String(b.moduleID || '')
    }
    return false
  }

  const take = (want) => {
    const key = blockIdentity(want)
    let idx = existing.findIndex((b, i) => !used.has(i) && blockIdentity(b) === key)
    if (idx < 0) {
      const candidates = existing
        .map((b, i) => ({ b, i }))
        .filter(({ b, i }) => !used.has(i) && b.kind === want.kind && sameLink(b, want))
      if (candidates.length === 1) idx = candidates[0].i
    }
    if (idx < 0) return null
    used.add(idx)
    return existing[idx]
  }

  const merged = []
  for (const want of desiredBlocks || []) {
    const have = take(want)
    if (!have) {
      merged.push({ ...want, blockID: '0' })
      continue
    }
    merged.push({
      ...want,
      blockID: have.blockID,
      title: reset ? want.title : (have.title || want.title),
      description: reset ? (want.description || '') : (have.description || want.description || ''),
      xywh: reset ? cloneXYWH(want.xywh) : cloneXYWH(have.xywh, want.xywh),
      style: reset ? want.style : (have.style || want.style),
      meta: reset ? want.meta : mergeBlockMeta(have.meta, want.meta),
      options: reset ? want.options : mergeBlockOptions(have.options, want.options),
    })
  }

  existing.forEach((b, i) => {
    if (!used.has(i)) merged.push(b)
  })
  return merged
}

function mergePageConfig (existing, desired, { reset = false } = {}) {
  if (reset || !existing || !Object.keys(existing).length) return desired || {}
  return {
    ...desired,
    ...existing,
    navItem: existing.navItem || desired?.navItem,
  }
}

async function mintToken () {
  if (process.env.TOKEN) return process.env.TOKEN.trim()

  const dsn = process.env.COMPOSE_DSN || 'postgres://postgres:Zse45rdx@127.0.0.1:5432/test9?sslmode=disable'
  const refresh = execFileSync('psql', [dsn, '-tA', '-c',
    "SELECT refresh FROM auth_oa2tokens WHERE expires_at > now() ORDER BY created_at DESC LIMIT 1",
  ], { encoding: 'utf8' }).trim()
  if (!refresh) {
    throw new Error('No live refresh token in DB; set TOKEN')
  }

  const authBase = (process.env.AUTH_API || 'http://localhost:3333').replace(/\/$/, '')
  const res = await fetch(authBase + '/auth/oauth2/default-client', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({ refresh_token: refresh, client_id: 'frontend-app' }),
    signal: AbortSignal.timeout(15000),
  })
  const json = await res.json().catch(() => ({}))
  if (!json.access_token) {
    throw new Error('token exchange failed: HTTP ' + res.status)
  }
  return json.access_token
}

async function detectBase (token) {
  if (process.env.COMPOSE_API) return process.env.COMPOSE_API.replace(/\/$/, '')
  const candidates = [
    'http://localhost:3333/api/compose',
    'http://localhost:3333/compose',
  ]
  for (const base of candidates) {
    try {
      const res = await fetch(base + '/namespace/', {
        headers: { Authorization: 'Bearer ' + token },
        signal: AbortSignal.timeout(5000),
      })
      if (res.ok || res.status === 401 || res.status === 403) return base
    } catch {}
  }
  throw new Error('Cannot reach Lowcode compose API on :3333 (if GoLand debugger is paused, Resume the process first)')
}

function snowflakeJSON (obj) {
  return JSON.stringify(obj).replace(
    /"(namespaceID|moduleID)":"(\d{15,})"/g,
    '"$1":$2',
  )
}

function apiFactory (base, token) {
  return async function api (method, path, body, stringify = JSON.stringify) {
    const url = base + path
    const res = await fetch(url, {
      method,
      headers: {
        Authorization: 'Bearer ' + token,
        'Content-Type': 'application/json',
      },
      body: body !== undefined ? stringify(body) : undefined,
      signal: AbortSignal.timeout(20000),
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

function setOf (payload) {
  if (!payload) return []
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.set)) return payload.set
  return []
}

async function ensureNamespace (api) {
  const list = setOf(await api('GET', '/namespace/?limit=200'))
  const existing = list.find(n => !n.deletedAt && (n.slug === 'cmdb' || String(n.name || '').toLowerCase() === 'cmdb'))
  if (existing) {
    console.log('namespace cmdb exists', existing.namespaceID)
    const needFix = !existing.enabled || existing.slug !== 'cmdb' || existing.name === 'cmdb'
    if (needFix) {
      await api('POST', `/namespace/${existing.namespaceID}`, {
        name: 'CMDB',
        slug: 'cmdb',
        enabled: true,
        meta: existing.meta && Object.keys(existing.meta).length ? existing.meta : {
          subtitle: 'Configuration management database',
          description: 'Network inventory, discovered devices, services, vulnerabilities and scan history from the CMDB agent.',
        },
        updatedAt: existing.updatedAt,
      })
    }
    return existing.namespaceID
  }
  const ns = await api('POST', '/namespace/', {
    name: 'CMDB',
    slug: 'cmdb',
    enabled: true,
    meta: {
      subtitle: 'Configuration management database',
      description: 'Network inventory, discovered devices, services, vulnerabilities and scan history from the CMDB agent.',
      prompt: 'This is the CMDB namespace: networks (CIDR scan targets), devices (configuration items), services (open ports), vulnerabilities (findings), and scans. Help operators inventory assets, triage HIGH/CRITICAL findings, and trigger network scans.',
    },
  })
  console.log('created namespace', ns.namespaceID)
  return ns.namespaceID
}

function deviceFields () {
  return [
    field('ip_address', 'IP Address', 'String', { required: true }),
    field('mac_address', 'MAC Address', 'String'),
    field('hostname', 'Hostname', 'String'),
    field('vendor', 'Vendor', 'String'),
    field('device_type', 'Device Type', 'Select', { options: selectOptions(DEVICE_TYPES) }),
    field('os', 'Operating System', 'String'),
    field('domain', 'Domain', 'String'),
    field('open_ports', 'Open Ports', 'String', {
      options: {
        displayType: 'json',
        jsonLayout: 'chips',
        jsonTemplate: '{{port}}/{{proto}} {{service}}',
        jsonFields: 'port,proto,service',
        jsonVariantField: 'port',
        jsonVariants: [
          { value: '21', variant: 'danger' },
          { value: '23', variant: 'danger' },
          { value: '22', variant: 'success' },
          { value: '80', variant: 'info' },
          { value: '443', variant: 'success' },
          { value: '3389', variant: 'warning' },
          { value: '445', variant: 'warning' },
        ],
      },
    }),
    field('services', 'Services (JSON)', 'String'),
    field('shares', 'Shares', 'String'),
    field('vulnerabilities', 'Vulnerabilities (JSON)', 'String'),
    field('last_seen', 'Last Seen', 'DateTime'),
    field('status', 'Status', 'Select', { options: selectOptions(DEVICE_STATUS) }),
    field('criticality', 'Criticality', 'Select', { options: selectOptions(CRITICALITY) }),
    field('notes', 'Notes', 'String'),
  ]
}

function networkFields () {
  return [
    field('name', 'Name', 'String', { required: true }),
    field('cidr', 'CIDR', 'String', { required: true }),
    field('description', 'Description', 'String'),
    field('enabled', 'Enabled', 'Bool', { options: { trueLabel: 'Yes', falseLabel: 'No', switch: true } }),
    field('last_scan', 'Last Scan', 'DateTime'),
    field('notes', 'Notes', 'String'),
  ]
}

function serviceFields (devicesID) {
  return [
    recordRel('device', 'Device', devicesID, 'hostname', ['hostname', 'ip_address'], true),
    field('port', 'Port', 'Number'),
    field('proto', 'Protocol', 'Select', { options: selectOptions(PROTO) }),
    field('service', 'Service', 'String'),
    field('version', 'Version', 'String'),
    field('banner', 'Banner', 'String'),
  ]
}

function vulnFields (devicesID) {
  return [
    recordRel('device', 'Device', devicesID, 'hostname', ['hostname', 'ip_address'], true),
    field('name', 'Name', 'String', { required: true }),
    field('severity', 'Severity', 'Select', { options: selectOptions(SEVERITY) }),
    field('cve', 'CVE', 'String'),
    field('description', 'Description', 'String'),
    field('remediation', 'Remediation', 'String'),
    field('status', 'Status', 'Select', { options: selectOptions(VULN_STATUS) }),
    field('detected_at', 'Detected At', 'DateTime'),
  ]
}

function scanFields (networksID) {
  return [
    field('target', 'Target', 'String', { required: true }),
    recordRel('network', 'Network', networksID, 'name', ['name', 'cidr']),
    field('status', 'Status', 'Select', { options: selectOptions(SCAN_STATUS) }),
    field('progress', 'Progress', 'Number'),
    field('found', 'Found', 'Number'),
    field('started_at', 'Started At', 'DateTime'),
    field('finished_at', 'Finished At', 'DateTime'),
    field('scanning_ip', 'Scanning IP', 'String'),
    field('error', 'Error', 'String'),
  ]
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
      options: { ...(prev.options || {}), ...(d.options || {}) },
    }
  })
}

async function ensureModule (api, nsID, { name, handle, fields }) {
  const list = setOf(await api('GET', `/namespace/${nsID}/module/?handle=${encodeURIComponent(handle)}&limit=50`))
  const existing = list.find(m => m.handle === handle && !m.deletedAt)
  const meta = {}
  if (!existing) {
    const created = await api('POST', `/namespace/${nsID}/module/`, { name, handle, fields, meta })
    console.log('created module', handle, created.moduleID)
    return created.moduleID
  }
  const full = await api('GET', `/namespace/${nsID}/module/${existing.moduleID}`)
  const merged = mergeFields(full.fields, fields)
  await api('POST', `/namespace/${nsID}/module/${existing.moduleID}`, {
    name,
    handle,
    fields: merged,
    meta: full.meta || {},
    config: full.config,
    updatedAt: full.updatedAt,
  })
  console.log('updated module', handle, existing.moduleID)
  return existing.moduleID
}

async function ensureChart (api, nsID, def) {
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

async function ensurePage (api, nsID, def) {
  const reset = applyResetPages()
  const list = setOf(await api('GET', `/namespace/${nsID}/page/?handle=${encodeURIComponent(def.handle)}&limit=200`))
  const existing = list.find(p => p.handle === def.handle && !p.deletedAt)
  let page
  let layoutSource = def.blocks
  if (!existing) {
    page = await api('POST', `/namespace/${nsID}/page/`, def)
    layoutSource = page.blocks?.length ? page.blocks : def.blocks
    console.log('created page', def.handle, page.pageID)
  } else {
    const full = await api('GET', `/namespace/${nsID}/page/${existing.pageID}`)
    const blocks = mergePageBlocks(full.blocks || [], def.blocks, { reset })
    page = await api('POST', `/namespace/${nsID}/page/${existing.pageID}`, {
      title: reset ? def.title : (full.title || def.title),
      handle: def.handle,
      moduleID: def.moduleID || full.moduleID || '0',
      visible: reset ? !!def.visible : (full.visible ?? def.visible),
      weight: reset ? (def.weight ?? full.weight) : (full.weight ?? def.weight),
      description: reset ? (def.description || '') : (full.description || def.description || ''),
      config: mergePageConfig(full.config, def.config, { reset }),
      meta: full.meta || def.meta || {},
      selfID: full.selfID,
      blocks,
      updatedAt: full.updatedAt,
    })
    if (!page.blocks?.length) {
      page = await api('GET', `/namespace/${nsID}/page/${existing.pageID}`)
    }
    layoutSource = page.blocks?.length ? page.blocks : blocks
    console.log(reset ? 'reset page' : 'updated page (kept builder layout)', def.handle, existing.pageID)
  }
  const layouts = setOf(await api('GET', `/namespace/${nsID}/page/${page.pageID}/layout/?limit=20`))
  const primary = layouts.find(l => l.handle === 'primary' && !l.deletedAt)
  const defaultLayoutConfig = {
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
  }
  const layoutBody = {
    handle: 'primary',
    meta: { title: page.title || def.title },
    blocks: layoutBlocks(layoutSource, primary?.blocks, { reset }),
    config: (!reset && primary?.config) ? primary.config : defaultLayoutConfig,
  }
  if (!primary) {
    await api('POST', `/namespace/${nsID}/page/${page.pageID}/layout/`, layoutBody)
    console.log('  layout primary')
  } else {
    await api('POST', `/namespace/${nsID}/page/${page.pageID}/layout/${primary.pageLayoutID}`, {
      ...layoutBody,
      updatedAt: primary.updatedAt,
    })
    console.log(reset ? '  layout reset' : '  layout kept')
  }
  return page.pageID
}

async function ensureRuleChain (api, chain) {
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

async function seedNetworks (api, nsID, moduleID) {
  const list = setOf(await api('GET', `/namespace/${nsID}/module/${moduleID}/record/?limit=5`))
  if (list.length) {
    console.log('networks already have records, skip seed')
    return
  }
  const samples = [
    { name: 'Office LAN', cidr: '192.168.0.0/24', description: 'Default office subnet', enabled: '1', notes: 'Used by the CMDB agent auto-scan' },
    { name: 'Lab', cidr: '10.0.0.0/24', description: 'Lab / test network', enabled: '1' },
  ]
  for (const s of samples) {
    const values = Object.entries(s).map(([name, value]) => ({ name, value: String(value) }))
    await api('POST', `/namespace/${nsID}/module/${moduleID}/record/`, { values })
  }
  console.log('seeded', samples.length, 'network records')
}

function buildPages ({ nsID, modules, charts }) {
  const { networks, devices, services, vulnerabilities, scans } = modules
  const recID = '${recordID}'

  return [
    {
      title: 'Dashboard',
      handle: 'dashboard',
      visible: true,
      weight: 0,
      description: 'CMDB overview',
      config: pageIcon('fas chart-pie'),
      blocks: withBlockIDs([
        metricBlock('At a glance', [0, 0, 48, 16], [
          metricItem('Devices', devices, '', { role: 'hero', color: '#2e59d9', fontSize: '28' }),
          metricItem('Online', devices, "status = 'online'", { role: 'meta', color: '#1cc88a' }),
          metricItem('Offline', devices, "status = 'offline'", { role: 'meta', color: '#858796' }),
          metricItem('Open findings', vulnerabilities, "status = 'open'", {
            role: 'balloon',
            color: '#e74a3b',
            thresholds: [
              { value: 0, variant: 'success', icon: 'arrow-up' },
              { value: 1, variant: 'warning', icon: 'arrow-right' },
              { value: 5, variant: 'danger', icon: 'alert' },
            ],
          }),
          metricItem('Critical / High', vulnerabilities, "(severity = 'CRITICAL' OR severity = 'HIGH') AND status = 'open'", {
            role: 'balloon',
            color: '#f6c23e',
            thresholds: [
              { value: 0, variant: 'success', icon: 'arrow-up' },
              { value: 1, variant: 'danger', icon: 'alert' },
            ],
          }),
        ]),
        ruleChain('Scan', [0, 16, 16, 12], {
          chainID: 'cmdb-trigger-scan',
          label: 'Scan Office LAN',
          variant: 'primary',
          icon: 'search',
          context: { cidr: 'auto' },
        }),
        ruleChain('Hygiene', [16, 16, 16, 12], {
          chainID: 'cmdb-stale-devices',
          label: 'Find stale online devices',
          variant: 'warning',
          icon: 'clock',
          context: { query: "status = 'online'" },
        }),
        progressBlock('Running scans', [32, 16, 16, 12], {
          value: { default: 0, moduleID: String(scans), filter: "status = 'running'", field: 'count', operation: 'count' },
          maxValue: { default: 1, moduleID: String(scans), filter: '', field: 'count', operation: 'count' },
          variant: 'info',
        }),
        block('Chart', 'Devices by type', [0, 28, 16, 20], { chartID: String(charts.byType) }),
        block('Chart', 'Devices by status', [16, 28, 16, 20], { chartID: String(charts.byStatus) }),
        block('Chart', 'Vulnerabilities by severity', [32, 28, 16, 20], { chartID: String(charts.bySeverity) }),
        recordList('Open vulnerabilities', [0, 48, 24, 22], vulnerabilities, ['name', 'severity', 'cve', 'status'], {
          prefilter: "status = 'open'",
          perPage: 10,
          hideAddButton: true,
        }),
        recordList('Recent scans', [24, 48, 24, 22], scans, ['target', 'status', 'found', 'started_at'], {
          perPage: 10,
          hideAddButton: true,
        }),
        recordList('Recently seen devices', [0, 70, 48, 20], devices, ['ip_address', 'hostname', 'device_type', 'status', 'last_seen'], {
          presort: 'last_seen DESC',
          perPage: 15,
        }),
      ]),
    },
    {
      title: 'Devices',
      handle: 'devices',
      visible: true,
      weight: 10,
      config: pageIcon('fas server'),
      blocks: withBlockIDs([
        metricBlock('Inventory', [0, 0, 48, 12], [
          metricItem('Devices', devices, '', { role: 'hero', color: '#2e59d9', fontSize: '28' }),
          metricItem('Online', devices, "status = 'online'", { role: 'meta', color: '#1cc88a' }),
          metricItem('Offline', devices, "status = 'offline'", { role: 'meta', color: '#858796' }),
          metricItem('Unknown', devices, "status = 'unknown'", { role: 'meta', color: '#f6c23e' }),
        ]),
        recordList('Devices', [0, 12, 48, 42], devices, ['ip_address', 'hostname', 'device_type', 'vendor', 'os', 'status', 'last_seen']),
      ]),
    },
    {
      title: 'Device',
      handle: 'device',
      moduleID: String(devices),
      visible: false,
      weight: 11,
      config: pageIcon('fas server'),
      blocks: withBlockIDs([
        recordBlock('Device', [0, 0, 32, 30], [
          'hostname', 'ip_address', 'status', 'device_type', 'criticality',
          'vendor', 'os', 'domain', 'mac_address', 'last_seen',
          'open_ports', 'shares', 'notes',
        ], {
          fieldRoles: {
            hostname: 'title',
            ip_address: 'subtitle',
            status: 'badge',
            device_type: 'badge',
            criticality: 'badge',
            vendor: 'meta',
            os: 'meta',
            domain: 'meta',
            mac_address: 'meta',
            last_seen: 'meta',
            open_ports: 'body',
            shares: 'body',
            notes: 'body',
          },
          sections: [
            { title: 'Exposure', fields: ['open_ports', 'shares'] },
            { title: 'Notes', fields: ['notes'] },
          ],
        }),
        metricBlock('This host', [32, 0, 16, 16], [
          metricItem('Open vulns', vulnerabilities, `device = ${recID} AND status = 'open'`, { role: 'hero', color: '#e74a3b', fontSize: '28' }),
          metricItem('Services', services, `device = ${recID}`, { role: 'meta', color: '#2e59d9' }),
          metricItem('High / Critical', vulnerabilities, `device = ${recID} AND status = 'open' AND (severity = 'CRITICAL' OR severity = 'HIGH')`, { role: 'balloon', color: '#f6c23e' }),
        ], { itemsPerRow: '1' }),
        ruleChain('Actions', [32, 16, 16, 14], {
          chainID: 'cmdb-flag-insecure-ports',
          label: 'Flag insecure ports',
          variant: 'warning',
          icon: 'exclamation-triangle',
        }),
        recordList('Services', [0, 30, 24, 22], services, ['port', 'proto', 'service', 'version'], {
          prefilter: `device = ${recID}`,
          refField: 'device',
          perPage: 20,
        }),
        recordList('Vulnerabilities', [24, 30, 24, 22], vulnerabilities, ['name', 'severity', 'cve', 'status'], {
          prefilter: `device = ${recID}`,
          refField: 'device',
          perPage: 20,
        }),
      ]),
    },
    {
      title: 'Networks',
      handle: 'networks',
      visible: true,
      weight: 20,
      config: pageIcon('fas sitemap'),
      blocks: withBlockIDs([
        recordList('Networks', [0, 0, 48, 50], networks, ['name', 'cidr', 'enabled', 'last_scan']),
      ]),
    },
    {
      title: 'Network',
      handle: 'network',
      moduleID: String(networks),
      visible: false,
      weight: 21,
      config: pageIcon('fas sitemap'),
      blocks: withBlockIDs([
        recordBlock('Network', [0, 0, 32, 22], ['name', 'cidr', 'enabled', 'description', 'last_scan', 'notes'], {
          fieldRoles: {
            name: 'title',
            cidr: 'subtitle',
            enabled: 'badge',
            description: 'body',
            last_scan: 'meta',
            notes: 'body',
          },
          sections: [
            { title: 'Notes', fields: ['description', 'notes'] },
          ],
        }),
        ruleChain('Scan', [32, 0, 16, 12], {
          chainID: 'cmdb-trigger-scan',
          label: 'Scan this network',
          variant: 'primary',
          icon: 'search',
        }),
        metricBlock('History', [32, 12, 16, 10], [
          metricItem('Scans', scans, `network = ${recID}`, { role: 'hero', color: '#2e59d9' }),
        ], { itemsPerRow: '1' }),
        recordList('Scans', [0, 22, 48, 24], scans, ['target', 'status', 'found', 'started_at'], {
          prefilter: `network = ${recID}`,
          refField: 'network',
        }),
      ]),
    },
    {
      title: 'Vulnerabilities',
      handle: 'vulnerabilities',
      visible: true,
      weight: 30,
      config: pageIcon('fas bug'),
      blocks: withBlockIDs([
        organizer('Open', [0, 0, 12, 28], vulnerabilities, {
          labelField: 'name',
          descriptionField: 'cve',
          groupField: 'status',
          group: 'open',
        }),
        organizer('Acknowledged', [12, 0, 12, 28], vulnerabilities, {
          labelField: 'name',
          descriptionField: 'cve',
          groupField: 'status',
          group: 'acknowledged',
        }),
        organizer('Fixed', [24, 0, 12, 28], vulnerabilities, {
          labelField: 'name',
          descriptionField: 'cve',
          groupField: 'status',
          group: 'fixed',
        }),
        organizer('False positive', [36, 0, 12, 28], vulnerabilities, {
          labelField: 'name',
          descriptionField: 'cve',
          groupField: 'status',
          group: 'false_positive',
        }),
        recordList('All findings', [0, 28, 48, 28], vulnerabilities, ['name', 'device', 'severity', 'cve', 'status', 'detected_at']),
      ]),
    },
    {
      title: 'Vulnerability',
      handle: 'vulnerability',
      moduleID: String(vulnerabilities),
      visible: false,
      weight: 31,
      config: pageIcon('fas bug'),
      blocks: withBlockIDs([
        recordBlock('Finding', [0, 0, 32, 30], ['name', 'cve', 'severity', 'status', 'device', 'detected_at', 'description', 'remediation'], {
          fieldRoles: {
            name: 'title',
            cve: 'subtitle',
            severity: 'badge',
            status: 'badge',
            device: 'meta',
            detected_at: 'meta',
            description: 'body',
            remediation: 'body',
          },
          sections: [
            { title: 'Details', fields: ['description'] },
            { title: 'Remediation', fields: ['remediation'] },
          ],
        }),
        ruleChain('Alert', [32, 0, 16, 10], {
          chainID: 'cmdb-high-vuln-alert',
          label: 'Notify if HIGH / CRITICAL',
          variant: 'danger',
          icon: 'envelope',
          context: { to: 'ops@localhost' },
        }),
        ruleChain('Triage', [32, 10, 16, 10], {
          chainID: 'cmdb-ack-finding',
          label: 'Acknowledge',
          variant: 'warning',
          icon: 'check',
        }),
        ruleChain('Close', [32, 20, 16, 10], {
          chainID: 'cmdb-close-finding',
          label: 'Mark as fixed',
          variant: 'success',
          icon: 'check-double',
        }),
      ]),
    },
    {
      title: 'Services',
      handle: 'services',
      visible: true,
      weight: 40,
      config: pageIcon('fas plug'),
      blocks: withBlockIDs([
        recordList('Services', [0, 0, 48, 50], services, ['device', 'port', 'proto', 'service', 'version']),
      ]),
    },
    {
      title: 'Service',
      handle: 'service',
      moduleID: String(services),
      visible: false,
      weight: 41,
      config: pageIcon('fas plug'),
      blocks: withBlockIDs([
        recordBlock('Service', [0, 0, 32, 22], ['service', 'port', 'proto', 'device', 'version', 'banner'], {
          fieldRoles: {
            service: 'title',
            port: 'subtitle',
            proto: 'badge',
            device: 'meta',
            version: 'meta',
            banner: 'body',
          },
          sections: [
            { title: 'Banner', fields: ['banner'] },
          ],
        }),
        ruleChain('Actions', [32, 0, 16, 14], {
          chainID: 'cmdb-insecure-service',
          label: 'Flag if Telnet / FTP',
          variant: 'warning',
          icon: 'exclamation-triangle',
        }),
      ]),
    },
    {
      title: 'Scans',
      handle: 'scans',
      visible: true,
      weight: 50,
      config: pageIcon('fas search'),
      blocks: withBlockIDs([
        metricBlock('Activity', [0, 0, 20, 12], [
          metricItem('Scans', scans, '', { role: 'hero', color: '#2e59d9' }),
          metricItem('Running', scans, "status = 'running'", { role: 'meta', color: '#36b9cc' }),
          metricItem('Failed', scans, "status = 'failed'", { role: 'balloon', color: '#e74a3b' }),
        ], { itemsPerRow: 'auto' }),
        ruleChain('Start scan', [20, 0, 16, 12], {
          chainID: 'cmdb-trigger-scan',
          label: 'Start scan',
          variant: 'primary',
          icon: 'search',
          context: { cidr: 'auto' },
        }),
        progressBlock('In flight', [36, 0, 12, 12], {
          value: { default: 0, moduleID: String(scans), filter: "status = 'running'", field: 'count', operation: 'count' },
          maxValue: { default: 1, moduleID: String(scans), filter: '', field: 'count', operation: 'count' },
        }),
        recordList('Scans', [0, 12, 48, 40], scans, ['target', 'network', 'status', 'progress', 'found', 'started_at', 'finished_at']),
      ]),
    },
    {
      title: 'Scan',
      handle: 'scan',
      moduleID: String(scans),
      visible: false,
      weight: 51,
      config: pageIcon('fas search'),
      blocks: withBlockIDs([
        recordBlock('Scan', [0, 0, 32, 26], ['target', 'status', 'network', 'progress', 'found', 'scanning_ip', 'started_at', 'finished_at', 'error'], {
          fieldRoles: {
            target: 'title',
            status: 'badge',
            network: 'subtitle',
            progress: 'meta',
            found: 'meta',
            scanning_ip: 'meta',
            started_at: 'meta',
            finished_at: 'meta',
            error: 'body',
          },
          sections: [
            { title: 'Error', fields: ['error'] },
          ],
        }),
        progressBlock('Progress', [32, 0, 16, 14], {
          value: {
            default: 0,
            moduleID: String(scans),
            filter: `recordID = ${recID}`,
            field: 'progress',
            operation: 'max',
          },
          maxValue: { default: 100, moduleID: '', filter: '', field: '', operation: '' },
          animated: true,
          variant: 'success',
        }),
        metricBlock('Found', [32, 14, 16, 12], [
          metricItem('Hosts', scans, `recordID = ${recID}`, { role: 'hero', field: 'found', operation: 'max', color: '#2e59d9' }),
        ], { itemsPerRow: '1' }),
      ]),
    },
  ]
}

function buildRuleChains ({ nsID, modules, agentUrl }) {
  const ns = String(nsID)
  const devices = String(modules.devices)
  const services = String(modules.services)
  const vulns = String(modules.vulnerabilities)
  const scans = String(modules.scans)

  return [
    {
      id: 'cmdb-trigger-scan',
      name: 'CMDB: trigger network scan',
      description: 'POST cidr to scan-cidr (or CMDB_AGENT_URL) and create a scans row. Needs cidr from the network record or block context.',
      entryNode: 'record_scan',
      nodes: [
        {
          id: 'record_scan',
          type: 'crud',
          label: 'Create scan record',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: scans,
            moduleHandle: 'scans',
            fields: {
              target: '{{cidr}}',
              network: '{{recordID}}',
              status: 'running',
              progress: '0',
              found: '0',
            },
          },
        },
        {
          id: 'http_scan',
          type: 'http',
          label: 'Start agent scan',
          config: {
            url: agentUrl + '/scan',
            method: 'POST',
            body: '{"cidr":"{{cidr}}","namespaceID":"{{namespaceID}}","token":"{{authToken}}","scanRecordID":"{{createdRecordID}}","callbackUrl":"{{callbackUrl}}"}',
            timeout: 30,
          },
        },
        {
          id: 'detach_poll',
          type: 'detach',
          label: 'Poll agent if no webhook',
          config: {
            kind: 'poll',
            ingestChainID: 'cmdb-ingest-scan',
            statusUrl: agentUrl + '/jobs/{{scanID}}',
            itemsUrl: agentUrl + '/jobs/{{scanID}}/items',
            interval: 2,
            timeout: 900,
            until: 'done,error,completed,failed',
          },
        },
      ],
      edges: [
        { from: 'record_scan', to: 'http_scan' },
        { from: 'http_scan', to: 'detach_poll' },
      ],
    },
    {
      id: 'cmdb-ingest-scan',
      name: 'CMDB: ingest agent job',
      description: 'Webhook/poll envelope → update scans row, upsert devices, services and vulnerabilities.',
      entryNode: 'update_scan',
      nodes: [
        {
          id: 'update_scan',
          type: 'crud',
          label: 'Update scan record',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: scans,
            moduleHandle: 'scans',
            recordID: '{{scanRecordID}}',
            omitEmpty: true,
            continueOnError: true,
            fields: {
              status: '{{status}}',
              progress: '{{progress}}',
              found: '{{found}}',
              error: '{{error}}',
              scanning_ip: '{{scanningIP}}',
              target: '{{target}}',
              started_at: '{{startedAt}}',
              finished_at: '{{finishedAt}}',
            },
          },
        },
        {
          id: 'foreach_items',
          type: 'foreach',
          label: 'Each device',
          config: { items: 'items', itemVar: 'item' },
        },
        {
          id: 'upsert_device',
          type: 'crud.upsert',
          label: 'Upsert device',
          config: {
            namespaceID: ns,
            moduleID: devices,
            moduleHandle: 'devices',
            matchBy: ['mac_address', 'ip_address', 'hostname'],
            omitEmpty: true,
            resultVar: 'deviceRecordID',
            fields: {
              ip_address: '{{item.ip}}',
              mac_address: '{{item.mac}}',
              hostname: '{{item.hostname}}',
              vendor: '{{item.vendor}}',
              device_type: '{{item.deviceType}}',
              os: '{{item.os}}',
              domain: '{{item.domain}}',
              open_ports: '{{item.openPorts}}',
              services: '{{item.services}}',
              shares: '{{item.shares}}',
              vulnerabilities: '{{item.vulnerabilities}}',
              last_seen: '{{item.lastSeen}}',
              status: '{{item.status}}',
            },
          },
        },
        {
          id: 'foreach_ports',
          type: 'foreach',
          label: 'Each open port',
          config: { items: 'item.openPorts', itemVar: 'port' },
        },
        {
          id: 'upsert_service',
          type: 'crud.upsert',
          label: 'Upsert service',
          config: {
            namespaceID: ns,
            moduleID: services,
            moduleHandle: 'services',
            matchBy: ['device', 'port', 'proto'],
            matchAll: true,
            omitEmpty: true,
            continueOnError: true,
            fields: {
              device: '{{deviceRecordID}}',
              port: '{{port.port}}',
              proto: '{{port.proto}}',
              service: '{{port.service}}',
              version: '{{port.version}}',
              banner: '{{port.banner}}',
            },
          },
        },
        {
          id: 'foreach_vulns',
          type: 'foreach',
          label: 'Each vulnerability',
          config: { items: 'item.vulnerabilities', itemVar: 'vuln' },
        },
        {
          id: 'upsert_vuln',
          type: 'crud.upsert',
          label: 'Upsert vulnerability',
          config: {
            namespaceID: ns,
            moduleID: vulns,
            moduleHandle: 'vulnerabilities',
            matchBy: ['device', 'name'],
            matchAll: true,
            omitEmpty: true,
            continueOnError: true,
            fields: {
              device: '{{deviceRecordID}}',
              name: '{{vuln.name}}',
              severity: '{{vuln.severity}}',
              cve: '{{vuln.cve}}',
              description: '{{vuln.description}}',
              remediation: '{{vuln.remediation}}',
              status: 'open',
              detected_at: '{{item.lastSeen}}',
            },
          },
        },
      ],
      edges: [
        { from: 'update_scan', to: 'foreach_items' },
        { from: 'foreach_items', to: 'upsert_device' },
        { from: 'foreach_items', to: 'foreach_ports' },
        { from: 'foreach_ports', to: 'upsert_service' },
        { from: 'foreach_items', to: 'foreach_vulns' },
        { from: 'foreach_vulns', to: 'upsert_vuln' },
      ],
    },
    {
      id: 'cmdb-high-vuln-alert',
      name: 'CMDB: high / critical vulnerability alert',
      description: 'Mail when the finding severity is HIGH or CRITICAL. Pass to (email) in block context.',
      entryNode: 'fork',
      nodes: [
        { id: 'fork', type: 'fork', label: 'HIGH or CRITICAL', config: { branches: 2 } },
        { id: 'is_high', type: 'condition', label: 'Severity HIGH', config: { field: 'severity', operator: 'eq', value: 'HIGH' } },
        { id: 'is_crit', type: 'condition', label: 'Severity CRITICAL', config: { field: 'severity', operator: 'eq', value: 'CRITICAL' } },
        {
          id: 'mail',
          type: 'mail',
          label: 'Notify',
          config: {
            to: '{{to}}',
            subject: '[CMDB] {{severity}} {{name}}',
            body: '<p>Severity: <b>{{severity}}</b></p><p>{{name}}</p><p>Device: {{device}}</p><p>CVE: {{cve}}</p>',
            contentType: 'html',
          },
        },
      ],
      edges: [
        { from: 'fork', to: 'is_high' },
        { from: 'fork', to: 'is_crit' },
        { from: 'is_high', to: 'mail', condition: 'is_high_result' },
        { from: 'is_crit', to: 'mail', condition: 'is_crit_result' },
      ],
    },
    {
      id: 'cmdb-insecure-service',
      name: 'CMDB: flag insecure service',
      description: 'Create an open HIGH finding when the service name looks like Telnet or FTP.',
      entryNode: 'fork',
      nodes: [
        { id: 'fork', type: 'fork', label: 'Telnet or FTP', config: { branches: 2 } },
        { id: 'is_telnet', type: 'condition', label: 'Telnet', config: { field: 'service', operator: 'contains', value: 'telnet' } },
        { id: 'is_ftp', type: 'condition', label: 'FTP', config: { field: 'service', operator: 'contains', value: 'ftp' } },
        {
          id: 'create_finding',
          type: 'crud',
          label: 'Create finding',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: vulns,
            moduleHandle: 'vulnerabilities',
            fields: {
              device: '{{device}}',
              name: 'Insecure service: {{service}}',
              severity: 'HIGH',
              description: '{{service}} on port {{port}} transmits data in cleartext.',
              remediation: 'Disable the service and use an encrypted alternative (SSH/SFTP).',
              status: 'open',
            },
          },
        },
      ],
      edges: [
        { from: 'fork', to: 'is_telnet' },
        { from: 'fork', to: 'is_ftp' },
        { from: 'is_telnet', to: 'create_finding', condition: 'is_telnet_result' },
        { from: 'is_ftp', to: 'create_finding', condition: 'is_ftp_result' },
      ],
    },
    {
      id: 'cmdb-flag-insecure-ports',
      name: 'CMDB: flag insecure ports on a device',
      description: 'On a device record, create a HIGH finding if open_ports/services mention Telnet (23) or FTP (21).',
      entryNode: 'fork',
      nodes: [
        { id: 'fork', type: 'fork', label: 'Port 21/23 or telnet/ftp', config: { branches: 4 } },
        { id: 'port_23', type: 'condition', label: 'TCP 23', config: { field: 'open_ports', operator: 'contains', value: '"port":23,' } },
        { id: 'port_21', type: 'condition', label: 'TCP 21', config: { field: 'open_ports', operator: 'contains', value: '"port":21,' } },
        { id: 'svc_telnet', type: 'condition', label: 'Telnet in services', config: { field: 'services', operator: 'contains', value: 'telnet' } },
        { id: 'svc_ftp', type: 'condition', label: 'FTP in services', config: { field: 'services', operator: 'contains', value: 'ftp' } },
        {
          id: 'create_finding',
          type: 'crud',
          label: 'Create finding',
          config: {
            operation: 'create',
            namespaceID: ns,
            moduleID: vulns,
            moduleHandle: 'vulnerabilities',
            fields: {
              device: '{{recordID}}',
              name: 'Insecure cleartext service on {{hostname}}',
              severity: 'HIGH',
              description: 'Host {{ip_address}} exposes Telnet and/or FTP (ports 21/23).',
              remediation: 'Disable the service and use an encrypted alternative (SSH/SFTP).',
              status: 'open',
            },
          },
        },
      ],
      edges: [
        { from: 'fork', to: 'port_23' },
        { from: 'fork', to: 'port_21' },
        { from: 'fork', to: 'svc_telnet' },
        { from: 'fork', to: 'svc_ftp' },
        { from: 'port_23', to: 'create_finding', condition: 'port_23_result' },
        { from: 'port_21', to: 'create_finding', condition: 'port_21_result' },
        { from: 'svc_telnet', to: 'create_finding', condition: 'svc_telnet_result' },
        { from: 'svc_ftp', to: 'create_finding', condition: 'svc_ftp_result' },
      ],
    },
    {
      id: 'cmdb-stale-devices',
      name: 'CMDB: list stale online devices',
      description: 'Search devices. Pass query (Compose QL) in block context, e.g. status = \'online\'.',
      entryNode: 'search',
      nodes: [
        {
          id: 'search',
          type: 'crud',
          label: 'Search devices',
          config: {
            operation: 'search',
            namespaceID: ns,
            moduleID: devices,
            moduleHandle: 'devices',
            query: '{{query}}',
            limit: 100,
          },
        },
      ],
      edges: [],
    },
    {
      id: 'cmdb-ack-finding',
      name: 'CMDB: acknowledge finding',
      description: 'Set the current vulnerability record status to acknowledged.',
      entryNode: 'upd',
      nodes: [
        {
          id: 'upd',
          type: 'crud',
          label: 'Acknowledge',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: vulns,
            moduleHandle: 'vulnerabilities',
            recordID: '{{recordID}}',
            fields: { status: 'acknowledged' },
          },
        },
      ],
      edges: [],
    },
    {
      id: 'cmdb-close-finding',
      name: 'CMDB: mark finding fixed',
      description: 'Set the current vulnerability record status to fixed.',
      entryNode: 'upd',
      nodes: [
        {
          id: 'upd',
          type: 'crud',
          label: 'Mark fixed',
          config: {
            operation: 'update',
            namespaceID: ns,
            moduleID: vulns,
            moduleHandle: 'vulnerabilities',
            recordID: '{{recordID}}',
            fields: { status: 'fixed' },
          },
        },
      ],
      edges: [],
    },
  ].map(c => ({ ...c, namespaceID: String(nsID) }))
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const agentUrl = (process.env.CMDB_AGENT_URL || 'http://localhost:8089/api').replace(/\/$/, '')

  console.log('API', base)

  const nsID = await ensureNamespace(api)

  const networks = await ensureModule(api, nsID, { name: 'Networks', handle: 'networks', fields: networkFields() })
  const devices = await ensureModule(api, nsID, { name: 'Devices', handle: 'devices', fields: deviceFields() })
  const services = await ensureModule(api, nsID, { name: 'Services', handle: 'services', fields: serviceFields(devices) })
  const vulnerabilities = await ensureModule(api, nsID, { name: 'Vulnerabilities', handle: 'vulnerabilities', fields: vulnFields(devices) })
  const scans = await ensureModule(api, nsID, { name: 'Scans', handle: 'scans', fields: scanFields(networks) })

  const modules = { networks, devices, services, vulnerabilities, scans }

  const byType = await ensureChart(api, nsID, doughnutChart('Devices by type', 'devices-by-type', devices, 'device_type'))
  const byStatus = await ensureChart(api, nsID, doughnutChart('Devices by status', 'devices-by-status', devices, 'status'))
  const bySeverity = await ensureChart(api, nsID, doughnutChart('Vulnerabilities by severity', 'vulns-by-severity', vulnerabilities, 'severity'))
  const charts = { byType, byStatus, bySeverity }

  const pageIDs = {}
  for (const page of buildPages({ nsID, modules, charts })) {
    pageIDs[page.handle] = await ensurePage(api, nsID, page)
  }

  const parentOf = {
    device: 'devices',
    network: 'networks',
    vulnerability: 'vulnerabilities',
    service: 'services',
    scan: 'scans',
  }
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

  for (const chain of buildRuleChains({ nsID, modules, agentUrl })) {
    await ensureRuleChain(api, chain)
  }

  try {
    await seedNetworks(api, nsID, networks)
  } catch (e) {
    console.warn('network seed skipped:', e.message)
  }

  const summary = {
    namespaceID: String(nsID),
    slug: 'cmdb',
    api: base,
    modules: Object.fromEntries(Object.entries(modules).map(([k, v]) => [k, String(v)])),
    charts: Object.fromEntries(Object.entries(charts).map(([k, v]) => [k, String(v)])),
    urls: {
      namespace: `/ns/cmdb`,
      dashboard: `/ns/cmdb/pages`,
    },
    agent: {
      flags: `--db=lowcode --api=${base.replace(/\/compose$/, '')} --namespace=${nsID}`,
      scan: `POST ${agentUrl}/scan {"cidr":"192.168.1.0/24"}`,
    },
  }
  writeFileSync(join(HERE, 'applied.json'), JSON.stringify(summary, null, 2))
  console.log('\nCMDB ready')
  console.log(JSON.stringify(summary, null, 2))
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
