/**
 * Страницы пространства «Стройконтроль ПТО».
 *
 * Навигация идёт «от объекта»: объект → карточка-рабочее место с вкладками
 * (обзор документов и версий, комплекты, ПД↔РД, замечания экспертизы, ИД, сметы).
 * Боковое меню — группы (страницы без модуля), внутри группы — списки; карточки записей
 * скрыты и открываются из списков. Служебные списки собраны в группе «Служебное».
 *
 * buildPages(m) → { defs, parentOf }:
 *   defs     — страницы в порядке создания (группы раньше своих детей); `parent` — handle
 *              группы, `tabbedBlockIDs` — блоки, живущие только внутри вкладок;
 *   parentOf — карточка → её список (для иерархии «список → карточка»).
 */
import {
  block, recordList, recordBlock, metricBlock, metricItem, ruleChain, pageIcon, withBlockIDs,
} from './helpers.mjs'

// фильтр «рабочие данные»: демо-записи в списках и метриках скрыты
export const REAL = "(data_kind != 'demo' OR data_kind IS NULL)"
// Описание ролей (создаются roles.mjs); участники назначаются в администрировании платформы → Роли.
const ROLES_HTML = '<p>Доступ к модулям пространства выдаётся ролями платформы (создаются <code>node compose/roles.mjs</code>). ' +
  'Участников в роли назначает администратор: <b>Администрирование платформы → Роли</b>.</p><ul>' +
  '<li><b>Наблюдатель</b> (<code>stroykontrol_view</code>) — читает всё, ничего не меняет;</li>' +
  '<li><b>Проектировщик</b> (<code>stroykontrol_des</code>) — загружает документацию и работает с замечаниями;</li>' +
  '<li><b>Специалист КК</b> (<code>stroykontrol_qc</code>) — как проектировщик, плюс решения по расхождениям, сравнениям и пакетам исправления;</li>' +
  '<li><b>Руководитель проекта</b> (<code>stroykontrol_mgr</code>) — пишет во все модули пространства.</li></ul>'

const AGENT = '/gateway/agents/stroykontrol/'

const iframe = (title, xywh, query) => block('IFrame', title, xywh, {
  src: `${AGENT}?${query}&namespaceID=\${namespaceID}`,
  srcField: '',
  displayAsImage: false,
})

function group (title, handle, weight, icon, text) {
  return {
    title, handle, visible: true, weight, group: true, config: pageIcon(icon),
    blocks: withBlockIDs([block('Content', title, [0, 0, 48, 10], { body: `<p>${text}</p>` })]),
  }
}

function listPage (title, handle, weight, icon, moduleID, fields, extra = {}) {
  return {
    title, handle, visible: true, weight, parent: extra.parent, config: pageIcon(icon),
    blocks: withBlockIDs([recordList(title, [0, 0, 48, 40], moduleID, fields, {
      hideAddButton: extra.hideAddButton, perPage: extra.perPage, presort: extra.presort, prefilter: extra.prefilter,
    })]),
  }
}

function card (title, handle, moduleID, weight, fields, fieldRoles = {}, extraBlocks = [], icon = 'fas file-lines') {
  const blocks = [recordBlock(title, [0, 0, 48, 28], fields, { fieldRoles }), ...extraBlocks]
  return { title, handle, moduleID: String(moduleID), visible: false, weight, config: pageIcon(icon), blocks: withBlockIDs(blocks) }
}

// Страница-рабочее место: шапка и кнопки сверху, ниже вкладки; блоки вкладок лежат в page.blocks,
// но не в раскладке (так устроен блок Tabs).
function workspace (title, handle, moduleID, weight, head, tabs, { tabsHeight = 100, icon = 'fas industry' } = {}) {
  const headBlocks = head.blocks
  const y = head.height
  const tabsBlock = block('Tabs', 'Вкладки', [0, y, 48, tabsHeight], {
    style: { appearance: 'tabs', alignment: 'start', justify: 'none', orientation: 'horizontal', position: 'start' },
    tabs: [],
    magnifyOption: '',
  })
  const all = withBlockIDs([...headBlocks, tabsBlock, ...tabs.map(t => t.block)])
  const tabbedBlockIDs = all.slice(headBlocks.length + 1).map(b => b.blockID)
  all[headBlocks.length].options.tabs = tabs.map((t, i) => ({ blockID: tabbedBlockIDs[i], title: t.title, lazy: true }))
  return { title, handle, moduleID: String(moduleID), visible: false, weight, config: pageIcon(icon), blocks: all, tabbedBlockIDs }
}

export function buildPages (m) {
  const defs = []
  const parentOf = {}
  const add = (list, card) => {
    defs.push(list)
    if (card) {
      defs.push(card)
      parentOf[card.handle] = list.handle
    }
    return list
  }

  // ---- Главная ----
  defs.push({
    title: 'Главная',
    handle: 'dashboard',
    visible: true,
    weight: 0,
    config: pageIcon('fas gauge-high'),
    blocks: withBlockIDs([
      metricBlock('Требует внимания', [0, 0, 48, 14], [
        metricItem('Замечания: ответ ждёт рассмотрения', m.expertise_remarks, `status = 'answer_review' AND ${REAL}`, { color: '#36b9cc' }),
        metricItem('Замечания: критичные, не устранены', m.expertise_remarks, `criticality = 'critical' AND status != 'resolved' AND ${REAL}`, { color: '#e74a3b' }),
        metricItem('Замечания: без ответа', m.expertise_remarks, `status = 'awaiting_answer' AND ${REAL}`, { color: '#858796' }),
        metricItem('Расхождения ПД/РД к подтверждению', m.pd_rd_discrepancies, `status = 'new' AND ${REAL}`, { color: '#f6c23e' }),
        metricItem('Сравнения в работе', m.pd_rd_comparisons, `(status = 'new' OR status = 'processing') AND ${REAL}`, { color: '#2e59d9' }),
        metricItem('Комплекты документов (действующие)', m.document_sets, `(status != 'superseded' OR status IS NULL) AND ${REAL}`, { color: '#1a7f37' }),
        metricItem('Пакеты ИД с замечаниями', m.packages, `acts_with_issues > 0 AND ${REAL}`, { color: '#e74a3b' }),
        metricItem('Расхождения ВОИСР', m.voisr_check_results, `status = 'mismatch' AND ${REAL}`, { color: '#e74a3b' }),
      ]),
      recordList('Объекты', [0, 14, 48, 30], m.objects, ['name', 'code', 'address', 'status'], {
        prefilter: REAL, presort: 'updatedAt DESC', perPage: 10, hideAddButton: true,
      }),
    ]),
  })

  // ---- Объекты: список + карточка-рабочее место ----
  const objectsList = listPage('Объекты', 'objects', 10, 'fas industry', m.objects, ['name', 'code', 'address', 'status'], { prefilter: REAL })
  const objectCard = workspace('Объект', 'object', m.objects, 11, {
    height: 22,
    blocks: [
      recordBlock('Объект', [0, 0, 48, 14], ['name', 'code', 'status', 'address', 'data_kind', 'notes', 'remarks_file'], {
        fieldRoles: { name: 'title', code: 'subtitle', status: 'badge', notes: 'body' },
      }),
      ruleChain('Проверить комплектность', [0, 14, 16, 4], {
        chainID: 'stroykontrol-check-sets', label: 'Проверить комплектность комплектов', variant: 'primary', icon: 'clipboard-check',
      }),
      ruleChain('Загрузить выгрузку замечаний', [16, 14, 16, 4], {
        chainID: 'stroykontrol-import-remarks', label: 'Загрузить выгрузку замечаний экспертизы', variant: 'outline-primary', icon: 'comments',
      }),
      ruleChain('Сверить ИД с реестром', [32, 14, 16, 4], {
        chainID: 'stroykontrol-check-id', label: 'Сверить ИД с реестром', variant: 'outline-secondary', icon: 'list-check',
      }),
    ],
  }, [
    { title: 'Обзор', block: iframe('Документы, комплектность и сравниваемые версии', [0, 0, 48, 90], 'view=object&objectID=${recordID}') },
    {
      title: 'Комплекты',
      block: recordList('Комплекты документов объекта', [0, 0, 48, 90], m.document_sets,
        ['stage', 'version', 'status', 'issued_at', 'volumes_present', 'volumes_total', 'completeness_pct', 'completeness_note'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'issued_at DESC', perPage: 20 }),
    },
    {
      title: 'Сооружения',
      block: recordList('Сооружения и участки объекта', [0, 0, 48, 90], m.structures,
        ['name', 'code', 'kind', 'location', 'note'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'name ASC', perPage: 50 }),
    },
    {
      title: 'Журнал загрузок',
      block: recordList('Журнал загрузок документации', [0, 0, 48, 90], m.upload_log,
        ['uploaded_at', 'uploaded_by', 'file_name', 'volume_code', 'kind', 'note', 'pages', 'set'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'uploaded_at DESC', perPage: 50, hideAddButton: true }),
    },
    {
      title: 'Ведомость томов',
      block: recordList('Ведомость состава томов', [0, 0, 48, 90], m.object_volumes,
        ['stage', 'kind', 'volume_code', 'title', 'parent_code', 'required', 'proposed', 'source'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'volume_code ASC', perPage: 100 }),
    },
    {
      title: 'ПД ↔ РД',
      block: recordList('Сравнения ПД/РД объекта', [0, 0, 48, 90], m.pd_rd_comparisons,
        ['title', 'sets_label', 'iteration_no', 'status', 'similarity_percent', 'differing_pages'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'createdAt DESC', perPage: 20 }),
    },
    {
      title: 'Пакеты исправления',
      block: recordList('Пакеты исправления объекта', [0, 0, 48, 90], m.correction_packages,
        ['title', 'iteration_no', 'n_documents', 'n_remarks', 'resolved', 'persisting', 'new_findings', 'runs', 'last_run'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'iteration_no DESC', perPage: 20, hideAddButton: true }),
    },
    { title: 'Замечания экспертизы', block: iframe('Рабочее место замечаний экспертизы', [0, 0, 48, 90], 'view=remarks&objectID=${recordID}') },
    {
      title: 'Задачи',
      block: recordList('Задачи объекта', [0, 0, 48, 90], m.tasks,
        ['title', 'remark', 'assignee', 'due_date', 'estimate_h', 'status', 'created_by', 'created_at'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'due_date ASC', perPage: 30 }),
    },
    {
      title: 'Исполнительная документация',
      block: recordList('Пакеты проверки ИД', [0, 0, 48, 90], m.packages,
        ['title', 'ks2_number', 'status', 'acts_total', 'acts_with_issues', 'findings_high', 'findings_medium', 'finished_at'],
        { prefilter: 'object_ref = ${recordID}', refField: 'object_ref', presort: 'createdAt DESC', perPage: 20 }),
    },
    {
      title: 'Сметы и ВОИСР',
      block: recordList('Позиции ВОИСР объекта', [0, 0, 48, 90], m.voisr_items,
        ['position_number', 'work_name', 'smeta_ref', 'volume', 'amount'],
        { prefilter: 'object = ${recordID}', refField: 'object', presort: 'position_number ASC', perPage: 50 }),
    },
  ])
  add(objectsList, objectCard)

  // ---- Проверки (сквозные списки по всем объектам) ----
  defs.push(group('Документация', 'docs', 15, 'fas folder-open', 'Комплекты документов, пакеты исправления, ведомость состава томов, сооружения и журнал загрузок.'))
  defs.push(group('Сравнение', 'compare', 25, 'fas code-compare', 'Сравнения ПД/РД, их расхождения и карточки замечаний по сравнению.'))
  defs.push(group('Замечания', 'remarks', 30, 'fas comments', 'Реестр замечаний экспертизы, журнал выгрузок, срезы реестра и уведомления. Рабочее место замечаний объекта — в карточке объекта.'))
  defs.push(group('Проверки', 'checks', 20, 'fas clipboard-check', 'Сквозные списки по исполнительной документации (ИД), сметам и ВОИСР. Работать удобнее из карточки объекта.'))
  const P = { parent: 'checks' }
  const D = { parent: 'docs' }
  const C = { parent: 'compare' }
  const M = { parent: 'remarks' }
  const A = { parent: 'admin' }

  add(listPage('Комплекты документов', 'document_sets', 21, 'fas boxes-stacked', m.document_sets,
    ['object', 'stage', 'version', 'status', 'issued_at', 'completeness_pct', 'completeness_note'], { ...D, prefilter: REAL }),
  card('Комплект документов', 'document_set', m.document_sets, 22,
    ['object', 'stage', 'version', 'status', 'issued_at', 'comment', 'volumes_present', 'volumes_total', 'completeness_pct', 'completeness_note'],
    { version: 'title', stage: 'badge', status: 'badge', comment: 'body' }, [
      recordList('Файлы комплекта (PDF)', [0, 28, 48, 24], m.cmp_documents,
        ['file', 'doc_version', 'volume_code', 'volume_title', 'pages', 'detected_by'],
        { prefilter: 'set = ${recordID}', refField: 'set', presort: 'createdAt ASC', perPage: 50 }),
    ]))

  add(listPage('Сравнения ПД/РД', 'pd_rd_comparisons', 23, 'fas file-invoice', m.pd_rd_comparisons,
    ['title', 'object', 'kind', 'sets_label', 'iteration_no', 'status', 'similarity_percent', 'differing_pages'], { ...C, prefilter: REAL }),
  card('Сравнение ПД/РД', 'pd_rd_comparison', m.pd_rd_comparisons, 24, [
    'title', 'object', 'kind', 'set_pd', 'set_rd', 'sets_label', 'iteration_no', 'parent_comparison', 'status', 'pd_file', 'rd_file', 'similarity_percent',
    'total_pages_pd', 'total_pages_rd', 'matching_pages', 'differing_pages', 'comment',
  ], { title: 'title', object: 'meta', status: 'badge', comment: 'body' }, [
    // файлы можно положить прямо в сравнение (старый способ) — комплекты удобнее: файлы загружаются один раз
    recordList('Файлы, приложенные прямо к сравнению (без комплекта)', [0, 28, 48, 12], m.cmp_documents,
      ['side', 'file', 'doc_version', 'volume_code', 'volume_title', 'pages', 'detected_by'],
      { prefilter: 'comparison = ${recordID}', refField: 'comparison', presort: 'side ASC, createdAt ASC', perPage: 20 }),
    ruleChain('Сравнить документы', [0, 40, 24, 8], {
      chainID: 'stroykontrol-compare-pd-rd', label: 'Сравнить документы (агент)', variant: 'primary', icon: 'code-compare',
    }),
    ruleChain('Сформировать замечания', [24, 40, 24, 8], {
      chainID: 'stroykontrol-issue-remarks', label: 'Сформировать замечания по сравнению', variant: 'outline-primary', icon: 'list-check',
    }),
    // Просмотрщик сравнения — отдельный агент (agents/stroykontrol), встроен IFrame-блоком.
    // src относительный: IFrameBase.vue сам добавляет window.CortezaAPI, а /gateway/agents/stroykontrol/*
    // — маршрут API Gateway (create_apigw_proxy.mjs), проксирующий на реальный адрес агента.
    block('IFrame', 'Просмотр сравнения (агент)', [0, 48, 48, 56], {
      src: `${AGENT}?recordID=\${recordID}&namespaceID=\${namespaceID}`, srcField: '', displayAsImage: false,
    }),
    recordList('Карточки замечаний', [0, 104, 48, 20], m.remark_cards,
      ['document_ref', 'sheet', 'what_diverges', 'final_text', 'status'],
      { prefilter: 'comparison = ${recordID}', presort: 'createdAt DESC', perPage: 20, hideAddButton: true }),
  ]))

  add(listPage('Пакеты исправления', 'correction_packages', 24.5, 'fas box-open', m.correction_packages,
    ['title', 'object', 'iteration_no', 'n_documents', 'n_remarks', 'resolved', 'persisting', 'new_findings', 'runs', 'last_run'], { ...D, hideAddButton: true, prefilter: REAL, presort: 'last_run DESC' }),
  card('Пакет исправления', 'correction_package', m.correction_packages, 24.6,
    ['title', 'object', 'comparison', 'parent_comparison', 'iteration_no', 'n_documents', 'n_remarks', 'resolved', 'persisting', 'new_findings', 'runs', 'last_run', 'history_json'],
    { title: 'title', history_json: 'body' }, [
      recordList('Исправленные документы', [0, 28, 48, 14], m.cmp_documents,
        ['file', 'volume_code', 'volume_title', 'change_no', 'pages'], { prefilter: 'comparison = ${record.values.comparison}', presort: 'createdAt ASC', perPage: 20, hideAddButton: true }),
    ], 'fas box-open'))

  add(listPage('Расхождения ПД/РД', 'pd_rd_discrepancies', 25, 'fas triangle-exclamation', m.pd_rd_discrepancies,
    ['comparison', 'severity', 'kind', 'status', 'iter_state', 'page_pd', 'page_rd', 'description'], { ...C, prefilter: REAL }),
  card('Расхождение ПД/РД', 'pd_rd_discrepancy', m.pd_rd_discrepancies, 26,
    ['comparison', 'severity', 'kind', 'status', 'page_pd', 'page_rd', 'description', 'anchor_text', 'evidence'],
    { discrepancy_type: 'title', severity: 'badge', status: 'badge', description: 'body' }))

  add(listPage('Замечания экспертизы', 'expertise_remarks', 27, 'fas comments', m.expertise_remarks,
    ['object', 'number', 'origin', 'criticality', 'status', 'discipline', 'expert', 'iterations', 'verdict', 'assignee', 'due_date', 'text'],
    { ...M, hideAddButton: true, perPage: 50, prefilter: REAL }),
  card('Замечание экспертизы', 'expertise_remark', m.expertise_remarks, 28, [
    'object', 'number', 'origin', 'source_comparison', 'criticality', 'status', 'assignee', 'due_date', 'discipline', 'expert', 'section', 'iterations', 'source_status',
    'text', 'refs_text', 'basis', 'answer', 'reject_reason', 'prev_answer', 'prev_reject_reason', 'proposals', 'expected_result', 'pto_note', 'estimate_h',
    'verdict', 'match_score', 'match_note', 'claim_mismatch', 'linked_documents', 'matched_discrepancies', 'ai_verdict', 'ai_note', 'ai_at', 'outcome_final', 'outcome_decision', 'outcome_comment', 'outcome_by', 'outcome_at',
  ], { number: 'title', status: 'badge', criticality: 'badge', text: 'body' }, [
    recordList('История итераций', [0, 28, 48, 24], m.expertise_remark_history,
      ['at', 'iteration', 'kind', 'status', 'source', 'note', 'answer', 'reject_reason', 'author'],
      { prefilter: 'remark = ${recordID}', presort: 'at ASC', perPage: 20, hideAddButton: true }),
  ]))

  add(listPage('Пакеты проверки ИД', 'packages', 29, 'fas file-signature', m.packages,
    ['title', 'object_ref', 'ks2_number', 'status', 'acts_total', 'acts_with_issues', 'findings_high', 'findings_medium', 'finished_at'], { ...P, prefilter: REAL }),
  packageCard(m))

  add(listPage('Замечания по ИД', 'findings', 30, 'fas triangle-exclamation', m.findings,
    ['package', 'check_no', 'severity', 'act', 'position', 'description', 'source'], { ...P, hideAddButton: true, perPage: 50 }),
  card('Замечание по ИД', 'finding', m.findings, 31, ['package', 'act', 'registry_row', 'registry_row_no', 'position', 'check_no', 'check_name', 'severity',
    'field', 'expected', 'actual', 'description', 'page', 'bbox', 'source'],
  { check_name: 'title', severity: 'badge', description: 'body' }, [idView('По месту', [0, 28, 48, 56], 'findingID')]))

  add(listPage('Проверка ВОИСР', 'voisr_check_results', 32, 'fas scale-balanced', m.voisr_check_results,
    ['voisr_item', 'smeta_volume', 'smeta_amount', 'volume_diff', 'amount_diff', 'amount_diff_percent', 'status'], { ...P, prefilter: REAL }),
  card('Результат проверки ВОИСР', 'voisr_check_result', m.voisr_check_results, 33,
    ['voisr_item', 'smeta_volume', 'smeta_amount', 'volume_diff', 'amount_diff', 'amount_diff_percent', 'status', 'note'],
    { voisr_item: 'title', status: 'badge', note: 'body' }))

  // ---- Отчёты ----
  defs.push({
    title: 'Отчёты', handle: 'reports', visible: true, weight: 35, config: pageIcon('fas file-export'),
    blocks: withBlockIDs([iframe('Выгрузки по объектам и сравнениям', [0, 0, 48, 90], 'view=reports')]),
  })

  // ---- Справочники ----
  defs.push(group('Справочники', 'refs', 40, 'fas book', 'Нормативные документы, юрлица, специалисты и ведомости состава томов.'))
  const R = { parent: 'refs' }
  add(listPage('Ведомость состава томов', 'object_volumes', 41, 'fas list-ol', m.object_volumes,
    ['object', 'stage', 'kind', 'volume_code', 'title', 'parent_code', 'required', 'proposed', 'source'], { ...D, perPage: 100, presort: 'volume_code ASC' }),
  card('Позиция ведомости', 'object_volume', m.object_volumes, 42, ['object', 'stage', 'kind', 'volume_code', 'title', 'parent_code', 'required', 'proposed', 'source', 'evidence'], { volume_code: 'title' }))
  add(listPage('Сооружения и участки', 'structures', 42.5, 'fas bridge', m.structures,
    ['object', 'name', 'code', 'kind', 'location'], { ...D, perPage: 100, presort: 'name ASC' }))
  add(listPage('Нормативные документы', 'norm_docs', 43, 'fas book', m.norm_docs,
    ['code', 'title', 'valid_from', 'valid_to', 'replaced_by', 'work_tags'], { ...R, perPage: 100, presort: 'code ASC' }),
  card('Нормативный документ', 'norm_doc', m.norm_docs, 44, ['code', 'title', 'valid_from', 'valid_to', 'replaced_by', 'work_tags'], { code: 'title' }))
  add(listPage('Юрлица (ЕГРЮЛ)', 'organizations', 45, 'fas building', m.organizations,
    ['name', 'ogrn', 'inn', 'egrul_status', 'egrul_short_name', 'egrul_region', 'checked_at'], R),
  card('Юрлицо', 'organization', m.organizations, 46, [
    'name', 'ogrn', 'inn', 'address', 'sro_name', 'egrul_status', 'egrul_name', 'egrul_short_name', 'egrul_inn',
    'egrul_kpp', 'egrul_region', 'egrul_reg_date', 'egrul_end_date', 'egrul_director', 'checked_at',
  ], { name: 'title', egrul_status: 'badge' }))
  add(listPage('Специалисты НРС', 'nrs_specialists', 47, 'fas id-badge', m.nrs_specialists,
    ['nrs_id', 'fio', 'status', 'status_text', 'checked_at'], R))

  // ---- Служебное ----
  defs.push(group('Администрирование', 'admin', 50, 'fas gear', 'Получатели уведомлений, роли и доступ, демо-данные. Для повседневной работы не нужны.'))
  defs.push({
    title: 'Роли и доступ', handle: 'roles', visible: true, weight: 60.5, parent: 'admin', config: pageIcon('fas user-shield'),
    blocks: withBlockIDs([block('Content', 'Роли и доступ', [0, 0, 48, 30], { body: ROLES_HTML })]),
  })
  add(listPage('Демо-данные (объекты)', 'demo_objects', 61, 'fas flask', m.objects,
    ['name', 'code', 'address', 'status', 'data_kind'], { ...A, prefilter: "data_kind = 'demo'" }))
  add(listPage('Документы сравнений', 'cmp_documents', 62, 'fas file-pdf', m.cmp_documents,
    ['set', 'comparison', 'side', 'file', 'doc_version', 'volume_code', 'volume_title', 'section', 'structure', 'change_no', 'pages', 'detected_by'], D),
  card('Документ', 'cmp_document', m.cmp_documents, 63,
    ['set', 'comparison', 'side', 'file', 'doc_version', 'volume_code', 'volume_title', 'section', 'structure', 'change_no', 'object_name', 'structure_object', 'subsection', 'part', 'book', 'pages', 'file_hash', 'released', 'detected_by'],
    { file: 'title', side: 'badge' }))
  add(listPage('Журнал загрузок документации', 'upload_log', 63.5, 'fas file-arrow-up', m.upload_log,
    ['uploaded_at', 'object', 'uploaded_by', 'file_name', 'volume_code', 'kind', 'note', 'pages'], { ...D, hideAddButton: true, presort: 'uploaded_at DESC' }))
  add(listPage('Журнал выгрузок замечаний', 'expertise_imports', 64, 'fas clock-rotate-left', m.expertise_imports,
    ['object', 'imported_at', 'file', 'total', 'new_count', 'changed_count', 'status_changes', 'reopened', 'archived_count'], { ...M, hideAddButton: true }))
  add(listPage('Срезы реестра замечаний', 'expertise_snapshots', 64.5, 'fas chart-line', m.expertise_snapshots,
    ['object', 'day', 'total', 'overdue', 'n_awaiting_answer', 'n_being_fixed', 'n_answer_review', 'n_rework', 'n_resolved'], { ...M, hideAddButton: true }))
  add(listPage('Задачи', 'tasks', 64.4, 'fas list-check', m.tasks,
    ['title', 'object', 'remark', 'assignee', 'due_date', 'estimate_h', 'status', 'created_by', 'done_at'], { ...S, prefilter: REAL, presort: 'due_date ASC' }),
  card('Задача', 'task', m.tasks, 64.41,
    ['title', 'object', 'remark', 'requirement', 'status', 'assignee', 'due_date', 'estimate_h', 'description', 'comment', 'created_by', 'created_at', 'done_at', 'closed_by'],
    { title: 'title', status: 'badge', description: 'body', comment: 'body' }, [
      recordList('История задачи', [0, 28, 48, 20], m.task_history,
        ['at', 'kind', 'source', 'note', 'author', 'status'],
        { prefilter: 'task = ${recordID}', presort: 'at ASC', perPage: 20, hideAddButton: true }),
    ], 'fas list-check'))
  add(listPage('Получатели уведомлений', 'notify_recipients', 64.6, 'fas bell', m.notify_recipients,
    ['name', 'email', 'role', 'object', 'events', 'active'], A))
  add(listPage('Уведомления', 'notifications', 64.7, 'fas envelope', m.notifications,
    ['created_at', 'object', 'number', 'event', 'recipient', 'title', 'delivery', 'read', 'error'], { ...M, hideAddButton: true }))
  add(listPage('Карточки замечаний (по сравнению)', 'remark_cards', 65, 'fas clipboard-list', m.remark_cards,
    ['comparison', 'document_ref', 'sheet', 'final_text', 'status'], C),
  card('Карточка замечания', 'remark_card', m.remark_cards, 66,
    ['comparison', 'document_ref', 'sheet', 'what_diverges', 'must_align', 'draft_text', 'final_text', 'status'],
    { document_ref: 'title', status: 'badge', what_diverges: 'body' }))
  add(listPage('Акты ИД', 'acts', 67, 'fas file-signature', m.acts,
    ['package', 'file_name', 'doc_number', 'doc_date', 'positions', 'ocr_status', 'check_status', 'issues_high', 'issues_medium'], { ...P, hideAddButton: true }),
  actCard(m))
  add(listPage('Строки реестра ИД (передача)', 'registry_rows', 69, 'fas table-list', m.registry_rows,
    ['package', 'row_no', 'kind', 'position', 'doc_number', 'doc_date', 'qty_ks2', 'qty_project', 'qty_fact'], { ...P, hideAddButton: true, perPage: 50 }),
  card('Строка реестра', 'registry_row', m.registry_rows, 70, [
    'package', 'row_no', 'seq', 'kind', 'position', 'work_name', 'doc_name', 'doc_number', 'doc_number_norm',
    'doc_date', 'doc_date_raw', 'unit', 'qty_vrc', 'qty_total', 'qty_ks2', 'qty_project', 'qty_fact',
  ]))
  add(listPage('Сводка по позициям (реестр ↔ КС-2)', 'position_summary', 71, 'fas scale-balanced', m.position_summary,
    ['package', 'position', 'name', 'unit', 'registry_sum', 'registry_ks2', 'ks2_qty', 'delta', 'status'], { ...P, hideAddButton: true, perPage: 100 }))
  add(listPage('Реестр ИД (комплектность)', 'id_registry', 72, 'fas list-check', m.id_registry,
    ['object', 'doc_name', 'doc_number', 'doc_version', 'doc_type', 'is_required'], P),
  card('Позиция реестра ИД', 'id_registry_item', m.id_registry, 73,
    ['object', 'doc_name', 'doc_number', 'doc_version', 'doc_type', 'is_required'], { doc_name: 'title', doc_number: 'subtitle', doc_type: 'badge' }))
  add(listPage('Файлы ИД', 'id_files', 74, 'fas file-arrow-up', m.id_files,
    ['object', 'file', 'recognized_name', 'matched_registry', 'match_status'], P),
  card('Файл ИД', 'id_file', m.id_files, 75, ['object', 'file', 'recognized_name', 'recognized_number', 'recognized_version',
    'matched_registry', 'match_status', 'uploaded_at', 'extracted_text'], { recognized_name: 'title', match_status: 'badge' }))
  add(listPage('Проверки комплектности ИД', 'id_check_runs', 76, 'fas clipboard-check', m.id_check_runs,
    ['object', 'run_date', 'total_registry_docs', 'uploaded_docs', 'missing_docs', 'mismatched_docs', 'overall_status'], P),
  card('Проверка комплектности ИД', 'id_check_run', m.id_check_runs, 77, ['object', 'run_date', 'total_registry_docs', 'uploaded_docs', 'missing_docs',
    'mismatched_docs', 'overall_status', 'report_file'], { object: 'title', overall_status: 'badge' }))
  add(listPage('Позиции смет', 'smeta_items', 78, 'fas file-invoice-dollar', m.smeta_items,
    ['object', 'smeta_number', 'item_number', 'work_name', 'volume', 'unit', 'total_amount'], P),
  card('Позиция сметы', 'smeta_item', m.smeta_items, 79,
    ['object', 'smeta_number', 'item_number', 'work_name', 'unit', 'volume', 'unit_price', 'total_amount'], { work_name: 'title', smeta_number: 'subtitle' }))
  add(listPage('Позиции ВОИСР', 'voisr_items', 80, 'fas table-list', m.voisr_items,
    ['object', 'position_number', 'work_name', 'smeta_ref', 'volume', 'amount'], P),
  card('Позиция ВОИСР', 'voisr_item', m.voisr_items, 81, ['object', 'position_number', 'work_name', 'smeta_ref', 'volume', 'amount'],
    { work_name: 'title', position_number: 'subtitle' }, [ruleChain('Проверить ВОИСР', [0, 20, 48, 8], {
      chainID: 'stroykontrol-check-voisr', label: 'Проверить со сметой (агент)', variant: 'primary', icon: 'scale-balanced',
    })]))

  function packageCard (m) {
    const byPackage = 'package = ${recordID}'
    const blocks = [
      recordBlock('Пакет', [0, 0, 32, 30], [
        'title', 'object_ref', 'object', 'ks2_number', 'period_from', 'period_to', 'status', 'progress',
        'registry_xlsx', 'registry_pdf', 'ks2_xlsx', 'ks2_pdf', 'act_files', 'comment', 'data_kind',
      ], { fieldRoles: { title: 'title', status: 'badge', progress: 'body' } }),
      ruleChain('Проверить', [32, 0, 16, 6], { chainID: 'stroykontrol-id-run', label: 'Запустить проверку ИД', variant: 'primary', icon: 'clipboard-check' }),
      recordBlock('Итоги', [32, 6, 16, 24], [
        'acts_total', 'acts_ok', 'acts_with_issues', 'findings_high', 'findings_medium', 'findings_low', 'started_at', 'finished_at',
      ], { horizontal: false }),
      idView('Замечания по месту', [0, 30, 48, 64], 'packageID'),
      recordList('Замечания (таблица)', [0, 94, 48, 30], m.findings,
        ['check_no', 'severity', 'act', 'position', 'registry_row_no', 'description', 'expected', 'actual', 'page', 'source'],
        { prefilter: byPackage, presort: 'check_no ASC, severity ASC', perPage: 50, hideAddButton: true }),
      recordList('Акты', [0, 124, 48, 30], m.acts,
        ['file_name', 'doc_number', 'doc_date', 'positions', 'p1_vol_project', 'p1_vol_fact', 'ocr_status', 'check_status', 'issues_high', 'issues_medium'],
        { prefilter: byPackage, presort: 'file_name ASC', perPage: 50, hideAddButton: true }),
      recordList('Реестр ↔ КС-2 (проверка 8)', [0, 154, 48, 30], m.position_summary,
        ['position', 'name', 'unit', 'registry_sum', 'registry_ks2', 'ks2_qty', 'delta', 'status'],
        { prefilter: byPackage, presort: 'position ASC', perPage: 100, hideAddButton: true }),
      recordList('Строки реестра', [0, 184, 48, 30], m.registry_rows,
        ['row_no', 'kind', 'position', 'doc_name', 'doc_number', 'doc_date', 'unit', 'qty_ks2', 'qty_project', 'qty_fact'],
        { prefilter: byPackage, presort: 'row_no ASC', perPage: 50, hideAddButton: true }),
    ]
    return { title: 'Пакет проверки ИД', handle: 'package', moduleID: String(m.packages), visible: false, weight: 30, config: pageIcon('fas boxes-stacked'), blocks: withBlockIDs(blocks) }
  }

  function actCard (m) {
    const blocks = [
      recordBlock('Акт', [0, 0, 48, 40], [
        'package', 'file_name', 'act_type', 'doc_number', 'doc_date', 'number', 'date', 'positions', 'registry_row',
        'check_status', 'issues_high', 'issues_medium', 'issues_low',
        'p1_work', 'p1_vol_project', 'p1_vol_fact', 'p1_unit', 'p2_rd_code', 'p2_text', 'p3_materials', 'p4_docs',
        'p5_start', 'p5_end', 'p6_norms', 'p6_rd_code', 'appendix', 'attached',
        'scheme_rd_code', 'scheme_vol_project', 'scheme_vol_fact', 'ocr_status', 'ocr_notes',
      ], { fieldRoles: { file_name: 'title', check_status: 'badge' } }),
      idView('Замечания по месту', [0, 40, 48, 64], 'actID'),
      recordList('Замечания по акту', [0, 104, 48, 24], m.findings,
        ['check_no', 'severity', 'description', 'expected', 'actual', 'page', 'source'],
        { prefilter: 'act = ${recordID}', presort: 'check_no ASC', perPage: 50, hideAddButton: true }),
      recordList('Представители', [0, 128, 48, 20], m.participants,
        ['role', 'fio', 'position_title', 'order_no', 'order_date', 'nrs_id'], { prefilter: 'act = ${recordID}', hideAddButton: true }),
      recordBlock('Извлечённые данные', [0, 148, 48, 20], ['pages_json', 'extracted_json']),
    ]
    return { title: 'Акт', handle: 'act', moduleID: String(m.acts), visible: false, weight: 68, config: pageIcon('fas file-signature'), blocks: withBlockIDs(blocks) }
  }

  return { defs, parentOf }
}

// Просмотрщик проверки ИД (веб-приложение агента, ?view=idcheck): скан с отмеченными замечаниями.
function idView (title, xywh, param) {
  return iframe(title, xywh, `view=idcheck&${param}=\${recordID}`)
}
