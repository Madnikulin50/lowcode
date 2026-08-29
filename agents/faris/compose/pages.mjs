import {
  block, recordList, recordBlock, metricBlock, metricItem, ruleChain, organizer,
  withBlockIDs, pageIcon,
} from './helpers.mjs'

const recID = '${recordID}'

export function buildPages ({ modules, charts }) {
  const m = modules
  return [
    dashboardPage(m, charts),
    vendorListPage(m, charts),
    approvedVendorsPage(m),
    purchaseListPage(m, charts),
    subsidiariesPage(m),
    erpMirrorPage(m),
    vendorCard(m),
    prCard(m),
    subsidiaryCard(m),
    erpVendorCard(m),
    erpBudgetCard(m),
    approvalLogCard(m),
  ]
}

function dashboardPage (m, charts) {
  return {
    title: 'Group dashboard',
    handle: 'dashboard',
    visible: true,
    weight: 0,
    description: 'Holding-level procurement visibility across eight Al-Faris subsidiaries.',
    config: {
      ...pageIcon('fas chart-pie'),
      prompt: `You are a group procurement controller for Al-Faris Holding.

This dashboard consolidates vendor onboarding and purchase requests across eight Saudi subsidiaries.

Explain what the KPIs and charts show, where cycle time or spend is concentrated, which requests are stalled, and what group procurement / finance should do this week. Use only the numbers on screen. Reply in English.`,
    },
    blocks: withBlockIDs([
      metricBlock('At a glance', [0, 0, 48, 12], [
        metricItem('Active onboarding', m.vendors, "status != 'draft' AND status != 'approved' AND status != 'rejected'", { role: 'hero', color: '#2e59d9', fontSize: '28' }),
        metricItem('Open purchase requests', m.purchase_requests, "status != 'approved' AND status != 'rejected'", { role: 'balloon', color: '#f6c23e' }),
        metricItem('Avg cycle (days)', m.purchase_requests, "status = 'approved'", { role: 'balloon', color: '#1cc88a', field: 'cycle_days', operation: 'avg', fmt: '0.0' }),
        metricItem('Stalled / overdue', m.purchase_requests, 'stalled = true', { role: 'meta', color: '#e74a3b' }),
      ]),
      block('Chart', 'Purchase requests by subsidiary', [0, 12, 24, 20], { chartID: String(charts.prBySubsidiary) }),
      block('Chart', 'Mock spend by category', [24, 12, 24, 20], { chartID: String(charts.spendByCategory) }),
      block('Chart', 'Vendor onboarding by status', [0, 32, 24, 18], { chartID: String(charts.vendorsByStatus) }),
      block('Chart', 'Spend by subsidiary', [24, 32, 24, 18], { chartID: String(charts.spendBySubsidiary) }),
      recordList('Overdue or stalled requests', [0, 50, 48, 18], m.purchase_requests, [
        'pr_number', 'title', 'subsidiary_code', 'estimated_value', 'status', 'due_at',
      ], {
        prefilter: 'stalled = true',
        perPage: 8,
        hideAddButton: true,
      }),
    ]),
  }
}

function vendorListPage (m, charts) {
  return {
    title: 'Vendor onboarding',
    handle: 'vendors',
    visible: true,
    weight: 10,
    config: pageIcon('fas handshake'),
    blocks: withBlockIDs([
      metricBlock('Pipeline', [0, 0, 48, 10], [
        metricItem('Incomplete packs', m.vendors, "status = 'incomplete'", { role: 'balloon', color: '#f6c23e' }),
        metricItem('In review', m.vendors, "status = 'procurement_review' OR status = 'compliance_review' OR status = 'finance_verify'", { role: 'balloon', color: '#2e59d9' }),
        metricItem('Approved', m.vendors, "status = 'approved'", { role: 'hero', color: '#1cc88a' }),
      ], { itemsPerRow: '3' }),
      organizer('Draft', [0, 10, 12, 22], m.vendors, {
        labelField: 'legal_name', descriptionField: 'subsidiary_code', groupField: 'status', group: 'draft',
      }),
      organizer('Submitted', [12, 10, 12, 22], m.vendors, {
        labelField: 'legal_name', descriptionField: 'subsidiary_code', groupField: 'status', group: 'submitted',
      }),
      organizer('Procurement', [24, 10, 12, 22], m.vendors, {
        labelField: 'legal_name', descriptionField: 'subsidiary_code', groupField: 'status', group: 'procurement_review',
      }),
      organizer('Compliance', [36, 10, 12, 22], m.vendors, {
        labelField: 'legal_name', descriptionField: 'subsidiary_code', groupField: 'status', group: 'compliance_review',
      }),
      recordList('All onboarding requests', [0, 32, 48, 24], m.vendors, [
        'legal_name', 'subsidiary_code', 'category', 'cr_number', 'status', 'submitted_at', 'due_at',
      ]),
    ]),
  }
}

function approvedVendorsPage (m) {
  return {
    title: 'Approved vendors',
    handle: 'approved-vendors',
    visible: true,
    weight: 15,
    config: pageIcon('fas check-circle'),
    blocks: withBlockIDs([
      recordList('Group approved vendor list', [0, 0, 48, 40], m.vendors, [
        'legal_name', 'trade_name', 'subsidiary_code', 'category', 'cr_number', 'vat_number', 'city', 'approved_at',
      ], {
        prefilter: "status = 'approved'",
        hideAddButton: true,
      }),
    ]),
  }
}

function purchaseListPage (m, charts) {
  return {
    title: 'Purchase requests',
    handle: 'purchase-requests',
    visible: true,
    weight: 20,
    config: pageIcon('fas file-invoice'),
    blocks: withBlockIDs([
      block('Chart', 'PR status mix', [0, 0, 24, 16], { chartID: String(charts.prByStatus) }),
      metricBlock('Queue', [24, 0, 24, 16], [
        metricItem('With finance', m.purchase_requests, "status = 'finance_approval'", { role: 'balloon', color: '#f6c23e' }),
        metricItem('On hold', m.purchase_requests, "status = 'on_hold'", { role: 'balloon', color: '#e74a3b' }),
        metricItem('Approved value', m.purchase_requests, "status = 'approved'", { role: 'hero', color: '#1cc88a', field: 'estimated_value', operation: 'sum', fmt: '0,0', suffix: ' SAR' }),
      ], { itemsPerRow: '1' }),
      organizer('Submitted', [0, 16, 12, 20], m.purchase_requests, {
        labelField: 'title', descriptionField: 'subsidiary_code', groupField: 'status', group: 'submitted',
      }),
      organizer('Procurement', [12, 16, 12, 20], m.purchase_requests, {
        labelField: 'title', descriptionField: 'subsidiary_code', groupField: 'status', group: 'procurement_review',
      }),
      organizer('Finance', [24, 16, 12, 20], m.purchase_requests, {
        labelField: 'title', descriptionField: 'subsidiary_code', groupField: 'status', group: 'finance_approval',
      }),
      organizer('On hold', [36, 16, 12, 20], m.purchase_requests, {
        labelField: 'title', descriptionField: 'estimated_value', groupField: 'status', group: 'on_hold',
      }),
      recordList('All purchase requests', [0, 36, 48, 24], m.purchase_requests, [
        'pr_number', 'title', 'subsidiary_code', 'item', 'estimated_value', 'status', 'budget_ok', 'due_at',
      ]),
    ]),
  }
}

function subsidiariesPage (m) {
  return {
    title: 'Subsidiaries',
    handle: 'subsidiaries',
    visible: true,
    weight: 30,
    config: pageIcon('fas building'),
    blocks: withBlockIDs([
      recordList('Al-Faris subsidiaries', [0, 0, 48, 36], m.subsidiaries, [
        'name', 'code', 'industry', 'erp_company_code', 'city',
      ]),
    ]),
  }
}

function erpMirrorPage (m) {
  return {
    title: 'ERP mirror',
    handle: 'erp-mirror',
    visible: true,
    weight: 40,
    config: pageIcon('fas globe'),
    blocks: withBlockIDs([
      recordList('ERP vendor master (mocked SAP)', [0, 0, 48, 22], m.erp_vendors, [
        'vendor_code', 'legal_name', 'cr_number', 'vat_number', 'erp_status', 'last_sync', 'source_system',
      ]),
      recordList('ERP budgets (mocked)', [0, 22, 48, 22], m.erp_budgets, [
        'subsidiary_code', 'category', 'cost_center', 'annual_budget', 'committed', 'remaining', 'fiscal_year',
      ]),
    ]),
  }
}

function vendorCard (m) {
  return {
    title: 'Vendor request',
    handle: 'vendor',
    moduleID: String(m.vendors),
    visible: false,
    weight: 11,
    config: pageIcon('fas handshake'),
    blocks: withBlockIDs([
      recordBlock('Vendor onboarding', [0, 0, 32, 30], [
        'legal_name', 'trade_name', 'status', 'subsidiary', 'category',
        'cr_number', 'vat_number', 'city', 'country',
        'contact_name', 'contact_email', 'contact_phone', 'bank_iban',
        'pack_complete', 'documents', 'erp_vendor',
        'submitted_at', 'approved_at', 'due_at', 'cycle_days', 'stalled', 'notes',
      ], {
        fieldRoles: {
          legal_name: 'title',
          trade_name: 'subtitle',
          status: 'badge',
          category: 'badge',
          subsidiary: 'meta',
          cr_number: 'meta',
          vat_number: 'meta',
        },
      }),
      ruleChain('Submit', [32, 0, 16, 8], { chainID: 'faris-submit-vendor', label: 'Submit to procurement', icon: 'play' }),
      ruleChain('Incomplete', [32, 8, 16, 8], { chainID: 'faris-vendor-incomplete', label: 'Return incomplete', variant: 'warning', icon: 'exclamation-triangle' }),
      ruleChain('Procurement', [32, 16, 16, 8], { chainID: 'faris-vendor-procurement', label: 'Procurement review', variant: 'info', icon: 'search' }),
      ruleChain('Compliance', [32, 24, 16, 8], { chainID: 'faris-vendor-compliance', label: 'Compliance check', variant: 'info', icon: 'cog' }),
      ruleChain('Finance', [32, 32, 16, 8], { chainID: 'faris-vendor-finance', label: 'Finance verify', variant: 'secondary', icon: 'percent' }),
      ruleChain('Approve', [32, 40, 8, 8], { chainID: 'faris-vendor-approve', label: 'Approve', variant: 'success', icon: 'check' }),
      ruleChain('Reject', [40, 40, 8, 8], { chainID: 'faris-vendor-reject', label: 'Reject', variant: 'danger', icon: 'times' }),
      recordList('Approval history', [0, 30, 32, 18], m.approval_log, [
        'step', 'decision', 'actor', 'decided_at', 'comment',
      ], { prefilter: `vendor = ${recID}`, refField: 'vendor', hideAddButton: true }),
    ]),
  }
}

function prCard (m) {
  return {
    title: 'Purchase request',
    handle: 'purchase-request',
    moduleID: String(m.purchase_requests),
    visible: false,
    weight: 21,
    config: pageIcon('fas file-invoice'),
    blocks: withBlockIDs([
      recordBlock('Purchase request', [0, 0, 32, 28], [
        'pr_number', 'title', 'status', 'subsidiary', 'item', 'quantity', 'uom',
        'estimated_value', 'vendor', 'category', 'budget_line',
        'budget_remaining', 'budget_ok', 'over_budget',
        'submitted_at', 'approved_at', 'due_at', 'cycle_days', 'stalled', 'justification',
      ], {
        fieldRoles: {
          title: 'title',
          pr_number: 'subtitle',
          status: 'badge',
          estimated_value: 'meta',
          budget_ok: 'badge',
          subsidiary: 'meta',
        },
      }),
      ruleChain('Submit', [32, 0, 16, 8], { chainID: 'faris-pr-submit', label: 'Submit', icon: 'play' }),
      ruleChain('Procurement', [32, 8, 16, 8], { chainID: 'faris-pr-procurement', label: 'Send to finance', variant: 'info', icon: 'search' }),
      ruleChain('Budget', [32, 16, 16, 8], { chainID: 'faris-pr-budget-check', label: 'Check ERP budget', variant: 'warning', icon: 'percent' }),
      ruleChain('Approve', [32, 24, 8, 8], { chainID: 'faris-pr-approve', label: 'Approve', variant: 'success', icon: 'check' }),
      ruleChain('Hold', [40, 24, 8, 8], { chainID: 'faris-pr-hold', label: 'Hold', variant: 'warning', icon: 'stop' }),
      ruleChain('Reject', [32, 32, 16, 8], { chainID: 'faris-pr-reject', label: 'Reject', variant: 'danger', icon: 'times' }),
      recordList('Approval history', [0, 28, 32, 18], m.approval_log, [
        'step', 'decision', 'actor', 'decided_at', 'days_in_step', 'comment',
      ], { prefilter: `purchase_request = ${recID}`, refField: 'purchase_request', hideAddButton: true }),
    ]),
  }
}

function subsidiaryCard (m) {
  return {
    title: 'Subsidiary',
    handle: 'subsidiary',
    moduleID: String(m.subsidiaries),
    visible: false,
    weight: 31,
    blocks: withBlockIDs([
      recordBlock('Subsidiary', [0, 0, 24, 18], [
        'name', 'code', 'industry', 'erp_company_code', 'city', 'notes',
      ], { fieldRoles: { name: 'title', code: 'subtitle', industry: 'badge' } }),
      recordList('Vendors', [24, 0, 24, 18], m.vendors, ['legal_name', 'status', 'category'], {
        prefilter: `subsidiary = ${recID}`, refField: 'subsidiary',
      }),
      recordList('Purchase requests', [0, 18, 48, 18], m.purchase_requests, [
        'pr_number', 'title', 'estimated_value', 'status',
      ], { prefilter: `subsidiary = ${recID}`, refField: 'subsidiary' }),
    ]),
  }
}

function erpVendorCard (m) {
  return {
    title: 'ERP vendor',
    handle: 'erp-vendor',
    moduleID: String(m.erp_vendors),
    visible: false,
    weight: 41,
    blocks: withBlockIDs([
      recordBlock('ERP vendor master', [0, 0, 48, 20], [
        'vendor_code', 'legal_name', 'cr_number', 'vat_number', 'city', 'country',
        'erp_status', 'last_sync', 'source_system',
      ], { fieldRoles: { legal_name: 'title', vendor_code: 'subtitle', erp_status: 'badge' } }),
    ]),
  }
}

function erpBudgetCard (m) {
  return {
    title: 'ERP budget',
    handle: 'erp-budget',
    moduleID: String(m.erp_budgets),
    visible: false,
    weight: 42,
    blocks: withBlockIDs([
      recordBlock('ERP budget line', [0, 0, 48, 18], [
        'subsidiary', 'subsidiary_code', 'category', 'cost_center',
        'annual_budget', 'committed', 'remaining', 'fiscal_year', 'currency',
      ], { fieldRoles: { cost_center: 'title', category: 'badge', remaining: 'meta' } }),
    ]),
  }
}

function approvalLogCard (m) {
  return {
    title: 'Approval log',
    handle: 'approval-log-item',
    moduleID: String(m.approval_log),
    visible: false,
    weight: 50,
    blocks: withBlockIDs([
      recordBlock('Approval step', [0, 0, 48, 18], [
        'subject', 'object_type', 'vendor', 'purchase_request', 'step', 'decision',
        'actor', 'decided_at', 'days_in_step', 'comment',
      ], { fieldRoles: { subject: 'title', decision: 'badge', step: 'badge' } }),
    ]),
  }
}
