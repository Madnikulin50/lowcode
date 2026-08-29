#!/usr/bin/env node
/**
 * Provision Al-Faris Holding procurement namespace (modules, charts, pages, rule chains, seed).
 *
 *   TOKEN=$(node /tmp/opencode/mint-token.mjs | head -1) \
 *   node apply.mjs
 *
 * Idempotent by handle. Seed runs only when subsidiaries are empty, or FARIS_RESEED=1.
 */
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { writeFileSync } from 'node:fs'

import {
  mintToken, detectBase, apiFactory, ensureNamespace, ensureModule, ensureChart,
  ensurePage, ensureRuleChain, parentPages, doughnutChart, barChart, withRevisions,
} from './helpers.mjs'
import {
  subsidiaryFields, categoryFields, erpVendorFields, erpBudgetFields,
  vendorFields, purchaseRequestFields, approvalLogFields,
} from './fields.mjs'
import { buildPages } from './pages.mjs'
import { buildRuleChains } from './chains.mjs'
import { seedIfEmpty } from './seed.mjs'
import { attachPrompts } from '../../../scripts/ai-prompts.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const NS_META = {
  subtitle: 'Vendor management & procurement for Al-Faris Holding Group',
  description: 'Unified vendor onboarding, purchase approvals and group spend visibility across eight Saudi subsidiaries. ERP vendor master and budgets are mocked (SAP-style).',
  prompt: `This is Al-Faris Holding Group procurement (eight subsidiaries).

You are a group procurement and finance controller. Use module records only — do not invent CR numbers, VAT IDs, budgets or cycle times.

Explain vendor onboarding bottlenecks, purchase-request queues, over-budget holds, and where spend concentrates by subsidiary or category. End with a prioritized action list for this week. Reply in English.`,
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  console.log('API', base)

  const nsID = await ensureNamespace(api, {
    name: 'Al-Faris Procurement',
    slug: 'faris',
    meta: NS_META,
  })

  const subsidiaries = await ensureModule(api, nsID, {
    name: 'Subsidiaries',
    handle: 'subsidiaries',
    fields: subsidiaryFields(),
  })
  const spend_categories = await ensureModule(api, nsID, {
    name: 'Spend categories',
    handle: 'spend_categories',
    fields: categoryFields(),
  })
  const erp_vendors = await ensureModule(api, nsID, {
    name: 'ERP vendor master',
    handle: 'erp_vendors',
    fields: erpVendorFields(),
  })
  const erp_budgets = await ensureModule(api, nsID, {
    name: 'ERP budgets',
    handle: 'erp_budgets',
    fields: erpBudgetFields(subsidiaries),
    config: {
      etl: {
        enabled: false,
        sourceType: 'rest',
        format: 'json',
        restUrl: 'https://sap.alfaris.example/api/budgets',
        restMethod: 'GET',
      },
    },
  })
  const vendors = await ensureModule(api, nsID, {
    name: 'Vendors',
    handle: 'vendors',
    fields: vendorFields(subsidiaries, erp_vendors),
    config: withRevisions(),
  })
  const purchase_requests = await ensureModule(api, nsID, {
    name: 'Purchase requests',
    handle: 'purchase_requests',
    fields: purchaseRequestFields(subsidiaries, vendors, erp_budgets),
    config: withRevisions(),
  })
  const approval_log = await ensureModule(api, nsID, {
    name: 'Approval log',
    handle: 'approval_log',
    fields: approvalLogFields(vendors, purchase_requests),
  })

  const modules = {
    subsidiaries, spend_categories, erp_vendors, erp_budgets,
    vendors, purchase_requests, approval_log,
  }

  const prBySubsidiary = await ensureChart(api, nsID, doughnutChart(
    'Purchase requests by subsidiary', 'pr-by-subsidiary', purchase_requests, 'subsidiary_code',
  ))
  const spendByCategory = await ensureChart(api, nsID, barChart(
    'Mock spend by category', 'spend-by-category', purchase_requests, 'category', 'estimated_value',
    { aggregate: 'SUM', label: 'Estimated SAR', filter: "status = 'approved'" },
  ))
  const vendorsByStatus = await ensureChart(api, nsID, doughnutChart(
    'Vendor onboarding by status', 'vendors-by-status', vendors, 'status',
  ))
  const spendBySubsidiary = await ensureChart(api, nsID, barChart(
    'Spend by subsidiary', 'spend-by-subsidiary', purchase_requests, 'subsidiary_code', 'estimated_value',
    { aggregate: 'SUM', label: 'Estimated SAR', filter: "status = 'approved'" },
  ))
  const prByStatus = await ensureChart(api, nsID, doughnutChart(
    'Purchase requests by status', 'pr-by-status', purchase_requests, 'status',
  ))
  const charts = { prBySubsidiary, spendByCategory, vendorsByStatus, spendBySubsidiary, prByStatus }

  const pages = attachPrompts(buildPages({ modules, charts }), 'faris')
  const pageIDs = {}
  for (const page of pages) {
    pageIDs[page.handle] = await ensurePage(api, nsID, page)
  }

  await parentPages(api, nsID, pageIDs, {
    vendor: 'vendors',
    'purchase-request': 'purchase-requests',
    subsidiary: 'subsidiaries',
    'erp-vendor': 'erp-mirror',
    'erp-budget': 'erp-mirror',
    'approval-log-item': 'purchase-requests',
  })

  for (const chain of buildRuleChains({ nsID, modules })) {
    await ensureRuleChain(api, chain)
  }

  try {
    await seedIfEmpty(api, nsID, modules, { force: process.env.FARIS_RESEED === '1' })
  } catch (e) {
    console.warn('seed skipped:', e.message)
  }

  const summary = {
    namespaceID: String(nsID),
    slug: 'faris',
    api: base,
    modules: Object.fromEntries(Object.entries(modules).map(([k, v]) => [k, String(v)])),
    charts: Object.fromEntries(Object.entries(charts).map(([k, v]) => [k, String(v)])),
    urls: {
      namespace: '/ns/faris',
      dashboard: '/ns/faris/pages',
    },
  }
  writeFileSync(join(HERE, 'applied.json'), JSON.stringify(summary, null, 2))
  console.log('\nAl-Faris Procurement ready')
  console.log(JSON.stringify(summary, null, 2))
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
