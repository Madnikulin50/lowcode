import {
  field, recordRel, fileField, dateField, moneyField, numberField, boolSwitch, selectOptions,
} from './helpers.mjs'

export const INDUSTRY = [
  ['construction', 'Construction'],
  ['energy', 'Energy'],
  ['retail', 'Retail'],
  ['logistics', 'Logistics'],
  ['healthcare', 'Healthcare'],
  ['food', 'Food'],
  ['real_estate', 'Real estate'],
  ['ict', 'ICT'],
]

export const CATEGORY = [
  ['it', 'IT'],
  ['facilities', 'Facilities'],
  ['professional_services', 'Professional services'],
  ['materials', 'Materials'],
  ['fleet', 'Fleet'],
  ['marketing', 'Marketing'],
]

export const ERP_STATUS = [
  ['active', 'Active', { backgroundColor: 'success', textColor: 'white' }],
  ['blocked', 'Blocked', { backgroundColor: 'danger', textColor: 'white' }],
  ['pending_sync', 'Pending sync', { backgroundColor: 'warning', textColor: 'dark' }],
]

export const VENDOR_STATUS = [
  ['draft', 'Draft', { backgroundColor: 'light', textColor: 'dark' }],
  ['incomplete', 'Incomplete', { backgroundColor: 'warning', textColor: 'dark' }],
  ['submitted', 'Submitted', { backgroundColor: 'info', textColor: 'white' }],
  ['procurement_review', 'Procurement review', { backgroundColor: 'primary', textColor: 'white' }],
  ['compliance_review', 'Compliance review', { backgroundColor: 'info', textColor: 'white' }],
  ['finance_verify', 'Finance verify', { backgroundColor: 'secondary', textColor: 'white' }],
  ['approved', 'Approved', { backgroundColor: 'success', textColor: 'white' }],
  ['rejected', 'Rejected', { backgroundColor: 'danger', textColor: 'white' }],
]

export const PR_STATUS = [
  ['draft', 'Draft', { backgroundColor: 'light', textColor: 'dark' }],
  ['submitted', 'Submitted', { backgroundColor: 'info', textColor: 'white' }],
  ['procurement_review', 'Procurement review', { backgroundColor: 'primary', textColor: 'white' }],
  ['finance_approval', 'Finance approval', { backgroundColor: 'warning', textColor: 'dark' }],
  ['approved', 'Approved', { backgroundColor: 'success', textColor: 'white' }],
  ['on_hold', 'On hold', { backgroundColor: 'warning', textColor: 'dark' }],
  ['rejected', 'Rejected', { backgroundColor: 'danger', textColor: 'white' }],
]

export const STEP = [
  ['submit', 'Submit'],
  ['procurement', 'Procurement'],
  ['compliance', 'Compliance'],
  ['finance', 'Finance'],
]

export const DECISION = [
  ['approved', 'Approved', { backgroundColor: 'success', textColor: 'white' }],
  ['rejected', 'Rejected', { backgroundColor: 'danger', textColor: 'white' }],
  ['returned', 'Returned', { backgroundColor: 'warning', textColor: 'dark' }],
  ['held', 'Held', { backgroundColor: 'secondary', textColor: 'white' }],
]

export function subsidiaryFields () {
  return [
    field('name', 'Name', 'String', { required: true }),
    field('code', 'Code', 'String', { required: true }),
    field('industry', 'Industry', 'Select', { options: selectOptions(INDUSTRY) }),
    field('erp_company_code', 'ERP company code', 'String'),
    field('city', 'City', 'String'),
    field('notes', 'Notes', 'String'),
  ]
}

export function categoryFields () {
  return [
    field('name', 'Name', 'String', { required: true }),
    field('code', 'Code', 'String', { required: true }),
    field('description', 'Description', 'String'),
  ]
}

export function erpVendorFields () {
  return [
    field('vendor_code', 'Vendor code', 'String', { required: true }),
    field('legal_name', 'Legal name', 'String', { required: true }),
    field('cr_number', 'CR number', 'String'),
    field('vat_number', 'VAT number', 'String'),
    field('city', 'City', 'String'),
    field('country', 'Country', 'String'),
    field('erp_status', 'ERP status', 'Select', { options: selectOptions(ERP_STATUS) }),
    dateField('last_sync', 'Last sync'),
    field('source_system', 'Source system', 'String'),
  ]
}

export function erpBudgetFields (subsidiaries) {
  return [
    recordRel('subsidiary', 'Subsidiary', subsidiaries, 'name', ['name', 'code'], true),
    field('subsidiary_code', 'Subsidiary code', 'String'),
    field('category', 'Category', 'Select', { options: selectOptions(CATEGORY) }),
    field('cost_center', 'Cost center', 'String'),
    moneyField('annual_budget', 'Annual budget'),
    moneyField('committed', 'Committed'),
    moneyField('remaining', 'Remaining'),
    field('fiscal_year', 'Fiscal year', 'String'),
    field('currency', 'Currency', 'String'),
  ]
}

export function vendorFields (subsidiaries, erpVendors) {
  return [
    field('legal_name', 'Legal name', 'String', { required: true }),
    field('trade_name', 'Trade name', 'String'),
    recordRel('subsidiary', 'Requesting subsidiary', subsidiaries, 'name', ['name', 'code'], true),
    field('subsidiary_code', 'Subsidiary code', 'String'),
    field('category', 'Category', 'Select', { options: selectOptions(CATEGORY) }),
    field('cr_number', 'CR number', 'String'),
    field('vat_number', 'VAT number', 'String'),
    field('city', 'City', 'String'),
    field('country', 'Country', 'String'),
    field('contact_name', 'Contact name', 'String'),
    field('contact_email', 'Contact email', 'String'),
    field('contact_phone', 'Contact phone', 'String'),
    field('bank_iban', 'Bank IBAN', 'String'),
    fileField('documents', 'Licenses & certificates', { multi: true, documents: true, images: true }),
    boolSwitch('pack_complete', 'Document pack complete'),
    field('status', 'Status', 'Select', { options: selectOptions(VENDOR_STATUS) }),
    recordRel('erp_vendor', 'ERP vendor master', erpVendors, 'legal_name', ['legal_name', 'vendor_code']),
    dateField('submitted_at', 'Submitted at'),
    dateField('approved_at', 'Approved at'),
    dateField('due_at', 'Due at'),
    numberField('cycle_days', 'Cycle days', { precision: 0, format: '0' }),
    boolSwitch('stalled', 'Stalled'),
    field('notes', 'Notes', 'String'),
  ]
}

export function purchaseRequestFields (subsidiaries, vendors, budgets) {
  return [
    field('title', 'Title', 'String', { required: true }),
    field('pr_number', 'PR number', 'String'),
    recordRel('subsidiary', 'Subsidiary', subsidiaries, 'name', ['name', 'code'], true),
    field('subsidiary_code', 'Subsidiary code', 'String'),
    field('item', 'Item', 'String', { required: true }),
    numberField('quantity', 'Quantity', { precision: 2, format: '0,0.00' }),
    field('uom', 'UoM', 'String'),
    moneyField('estimated_value', 'Estimated value'),
    recordRel('vendor', 'Vendor', vendors, 'legal_name', ['legal_name', 'trade_name']),
    field('category', 'Category', 'Select', { options: selectOptions(CATEGORY) }),
    recordRel('budget_line', 'ERP budget line', budgets, 'cost_center', ['cost_center', 'category']),
    moneyField('budget_remaining', 'Budget remaining (ERP)'),
    boolSwitch('budget_ok', 'Within budget'),
    boolSwitch('over_budget', 'Over budget'),
    field('status', 'Status', 'Select', { options: selectOptions(PR_STATUS) }),
    dateField('submitted_at', 'Submitted at'),
    dateField('approved_at', 'Approved at'),
    dateField('due_at', 'Due at'),
    numberField('cycle_days', 'Cycle days', { precision: 0, format: '0' }),
    boolSwitch('stalled', 'Stalled / overdue'),
    field('justification', 'Justification', 'String'),
  ]
}

export function approvalLogFields (vendors, prs) {
  return [
    field('subject', 'Subject', 'String', { required: true }),
    field('object_type', 'Object type', 'Select', {
      options: selectOptions([['vendor', 'Vendor'], ['purchase_request', 'Purchase request']]),
    }),
    recordRel('vendor', 'Vendor', vendors, 'legal_name', ['legal_name']),
    recordRel('purchase_request', 'Purchase request', prs, 'title', ['title', 'pr_number']),
    field('step', 'Step', 'Select', { options: selectOptions(STEP) }),
    field('decision', 'Decision', 'Select', { options: selectOptions(DECISION) }),
    field('actor', 'Actor', 'String'),
    dateField('decided_at', 'Decided at'),
    numberField('days_in_step', 'Days in step', { precision: 0, format: '0' }),
    field('comment', 'Comment', 'String'),
  ]
}
