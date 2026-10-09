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
  ensurePage, ensureRuleChain, parentPages, doughnutChart, ganttChart, barChart,
  withRevisions, detectSystem, ensureRole, moduleResource, recordResource, fieldResource,
} from './helpers.mjs'
import {
  documentTypeFields, counterpartyFields, materialFields, laborNormFields,
  constructionTypeFields, wbsTemplateFields, phaseRequirementFields,
  projectFields, projectMemberFields, wbsFields, contractFields, documentFields,
  documentVersionFields, approvalFields, commentFields, riskFields, rfcFields,
  changeLogFields, budgetLineFields, cashflowFields, progressFactFields, aiAdvisorFields,
} from './fields.mjs'
import { buildPages } from './pages.mjs'
import { buildRuleChains } from './chains.mjs'
import { seedIfEmpty, userFromToken } from './seed.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const NS_META = {
  subtitle: 'Сопровождение инвестиционных проектов',
  description: 'Единое пространство PMO: документы, WBS, договоры, бюджет, EVM, RFC и ИИ-советчики.',
  prompt: 'Это пространство сопровождения инвестиционных проектов (единый источник правды). Модули: projects, wbs_items, documents, document_versions, approvals, contracts, risks, change_requests, change_log, budget_lines, cashflow_items, progress_facts, project_members, document_types, counterparties, materials, labor_norms, construction_types, wbs_templates, phase_requirements, ai_advisors. ИИ — только советчик (human-in-the-loop): рекомендации, решение утверждает человек. Юрист — договоры и документы. Финконтролёр — бюджет, EVM, отклонения. Не выдумывай цифры — читай записи инструментами.',
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  const engineUrl = (process.env.INVEST_ENGINE_URL || 'http://localhost:8086/api').replace(/\/$/, '')
  const evmUrl = (process.env.CALC_EVM_URL || 'http://localhost:8088/api').replace(/\/$/, '')

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
  const construction_types = await ensureModule(api, nsID, { name: 'Типы конструкций', handle: 'construction_types', fields: constructionTypeFields() })
  const wbs_templates = await ensureModule(api, nsID, { name: 'Шаблоны WBS', handle: 'wbs_templates', fields: wbsTemplateFields(construction_types) })
  const phase_requirements = await ensureModule(api, nsID, { name: 'Документы фазы', handle: 'phase_requirements', fields: phaseRequirementFields(document_types) })
  const projects = await ensureModule(api, nsID, {
    name: 'Проекты',
    handle: 'projects',
    fields: projectFields(construction_types),
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
    construction_types, wbs_templates, phase_requirements,
    projects, project_members, wbs_items, contracts, documents,
    document_versions, approvals, document_comments, risks,
    change_requests, change_log, budget_lines, cashflow_items,
    progress_facts, ai_advisors,
  }

  const docsByStatus = await ensureChart(api, nsID, doughnutChart('Документы по статусу', 'docs-by-status', documents, 'status'))
  const rfcByStatus = await ensureChart(api, nsID, doughnutChart('RFC по статусу', 'rfc-by-status', change_requests, 'status'))
  const risksByImpact = await ensureChart(api, nsID, doughnutChart('Риски по влиянию', 'risks-by-impact', risks, 'impact'))
  const risksByScore = await ensureChart(api, nsID, doughnutChart('Риски по баллу', 'risks-by-score', risks, 'score'))
  const cashflowByDir = await ensureChart(api, nsID, barChart('Денежный поток', 'cashflow-by-dir', cashflow_items, 'direction', 'amount', { aggregate: 'SUM', type: 'bar' }))
  const cashflowByDate = await ensureChart(api, nsID, barChart('Поток по датам', 'cashflow-by-date', cashflow_items, 'date', 'amount', { aggregate: 'SUM', type: 'line' }))
  const wbsGantt = await ensureChart(api, nsID, ganttChart('График WBS', 'wbs-gantt', wbs_items, 'name', 'start_planned', 'end_planned'))
  const charts = { docsByStatus, rfcByStatus, risksByImpact, risksByScore, cashflowByDir, cashflowByDate, wbsGantt }

  const roles = await provisionRoles(token, base, nsID, modules)

  const pageIDs = {}
  for (const page of buildPages({ modules, charts, roles })) {
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
    'construction-type': 'nsi',
    'wbs-template': 'nsi',
    'phase-requirement': 'nsi',
    counterparty: 'nsi',
    material: 'nsi',
    'labor-norm': 'nsi',
  })

  for (const chain of buildRuleChains({ nsID, modules, engineUrl, evmUrl })) {
    await ensureRuleChain(api, chain)
  }

  try {
    await seedIfEmpty(api, nsID, modules, userFromToken(token), { token, base })
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
      recalculate: `POST ${evmUrl}/call/evm`,
      simulateRfc: `POST ${engineUrl}/simulate-rfc`,
      submitApproval: `POST ${engineUrl}/submit-approval`,
    },
    roles,
  }
  writeFileSync(join(HERE, 'applied.json'), JSON.stringify(summary, null, 2))
  console.log('\nInvest ready')
  console.log(JSON.stringify(summary, null, 2))
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})

function allow (roleID, resource, operation) {
  return { roleID: String(roleID), resource, operation, access: 'allow' }
}

function moduleRules (roleID, nsID, moduleID, write) {
  const rules = [
    allow(roleID, moduleResource(nsID, moduleID), 'read'),
    allow(roleID, moduleResource(nsID, moduleID), 'records.search'),
    allow(roleID, recordResource(nsID, moduleID), 'read'),
    allow(roleID, fieldResource(nsID, moduleID), 'record.value.read'),
  ]
  if (write) {
    rules.push(
      allow(roleID, moduleResource(nsID, moduleID), 'record.create'),
      allow(roleID, recordResource(nsID, moduleID), 'update'),
      allow(roleID, fieldResource(nsID, moduleID), 'record.value.update'),
    )
  }
  return rules
}

async function provisionRoles (token, composeBase, nsID, modules) {
  const sysBase = await detectSystem(token, composeBase)
  if (!sysBase) {
    console.warn('system API not found — skip RBAC roles')
    return {}
  }
  const sys = apiFactory(sysBase, token)
  const defs = [
    { key: 'investor', handle: 'invest-investor', name: 'Инвестор / Заказчик' },
    { key: 'bank', handle: 'invest-bank', name: 'Банк / Кредитор' },
    { key: 'contractor', handle: 'invest-contractor', name: 'Генподрядчик' },
    { key: 'designer', handle: 'invest-designer', name: 'ГИП / Проектировщик' },
    { key: 'government', handle: 'invest-government', name: 'Гос. органы' },
    { key: 'pmo', handle: 'invest-pmo', name: 'PMO' },
  ]
  const roles = {}
  for (const d of defs) {
    try {
      roles[d.key] = await ensureRole(sys, d)
    } catch (e) {
      console.warn('role', d.handle, e.message)
    }
  }

  const allHandles = Object.keys(modules)
  const finance = ['projects', 'budget_lines', 'cashflow_items', 'risks', 'change_requests', 'change_log']
  const specs = {
    investor: Object.fromEntries(allHandles.map(h => [h, true])),
    pmo: Object.fromEntries(allHandles.map(h => [h, true])),
    bank: Object.fromEntries(finance.map(h => [h, false])),
    contractor: {
      projects: false, wbs_items: false, progress_facts: true, documents: true,
      document_versions: true, approvals: true, contracts: false, risks: false,
    },
    designer: {
      projects: false, wbs_items: true, documents: true, document_versions: true,
      approvals: true, contracts: false, document_types: false,
    },
    government: {
      projects: false, documents: false, document_types: false, risks: false,
    },
  }

  for (const [key, spec] of Object.entries(specs)) {
    const roleID = roles[key]
    if (!roleID) continue
    const rules = [
      allow(roleID, `corteza::compose:namespace/${nsID}`, 'read'),
      allow(roleID, `corteza::compose:namespace/${nsID}`, 'pages.search'),
      allow(roleID, `corteza::compose:namespace/${nsID}`, 'modules.search'),
      allow(roleID, `corteza::compose:namespace/${nsID}`, 'charts.search'),
    ]
    for (const [handle, write] of Object.entries(spec)) {
      if (!modules[handle]) continue
      rules.push(...moduleRules(roleID, nsID, modules[handle], write))
    }
    try {
      await sys('PATCH', `/permissions/${roleID}/rules`, { rules })
      console.log('granted', key, rules.length, 'rules')
    } catch (e) {
      console.warn('grant', key, e.message)
    }
  }
  return roles
}
