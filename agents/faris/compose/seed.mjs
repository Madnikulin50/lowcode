import { createRecord, setOf } from './helpers.mjs'

const SUBS = [
  { name: 'Al-Faris Construction', code: 'AFC', industry: 'construction', erp_company_code: 'SA-1000', city: 'Riyadh' },
  { name: 'Al-Faris Energy', code: 'AFE', industry: 'energy', erp_company_code: 'SA-1100', city: 'Dhahran' },
  { name: 'Al-Faris Retail', code: 'AFR', industry: 'retail', erp_company_code: 'SA-1200', city: 'Jeddah' },
  { name: 'Al-Faris Logistics', code: 'AFL', industry: 'logistics', erp_company_code: 'SA-1300', city: 'Dammam' },
  { name: 'Al-Faris Healthcare', code: 'AFH', industry: 'healthcare', erp_company_code: 'SA-1400', city: 'Riyadh' },
  { name: 'Al-Faris Food', code: 'AFF', industry: 'food', erp_company_code: 'SA-1500', city: 'Jeddah' },
  { name: 'Al-Faris Real Estate', code: 'AFRE', industry: 'real_estate', erp_company_code: 'SA-1600', city: 'Riyadh' },
  { name: 'Al-Faris ICT', code: 'AFI', industry: 'ict', erp_company_code: 'SA-1700', city: 'Riyadh' },
]

const CATS = [
  { name: 'IT', code: 'it', description: 'Hardware, software, cloud and telecom' },
  { name: 'Facilities', code: 'facilities', description: 'Buildings, utilities, maintenance' },
  { name: 'Professional services', code: 'professional_services', description: 'Legal, audit, consulting' },
  { name: 'Materials', code: 'materials', description: 'Construction and production materials' },
  { name: 'Fleet', code: 'fleet', description: 'Vehicles and spare parts' },
  { name: 'Marketing', code: 'marketing', description: 'Campaigns, agencies, media' },
]

function recID (row) {
  return String(row.recordID || row.ID)
}

async function put (api, nsID, moduleID, values) {
  const compact = Object.fromEntries(Object.entries(values).filter(([, v]) => v !== '' && v != null && v !== undefined))
  return createRecord(api, nsID, moduleID, compact)
}

export async function seedIfEmpty (api, nsID, modules, { force = false } = {}) {
  const existing = setOf(await api('GET', `/namespace/${nsID}/module/${modules.subsidiaries}/record/?limit=5`))
  if (existing.length && !force) {
    console.log('subsidiaries already have records, skip seed')
    return
  }

  const subIDs = {}
  for (const row of SUBS) {
    const rec = await put(api, nsID, modules.subsidiaries, {
      ...row,
      notes: 'Al-Faris Holding Group subsidiary — demo record.',
    })
    subIDs[row.code] = recID(rec)
    console.log('  subsidiary', row.code, subIDs[row.code])
  }

  for (const row of CATS) {
    await put(api, nsID, modules.spend_categories, row)
  }

  const erpVendors = [
    { vendor_code: 'ERP-41001', legal_name: 'Najd Industrial Supplies LLC', cr_number: '1010123456', vat_number: '300123456700003', city: 'Riyadh', country: 'Saudi Arabia', erp_status: 'active' },
    { vendor_code: 'ERP-41002', legal_name: 'Hijaz Facilities Co.', cr_number: '4030987654', vat_number: '300987654300003', city: 'Jeddah', country: 'Saudi Arabia', erp_status: 'active' },
    { vendor_code: 'ERP-41003', legal_name: 'Eastern Gulf Logistics LLC', cr_number: '2050111222', vat_number: '300111222300003', city: 'Dammam', country: 'Saudi Arabia', erp_status: 'active' },
    { vendor_code: 'ERP-41004', legal_name: 'Riyadh Cloud Systems', cr_number: '1010555666', vat_number: '300555666700003', city: 'Riyadh', country: 'Saudi Arabia', erp_status: 'active' },
    { vendor_code: 'ERP-41005', legal_name: 'Oasis Clinical Equipment', cr_number: '1010777888', vat_number: '300777888900003', city: 'Riyadh', country: 'Saudi Arabia', erp_status: 'active' },
    { vendor_code: 'ERP-41006', legal_name: 'Red Sea Catering Group', cr_number: '4030444555', vat_number: '300444555600003', city: 'Jeddah', country: 'Saudi Arabia', erp_status: 'active' },
    { vendor_code: 'ERP-41007', legal_name: 'Desert Fleet Services', cr_number: '1010333444', vat_number: '300333444500003', city: 'Riyadh', country: 'Saudi Arabia', erp_status: 'blocked' },
    { vendor_code: 'ERP-41008', legal_name: 'Qiddiya Fit-Out Partners', cr_number: '1010999000', vat_number: '300999000100003', city: 'Riyadh', country: 'Saudi Arabia', erp_status: 'pending_sync' },
    { vendor_code: 'ERP-41009', legal_name: 'Bahrain Tech Distributors WLL', cr_number: 'BH-88221', vat_number: 'BH-VAT-88221', city: 'Manama', country: 'Bahrain', erp_status: 'active' },
    { vendor_code: 'ERP-41010', legal_name: 'Global Audit Partners ME', cr_number: '1010666777', vat_number: '300666777800003', city: 'Riyadh', country: 'Saudi Arabia', erp_status: 'active' },
  ]
  const erpIDs = {}
  for (const row of erpVendors) {
    const rec = await put(api, nsID, modules.erp_vendors, {
      ...row,
      last_sync: '2026-08-20',
      source_system: 'SAP S/4HANA (mock)',
    })
    erpIDs[row.vendor_code] = recID(rec)
  }

  const budgetSpecs = [
    ['AFC', 'materials', 'CC-AFC-MAT', 18000000, 12400000],
    ['AFC', 'professional_services', 'CC-AFC-PS', 2500000, 2100000],
    ['AFE', 'it', 'CC-AFE-IT', 4200000, 3100000],
    ['AFE', 'facilities', 'CC-AFE-FAC', 6000000, 5400000],
    ['AFR', 'marketing', 'CC-AFR-MKT', 3500000, 1800000],
    ['AFR', 'it', 'CC-AFR-IT', 1800000, 900000],
    ['AFL', 'fleet', 'CC-AFL-FLT', 8000000, 6100000],
    ['AFL', 'it', 'CC-AFL-IT', 900000, 420000],
    ['AFH', 'it', 'CC-AFH-IT', 2200000, 700000],
    ['AFH', 'professional_services', 'CC-AFH-PS', 1600000, 400000],
    ['AFF', 'facilities', 'CC-AFF-FAC', 2700000, 1900000],
    ['AFF', 'materials', 'CC-AFF-MAT', 5100000, 4300000],
    ['AFRE', 'facilities', 'CC-AFRE-FAC', 9400000, 7200000],
    ['AFRE', 'professional_services', 'CC-AFRE-PS', 1200000, 350000],
    ['AFI', 'it', 'CC-AFI-IT', 7600000, 4100000],
    ['AFI', 'professional_services', 'CC-AFI-PS', 2100000, 800000],
  ]
  const budgetIDs = {}
  for (const [code, category, cost_center, annual, committed] of budgetSpecs) {
    const remaining = annual - committed
    const rec = await put(api, nsID, modules.erp_budgets, {
      subsidiary: subIDs[code],
      subsidiary_code: code,
      category,
      cost_center,
      annual_budget: String(annual),
      committed: String(committed),
      remaining: String(remaining),
      fiscal_year: '2026',
      currency: 'SAR',
    })
    budgetIDs[cost_center] = { id: recID(rec), remaining, code, category }
  }

  const vendors = [
    { legal_name: 'Najd Industrial Supplies LLC', trade_name: 'Najd Supplies', subsidiary: 'AFC', category: 'materials', cr_number: '1010123456', vat_number: '300123456700003', city: 'Riyadh', status: 'approved', erp: 'ERP-41001', submitted_at: '2026-06-02', approved_at: '2026-06-18', cycle_days: 16, pack: true },
    { legal_name: 'Hijaz Facilities Co.', trade_name: 'Hijaz FM', subsidiary: 'AFR', category: 'facilities', cr_number: '4030987654', vat_number: '300987654300003', city: 'Jeddah', status: 'approved', erp: 'ERP-41002', submitted_at: '2026-05-11', approved_at: '2026-05-28', cycle_days: 17, pack: true },
    { legal_name: 'Riyadh Cloud Systems', trade_name: 'RCS', subsidiary: 'AFI', category: 'it', cr_number: '1010555666', vat_number: '300555666700003', city: 'Riyadh', status: 'approved', erp: 'ERP-41004', submitted_at: '2026-04-03', approved_at: '2026-04-14', cycle_days: 11, pack: true },
    { legal_name: 'Oasis Clinical Equipment', trade_name: 'Oasis Med', subsidiary: 'AFH', category: 'materials', cr_number: '1010777888', vat_number: '300777888900003', city: 'Riyadh', status: 'approved', erp: 'ERP-41005', submitted_at: '2026-03-20', approved_at: '2026-04-09', cycle_days: 20, pack: true },
    { legal_name: 'Eastern Gulf Logistics LLC', trade_name: 'EGL', subsidiary: 'AFL', category: 'fleet', cr_number: '2050111222', vat_number: '300111222300003', city: 'Dammam', status: 'approved', erp: 'ERP-41003', submitted_at: '2026-07-01', approved_at: '2026-07-12', cycle_days: 11, pack: true },
    { legal_name: 'Red Sea Catering Group', trade_name: 'Red Sea Catering', subsidiary: 'AFF', category: 'facilities', cr_number: '4030444555', vat_number: '300444555600003', city: 'Jeddah', status: 'finance_verify', erp: 'ERP-41006', submitted_at: '2026-08-10', due_at: '2026-08-24', cycle_days: 15, pack: true },
    { legal_name: 'Qiddiya Fit-Out Partners', trade_name: 'Qiddiya Fit-Out', subsidiary: 'AFRE', category: 'facilities', cr_number: '1010999000', vat_number: '300999000100003', city: 'Riyadh', status: 'compliance_review', erp: 'ERP-41008', submitted_at: '2026-08-12', due_at: '2026-08-26', pack: true },
    { legal_name: 'Global Audit Partners ME', trade_name: 'GAP ME', subsidiary: 'AFE', category: 'professional_services', cr_number: '1010666777', vat_number: '300666777800003', city: 'Riyadh', status: 'procurement_review', erp: 'ERP-41010', submitted_at: '2026-08-18', due_at: '2026-09-01', pack: true },
    { legal_name: 'Neom Steel Trading', trade_name: 'Neom Steel', subsidiary: 'AFC', category: 'materials', cr_number: '1010222333', vat_number: '300222333400003', city: 'Riyadh', status: 'submitted', submitted_at: '2026-08-21', due_at: '2026-09-04', pack: true },
    { legal_name: 'Asir Digital Media', trade_name: 'Asir Media', subsidiary: 'AFR', category: 'marketing', city: 'Abha', status: 'incomplete', submitted_at: '2026-08-19', due_at: '2026-08-22', stalled: true, pack: false, notes: 'CR scan missing; VAT certificate expired.' },
    { legal_name: 'Tabuk HVAC Specialists', trade_name: 'Tabuk HVAC', subsidiary: 'AFE', category: 'facilities', cr_number: '3550111222', city: 'Tabuk', status: 'draft', pack: false },
    { legal_name: 'Desert Fleet Services', trade_name: 'Desert Fleet', subsidiary: 'AFL', category: 'fleet', cr_number: '1010333444', vat_number: '300333444500003', city: 'Riyadh', status: 'rejected', erp: 'ERP-41007', submitted_at: '2026-07-20', approved_at: '2026-07-25', cycle_days: 5, pack: true, notes: 'ERP master is blocked — sanctions list hit (demo).' },
    { legal_name: 'Bahrain Tech Distributors WLL', trade_name: 'BTD', subsidiary: 'AFI', category: 'it', cr_number: 'BH-88221', vat_number: 'BH-VAT-88221', city: 'Manama', country: 'Bahrain', status: 'procurement_review', erp: 'ERP-41009', submitted_at: '2026-08-08', due_at: '2026-08-20', stalled: true, pack: true, notes: 'GCC vendor — extra compliance questionnaire outstanding.' },
    { legal_name: 'Hail Smart Meters Co.', trade_name: 'Hail Meters', subsidiary: 'AFE', category: 'it', cr_number: '3350444555', vat_number: '300444555600013', city: 'Hail', status: 'submitted', submitted_at: '2026-08-22', due_at: '2026-09-05', pack: true },
  ]

  const vendorIDs = {}
  for (const v of vendors) {
    const rec = await put(api, nsID, modules.vendors, {
      legal_name: v.legal_name,
      trade_name: v.trade_name,
      subsidiary: subIDs[v.subsidiary],
      subsidiary_code: v.subsidiary,
      category: v.category,
      cr_number: v.cr_number || '',
      vat_number: v.vat_number || '',
      city: v.city,
      country: v.country || 'Saudi Arabia',
      contact_name: 'Procurement desk',
      contact_email: `vendors@${v.subsidiary.toLowerCase()}.alfaris.sa`,
      contact_phone: '+966-11-400-0000',
      bank_iban: 'SA44 2000 0000 1234 5678 9012',
      pack_complete: v.pack ? '1' : '0',
      status: v.status,
      erp_vendor: v.erp ? erpIDs[v.erp] : '',
      submitted_at: v.submitted_at || '',
      approved_at: v.approved_at || '',
      due_at: v.due_at || '',
      cycle_days: v.cycle_days != null ? String(v.cycle_days) : '',
      stalled: v.stalled ? '1' : '0',
      notes: v.notes || '',
    })
    vendorIDs[v.legal_name] = recID(rec)
  }

  const prs = [
    { n: 'PR-2026-0142', title: 'Rebar for King Abdullah project', sub: 'AFC', item: 'B500B rebar 16mm', qty: 240, uom: 'ton', value: 1860000, vendor: 'Najd Industrial Supplies LLC', cat: 'materials', cc: 'CC-AFC-MAT', status: 'approved', submitted_at: '2026-07-02', approved_at: '2026-07-11', cycle: 9, ok: true },
    { n: 'PR-2026-0188', title: 'Site diesel generators (rental)', sub: 'AFC', item: '500 kVA generator rental', qty: 4, uom: 'month', value: 420000, vendor: 'Najd Industrial Supplies LLC', cat: 'facilities', cc: 'CC-AFC-MAT', status: 'finance_approval', submitted_at: '2026-08-18', due_at: '2026-08-28', ok: true },
    { n: 'PR-2026-0201', title: 'SAP licenses top-up', sub: 'AFI', item: 'SAP named users', qty: 40, uom: 'license', value: 960000, vendor: 'Riyadh Cloud Systems', cat: 'it', cc: 'CC-AFI-IT', status: 'approved', submitted_at: '2026-06-01', approved_at: '2026-06-06', cycle: 5, ok: true },
    { n: 'PR-2026-0214', title: 'SOC monitoring 12 months', sub: 'AFI', item: 'Managed SOC', qty: 12, uom: 'month', value: 1440000, vendor: 'Riyadh Cloud Systems', cat: 'it', cc: 'CC-AFI-IT', status: 'on_hold', submitted_at: '2026-08-04', due_at: '2026-08-15', stalled: true, ok: false, over: true },
    { n: 'PR-2026-0220', title: 'Mall campaign Ramadan 2027', sub: 'AFR', item: 'OOH + digital campaign', qty: 1, uom: 'lot', value: 2100000, vendor: 'Hijaz Facilities Co.', cat: 'marketing', cc: 'CC-AFR-MKT', status: 'on_hold', submitted_at: '2026-08-11', due_at: '2026-08-20', stalled: true, ok: false, over: true },
    { n: 'PR-2026-0228', title: 'Store HVAC overhaul — Tahlia', sub: 'AFR', item: 'VRV replacement', qty: 6, uom: 'unit', value: 380000, vendor: 'Hijaz Facilities Co.', cat: 'facilities', cc: 'CC-AFR-IT', status: 'procurement_review', submitted_at: '2026-08-20', due_at: '2026-09-03', ok: true },
    { n: 'PR-2026-0233', title: 'Reefer trailers', sub: 'AFL', item: 'Reefer trailer 40ft', qty: 8, uom: 'unit', value: 2720000, vendor: 'Eastern Gulf Logistics LLC', cat: 'fleet', cc: 'CC-AFL-FLT', status: 'approved', submitted_at: '2026-07-15', approved_at: '2026-07-22', cycle: 7, ok: true },
    { n: 'PR-2026-0241', title: 'TMS cloud subscription', sub: 'AFL', item: 'TMS SaaS', qty: 12, uom: 'month', value: 540000, vendor: 'Riyadh Cloud Systems', cat: 'it', cc: 'CC-AFL-IT', status: 'finance_approval', submitted_at: '2026-08-16', due_at: '2026-08-30', ok: false, over: true },
    { n: 'PR-2026-0248', title: 'MRI coil replacement', sub: 'AFH', item: 'MRI head coil', qty: 2, uom: 'unit', value: 890000, vendor: 'Oasis Clinical Equipment', cat: 'materials', cc: 'CC-AFH-IT', status: 'submitted', submitted_at: '2026-08-23', due_at: '2026-09-06', ok: true },
    { n: 'PR-2026-0252', title: 'JCI mock survey', sub: 'AFH', item: 'Accreditation consultancy', qty: 1, uom: 'lot', value: 620000, vendor: 'Global Audit Partners ME', cat: 'professional_services', cc: 'CC-AFH-PS', status: 'finance_approval', submitted_at: '2026-08-14', due_at: '2026-08-21', stalled: true, ok: false, over: true },
    { n: 'PR-2026-0259', title: 'Cold-room upgrade — factory 2', sub: 'AFF', item: 'Cold room panels', qty: 1, uom: 'lot', value: 410000, vendor: 'Red Sea Catering Group', cat: 'facilities', cc: 'CC-AFF-FAC', status: 'approved', submitted_at: '2026-06-20', approved_at: '2026-06-29', cycle: 9, ok: true },
    { n: 'PR-2026-0266', title: 'Palm oil bulk Q3', sub: 'AFF', item: 'Refined palm oil', qty: 200, uom: 'ton', value: 980000, cat: 'materials', cc: 'CC-AFF-MAT', status: 'draft', ok: true },
    { n: 'PR-2026-0271', title: 'Lobby fit-out — Olaya tower', sub: 'AFRE', item: 'Joinery and stone', qty: 1, uom: 'lot', value: 3100000, vendor: 'Qiddiya Fit-Out Partners', cat: 'facilities', cc: 'CC-AFRE-FAC', status: 'procurement_review', submitted_at: '2026-08-17', due_at: '2026-09-01', ok: true },
    { n: 'PR-2026-0277', title: 'Legal DD — land plot 14', sub: 'AFRE', item: 'Legal due diligence', qty: 1, uom: 'lot', value: 275000, vendor: 'Global Audit Partners ME', cat: 'professional_services', cc: 'CC-AFRE-PS', status: 'submitted', submitted_at: '2026-08-24', due_at: '2026-09-07', ok: true },
    { n: 'PR-2026-0284', title: 'Plant DCS upgrade', sub: 'AFE', item: 'DCS controllers', qty: 12, uom: 'unit', value: 1950000, vendor: 'Riyadh Cloud Systems', cat: 'it', cc: 'CC-AFE-IT', status: 'approved', submitted_at: '2026-05-02', approved_at: '2026-05-19', cycle: 17, ok: true },
    { n: 'PR-2026-0290', title: 'Camp facilities — Shaybah', sub: 'AFE', item: 'Modular camp units', qty: 20, uom: 'unit', value: 2400000, cat: 'facilities', cc: 'CC-AFE-FAC', status: 'rejected', submitted_at: '2026-07-08', approved_at: '2026-07-10', cycle: 2, ok: false, notes: 'Scope belongs to a JV partner.' },
    { n: 'PR-2026-0296', title: 'Laptop refresh — HQ', sub: 'AFI', item: 'Business laptops', qty: 120, uom: 'unit', value: 540000, vendor: 'Bahrain Tech Distributors WLL', cat: 'it', cc: 'CC-AFI-IT', status: 'submitted', submitted_at: '2026-08-22', due_at: '2026-09-05', ok: true },
    { n: 'PR-2026-0302', title: 'Warehouse WMS scanners', sub: 'AFL', item: 'Handheld scanners', qty: 80, uom: 'unit', value: 192000, vendor: 'Riyadh Cloud Systems', cat: 'it', cc: 'CC-AFL-IT', status: 'approved', submitted_at: '2026-08-01', approved_at: '2026-08-05', cycle: 4, ok: true },
  ]

  const prIDs = {}
  for (const p of prs) {
    const b = budgetIDs[p.cc]
    const rec = await put(api, nsID, modules.purchase_requests, {
      title: p.title,
      pr_number: p.n,
      subsidiary: subIDs[p.sub],
      subsidiary_code: p.sub,
      item: p.item,
      quantity: String(p.qty),
      uom: p.uom,
      estimated_value: String(p.value),
      vendor: p.vendor ? vendorIDs[p.vendor] : '',
      category: p.cat,
      budget_line: b ? b.id : '',
      budget_remaining: b ? String(b.remaining) : '',
      budget_ok: p.ok ? '1' : '0',
      over_budget: p.over ? '1' : '0',
      status: p.status,
      submitted_at: p.submitted_at || '',
      approved_at: p.approved_at || '',
      due_at: p.due_at || '',
      cycle_days: p.cycle != null ? String(p.cycle) : '',
      stalled: p.stalled ? '1' : '0',
      justification: p.notes || 'Group framework request — demo data.',
    })
    prIDs[p.n] = recID(rec)
  }

  const logs = [
    ['Approved vendor: Najd Industrial Supplies LLC', 'vendor', 'Najd Industrial Supplies LLC', null, 'finance', 'approved', 'Group finance', '2026-06-18', 4, 'CR and VAT matched SAP vendor 41001.'],
    ['Approved vendor: Riyadh Cloud Systems', 'vendor', 'Riyadh Cloud Systems', null, 'finance', 'approved', 'Group finance', '2026-04-14', 2, 'Preferred ICT supplier.'],
    ['Incomplete pack: Asir Digital Media', 'vendor', 'Asir Digital Media', null, 'submit', 'returned', 'Central procurement', '2026-08-19', 0, 'Missing CR scan.'],
    ['Rejected: Desert Fleet Services', 'vendor', 'Desert Fleet Services', null, 'compliance', 'rejected', 'Compliance', '2026-07-25', 3, 'ERP master blocked.'],
    ['PR-2026-0142 approved', 'purchase_request', null, 'PR-2026-0142', 'finance', 'approved', 'Group finance', '2026-07-11', 3, 'Within CC-AFC-MAT remaining.'],
    ['PR-2026-0201 approved', 'purchase_request', null, 'PR-2026-0201', 'finance', 'approved', 'Group finance', '2026-06-06', 2, 'Covered by ICT budget.'],
    ['PR-2026-0214 held', 'purchase_request', null, 'PR-2026-0214', 'finance', 'held', 'Group finance', '2026-08-15', 11, 'Exceeds remaining IT budget for AFI.'],
    ['PR-2026-0220 held', 'purchase_request', null, 'PR-2026-0220', 'finance', 'held', 'Group finance', '2026-08-20', 9, 'Campaign estimate above marketing remaining.'],
    ['PR-2026-0233 approved', 'purchase_request', null, 'PR-2026-0233', 'finance', 'approved', 'Group finance', '2026-07-22', 4, 'Fleet capex approved.'],
    ['PR-2026-0252 stalled at finance', 'purchase_request', null, 'PR-2026-0252', 'finance', 'held', 'Group finance', '2026-08-21', 7, 'Waiting for hospital board minute.'],
    ['PR-2026-0290 rejected', 'purchase_request', null, 'PR-2026-0290', 'procurement', 'rejected', 'Central procurement', '2026-07-10', 2, 'Out of scope for AFE.'],
    ['PR-2026-0284 approved', 'purchase_request', null, 'PR-2026-0284', 'finance', 'approved', 'Group finance', '2026-05-19', 8, 'DCS upgrade — energy subsidiary.'],
    ['Onboarding RCS to procurement', 'vendor', 'Riyadh Cloud Systems', null, 'procurement', 'approved', 'Central procurement', '2026-04-08', 5, 'Three references checked.'],
    ['Qiddiya Fit-Out → compliance', 'vendor', 'Qiddiya Fit-Out Partners', null, 'compliance', 'approved', 'Compliance', '2026-08-14', 2, 'Pending ZATCA e-invoicing proof.'],
    ['EGL approved', 'vendor', 'Eastern Gulf Logistics LLC', null, 'finance', 'approved', 'Group finance', '2026-07-12', 3, 'IBAN verified.'],
  ]
  for (const [subject, object_type, vName, prNo, step, decision, actor, decided_at, days, comment] of logs) {
    await put(api, nsID, modules.approval_log, {
      subject,
      object_type,
      vendor: vName ? vendorIDs[vName] : '',
      purchase_request: prNo ? prIDs[prNo] : '',
      step,
      decision,
      actor,
      decided_at,
      days_in_step: String(days),
      comment,
    })
  }

  console.log('seeded Al-Faris demo records')
}
