/**
 * Demo records for the invest namespace.
 *
 * Idempotent by unique keys (project code, document number, contract number, …).
 * Safe to re-run: existing records are reused, missing ones are created.
 *
 *   node seed.mjs
 */
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { createRecord, setOf, detectBase, apiFactory } from './helpers.mjs'
import { mintToken } from '../../backup/compose/helpers.mjs'
import { seedDocumentFiles } from './seed_files.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

function recID (row) {
  return String(row.recordID || row.ID)
}

function recVal (row, name) {
  const hit = (row.values || []).find(v => v.name === name)
  return hit ? String(hit.value ?? '') : ''
}

export function userFromToken (token) {
  try {
    return String(JSON.parse(Buffer.from(String(token).split('.')[1], 'base64url')).sub || '')
  } catch {
    return ''
  }
}

async function patchRecord (api, nsID, moduleID, recordID, patch) {
  const rec = await api('GET', `/namespace/${nsID}/module/${moduleID}/record/${recordID}`)
  const current = {}
  for (const v of rec.values || []) {
    if (v.name) current[v.name] = v.value == null ? '' : String(v.value)
  }
  let changed = false
  for (const [k, v] of Object.entries(patch)) {
    if (v == null || v === '') continue
    if (current[k] !== String(v)) {
      current[k] = String(v)
      changed = true
    }
  }
  if (!changed) return false
  const payload = Object.entries(current)
    .filter(([, v]) => v !== '' && v != null)
    .map(([name, value]) => ({ name, value }))
  await api('POST', `/namespace/${nsID}/module/${moduleID}/record/${recordID}`, {
    values: payload,
    updatedAt: rec.updatedAt,
  })
  return true
}

async function listAll (api, nsID, moduleID) {
  return setOf(await api('GET', `/namespace/${nsID}/module/${moduleID}/record/?limit=500`))
}

function indexBy (rows, field) {
  const map = new Map()
  for (const r of rows) {
    const k = recVal(r, field)
    if (k) map.set(k, recID(r))
  }
  return map
}

async function put (api, nsID, moduleID, values) {
  const compact = Object.fromEntries(
    Object.entries(values).filter(([, v]) => v !== '' && v != null && v !== undefined),
  )
  return createRecord(api, nsID, moduleID, compact)
}

async function ensure (api, nsID, moduleID, index, key, values) {
  if (key && index.has(key)) return index.get(key)
  const rec = await put(api, nsID, moduleID, values)
  const id = recID(rec)
  if (key) index.set(key, id)
  return id
}

const DOC_TYPES = [
  { name: 'Договор', code: 'CONTRACT' },
  { name: 'ПСД', code: 'PSD' },
  { name: 'Смета', code: 'ESTIMATE' },
  { name: 'Акт КС-2', code: 'KS2' },
  { name: 'Акт КС-3', code: 'KS3' },
  { name: 'Исполнительная документация', code: 'AS_BUILT' },
  { name: 'Приказ', code: 'ORDER' },
  { name: 'ТЭО', code: 'TEO', description: 'Технико-экономическое обоснование' },
  { name: 'Изыскания', code: 'SURVEY', description: 'Инженерные изыскания' },
  { name: 'Рабочая документация', code: 'RD' },
  { name: 'Акт ПНР', code: 'PNR', description: 'Пусконаладочные работы' },
  { name: 'Разрешение / согласование', code: 'PERMIT' },
]

const COUNTERPARTIES = [
  { name: 'ООО Генподрядчик-1', inn: '7700000001', role: 'contractor' },
  { name: 'АО РосИнвест', inn: '7700000002', role: 'investor', notes: 'Стратегический инвестор портфеля.' },
  { name: 'ПАО ЭнергоБанк', inn: '7700000003', role: 'bank', notes: 'Кредитная линия проектного финансирования.' },
  { name: 'ООО ГеоИзыскания', inn: '7700000004', role: 'designer' },
  { name: 'ООО ПроектГидро', inn: '7700000005', role: 'designer' },
  { name: 'ООО СтройСевер', inn: '7700000006', role: 'contractor' },
  { name: 'ООО ПНР-Сервис', inn: '7700000007', role: 'contractor' },
  { name: 'Минэнерго (демо)', inn: '7700000008', role: 'government' },
  { name: 'ООО СолнцеСтрой', inn: '7700000009', role: 'contractor' },
]

const MATERIALS = [
  { name: 'Бетон B25', unit: 'м³', unit_price: '8500', gost: 'ГОСТ 26633' },
  { name: 'Арматура А500С', unit: 'т', unit_price: '62000', gost: 'ГОСТ 34028' },
  { name: 'Кабель АС-400', unit: 'км', unit_price: '2800000' },
  { name: 'Трансформатор 220/110', unit: 'шт', unit_price: '85000000' },
]

const LABOR = [
  { name: 'Укладка бетона', unit: 'м³', hours: '2.5' },
  { name: 'Монтаж металлоконструкций', unit: 'т', hours: '18' },
  { name: 'ПНР ОРУ', unit: 'яч', hours: '120' },
]

const CONSTRUCTION_TYPES = [
  { code: 'TES', name: 'Тепловая электростанция', description: 'ТЭС / энергоблок' },
  { code: 'WIND', name: 'Ветропарк', description: 'ВЭС и подключение' },
  { code: 'GRID', name: 'ЛЭП / сеть', description: 'Линии и подстанции' },
  { code: 'DC', name: 'ЦОД', description: 'Центр обработки данных' },
]

const WBS_TEMPLATES = [
  { type: 'TES', code: '1', name: 'Площадка и фундаменты', level: 'stage', budget_planned: '500000000', duration_days: '180' },
  { type: 'TES', code: '1.1', name: 'Земляные работы', level: 'work', parent_code: '1', budget_planned: '120000000', duration_days: '60' },
  { type: 'TES', code: '1.2', name: 'Фундаменты', level: 'work', parent_code: '1', predecessor_code: '1.1', budget_planned: '220000000', duration_days: '90' },
  { type: 'TES', code: '2', name: 'Основное оборудование', level: 'stage', budget_planned: '1800000000', duration_days: '360' },
  { type: 'WIND', code: '1', name: 'Концепция и ТЭО', level: 'stage', budget_planned: '45000000', duration_days: '270' },
  { type: 'WIND', code: '1.1', name: 'Сбор исходных данных', level: 'work', parent_code: '1', budget_planned: '12000000', duration_days: '90' },
  { type: 'GRID', code: '1', name: 'Изыскания', level: 'stage', budget_planned: '420000000', duration_days: '540' },
  { type: 'GRID', code: '1.1', name: 'Трасса и геология', level: 'work', parent_code: '1', budget_planned: '180000000', duration_days: '360' },
  { type: 'DC', code: '1', name: 'Концепция', level: 'stage', budget_planned: '80000000', duration_days: '240' },
  { type: 'DC', code: '1.1', name: 'ТЭО и финмодель', level: 'work', parent_code: '1', budget_planned: '35000000', duration_days: '180' },
]

const PHASE_REQ = [
  { phase: 'concept', type: 'TEO' },
  { phase: 'survey', type: 'SURVEY' },
  { phase: 'design', type: 'PSD' },
  { phase: 'design', type: 'ESTIMATE' },
  { phase: 'construction', type: 'KS2' },
  { phase: 'construction', type: 'KS3' },
  { phase: 'commissioning', type: 'PNR' },
  { phase: 'operations', type: 'AS_BUILT' },
]

function riskScore (probability, impact) {
  const m = { low: 1, medium: 2, high: 3, critical: 4 }
  return String((m[probability] || 1) * (m[impact] || 1))
}

const ADVISORS = [
  {
    name: 'Юрист',
    role: 'lawyer',
    enabled: '1',
    modules: 'contracts,documents',
    prompt: 'Human-in-the-loop юрист: анализируй договоры и документы, не утверждай сам.',
  },
  {
    name: 'Финконтролёр',
    role: 'fincontroller',
    enabled: '1',
    modules: 'budget_lines,cashflow_items,wbs_items,change_requests',
    prompt: 'Human-in-the-loop финконтролёр: план/факт, EVM, RFC. Не меняй цифры сам.',
  },
]

/**
 * Portfolio: every phase (1–6) and every project status.
 * Related docs/contracts/RFC/risks cover the remaining enums.
 */
const PROJECTS = [
  {
    code: 'WIND-001',
    name: 'Ветропарк Приморье',
    phase: 'concept',
    status: 'draft',
    type: 'WIND',
    investor: 'АО РосИнвест',
    start_planned: '2027-03-01',
    end_planned: '2030-12-31',
    budget_planned: '2400000000',
    budget_actual: '0',
    description: 'Черновик концепции ВЭС 200 МВт. На инвесткомитет ещё не выносился.',
    wbs: [
      { code: '1', name: 'Концепция и ТЭО', level: 'stage', start_planned: '2027-03-01', end_planned: '2027-12-31', budget_planned: '45000000', percent_complete: '5' },
      { code: '1.1', name: 'Сбор исходных данных', level: 'work', parent: '1', start_planned: '2027-03-01', end_planned: '2027-06-30', budget_planned: '12000000', percent_complete: '10' },
    ],
    contracts: [
      { number: 'КОНЦ-WIND-001', title: 'Предконтракт на ТЭО', cp: '7700000005', amount: '18000000', start_date: '2027-03-01', end_date: '2027-12-31', status: 'draft', terms: 'Черновик, не подписан.' },
    ],
    documents: [
      { title: 'Концепция ВЭС Приморье', number: 'TEO-WIND-001-D', type: 'TEO', status: 'draft', sign: 'unsigned', due_date: '2027-06-01', notes: 'Первая редакция, без цифр CAPEX.' },
    ],
    risks: [
      { title: 'Нет сетки присоединения', probability: 'high', impact: 'critical', status: 'open', mitigation: 'Запрос ТУ в сетевую компанию на этапе концепции.' },
    ],
    rfc: [],
    budget: [
      { article: 'WIND-001 ТЭО', planned: '45000000', actual: '0', reserve: '5000000' },
    ],
  },
  {
    code: 'DC-001',
    name: 'ЦОД Сибирь',
    phase: 'concept',
    status: 'active',
    type: 'DC',
    investor: 'АО РосИнвест',
    start_planned: '2026-06-01',
    end_planned: '2029-06-30',
    budget_planned: '3100000000',
    budget_actual: '28000000',
    spi: '1.05',
    cpi: '0.97',
    description: 'Активная проработка концепции ЦОД 40 МВт. ТЭО на согласовании банка.',
    wbs: [
      { code: '1', name: 'Концепция', level: 'stage', start_planned: '2026-06-01', end_planned: '2027-02-28', budget_planned: '80000000', percent_complete: '40', start_actual: '2026-06-10' },
      { code: '1.1', name: 'ТЭО и финмодель', level: 'work', parent: '1', start_planned: '2026-06-01', end_planned: '2026-11-30', budget_planned: '35000000', actual_cost: '28000000', percent_complete: '70' },
      { code: '1.2', name: 'Выбор площадки', level: 'work', parent: '1', predecessor: '1.1', start_planned: '2026-10-01', end_planned: '2027-02-28', budget_planned: '25000000', percent_complete: '15' },
    ],
    contracts: [
      { number: 'ТЭО-DC-001', title: 'Разработка ТЭО ЦОД', cp: '7700000005', amount: '35000000', start_date: '2026-06-15', end_date: '2026-12-31', status: 'active' },
    ],
    documents: [
      { title: 'ТЭО ЦОД Сибирь v0.9', number: 'TEO-DC-001', type: 'TEO', status: 'in_review', sign: 'unsigned', due_date: '2026-09-15', notes: 'На согласовании у банка.' },
      { title: 'Разрешение на размещение', number: 'PERM-DC-001', type: 'PERMIT', status: 'draft', sign: 'unsigned', due_date: '2026-12-01' },
    ],
    risks: [
      { title: 'Дефицит мощности в узле', probability: 'medium', impact: 'high', status: 'mitigating', mitigation: 'Параллельный запрос ТУ на 2 площадки.' },
    ],
    rfc: [
      { title: 'Увеличение мощности ЦОД до 48 МВт', rfc_type: 'scope', status: 'draft', delta_budget: '420000000', delta_days: '90', justification: 'Запрос якорного клиента.' },
    ],
    budget: [
      { article: 'DC-001 ТЭО', planned: '35000000', actual: '28000000', reserve: '4000000' },
    ],
    cashflow: [
      { article: 'DC-001 ТЭО', date: '2026-07-15', amount: '14000000', direction: 'outflow', description: 'Аванс проектировщику ТЭО' },
      { article: 'DC-001 ТЭО', date: '2026-08-20', amount: '14000000', direction: 'outflow', description: 'Второй платёж ТЭО' },
    ],
  },
  {
    code: 'GRID-001',
    name: 'ЛЭП 500 кВ Урал',
    phase: 'survey',
    status: 'active',
    type: 'GRID',
    investor: 'Инвестор А',
    start_planned: '2025-09-01',
    end_planned: '2029-12-31',
    budget_planned: '8500000000',
    budget_actual: '210000000',
    spi: '0.94',
    cpi: '1.02',
    eac: '8400000000',
    description: 'Инженерные изыскания коридора ВЛ 500 кВ. Полевые работы идут.',
    wbs: [
      { code: '1', name: 'Изыскания', level: 'stage', start_planned: '2025-09-01', end_planned: '2027-03-31', budget_planned: '420000000', percent_complete: '35', start_actual: '2025-09-12' },
      { code: '1.1', name: 'Трасса и геология', level: 'work', parent: '1', start_planned: '2025-09-01', end_planned: '2026-08-31', budget_planned: '180000000', actual_cost: '120000000', percent_complete: '55', start_actual: '2025-09-12' },
      { code: '1.2', name: 'Экология и ОВОС', level: 'work', parent: '1', start_planned: '2026-01-01', end_planned: '2026-12-31', budget_planned: '90000000', actual_cost: '40000000', percent_complete: '40' },
    ],
    contracts: [
      { number: 'ИЗ-GRID-001', title: 'Инженерные изыскания ВЛ 500', cp: '7700000004', amount: '180000000', start_date: '2025-09-01', end_date: '2027-03-31', status: 'active' },
    ],
    documents: [
      { title: 'Программа изысканий', number: 'SURV-GRID-001', type: 'SURVEY', status: 'approved', sign: 'signed', due_date: '2025-10-01' },
      { title: 'Промежуточный отчёт геологии', number: 'SURV-GRID-001-R1', type: 'SURVEY', status: 'in_review', sign: 'pending', due_date: '2026-09-10' },
    ],
    risks: [
      { title: 'Особо охраняемые территории на трассе', probability: 'medium', impact: 'high', status: 'open', mitigation: 'Альтернативный коридор + согласование с Минприроды.' },
    ],
    rfc: [
      { title: 'Смещение коридора на 12 км', rfc_type: 'scope', status: 'in_review', delta_budget: '65000000', delta_days: '45', justification: 'Обход заказника.' },
    ],
    budget: [
      { article: 'GRID-001 изыскания', planned: '180000000', actual: '120000000', reserve: '15000000' },
    ],
    progress: [
      { wbs: '1.1', quantity: '84', unit: 'км', percent: '55', cost: '120000000', recorded_at: '2026-08-01T10:00:00.000Z', notes: 'Пробурено 84 из 152 скважин.' },
    ],
  },
  {
    code: 'TES-001',
    name: 'ТЭС-мини Казань',
    phase: 'design',
    status: 'active',
    type: 'TES',
    investor: 'АО РосИнвест',
    start_planned: '2025-01-15',
    end_planned: '2028-06-30',
    budget_planned: '4200000000',
    budget_actual: '310000000',
    spi: '0.91',
    cpi: '0.89',
    eac: '4600000000',
    description: 'Стадия проектирования. ПСД на согласовании, смета отклонена по резервам.',
    wbs: [
      { code: '1', name: 'Проектирование', level: 'stage', start_planned: '2025-01-15', end_planned: '2026-12-31', budget_planned: '380000000', percent_complete: '62', start_actual: '2025-01-20' },
      { code: '1.1', name: 'ПСД (стадия П)', level: 'work', parent: '1', start_planned: '2025-01-15', end_planned: '2026-04-30', budget_planned: '210000000', actual_cost: '190000000', percent_complete: '85' },
      { code: '1.2', name: 'Рабочая документация', level: 'work', parent: '1', predecessor: '1.1', start_planned: '2026-03-01', end_planned: '2026-12-31', budget_planned: '140000000', actual_cost: '80000000', percent_complete: '40' },
    ],
    contracts: [
      { number: 'П-TES-001', title: 'Проектирование ТЭС-мини', cp: '7700000005', amount: '380000000', start_date: '2025-01-20', end_date: '2026-12-31', status: 'active' },
    ],
    documents: [
      { title: 'ПСД ТЭС-мини Казань', number: 'PSD-TES-001', type: 'PSD', wbs: '1.1', status: 'in_review', sign: 'pending', due_date: '2026-09-05', notes: 'Экспертиза промышленной безопасности.' },
      { title: 'Сводная смета', number: 'EST-TES-001', type: 'ESTIMATE', wbs: '1.1', status: 'rejected', sign: 'unsigned', due_date: '2026-08-20', notes: 'Завышен резерв непредвиденных — вернуть в работу.' },
      { title: 'РД котлоагрегат', number: 'RD-TES-001-KA', type: 'RD', wbs: '1.2', status: 'draft', sign: 'unsigned', due_date: '2026-10-15' },
    ],
    risks: [
      { title: 'CPI ниже 0.9 на проектировании', probability: 'high', impact: 'medium', status: 'mitigating', mitigation: 'Пересчёт сметы, сокращение объёма РД 1-й очереди.' },
    ],
    rfc: [
      { title: 'Доп. объём РД по золоудалению', rfc_type: 'budget', status: 'in_review', delta_budget: '28000000', delta_days: '21', eac_before: '4200000000', eac_after: '4228000000', justification: 'Требование экспертизы.' },
    ],
    budget: [
      { article: 'TES-001 проектирование', planned: '380000000', actual: '270000000', reserve: '20000000' },
    ],
    cashflow: [
      { article: 'TES-001 проектирование', date: '2026-04-01', amount: '500000000', direction: 'inflow', description: 'Транш банка — проектирование' },
      { article: 'TES-001 проектирование', date: '2026-04-10', amount: '90000000', direction: 'outflow', description: 'Оплата ПСД' },
    ],
  },
  {
    code: 'SOL-001',
    name: 'СЭС Астрахань',
    phase: 'design',
    status: 'on_hold',
    investor: 'Инвестор А',
    start_planned: '2025-04-01',
    end_planned: '2027-12-31',
    budget_planned: '1800000000',
    budget_actual: '95000000',
    spi: '0.62',
    cpi: '0.84',
    description: 'Проектирование приостановлено: нет регистрации права на земельный участок.',
    wbs: [
      { code: '1', name: 'Проектирование СЭС', level: 'stage', start_planned: '2025-04-01', end_planned: '2026-10-31', budget_planned: '120000000', percent_complete: '30' },
      { code: '1.1', name: 'Генплан и ОРУ', level: 'work', parent: '1', start_planned: '2025-04-01', end_planned: '2026-02-28', budget_planned: '55000000', actual_cost: '48000000', percent_complete: '70' },
    ],
    contracts: [
      { number: 'П-SOL-001', title: 'Проект СЭС 80 МВт', cp: '7700000005', amount: '120000000', start_date: '2025-04-15', end_date: '2026-10-31', status: 'active', terms: 'Работы приостановлены уведомлением от 2026-07-01.' },
    ],
    documents: [
      { title: 'ПСД СЭС Астрахань', number: 'PSD-SOL-001', type: 'PSD', status: 'draft', sign: 'unsigned', notes: 'Заморожен до земли.' },
      { title: 'Приказ о приостановке', number: 'ORD-SOL-001-HOLD', type: 'ORDER', status: 'approved', sign: 'signed', due_date: '2026-07-02' },
    ],
    risks: [
      { title: 'Срыв окна поставки инверторов', probability: 'high', impact: 'high', status: 'open', mitigation: 'Опцион у поставщика до Q1 2027.' },
    ],
    rfc: [
      { title: 'Перенос СМР на 2027', rfc_type: 'schedule', status: 'approved', delta_budget: '0', delta_days: '180', justification: 'Ожидание земли. Утверждено PMO.' },
    ],
    changelog: [
      { rfc: 'Перенос СМР на 2027', summary: 'Финиш сдвинут на 180 дней из‑за земли', old_end: '2027-12-31', new_end: '2028-06-30', changed_at: '2026-07-03T12:00:00.000Z' },
    ],
    budget: [
      { article: 'SOL-001 проектирование', planned: '120000000', actual: '95000000', reserve: '8000000' },
    ],
  },
  {
    code: 'ENRG-001',
    name: 'Пилотный энергообъект',
    phase: 'construction',
    status: 'active',
    investor: 'Инвестор А',
    start_planned: '2026-01-01',
    end_planned: '2027-12-31',
    budget_planned: '1500000000',
    budget_actual: '120000000',
    description: 'Демонстрационный инвестиционный проект для пространства документооборота.',
    wbs: [
      { code: '1', name: 'Строительство', level: 'stage', start_planned: '2026-03-01', end_planned: '2027-06-30', budget_planned: '900000000', percent_complete: '12' },
      { code: '1.1', name: 'Фундамент', level: 'work', parent: '1', start_planned: '2026-03-01', end_planned: '2026-06-30', budget_planned: '180000000', actual_cost: '40000000', percent_complete: '25' },
      { code: '1.2', name: 'Каркас', level: 'work', parent: '1', predecessor: '1.1', start_planned: '2026-07-01', end_planned: '2026-12-31', budget_planned: '320000000', actual_cost: '0', percent_complete: '0' },
    ],
    contracts: [
      { number: 'ГП-001/2026', title: 'Генеральный подряд', cp: '7700000001', amount: '900000000', start_date: '2026-02-01', end_date: '2027-12-31', status: 'active', terms: 'Оплата по актам КС-2/КС-3. Штраф за просрочку 0.1% в день.' },
    ],
    documents: [
      { title: 'Договор генподряда', number: 'ГП-001/2026', type: 'CONTRACT', contract: 'ГП-001/2026', status: 'in_review', sign: 'unsigned', due_date: '2026-09-01', notes: 'Черновик на согласовании юриста.' },
      { title: 'ПСД пилотного объекта', number: 'PSD-ENRG-001', type: 'PSD', status: 'approved', sign: 'signed', due_date: '2026-02-15' },
      { title: 'Смета фундамент', number: 'EST-ENRG-001-1.1', type: 'ESTIMATE', wbs: '1.1', status: 'approved', sign: 'signed' },
      { title: 'Акт КС-2 фундамент №1', number: 'KS2-ENRG-001-01', type: 'KS2', wbs: '1.1', contract: 'ГП-001/2026', status: 'in_review', sign: 'pending', due_date: '2026-09-12' },
      { title: 'Акт КС-3 март', number: 'KS3-ENRG-001-03', type: 'KS3', contract: 'ГП-001/2026', status: 'draft', sign: 'unsigned' },
      { title: 'Приказ о старте СМР', number: 'ORD-ENRG-001', type: 'ORDER', status: 'archived', sign: 'signed' },
    ],
    risks: [
      { title: 'Срыв поставки арматуры', probability: 'medium', impact: 'high', status: 'open', mitigation: 'Дублирующий поставщик, запас 14 дней.' },
      { title: 'Паводок на котловане', probability: 'low', impact: 'medium', status: 'accepted', mitigation: 'Принят: обвалование уже в смете.' },
    ],
    rfc: [
      { title: 'Увеличение объёма фундамента', rfc_type: 'scope', status: 'draft', delta_budget: '25000000', delta_days: '14', justification: 'По результатам изысканий нужна доп. подушка.' },
    ],
    budget: [
      { article: 'Фундамент — бетон', planned: '180000000', actual: '40000000', reserve: '15000000', wbs: '1.1' },
    ],
    progress: [
      { wbs: '1.1', quantity: '120', unit: 'м³', percent: '25', cost: '40000000', recorded_at: '2026-08-15T09:00:00.000Z', notes: 'Первая заливка. Фото с площадки — веб-форма.' },
    ],
  },
  {
    code: 'HPP-001',
    name: 'МГЭС Карелия',
    phase: 'construction',
    status: 'on_hold',
    investor: 'ПАО ЭнергоБанк',
    start_planned: '2024-04-01',
    end_planned: '2027-10-31',
    budget_planned: '2700000000',
    budget_actual: '810000000',
    spi: '0.55',
    cpi: '0.78',
    eac: '3400000000',
    description: 'СМР заморожены: банк приостановил выборку транша после превышения EAC.',
    wbs: [
      { code: '1', name: 'Строительство МГЭС', level: 'stage', start_planned: '2024-04-01', end_planned: '2027-06-30', budget_planned: '2100000000', percent_complete: '28', start_actual: '2024-04-20' },
      { code: '1.1', name: 'Плотина', level: 'work', parent: '1', start_planned: '2024-04-01', end_planned: '2026-09-30', budget_planned: '980000000', actual_cost: '620000000', percent_complete: '45' },
      { code: '1.2', name: 'Здание ГЭС', level: 'work', parent: '1', predecessor: '1.1', start_planned: '2025-06-01', end_planned: '2027-03-31', budget_planned: '540000000', actual_cost: '190000000', percent_complete: '18' },
    ],
    contracts: [
      { number: 'ГП-HPP-001', title: 'Генподряд МГЭС', cp: '7700000006', amount: '2100000000', start_date: '2024-04-01', end_date: '2027-10-31', status: 'active', terms: 'Приостановка СМР по уведомлению банка.' },
    ],
    documents: [
      { title: 'Договор генподряда МГЭС', number: 'ГП-HPP-001', type: 'CONTRACT', contract: 'ГП-HPP-001', status: 'approved', sign: 'signed' },
      { title: 'Акт КС-2 плотина №12', number: 'KS2-HPP-001-12', type: 'KS2', wbs: '1.1', contract: 'ГП-HPP-001', status: 'rejected', sign: 'failed', due_date: '2026-07-20', notes: 'Объёмы не сходятся с исполнительной.' },
      { title: 'Исполнительная плотина', number: 'AB-HPP-001', type: 'AS_BUILT', wbs: '1.1', status: 'in_review', sign: 'unsigned', due_date: '2026-09-01' },
    ],
    risks: [
      { title: 'Остановка финансирования', probability: 'critical', impact: 'critical', status: 'open', mitigation: 'Переговоры с банком, урезание 2-й очереди.' },
      { title: 'Размыв перемычки', probability: 'medium', impact: 'high', status: 'closed', mitigation: 'Укреплено в 2025, риск закрыт.' },
    ],
    rfc: [
      { title: 'Консервация площадки на зиму', rfc_type: 'schedule', status: 'in_review', delta_budget: '45000000', delta_days: '120', justification: 'Содержание при простое.' },
      { title: 'Смена генподрядчика', rfc_type: 'scope', status: 'rejected', delta_budget: '180000000', delta_days: '60', justification: 'Отклонено: штраф за расторжение выше экономии.' },
    ],
    budget: [
      { article: 'HPP-001 плотина', planned: '980000000', actual: '620000000', reserve: '40000000', wbs: '1.1' },
    ],
    cashflow: [
      { article: 'HPP-001 плотина', date: '2026-03-01', amount: '400000000', direction: 'inflow', description: 'Последний транш банка до холда' },
      { article: 'HPP-001 плотина', date: '2026-03-15', amount: '210000000', direction: 'outflow', description: 'КС-2 плотина' },
    ],
    progress: [
      { wbs: '1.1', quantity: '42000', unit: 'м³', percent: '45', cost: '620000000', recorded_at: '2026-06-20T11:00:00.000Z', notes: 'Последняя фиксация до холда.' },
    ],
  },
  {
    code: 'SUB-001',
    name: 'ПС Север 220 кВ',
    phase: 'commissioning',
    status: 'active',
    investor: 'Инвестор А',
    start_planned: '2023-02-01',
    end_planned: '2026-11-30',
    budget_planned: '920000000',
    budget_actual: '875000000',
    spi: '0.98',
    cpi: '0.96',
    eac: '940000000',
    description: 'Стройка завершена, идёт ПНР и комплексное опробование.',
    wbs: [
      { code: '1', name: 'Ввод в эксплуатацию', level: 'stage', start_planned: '2026-05-01', end_planned: '2026-11-30', budget_planned: '85000000', percent_complete: '55', start_actual: '2026-05-12' },
      { code: '1.1', name: 'ПНР ОРУ 220', level: 'work', parent: '1', start_planned: '2026-05-01', end_planned: '2026-09-30', budget_planned: '42000000', actual_cost: '31000000', percent_complete: '70', start_actual: '2026-05-12' },
      { code: '1.2', name: 'Комплексное опробование', level: 'work', parent: '1', predecessor: '1.1', start_planned: '2026-09-15', end_planned: '2026-11-30', budget_planned: '18000000', actual_cost: '4000000', percent_complete: '20' },
    ],
    contracts: [
      { number: 'ПНР-SUB-001', title: 'Пусконаладка ПС Север', cp: '7700000007', amount: '85000000', start_date: '2026-05-01', end_date: '2026-11-30', status: 'active' },
      { number: 'ГП-SUB-001', title: 'СМР ПС Север (закрыт)', cp: '7700000006', amount: '720000000', start_date: '2023-02-01', end_date: '2026-04-30', status: 'completed' },
    ],
    documents: [
      { title: 'Акт ПНР ОРУ яч.1', number: 'PNR-SUB-001-01', type: 'PNR', wbs: '1.1', status: 'approved', sign: 'signed' },
      { title: 'Акт ПНР ОРУ яч.2', number: 'PNR-SUB-001-02', type: 'PNR', wbs: '1.1', status: 'in_review', sign: 'pending', due_date: '2026-09-08' },
      { title: 'Исполнительная ПС Север', number: 'AB-SUB-001', type: 'AS_BUILT', status: 'approved', sign: 'signed' },
      { title: 'Разрешение на допуск', number: 'PERM-SUB-001', type: 'PERMIT', status: 'in_review', sign: 'unsigned', due_date: '2026-10-01', notes: 'Ростехнадзор.' },
    ],
    risks: [
      { title: 'Задержка допуска Ростехнадзора', probability: 'medium', impact: 'high', status: 'mitigating', mitigation: 'Предварительная проверка чек-листа ПНР.' },
    ],
    rfc: [
      { title: 'Доп. испытания РЗА', rfc_type: 'scope', status: 'approved', delta_budget: '4200000', delta_days: '10', eac_before: '920000000', eac_after: '924200000', justification: 'Требование сетевой компании.' },
    ],
    changelog: [
      { rfc: 'Доп. испытания РЗА', summary: 'Утверждены доп. испытания РЗА (+4.2 млн)', old_budget: '920000000', new_budget: '924200000', changed_at: '2026-07-18T09:30:00.000Z' },
    ],
    budget: [
      { article: 'SUB-001 ПНР', planned: '85000000', actual: '35000000', reserve: '7000000', wbs: '1' },
    ],
    progress: [
      { wbs: '1.1', quantity: '7', unit: 'яч', percent: '70', cost: '31000000', recorded_at: '2026-08-22T14:00:00.000Z', notes: '7 из 10 ячеек ОРУ под напряжением.' },
    ],
  },
  {
    code: 'HEAT-001',
    name: 'Котельная Центр',
    phase: 'operations',
    status: 'completed',
    investor: 'Инвестор А',
    start_planned: '2022-03-01',
    end_planned: '2025-09-30',
    budget_planned: '450000000',
    budget_actual: '438000000',
    spi: '1.02',
    cpi: '1.03',
    eac: '438000000',
    description: 'Объект сдан и принят. Архив документов, договор генподряда исполнен.',
    wbs: [
      { code: '1', name: 'Строительство котельной', level: 'stage', start_planned: '2022-03-01', end_planned: '2025-06-30', budget_planned: '390000000', actual_cost: '382000000', percent_complete: '100', start_actual: '2022-03-10', end_actual: '2025-06-12' },
      { code: '1.1', name: 'СМР', level: 'work', parent: '1', start_planned: '2022-03-01', end_planned: '2025-03-31', budget_planned: '310000000', actual_cost: '305000000', percent_complete: '100', start_actual: '2022-03-10', end_actual: '2025-03-20' },
      { code: '2', name: 'Эксплуатация (гарантия)', level: 'stage', start_planned: '2025-07-01', end_planned: '2027-06-30', budget_planned: '18000000', percent_complete: '40' },
    ],
    contracts: [
      { number: 'ГП-HEAT-001', title: 'Генподряд котельная', cp: '7700000001', amount: '390000000', start_date: '2022-03-01', end_date: '2025-09-30', status: 'completed' },
    ],
    documents: [
      { title: 'Акт ввода в эксплуатацию', number: 'PNR-HEAT-001-IN', type: 'PNR', status: 'approved', sign: 'signed' },
      { title: 'Исполнительная котельная', number: 'AB-HEAT-001', type: 'AS_BUILT', status: 'archived', sign: 'signed' },
      { title: 'КС-3 итоговый', number: 'KS3-HEAT-001-FIN', type: 'KS3', contract: 'ГП-HEAT-001', status: 'archived', sign: 'signed' },
    ],
    risks: [
      { title: 'Гарантийные дефекты котла', probability: 'low', impact: 'low', status: 'closed', mitigation: 'Устранены подрядчиком в 2025 Q4.' },
    ],
    rfc: [
      { title: 'Замена горелок (гарантия)', rfc_type: 'scope', status: 'approved', delta_budget: '0', delta_days: '0', justification: 'За счёт подрядчика.' },
    ],
    budget: [
      { article: 'HEAT-001 СМР', planned: '310000000', actual: '305000000', reserve: '0', wbs: '1.1' },
    ],
    cashflow: [
      { article: 'HEAT-001 СМР', date: '2025-09-15', amount: '12000000', direction: 'outflow', description: 'Гарантийное удержание, возврат подрядчику' },
    ],
  },
  {
    code: 'CHP-001',
    name: 'Когенерация Юг',
    phase: 'operations',
    status: 'active',
    investor: 'АО РосИнвест',
    start_planned: '2020-01-01',
    end_planned: '2024-12-31',
    budget_planned: '1200000000',
    budget_actual: '1185000000',
    spi: '1.00',
    cpi: '1.01',
    eac: '1185000000',
    description: 'Действующая когенерация. CAPEX закрыт, идут эксплуатационные договоры и ремонты.',
    wbs: [
      { code: '1', name: 'Эксплуатация', level: 'stage', start_planned: '2025-01-01', end_planned: '2028-12-31', budget_planned: '240000000', percent_complete: '35' },
      { code: '1.1', name: 'Капремонт 2026', level: 'work', parent: '1', start_planned: '2026-05-01', end_planned: '2026-07-15', budget_planned: '48000000', actual_cost: '21000000', percent_complete: '40', start_actual: '2026-05-04' },
    ],
    contracts: [
      { number: 'ЭКСП-CHP-001', title: 'Сервис когенерации 2026–2028', cp: '7700000007', amount: '96000000', start_date: '2026-01-01', end_date: '2028-12-31', status: 'active' },
    ],
    documents: [
      { title: 'График капремонта 2026', number: 'ORD-CHP-001-KR', type: 'ORDER', status: 'approved', sign: 'signed' },
      { title: 'Смета капремонта', number: 'EST-CHP-001-KR', type: 'ESTIMATE', wbs: '1.1', status: 'approved', sign: 'signed' },
      { title: 'Акт КС-2 капремонт №2', number: 'KS2-CHP-001-02', type: 'KS2', wbs: '1.1', status: 'in_review', sign: 'pending', due_date: '2026-09-05' },
    ],
    risks: [
      { title: 'Срыв окна капремонта', probability: 'medium', impact: 'medium', status: 'mitigating', mitigation: 'Подменный модуль на складе.' },
    ],
    rfc: [
      { title: 'Доп. замена теплообменника', rfc_type: 'budget', status: 'draft', delta_budget: '8700000', delta_days: '7', justification: 'Дефектация на останове.' },
    ],
    budget: [
      { article: 'CHP-001 капремонт', planned: '48000000', actual: '21000000', reserve: '5000000', wbs: '1.1' },
    ],
    cashflow: [
      { article: 'CHP-001 капремонт', date: '2026-05-10', amount: '18000000', direction: 'outflow', description: 'Аванс капремонт' },
      { article: 'CHP-001 капремонт', date: '2026-06-01', amount: '25000000', direction: 'inflow', description: 'Выручка за май (эксплуатация)' },
    ],
    progress: [
      { wbs: '1.1', quantity: '2', unit: 'агр', percent: '40', cost: '21000000', recorded_at: '2026-08-10T08:00:00.000Z', notes: '2 агрегата из 5 вскрыты.' },
    ],
  },
  {
    code: 'LNG-001',
    name: 'СПГ-терминал Балтика',
    phase: 'survey',
    status: 'cancelled',
    investor: 'АО РосИнвест',
    start_planned: '2024-01-01',
    end_planned: '2031-12-31',
    budget_planned: '12000000000',
    budget_actual: '185000000',
    description: 'Проект отменён инвесткомитетом: нет рынка сбыта и разрешения на акваторию.',
    wbs: [
      { code: '1', name: 'Изыскания (остановлены)', level: 'stage', start_planned: '2024-01-01', end_planned: '2026-06-30', budget_planned: '320000000', actual_cost: '185000000', percent_complete: '20', start_actual: '2024-02-01', end_actual: '2026-04-15' },
    ],
    contracts: [
      { number: 'ИЗ-LNG-001', title: 'Изыскания акватории', cp: '7700000004', amount: '220000000', start_date: '2024-02-01', end_date: '2026-06-30', status: 'terminated', terms: 'Расторгнут 2026-04-15, выплачена компенсация.' },
    ],
    documents: [
      { title: 'Приказ об отмене проекта', number: 'ORD-LNG-001-CAN', type: 'ORDER', status: 'approved', sign: 'signed' },
      { title: 'Отчёт изысканий (архив)', number: 'SURV-LNG-001', type: 'SURVEY', status: 'archived', sign: 'signed' },
      { title: 'Договор изысканий', number: 'ИЗ-LNG-001', type: 'CONTRACT', contract: 'ИЗ-LNG-001', status: 'rejected', sign: 'unsigned', notes: 'Расторжение, юридическая папка.' },
    ],
    risks: [
      { title: 'Отсутствие рынка СПГ', probability: 'critical', impact: 'critical', status: 'accepted', mitigation: 'Проект закрыт, риск принят инвесткомитетом.' },
    ],
    rfc: [
      { title: 'Полная отмена CAPEX', rfc_type: 'scope', status: 'approved', delta_budget: '-11815000000', delta_days: '0', justification: 'Решение инвесткомитета 2026-04-10.' },
    ],
    changelog: [
      { rfc: 'Полная отмена CAPEX', summary: 'Проект отменён, бюджет обнулён кроме изысканий', old_budget: '12000000000', new_budget: '185000000', changed_at: '2026-04-10T16:00:00.000Z' },
    ],
    budget: [
      { article: 'LNG-001 изыскания (закрыто)', planned: '320000000', actual: '185000000', reserve: '0' },
    ],
  },
]

export async function seedIfEmpty (api, nsID, modules, userID = '', opts = {}) {
  const typeRows = await listAll(api, nsID, modules.document_types)
  const typesByCode = indexBy(typeRows, 'code')
  for (const row of DOC_TYPES) {
    await ensure(api, nsID, modules.document_types, typesByCode, row.code, row)
  }

  const cpRows = await listAll(api, nsID, modules.counterparties)
  const cpByInn = indexBy(cpRows, 'inn')
  for (const row of COUNTERPARTIES) {
    await ensure(api, nsID, modules.counterparties, cpByInn, row.inn, row)
  }

  const matRows = await listAll(api, nsID, modules.materials)
  const matByName = indexBy(matRows, 'name')
  for (const row of MATERIALS) {
    await ensure(api, nsID, modules.materials, matByName, row.name, row)
  }

  const laborRows = await listAll(api, nsID, modules.labor_norms)
  const laborByName = indexBy(laborRows, 'name')
  for (const row of LABOR) {
    await ensure(api, nsID, modules.labor_norms, laborByName, row.name, row)
  }

  const ctypeRows = modules.construction_types ? await listAll(api, nsID, modules.construction_types) : []
  const ctypeByCode = indexBy(ctypeRows, 'code')
  if (modules.construction_types) {
    for (const row of CONSTRUCTION_TYPES) {
      await ensure(api, nsID, modules.construction_types, ctypeByCode, row.code, row)
    }
  }

  const tplRows = modules.wbs_templates ? await listAll(api, nsID, modules.wbs_templates) : []
  const tplByKey = new Map()
  for (const r of tplRows) {
    const key = `${recVal(r, 'construction_type')}:${recVal(r, 'code')}`
    if (recVal(r, 'code')) tplByKey.set(key, recID(r))
  }
  if (modules.wbs_templates) {
    for (const row of WBS_TEMPLATES) {
      const typeID = ctypeByCode.get(row.type)
      if (!typeID) continue
      const key = `${typeID}:${row.code}`
      await ensure(api, nsID, modules.wbs_templates, tplByKey, key, {
        construction_type: typeID,
        code: row.code,
        name: row.name,
        level: row.level,
        parent_code: row.parent_code,
        predecessor_code: row.predecessor_code,
        budget_planned: row.budget_planned,
        duration_days: row.duration_days,
      })
    }
  }

  if (modules.phase_requirements) {
    const preqRows = await listAll(api, nsID, modules.phase_requirements)
    const preqByKey = new Map()
    for (const r of preqRows) {
      preqByKey.set(`${recVal(r, 'phase')}:${recVal(r, 'doc_type')}`, recID(r))
    }
    for (const row of PHASE_REQ) {
      const typeID = typesByCode.get(row.type)
      if (!typeID) continue
      const key = `${row.phase}:${typeID}`
      await ensure(api, nsID, modules.phase_requirements, preqByKey, key, {
        phase: row.phase,
        doc_type: typeID,
        required: '1',
      })
    }
  }

  const advRows = await listAll(api, nsID, modules.ai_advisors)
  const advByName = indexBy(advRows, 'name')
  for (const row of ADVISORS) {
    await ensure(api, nsID, modules.ai_advisors, advByName, row.name, row)
  }

  const projectRows = await listAll(api, nsID, modules.projects)
  const projectByCode = indexBy(projectRows, 'code')

  const memberRows = modules.project_members ? await listAll(api, nsID, modules.project_members) : []
  const memberByKey = new Map()
  for (const r of memberRows) {
    const key = `${recVal(r, 'project')}:${recVal(r, 'user')}:${recVal(r, 'role')}`
    if (recVal(r, 'project') && recVal(r, 'user')) memberByKey.set(key, recID(r))
  }

  const wbsRows = await listAll(api, nsID, modules.wbs_items)
  const wbsByKey = new Map()
  for (const r of wbsRows) {
    const key = `${recVal(r, 'project')}:${recVal(r, 'code')}`
    if (recVal(r, 'code')) wbsByKey.set(key, recID(r))
  }

  const contractRows = await listAll(api, nsID, modules.contracts)
  const contractByNumber = indexBy(contractRows, 'number')

  const docRows = await listAll(api, nsID, modules.documents)
  const docByNumber = indexBy(docRows, 'number')

  const riskRows = await listAll(api, nsID, modules.risks)
  const riskByTitle = indexBy(riskRows, 'title')

  const rfcRows = await listAll(api, nsID, modules.change_requests)
  const rfcByTitle = indexBy(rfcRows, 'title')

  const logRows = await listAll(api, nsID, modules.change_log)
  const logBySummary = indexBy(logRows, 'summary')

  const budgetRows = await listAll(api, nsID, modules.budget_lines)
  const budgetByArticle = indexBy(budgetRows, 'article')

  const cashRows = await listAll(api, nsID, modules.cashflow_items)
  const cashByDesc = indexBy(cashRows, 'description')

  const progressRows = await listAll(api, nsID, modules.progress_facts)
  const progressByNotes = indexBy(progressRows, 'notes')

  let createdProjects = 0
  for (const spec of PROJECTS) {
    const existed = projectByCode.has(spec.code)
    const typeID = spec.type ? (ctypeByCode.get(spec.type) || '') : ''
    const projectID = await ensure(api, nsID, modules.projects, projectByCode, spec.code, {
      name: spec.name,
      code: spec.code,
      phase: spec.phase,
      status: spec.status,
      investor: spec.investor,
      start_planned: spec.start_planned,
      end_planned: spec.end_planned,
      budget_planned: spec.budget_planned,
      budget_actual: spec.budget_actual,
      eac: spec.eac,
      spi: spec.spi,
      cpi: spec.cpi,
      construction_type: typeID,
      description: spec.description,
    })
    if (typeID) {
      await patchRecord(api, nsID, modules.projects, projectID, { construction_type: typeID })
    }
    if (!existed) {
      createdProjects++
      console.log('  project', spec.code, spec.phase, spec.status, projectID)
    }

    const wbsIDs = {}
    for (const item of spec.wbs || []) {
      const key = `${projectID}:${item.code}`
      const parentID = item.parent ? wbsIDs[item.parent] || wbsByKey.get(`${projectID}:${item.parent}`) : ''
      const predID = item.predecessor ? wbsIDs[item.predecessor] || wbsByKey.get(`${projectID}:${item.predecessor}`) : ''
      const id = await ensure(api, nsID, modules.wbs_items, wbsByKey, key, {
        project: projectID,
        parent: parentID,
        predecessor: predID,
        code: item.code,
        name: item.name,
        level: item.level,
        start_planned: item.start_planned,
        end_planned: item.end_planned,
        start_actual: item.start_actual,
        end_actual: item.end_actual,
        budget_planned: item.budget_planned,
        actual_cost: item.actual_cost,
        percent_complete: item.percent_complete,
        notes: item.notes,
      })
      wbsIDs[item.code] = id
    }

    for (const c of spec.contracts || []) {
      await ensure(api, nsID, modules.contracts, contractByNumber, c.number, {
        project: projectID,
        number: c.number,
        title: c.title,
        counterparty: cpByInn.get(c.cp) || '',
        amount: c.amount,
        start_date: c.start_date,
        end_date: c.end_date,
        status: c.status,
        terms: c.terms,
      })
    }

    for (const d of spec.documents || []) {
      const docID = await ensure(api, nsID, modules.documents, docByNumber, d.number, {
        title: d.title,
        number: d.number,
        doc_type: typesByCode.get(d.type) || '',
        project: projectID,
        wbs: d.wbs ? (wbsIDs[d.wbs] || wbsByKey.get(`${projectID}:${d.wbs}`) || '') : '',
        contract: d.contract ? (contractByNumber.get(d.contract) || '') : '',
        status: d.status,
        sign_status: d.sign,
        due_date: d.due_date,
        notes: d.notes,
        author: userID,
      })
      if (userID) {
        await patchRecord(api, nsID, modules.documents, docID, { author: userID })
      }
    }

    if (userID && modules.project_members) {
      for (const role of ['pmo', 'investor']) {
        const key = `${projectID}:${userID}:${role}`
        await ensure(api, nsID, modules.project_members, memberByKey, key, {
          project: projectID,
          user: userID,
          role,
        })
      }
    }

    for (const r of spec.risks || []) {
      await ensure(api, nsID, modules.risks, riskByTitle, r.title, {
        project: projectID,
        wbs: r.wbs ? (wbsIDs[r.wbs] || '') : '',
        title: r.title,
        probability: r.probability,
        impact: r.impact,
        score: riskScore(r.probability, r.impact),
        status: r.status,
        mitigation: r.mitigation,
        description: r.description,
      })
    }

    for (const r of spec.rfc || []) {
      await ensure(api, nsID, modules.change_requests, rfcByTitle, r.title, {
        project: projectID,
        wbs: r.wbs ? (wbsIDs[r.wbs] || '') : '',
        title: r.title,
        rfc_type: r.rfc_type,
        status: r.status,
        delta_budget: r.delta_budget,
        delta_days: r.delta_days,
        eac_before: r.eac_before,
        eac_after: r.eac_after,
        justification: r.justification,
      })
    }

    for (const row of spec.changelog || []) {
      await ensure(api, nsID, modules.change_log, logBySummary, row.summary, {
        rfc: rfcByTitle.get(row.rfc) || '',
        project: projectID,
        summary: row.summary,
        old_budget: row.old_budget,
        new_budget: row.new_budget,
        old_end: row.old_end,
        new_end: row.new_end,
        changed_at: row.changed_at,
      })
    }

    for (const b of spec.budget || []) {
      await ensure(api, nsID, modules.budget_lines, budgetByArticle, b.article, {
        project: projectID,
        wbs: b.wbs ? (wbsIDs[b.wbs] || wbsByKey.get(`${projectID}:${b.wbs}`) || '') : '',
        article: b.article,
        planned: b.planned,
        actual: b.actual,
        reserve: b.reserve,
        notes: b.notes,
      })
    }

    for (const c of spec.cashflow || []) {
      await ensure(api, nsID, modules.cashflow_items, cashByDesc, c.description, {
        project: projectID,
        budget_line: budgetByArticle.get(c.article) || '',
        date: c.date,
        amount: c.amount,
        direction: c.direction,
        description: c.description,
      })
    }

    for (const p of spec.progress || []) {
      await ensure(api, nsID, modules.progress_facts, progressByNotes, p.notes, {
        project: projectID,
        wbs: wbsIDs[p.wbs] || wbsByKey.get(`${projectID}:${p.wbs}`) || '',
        quantity: p.quantity,
        unit: p.unit,
        percent: p.percent,
        cost: p.cost,
        recorded_at: p.recorded_at,
        notes: p.notes,
      })
    }
  }

  console.log(`seeded invest demo (${createdProjects} new projects, rest reused by code)`)
  try {
    await seedDocumentFiles(api, nsID, modules, opts)
  } catch (e) {
    console.warn('document files:', e.message || e)
  }
}

async function main () {
  const applied = JSON.parse(readFileSync(join(HERE, 'applied.json'), 'utf8'))
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  console.log('API', base, 'ns', applied.namespaceID)
  await seedIfEmpty(api, applied.namespaceID, applied.modules, userFromToken(token), { token, base })
}

const isDirect = process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]
if (isDirect) {
  main().catch(err => {
    console.error(err)
    process.exit(1)
  })
}
