/**
 * Demo contour for the DSPM namespace. Reuses discovered stores and datasets.
 * Idempotent by name (and by dataset+class / identity+dataset for links).
 *
 *   COMPOSE_API=http://127.0.0.1:3333/compose AUTH_API=http://127.0.0.1:3333 \
 *   COMPOSE_DSN=postgres://... node seed.mjs
 */
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mintToken, detectBase, apiFactory, setOf, createRecord } from '../../invest/compose/helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
const ns = applied.namespaceID
const m = applied.modules

function recID (row) {
  return String(row.recordID || row.ID || '')
}

function recVal (row, name) {
  const hit = (row.values || []).find(v => v.name === name)
  return hit ? String(hit.value ?? '') : ''
}

function userFromToken (token) {
  try {
    return String(JSON.parse(Buffer.from(String(token).split('.')[1], 'base64url')).sub || '')
  } catch {
    return ''
  }
}

async function listAll (api, moduleID) {
  const rows = []
  for (let page = 1; page <= 20; page++) {
    const set = setOf(await api('GET', `/namespace/${ns}/module/${moduleID}/record/?limit=200&page=${page}`))
    rows.push(...set)
    if (set.length < 200) break
  }
  return rows
}

async function patch (api, moduleID, recordID, fields) {
  const rec = await api('GET', `/namespace/${ns}/module/${moduleID}/record/${recordID}`)
  const current = {}
  for (const v of rec.values || []) {
    if (v.name) current[v.name] = v.value == null ? '' : String(v.value)
  }
  let changed = false
  for (const [k, v] of Object.entries(fields)) {
    if (String(current[k] ?? '') !== String(v)) {
      current[k] = String(v)
      changed = true
    }
  }
  if (!changed) return
  const values = Object.entries(current)
    .filter(([, v]) => v !== '' && v != null)
    .map(([name, value]) => ({ name, value }))
  await api('POST', `/namespace/${ns}/module/${moduleID}/record/${recordID}`, {
    values,
    updatedAt: rec.updatedAt,
  })
}

function indexBy (rows, field) {
  const map = new Map()
  for (const row of rows) {
    const key = recVal(row, field)
    if (key && !map.has(key)) map.set(key, recID(row))
  }
  return map
}

async function ensure (api, moduleID, keyField, key, values, cache) {
  if (cache.has(key)) return cache.get(key)
  const id = recID(await createRecord(api, ns, moduleID, values))
  cache.set(key, id)
  return id
}

function pickDataset (datasets, storesByID, storeName, hints) {
  const wanted = new Set(hints)
  const inStore = datasets.filter(d => storesByID.get(recVal(d, 'data_store')) === storeName)
  const hit = inStore.find(d => wanted.has(recVal(d, 'name')))
  return hit || inStore[0] || null
}

async function main () {
  const token = await mintToken()
  const api = apiFactory(await detectBase(token), token)
  const userID = userFromToken(token)
  const now = new Date().toISOString()
  const yesterday = new Date(Date.now() - 86400000).toISOString()

  const stores = await listAll(api, m.data_stores)
  const datasets = await listAll(api, m.datasets)
  const classes = await listAll(api, m.data_classes)
  if (!stores.length || !datasets.length) {
    throw new Error('Сначала нажмите «Обнаружение»: нет хранилищ или датасетов')
  }
  const classByCode = indexBy(classes, 'code')
  const storeNameByID = new Map(stores.map(s => [recID(s), recVal(s, 'name')]))

  const sensitivity = { test12: 'restricted', test9: 'confidential', corteza: 'internal' }
  const exposure = { test12: 'internal', test9: 'org_wide', corteza: 'internal' }
  for (const store of stores) {
    const name = recVal(store, 'name')
    if (!sensitivity[name]) continue
    await patch(api, m.data_stores, recID(store), {
      sensitivity: sensitivity[name],
      exposure: exposure[name],
      criticality: name === 'test12' ? 'high' : 'medium',
    })
  }

  const ds = (store, hints) => {
    const row = pickDataset(datasets, storeNameByID, store, hints)
    if (!row) throw new Error(`нет датасета в ${store}`)
    return { id: recID(row), name: recVal(row, 'name'), path: recVal(row, 'path'), store }
  }
  const users = ds('test12', ['users', 'compose_user', 'auth_users'])
  const sessions = ds('test12', ['auth_sessions', 'sessions'])
  const tokens = ds('test12', ['auth_oa2tokens', 'tokens'])
  const clients = ds('test9', ['auth_clients', 'clients'])
  const attachments = ds('test9', ['attachments', 'compose_attachment'])

  const masks = indexBy(await listAll(api, m.masking_profiles), 'name')
  const filters = indexBy(await listAll(api, m.row_filters), 'name')
  const maskEmail = await ensure(api, m.masking_profiles, 'name', 'Частичная почта', {
    name: 'Частичная почта', method: 'partial', pattern: 'a***@domain', description: 'Оставляет домен, скрывает локальную часть',
  }, masks)
  const maskPii = await ensure(api, m.masking_profiles, 'name', 'Редакция ПДн', {
    name: 'Редакция ПДн', method: 'redact', pattern: '[REDACTED]', description: 'Полностью скрывает персональные поля',
  }, masks)
  const maskHash = await ensure(api, m.masking_profiles, 'name', 'Хеш секрета', {
    name: 'Хеш секрета', method: 'hash', pattern: 'sha256', description: 'Вместо токена отдаёт хеш',
  }, masks)

  const filterDept = await ensure(api, m.row_filters, 'name', 'Только своё подразделение', {
    name: 'Только своё подразделение', expression: "department = current_user.department", description: 'Строки чужого подразделения не видны',
  }, filters)
  await ensure(api, m.row_filters, 'name', 'Только prod', {
    name: 'Только prod', expression: "environment = 'prod'", description: 'Сервисам закрыт доступ к prod-строкам вне окна',
  }, filters)

  const policies = indexBy(await listAll(api, m.policies), 'name')
  const polPii = await ensure(api, m.policies, 'name', 'ПДн: маска для AI', {
    name: 'ПДн: маска для AI', mode: 'mask', data_class: classByCode.get('pii') || '', identity_kind: 'ai_agent',
    masking_profile: maskPii, row_filter: filterDept, enabled: '1',
    description: 'AI-агенты видят персональные данные только в замаскированном виде',
  }, policies)
  const polPci = await ensure(api, m.policies, 'name', 'Платежи: блок сервисных учёток', {
    name: 'Платежи: блок сервисных учёток', mode: 'block', data_class: classByCode.get('pci') || '', identity_kind: 'service',
    masking_profile: maskHash, enabled: '1',
    description: 'Сервисные учётки не читают платёжные поля',
  }, policies)
  const polCred = await ensure(api, m.policies, 'name', 'Секреты: наблюдение', {
    name: 'Секреты: наблюдение', mode: 'monitor', data_class: classByCode.get('credentials') || '', identity_kind: 'any',
    masking_profile: maskHash, enabled: '1',
    description: 'Доступ к токенам фиксируется, исполнение ещё не включено',
  }, policies)
  const polFin = await ensure(api, m.policies, 'name', 'Финансы: предупреждение', {
    name: 'Финансы: предупреждение', mode: 'alert', data_class: classByCode.get('financial') || '', identity_kind: 'human',
    masking_profile: maskEmail, enabled: '1',
    description: 'Выгрузка финансовых таблиц поднимает алерт',
  }, policies)

  const bindings = new Set((await listAll(api, m.policy_bindings)).map(r => `${recVal(r, 'policy')}|${recVal(r, 'dataset')}`))
  async function bind (policy, dataset) {
    const key = `${policy}|${dataset}`
    if (bindings.has(key)) return
    await createRecord(api, ns, m.policy_bindings, { policy, dataset, enabled: '1' })
    bindings.add(key)
  }
  await bind(polPii, users.id)
  await bind(polPci, clients.id)
  await bind(polCred, tokens.id)
  await bind(polFin, sessions.id)

  const classesOn = new Set((await listAll(api, m.classifications)).map(r => `${recVal(r, 'dataset')}|${recVal(r, 'data_class')}`))
  async function classify (dataset, code, extra) {
    const dataClass = classByCode.get(code)
    if (!dataClass) return
    const key = `${dataset}|${dataClass}`
    if (classesOn.has(key)) return
    await createRecord(api, ns, m.classifications, {
      dataset, data_class: dataClass, ...extra, reviewer: extra.status === 'confirmed' ? userID : '',
    })
    classesOn.add(key)
  }
  await classify(users.id, 'pii', { confidence: '0.92', source: 'model', status: 'pending', location_hint: users.path, masked_sample: 'a***@example.com' })
  await classify(sessions.id, 'credentials', { confidence: '0.81', source: 'rule', status: 'pending', location_hint: sessions.path, masked_sample: 'sess_***' })
  await classify(tokens.id, 'credentials', { confidence: '0.97', source: 'rule', status: 'confirmed', location_hint: tokens.path, masked_sample: 'sha256:ab12…', reviewed_at: yesterday })
  await classify(clients.id, 'pci', { confidence: '0.74', source: 'model', status: 'pending', location_hint: clients.path, masked_sample: '**** **** **** 4242' })
  await classify(attachments.id, 'ip', { confidence: '0.66', source: 'manual', status: 'confirmed', location_hint: attachments.path, masked_sample: 'contract-***.pdf', reviewed_at: yesterday })

  const identities = indexBy(await listAll(api, m.identities), 'name')
  const anna = await ensure(api, m.identities, 'name', 'Анна Соколова', {
    name: 'Анна Соколова', kind: 'human', source: 'idp', external_id: 'idp:anna.sokolova', email: 'anna.sokolova@example.com',
    department: 'Финансы', status: 'active', last_seen: now,
  }, identities)
  const igor = await ensure(api, m.identities, 'name', 'Игорь Петров', {
    name: 'Игорь Петров', kind: 'human', source: 'directory', external_id: 'dir:igor.petrov', email: 'igor.petrov@example.com',
    department: 'Безопасность', status: 'active', last_seen: yesterday,
  }, identities)
  const etl = await ensure(api, m.identities, 'name', 'svc-etl', {
    name: 'svc-etl', kind: 'service', source: 'discovered', external_id: 'svc:etl', department: 'Данные', status: 'active', last_seen: now,
  }, identities)
  const copilot = await ensure(api, m.identities, 'name', 'copilot-finance', {
    name: 'copilot-finance', kind: 'ai_agent', source: 'discovered', external_id: 'ai:copilot-finance', department: 'Финансы', status: 'active', last_seen: now,
  }, identities)
  const ghost = await ensure(api, m.identities, 'name', 'Подрядчик (призрак)', {
    name: 'Подрядчик (призрак)', kind: 'human', source: 'local', external_id: 'local:contractor', email: 'contractor@external.example',
    department: 'Подряд', status: 'ghost', last_seen: yesterday,
  }, identities)

  const grants = new Set((await listAll(api, m.entitlements)).map(r => `${recVal(r, 'identity')}|${recVal(r, 'dataset')}|${recVal(r, 'access_level')}`))
  async function grant (identity, dataset, values) {
    const key = `${identity}|${dataset}|${values.access_level}`
    if (grants.has(key)) return
    await createRecord(api, ns, m.entitlements, { identity, dataset, ...values })
    grants.add(key)
  }
  await grant(anna, users.id, { access_level: 'write', origin: 'role', origin_name: 'finance-editor', used: '1', last_used: now, status: 'excessive' })
  await grant(igor, tokens.id, { access_level: 'read', origin: 'direct', origin_name: 'break-glass', used: '1', last_used: yesterday, status: 'active' })
  await grant(etl, tokens.id, { access_level: 'admin', origin: 'role', origin_name: 'etl-runner', used: '1', last_used: now, status: 'excessive' })
  await grant(copilot, users.id, { access_level: 'read', origin: 'link', origin_name: 'copilot-share', used: '1', last_used: now, status: 'active' })
  await grant(ghost, attachments.id, { access_level: 'share', origin: 'link', origin_name: 'старая ссылка', used: '0', status: 'unused' })
  await grant(anna, sessions.id, { access_level: 'read', origin: 'group', origin_name: 'finance', used: '1', last_used: yesterday, status: 'active' })

  const findings = indexBy(await listAll(api, m.findings), 'title')
  const storeID = name => {
    const row = stores.find(s => recVal(s, 'name') === name)
    return row ? recID(row) : ''
  }
  const fExcess = await ensure(api, m.findings, 'title', 'svc-etl администрирует токены', {
    title: 'svc-etl администрирует токены', kind: 'excessive_access', severity: 'critical', status: 'open',
    data_store: storeID('test12'), dataset: tokens.id, identity: etl, policy: polCred, owner: userID,
    detected_at: now, summary: 'Сервисная учётка имеет admin на таблицу токенов. Политика в режиме наблюдения.',
  }, findings)
  const fAi = await ensure(api, m.findings, 'title', 'Теневой copilot читает пользователей', {
    title: 'Теневой copilot читает пользователей', kind: 'shadow_ai', severity: 'high', status: 'with_owner',
    data_store: storeID('test12'), dataset: users.id, identity: copilot, policy: polPii, owner: userID,
    detected_at: yesterday, summary: 'Несанкционированный агент читает персональные данные. Маска для AI ещё не подтверждена.',
  }, findings)
  await ensure(api, m.findings, 'title', 'Ссылка подрядчика на вложения', {
    title: 'Ссылка подрядчика на вложения', kind: 'stale_link', severity: 'medium', status: 'open',
    data_store: storeID('test9'), dataset: attachments.id, identity: ghost, owner: userID,
    detected_at: yesterday, summary: 'Призрачная учётка сохраняет share на вложения и не пользовалась им.',
  }, findings)
  await ensure(api, m.findings, 'title', 'test9 открыт на организацию', {
    title: 'test9 открыт на организацию', kind: 'public_exposure', severity: 'high', status: 'in_progress',
    data_store: storeID('test9'), dataset: clients.id, policy: polPci, owner: userID,
    detected_at: now, summary: 'Хранилище test9 помечено как доступное на всю организацию, в нём есть платёжный класс.',
  }, findings)
  await ensure(api, m.findings, 'title', 'Неиспользуемый share подрядчика', {
    title: 'Неиспользуемый share подрядчика', kind: 'unused_entitlement', severity: 'low', status: 'open',
    data_store: storeID('test9'), dataset: attachments.id, identity: ghost, owner: userID,
    detected_at: yesterday, summary: 'Право share не использовалось. Кандидат на отзыв.',
  }, findings)

  const remediations = indexBy(await listAll(api, m.remediations), 'result')
  await ensure(api, m.remediations, 'result', 'Снять admin у svc-etl', {
    finding: fExcess, action: 'revoke', status: 'proposed', actor: userID, reversible: '1',
    result: 'Снять admin у svc-etl',
  }, remediations)
  await ensure(api, m.remediations, 'result', 'Поставить маску на copilot', {
    finding: fAi, action: 'mask', status: 'approved', actor: userID, reversible: '1',
    result: 'Поставить маску на copilot',
  }, remediations)

  const requests = indexBy(await listAll(api, m.access_requests), 'justification')
  const week = new Date(Date.now() + 7 * 86400000).toISOString()
  await ensure(api, m.access_requests, 'justification', 'Сверка выгрузки клиентов за квартал', {
    requester: userID, identity: anna, dataset: clients.id, justification: 'Сверка выгрузки клиентов за квартал',
    starts_at: now, ends_at: week, status: 'pending', policy: polPci,
  }, requests)
  await ensure(api, m.access_requests, 'justification', 'Расследование токенов после инцидента', {
    requester: userID, identity: igor, dataset: tokens.id, justification: 'Расследование токенов после инцидента',
    starts_at: yesterday, ends_at: week, approver: userID, status: 'approved', policy: polCred,
  }, requests)
  await ensure(api, m.access_requests, 'justification', 'Разовый доступ подрядчика к вложениям', {
    requester: userID, identity: ghost, dataset: attachments.id, justification: 'Разовый доступ подрядчика к вложениям',
    starts_at: yesterday, ends_at: now, status: 'rejected',
  }, requests)

  const events = indexBy(await listAll(api, m.activity_events), 'summary')
  const evMask = await ensure(api, m.activity_events, 'summary', 'copilot-finance получил маскированный фрагмент пользователей', {
    occurred_at: now, identity: copilot, dataset: users.id, data_class: classByCode.get('pii') || '',
    tool: 'copilot', policy: polPii, outcome: 'mask', summary: 'copilot-finance получил маскированный фрагмент пользователей',
  }, events)
  await ensure(api, m.activity_events, 'summary', 'svc-etl прочитал таблицу токенов', {
    occurred_at: now, identity: etl, dataset: tokens.id, data_class: classByCode.get('credentials') || '',
    tool: 'psql', policy: polCred, outcome: 'allow', summary: 'svc-etl прочитал таблицу токенов',
  }, events)
  await ensure(api, m.activity_events, 'summary', 'Выгрузка сессий Анной подняла предупреждение', {
    occurred_at: yesterday, identity: anna, dataset: sessions.id, data_class: classByCode.get('credentials') || '',
    tool: 'compose', policy: polFin, outcome: 'alert', summary: 'Выгрузка сессий Анной подняла предупреждение',
  }, events)

  const alerts = indexBy(await listAll(api, m.alerts), 'title')
  await ensure(api, m.alerts, 'title', 'Маскирование copilot на пользователях', {
    title: 'Маскирование copilot на пользователях', severity: 'high', status: 'investigating',
    identity: copilot, dataset: users.id, activity_event: evMask, detected_at: now,
    summary: 'Политика маски сработала на чтении ПДн несанкционированным агентом',
  }, alerts)
  await ensure(api, m.alerts, 'title', 'Сервис прочитал секреты', {
    title: 'Сервис прочитал секреты', severity: 'critical', status: 'open',
    identity: etl, dataset: tokens.id, detected_at: now,
    summary: 'svc-etl прошёл в таблицу токенов при политике в режиме наблюдения',
  }, alerts)

  const ai = indexBy(await listAll(api, m.ai_assets), 'name')
  await ensure(api, m.ai_assets, 'name', 'Copilot Finance', {
    name: 'Copilot Finance', kind: 'copilot', sanctioned: '0', identity: copilot, status: 'discovered', owner: userID,
    notes: 'Найден по доступу к таблице пользователей. Санкция не выдана.',
  }, ai)
  await ensure(api, m.ai_assets, 'name', 'База знаний безопасности', {
    name: 'База знаний безопасности', kind: 'knowledge_base', sanctioned: '1', identity: igor, status: 'approved', owner: userID,
    notes: 'Внутренняя база, владелец — служба безопасности.',
  }, ai)

  const controls = indexBy(await listAll(api, m.compliance_controls), 'name')
  await ensure(api, m.compliance_controls, 'name', 'Ограничение доступа к ПДн', {
    framework: '152-fz', name: 'Ограничение доступа к ПДн', data_class: classByCode.get('pii') || '', policy: polPii,
    status: 'partial', evidence: 'Маска для AI задана, класс пользователей ещё на проверке',
  }, controls)
  await ensure(api, m.compliance_controls, 'name', 'Минимизация платёжных данных', {
    framework: 'pci-dss', name: 'Минимизация платёжных данных', data_class: classByCode.get('pci') || '', policy: polPci,
    status: 'gap', evidence: 'Класс клиентов не подтверждён, хранилище открыто на организацию',
  }, controls)
  await ensure(api, m.compliance_controls, 'name', 'Учёт секретов', {
    framework: 'iso27001', name: 'Учёт секретов', data_class: classByCode.get('credentials') || '', policy: polCred,
    status: 'covered', evidence: 'Токены классифицированы и подтверждены, доступ журналируется',
  }, controls)
  await ensure(api, m.compliance_controls, 'name', 'Контроль финансовой выгрузки', {
    framework: 'sox', name: 'Контроль финансовой выгрузки', data_class: classByCode.get('financial') || '', policy: polFin,
    status: 'partial', evidence: 'Политика предупреждает, автоматического блока нет',
  }, controls)

  console.log(JSON.stringify({
    users: users.path,
    sessions: sessions.path,
    tokens: tokens.path,
    clients: clients.path,
    attachments: attachments.path,
    identities: identities.size,
    findings: findings.size,
  }, null, 2))
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
