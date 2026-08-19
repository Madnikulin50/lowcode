#!/usr/bin/env node
/**
 * Provision the Invest Corteza namespace (modules, charts, pages, layouts, rule chains).
 *
 *   COMPOSE_API=http://localhost:3333/api/compose \
 *   TOKEN=$(node /tmp/opencode/mint-token.mjs | head -1) \
 *   node apply.mjs
 *
 * Idempotent: existing resources are reused / updated by handle.
 */
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { writeFileSync } from 'node:fs'

import {
  mintToken, detectBase, apiFactory, ensureNamespace, ensureModule, ensureChart,
  ensurePage, ensureRuleChain, parentPages, doughnutChart, ganttChart,
  withRevisions, createRecord, setOf,
} from './helpers.mjs'
import {
  documentTypeFields, counterpartyFields, materialFields, laborNormFields,
  projectFields, projectMemberFields, wbsFields, contractFields, documentFields,
  documentVersionFields, approvalFields, commentFields, riskFields, rfcFields,
  changeLogFields, budgetLineFields, cashflowFields, progressFactFields, aiAdvisorFields,
} from './fields.mjs'
import { buildPages } from './pages.mjs'
import { buildRuleChains } from './chains.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const NS_META = {
  subtitle: 'Сопровождение инвестиционных проектов',
  description: 'Единое пространство PMO: документы, WBS, договоры, бюджет, EVM, RFC и ИИ-советчики.',
  prompt: 'Это пространство сопровождения инвестиционных проектов (единый источник правды). Модули: projects, wbs_items, documents, document_versions, approvals, contracts, risks, change_requests, change_log, budget_lines, cashflow_items, progress_facts, project_members, document_types, counterparties, materials, labor_norms, ai_advisors. ИИ — только советчик (human-in-the-loop): рекомендации, решение утверждает человек. Юрист — договоры и документы. Финконтролёр — бюджет, EVM, отклонения. Не выдумывай цифры — читай записи инструментами.',
}

async function seedIfEmpty (api, nsID, modules) {
  const existing = setOf(await api('GET', `/namespace/${nsID}/module/${modules.projects}/record/?limit=5`))
  if (existing.length) {
    console.log('projects already have records, skip seed')
    return
  }

  for (const row of [
    { name: 'Договор', code: 'CONTRACT' },
    { name: 'ПСД', code: 'PSD' },
    { name: 'Смета', code: 'ESTIMATE' },
    { name: 'Акт КС-2', code: 'KS2' },
    { name: 'Акт КС-3', code: 'KS3' },
    { name: 'Исполнительная документация', code: 'AS_BUILT' },
    { name: 'Приказ', code: 'ORDER' },
  ]) {
    await createRecord(api, nsID, modules.document_types, row)
  }

  const cp = await createRecord(api, nsID, modules.counterparties, {
    name: 'ООО Генподрядчик-1', inn: '7700000001', role: 'contractor',
  })
  await createRecord(api, nsID, modules.materials, {
    name: 'Бетон B25', unit: 'м³', unit_price: '8500', gost: 'ГОСТ 26633',
  })
  await createRecord(api, nsID, modules.labor_norms, {
    name: 'Укладка бетона', unit: 'м³', hours: '2.5',
  })

  const project = await createRecord(api, nsID, modules.projects, {
    name: 'Пилотный энергообъект',
    code: 'ENRG-001',
    phase: 'construction',
    status: 'active',
    investor: 'Инвестор А',
    start_planned: '2026-01-01',
    end_planned: '2027-12-31',
    budget_planned: '1500000000',
    budget_actual: '120000000',
    description: 'Демонстрационный инвестиционный проект для пространства документооборота.',
  })
  const projectID = String(project.recordID || project.ID)

  const stage = await createRecord(api, nsID, modules.wbs_items, {
    project: projectID,
    code: '1',
    name: 'Строительство',
    level: 'stage',
    start_planned: '2026-03-01',
    end_planned: '2027-06-30',
    budget_planned: '900000000',
    percent_complete: '12',
  })
  const stageID = String(stage.recordID || stage.ID)
  const workA = await createRecord(api, nsID, modules.wbs_items, {
    project: projectID,
    parent: stageID,
    code: '1.1',
    name: 'Фундамент',
    level: 'work',
    start_planned: '2026-03-01',
    end_planned: '2026-06-30',
    budget_planned: '180000000',
    actual_cost: '40000000',
    percent_complete: '25',
  })
  const workAID = String(workA.recordID || workA.ID)
  await createRecord(api, nsID, modules.wbs_items, {
    project: projectID,
    parent: stageID,
    predecessor: workAID,
    code: '1.2',
    name: 'Каркас',
    level: 'work',
    start_planned: '2026-07-01',
    end_planned: '2026-12-31',
    budget_planned: '320000000',
    actual_cost: '0',
    percent_complete: '0',
  })

  const contract = await createRecord(api, nsID, modules.contracts, {
    project: projectID,
    number: 'ГП-001/2026',
    title: 'Генеральный подряд',
    counterparty: String(cp.recordID || cp.ID),
    amount: '900000000',
    start_date: '2026-02-01',
    end_date: '2027-12-31',
    status: 'active',
    terms: 'Оплата по актам КС-2/КС-3. Штраф за просрочку 0.1% в день.',
  })

  await createRecord(api, nsID, modules.documents, {
    title: 'Договор генподряда',
    number: 'ГП-001/2026',
    project: projectID,
    contract: String(contract.recordID || contract.ID),
    status: 'in_review',
    sign_status: 'unsigned',
    due_date: '2026-09-01',
    notes: 'Черновик на согласовании юриста.',
  })

  await createRecord(api, nsID, modules.budget_lines, {
    project: projectID,
    wbs: workAID,
    article: 'Фундамент — бетон',
    planned: '180000000',
    actual: '40000000',
    reserve: '15000000',
  })

  await createRecord(api, nsID, modules.risks, {
    project: projectID,
    wbs: workAID,
    title: 'Срыв поставки арматуры',
    probability: 'medium',
    impact: 'high',
    status: 'open',
    mitigation: 'Дублирующий поставщик, запас 14 дней.',
  })

  await createRecord(api, nsID, modules.change_requests, {
    project: projectID,
    wbs: workAID,
    title: 'Увеличение объёма фундамента',
    rfc_type: 'scope',
    status: 'draft',
    delta_budget: '25000000',
    delta_days: '14',
    justification: 'По результатам изысканий нужна доп. подушка.',
  })

  await createRecord(api, nsID, modules.progress_facts, {
    project: projectID,
    wbs: workAID,
    quantity: '120',
    unit: 'м³',
    percent: '25',
    cost: '40000000',
    recorded_at: new Date().toISOString(),
    notes: 'Первая заливка. Фото с площадки — веб-форма (мобильный офлайн — отдельное согласование).',
  })

  await createRecord(api, nsID, modules.ai_advisors, {
    name: 'Юрист',
    role: 'lawyer',
    enabled: '1',
    modules: 'contracts,documents',
    prompt: 'Human-in-the-loop юрист: анализируй договоры и документы, не утверждай сам.',
  })
  await createRecord(api, nsID, modules.ai_advisors, {
    name: 'Финконтролёр',
    role: 'fincontroller',
    enabled: '1',
    modules: 'budget_lines,cashflow_items,wbs_items,change_requests',
    prompt: 'Human-in-the-loop финконтролёр: план/факт, EVM, RFC. Не меняй цифры сам.',
  })

  console.log('seeded demo project ENRG-001')
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const engineUrl = (process.env.INVEST_ENGINE_URL || 'http://localhost:8086/api').replace(/\/$/, '')

  console.log('API', base)

  const nsID = await ensureNamespace(api, {
    name: 'Инвестпроекты',
    slug: 'invest',
    meta: NS_META,
  })

  const document_types = await ensureModule(api, nsID, { name: 'Типы документов', handle: 'document_types', fields: documentTypeFields() })
  const counterparties = await ensureModule(api, nsID, { name: 'Контрагенты', handle: 'counterparties', fields: counterpartyFields() })
  const materials = await ensureModule(api, nsID, { name: 'Материалы', handle: 'materials', fields: materialFields() })
  const labor_norms = await ensureModule(api, nsID, { name: 'Нормы времени', handle: 'labor_norms', fields: laborNormFields() })
  const projects = await ensureModule(api, nsID, {
    name: 'Проекты',
    handle: 'projects',
    fields: projectFields(),
    config: withRevisions(),
  })
  const project_members = await ensureModule(api, nsID, { name: 'Участники проекта', handle: 'project_members', fields: projectMemberFields(projects) })

  let wbs_items = await ensureModule(api, nsID, {
    name: 'WBS',
    handle: 'wbs_items',
    fields: wbsFields(projects, '0'),
    config: withRevisions(),
  })
  wbs_items = await ensureModule(api, nsID, {
    name: 'WBS',
    handle: 'wbs_items',
    fields: wbsFields(projects, wbs_items),
    config: withRevisions(),
  })

  const contracts = await ensureModule(api, nsID, {
    name: 'Договоры',
    handle: 'contracts',
    fields: contractFields(projects, counterparties),
    config: withRevisions(),
  })
  const documents = await ensureModule(api, nsID, {
    name: 'Документы',
    handle: 'documents',
    fields: documentFields(projects, wbs_items, contracts, document_types),
    config: withRevisions(),
  })
  const document_versions = await ensureModule(api, nsID, { name: 'Версии документов', handle: 'document_versions', fields: documentVersionFields(documents) })
  const approvals = await ensureModule(api, nsID, { name: 'Согласования', handle: 'approvals', fields: approvalFields(documents) })
  const document_comments = await ensureModule(api, nsID, { name: 'Комментарии к документам', handle: 'document_comments', fields: commentFields(documents) })
  const risks = await ensureModule(api, nsID, { name: 'Риски', handle: 'risks', fields: riskFields(projects, wbs_items) })
  const change_requests = await ensureModule(api, nsID, {
    name: 'RFC',
    handle: 'change_requests',
    fields: rfcFields(projects, wbs_items),
    config: withRevisions(),
  })
  const change_log = await ensureModule(api, nsID, { name: 'Журнал изменений', handle: 'change_log', fields: changeLogFields(change_requests, projects) })
  const budget_lines = await ensureModule(api, nsID, {
    name: 'Статьи бюджета',
    handle: 'budget_lines',
    fields: budgetLineFields(projects, wbs_items),
    config: withRevisions({
      etl: {
        enabled: false,
        sourceType: 'rest',
        format: 'json',
        restUrl: 'http://1c.local/odata/standard.odata/',
        restMethod: 'GET',
      },
    }),
  })
  const cashflow_items = await ensureModule(api, nsID, { name: 'Денежный поток', handle: 'cashflow_items', fields: cashflowFields(projects, budget_lines) })
  const progress_facts = await ensureModule(api, nsID, { name: 'Факты прогресса', handle: 'progress_facts', fields: progressFactFields(projects, wbs_items) })
  const ai_advisors = await ensureModule(api, nsID, { name: 'ИИ-советчики', handle: 'ai_advisors', fields: aiAdvisorFields() })

  const modules = {
    document_types, counterparties, materials, labor_norms,
    projects, project_members, wbs_items, contracts, documents,
    document_versions, approvals, document_comments, risks,
    change_requests, change_log, budget_lines, cashflow_items,
    progress_facts, ai_advisors,
  }

  const docsByStatus = await ensureChart(api, nsID, doughnutChart('Документы по статусу', 'docs-by-status', documents, 'status'))
  const rfcByStatus = await ensureChart(api, nsID, doughnutChart('RFC по статусу', 'rfc-by-status', change_requests, 'status'))
  const risksByImpact = await ensureChart(api, nsID, doughnutChart('Риски по влиянию', 'risks-by-impact', risks, 'impact'))
  const wbsGantt = await ensureChart(api, nsID, ganttChart('График WBS', 'wbs-gantt', wbs_items, 'name', 'start_planned', 'end_planned'))
  const charts = { docsByStatus, rfcByStatus, risksByImpact, wbsGantt }

  const pageIDs = {}
  for (const page of buildPages({ modules, charts })) {
    pageIDs[page.handle] = await ensurePage(api, nsID, page)
  }

  await parentPages(api, nsID, pageIDs, {
    project: 'projects',
    'wbs-item': 'wbs',
    document: 'documents',
    'document-version': 'documents',
    approval: 'documents',
    contract: 'contracts',
    risk: 'risks',
    'change-request': 'changes',
    'change-log-item': 'changes',
    'budget-line': 'budget',
    'cashflow-item': 'budget',
    'progress-fact': 'progress',
    'project-member': 'projects',
    'document-type': 'nsi',
    counterparty: 'nsi',
    material: 'nsi',
    'labor-norm': 'nsi',
  })

  for (const chain of buildRuleChains({ nsID, modules, engineUrl })) {
    await ensureRuleChain(api, chain)
  }

  try {
    await seedIfEmpty(api, nsID, modules)
  } catch (e) {
    console.warn('seed skipped:', e.message)
  }

  const summary = {
    namespaceID: String(nsID),
    slug: 'invest',
    api: base,
    modules: Object.fromEntries(Object.entries(modules).map(([k, v]) => [k, String(v)])),
    charts: Object.fromEntries(Object.entries(charts).map(([k, v]) => [k, String(v)])),
    urls: {
      namespace: '/ns/invest',
      dashboard: '/ns/invest/pages',
    },
    engine: {
      flags: `--api=${base.replace(/\/compose$/, '')} --namespace=${nsID}`,
      recalculate: `POST ${engineUrl}/recalculate-evm`,
    },
  }
  writeFileSync(join(HERE, 'applied.json'), JSON.stringify(summary, null, 2))
  console.log('\nInvest ready')
  console.log(JSON.stringify(summary, null, 2))
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
