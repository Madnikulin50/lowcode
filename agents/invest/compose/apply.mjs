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
  withRevisions,
} from './helpers.mjs'
import {
  documentTypeFields, counterpartyFields, materialFields, laborNormFields,
  projectFields, projectMemberFields, wbsFields, contractFields, documentFields,
  documentVersionFields, approvalFields, commentFields, riskFields, rfcFields,
  changeLogFields, budgetLineFields, cashflowFields, progressFactFields, aiAdvisorFields,
} from './fields.mjs'
import { buildPages } from './pages.mjs'
import { buildRuleChains } from './chains.mjs'
import { seedIfEmpty } from './seed.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const NS_META = {
  subtitle: 'Сопровождение инвестиционных проектов',
  description: 'Единое пространство PMO: документы, WBS, договоры, бюджет, EVM, RFC и ИИ-советчики.',
  prompt: 'Это пространство сопровождения инвестиционных проектов (единый источник правды). Модули: projects, wbs_items, documents, document_versions, approvals, contracts, risks, change_requests, change_log, budget_lines, cashflow_items, progress_facts, project_members, document_types, counterparties, materials, labor_norms, ai_advisors. ИИ — только советчик (human-in-the-loop): рекомендации, решение утверждает человек. Юрист — договоры и документы. Финконтролёр — бюджет, EVM, отклонения. Не выдумывай цифры — читай записи инструментами.',
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
