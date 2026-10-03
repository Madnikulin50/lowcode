/**
 * Пространство «Проверка ИД (АОСР)»: проверка АОСР/АООК против реестра
 * передачи ИД, формы КС-2 и нормативных требований (8 проверок ТЗ).
 *
 * Сама проверка — в Go-агенте agents/stroykontrol (пакет idcheck); здесь
 * только структура данных, страницы и rule chain, которая по кнопке
 * отправляет пакет агенту. Идемпотентно (обновляет по handle).
 *
 *   cd agents/stroykontrol/compose_id
 *   COMPOSE_DSN='postgres://…/test11?sslmode=disable' node apply_id.mjs
 *
 * IDCHECK_AGENT_URL — где агент слушает (по умолчанию http://localhost:8092).
 */
import { writeFileSync, existsSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  selectOptions, field, recordRel, fileField, dateField, numberField,
  block, recordList, recordBlock, metricItem, metricBlock, ruleChain, pageIcon, withBlockIDs,
  mintToken, detectBase, apiFactory, ensureNamespace, ensureModule, ensurePage, ensureRuleChain,
  parentPages, createRecord, setOf,
} from '../compose/helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))
const AGENT_URL = (process.env.IDCHECK_AGENT_URL || 'http://localhost:8092').replace(/\/$/, '')

const PACKAGE_STATUS = [
  ['new', 'Новый', { variant: 'secondary' }],
  ['processing', 'Проверяется', { variant: 'warning' }],
  ['done', 'Проверен', { variant: 'success' }],
  ['failed', 'Ошибка', { variant: 'danger' }],
]
const ROW_KIND = [['section', 'Раздел'], ['position', 'Позиция ВРЦ'], ['act', 'Акт']]
const OCR_STATUS = [['processing', 'Распознаётся', { variant: 'warning' }], ['done', 'Распознан', { variant: 'success' }], ['failed', 'Ошибка', { variant: 'danger' }]]
const CHECK_STATUS = [['ok', 'Без замечаний', { variant: 'success' }], ['issues', 'Есть замечания', { variant: 'danger' }]]
const SEVERITY = [
  ['high', 'Высокая', { variant: 'danger' }],
  ['medium', 'Средняя', { variant: 'warning' }],
  ['low', 'Низкая', { variant: 'info' }],
  ['info', 'Инфо', { variant: 'secondary' }],
]
const CHECK_NO = [
  ['1', '1. Акт ↔ реестр'],
  ['2', '2. Приложения / сканы'],
  ['3', '3. Шифр РД'],
  ['4', '4. Объём ↔ схема'],
  ['5', '5. Нормативка п.6'],
  ['6', '6. Представители / НРС'],
  ['7', '7. Юрлица ↔ ЕГРЮЛ'],
  ['8', '8. Реестр ↔ КС-2'],
]
const SOURCE = [['det', 'Детерминированно'], ['ai', 'ИИ'], ['ext', 'Внешний реестр']]
const POSITION_STATUS = [
  ['ok', 'Совпадает', { variant: 'success' }],
  ['mismatch', 'Расхождение', { variant: 'danger' }],
  ['missing_ks2', 'Нет в КС-2', { variant: 'danger' }],
  ['missing_registry', 'Нет в реестре', { variant: 'danger' }],
]
const EGRUL_STATUS = [['active', 'Действует', { variant: 'success' }], ['ceased', 'Прекращено', { variant: 'danger' }], ['not_found', 'Не найдено', { variant: 'danger' }]]
const NRS_STATUS = [['active', 'Действует', { variant: 'success' }], ['excluded', 'Исключён', { variant: 'danger' }], ['not_found', 'Не найден', { variant: 'danger' }]]

const text = (name, label, extra = {}) => field(name, label, 'String', { ...extra, options: { multiLine: true, useRichTextEditor: false } })
const str = (name, label, extra = {}) => field(name, label, 'String', extra)
const sel = (name, label, opts) => field(name, label, 'Select', { options: selectOptions(opts) })
const qty = (name, label) => numberField(name, label, { precision: 4 })
const int = (name, label) => numberField(name, label, { precision: 0 })

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  console.log('API', base)

  const nsID = await ensureNamespace(api, {
    name: 'Проверка ИД (АОСР)',
    slug: 'stroykontrol-id',
    meta: {
      subtitle: 'АОСР/АООК ↔ реестр ИД ↔ КС-2 ↔ нормативные требования',
      description: 'Проверка первичной исполнительной документации: соответствие реестру и КС-2, комплектность, шифры, объёмы, нормативка, представители (НРС), юрлица (ЕГРЮЛ).',
    },
  })

  const m = {}
  m.packages = await ensureModule(api, nsID, {
    name: 'Пакеты проверки',
    handle: 'packages',
    fields: [
      str('title', 'Название', { required: true }),
      str('object', 'Объект'),
      str('ks2_number', 'КС-2 №'),
      dateField('period_from', 'Период с'),
      dateField('period_to', 'Период по'),
      fileField('registry_xlsx', 'Реестр ИД (xlsx)', { mimetypes: '.xlsx' }),
      fileField('registry_pdf', 'Реестр ИД (pdf)'),
      fileField('ks2_xlsx', 'КС-2 (xlsx)', { mimetypes: '.xlsx' }),
      fileField('ks2_pdf', 'КС-2 (pdf)'),
      fileField('act_files', 'Акты АОСР/АООК (pdf)', { multi: true }),
      sel('status', 'Статус', PACKAGE_STATUS),
      text('progress', 'Ход проверки'),
      int('acts_total', 'Актов'),
      int('acts_ok', 'Без замечаний'),
      int('acts_with_issues', 'С замечаниями'),
      int('findings_high', 'Замечаний: высокие'),
      int('findings_medium', 'Замечаний: средние'),
      int('findings_low', 'Замечаний: низкие'),
      dateField('started_at', 'Начало проверки', { onlyDate: false }),
      dateField('finished_at', 'Окончание проверки', { onlyDate: false }),
      text('comment', 'Комментарий'),
    ],
  })

  m.registry_rows = await ensureModule(api, nsID, {
    name: 'Строки реестра',
    handle: 'registry_rows',
    fields: [
      recordRel('package', 'Пакет', m.packages, 'title', ['title']),
      int('row_no', 'Строка'),
      str('seq', '№ п/п'),
      sel('kind', 'Тип строки', ROW_KIND),
      str('position', 'Позиция ВРЦ'),
      text('work_name', 'Наименование работ'),
      text('doc_name', 'Наименование документа'),
      str('doc_number', '№ документа'),
      str('doc_number_norm', '№ (нормализованный)'),
      dateField('doc_date', 'Дата'),
      str('doc_date_raw', 'Дата (как в реестре)'),
      str('unit', 'Ед. изм.'),
      qty('qty_vrc', 'Кол-во по ВРЦ (9)'),
      qty('qty_total', 'Итого с начала (10)'),
      qty('qty_ks2', 'КС-2 (11)'),
      qty('qty_project', 'Проект по АОСР (12)'),
      qty('qty_fact', 'Факт по АОСР (13)'),
    ],
  })

  m.acts = await ensureModule(api, nsID, {
    name: 'Акты',
    handle: 'acts',
    fields: [
      recordRel('package', 'Пакет', m.packages, 'title', ['title']),
      str('file_name', 'Файл'),
      str('attachment_id', 'ID вложения'),
      str('number', '№ (из имени файла)'),
      str('number_norm', '№ (нормализованный)'),
      dateField('date', 'Дата (из имени файла)'),
      str('positions', 'Позиции ВРЦ'),
      recordRel('registry_row', 'Строка реестра', m.registry_rows, 'doc_number', ['doc_number', 'position']),
      str('act_type', 'Тип акта'),
      str('doc_number', '№ в акте'),
      dateField('doc_date', 'Дата в акте'),
      text('p1_work', 'п.1 Работы'),
      qty('p1_vol_project', 'п.1 Vпр'),
      qty('p1_vol_fact', 'п.1 Vф'),
      str('p1_unit', 'п.1 Ед.'),
      str('p2_rd_code', 'п.2 Шифр РД'),
      text('p2_text', 'п.2 Текст'),
      text('p3_materials', 'п.3 Материалы'),
      text('p4_docs', 'п.4 Документы'),
      dateField('p5_start', 'п.5 Начало работ'),
      dateField('p5_end', 'п.5 Окончание работ'),
      text('p6_norms', 'п.6 Нормативные документы'),
      str('p6_rd_code', 'п.6 Шифр РД'),
      text('appendix', 'Приложения'),
      text('attached', 'Вложенные сканы'),
      str('scheme_rd_code', 'Шифр на исп. схеме'),
      qty('scheme_vol_project', 'Схема: проект'),
      qty('scheme_vol_fact', 'Схема: факт'),
      text('pages_json', 'Страницы (классификация)'),
      sel('ocr_status', 'Распознавание', OCR_STATUS),
      text('ocr_notes', 'Примечания распознавания'),
      text('extracted_json', 'Извлечённые данные (JSON)'),
      sel('check_status', 'Результат', CHECK_STATUS),
      int('issues_high', 'Высоких'),
      int('issues_medium', 'Средних'),
      int('issues_low', 'Низких'),
    ],
  })

  m.participants = await ensureModule(api, nsID, {
    name: 'Представители',
    handle: 'participants',
    fields: [
      recordRel('package', 'Пакет', m.packages, 'title', ['title']),
      recordRel('act', 'Акт', m.acts, 'file_name', ['file_name', 'doc_number']),
      text('role', 'Роль'),
      str('fio', 'ФИО'),
      str('position_title', 'Должность'),
      str('org', 'Организация'),
      str('order_no', 'Приказ'),
      dateField('order_date', 'Дата приказа'),
      str('nrs_id', 'ID в НРС'),
    ],
  })

  m.organizations = await ensureModule(api, nsID, {
    name: 'Юрлица (ЕГРЮЛ)',
    handle: 'organizations',
    fields: [
      str('name', 'Наименование в акте'),
      str('ogrn', 'ОГРН'),
      str('inn', 'ИНН'),
      text('address', 'Адрес в акте'),
      str('sro_name', 'СРО'),
      sel('egrul_status', 'Статус ЕГРЮЛ', EGRUL_STATUS),
      text('egrul_name', 'Полное наименование (ЕГРЮЛ)'),
      str('egrul_short_name', 'Краткое наименование (ЕГРЮЛ)'),
      str('egrul_inn', 'ИНН (ЕГРЮЛ)'),
      str('egrul_kpp', 'КПП (ЕГРЮЛ)'),
      str('egrul_region', 'Регион (ЕГРЮЛ)'),
      dateField('egrul_reg_date', 'Дата регистрации'),
      dateField('egrul_end_date', 'Дата прекращения'),
      text('egrul_director', 'Руководитель'),
      dateField('checked_at', 'Проверено', { onlyDate: false }),
    ],
  })

  m.nrs_specialists = await ensureModule(api, nsID, {
    name: 'Специалисты НРС',
    handle: 'nrs_specialists',
    fields: [
      str('nrs_id', 'Идентификационный номер'),
      str('fio', 'ФИО (из акта)'),
      sel('status', 'Статус', NRS_STATUS),
      str('status_text', 'Статус (НРС)'),
      text('work_type', 'Вид работ'),
      dateField('checked_at', 'Проверено', { onlyDate: false }),
    ],
  })

  m.norm_docs = await ensureModule(api, nsID, {
    name: 'Нормативные документы',
    handle: 'norm_docs',
    fields: [
      str('code', 'Обозначение', { required: true }),
      text('title', 'Наименование'),
      dateField('valid_from', 'Действует с'),
      dateField('valid_to', 'Утратил силу с'),
      str('replaced_by', 'Заменён на'),
      str('work_tags', 'Требуется для работ (ключевые слова через запятую, * — для всех)'),
    ],
  })

  m.findings = await ensureModule(api, nsID, {
    name: 'Замечания',
    handle: 'findings',
    fields: [
      recordRel('package', 'Пакет', m.packages, 'title', ['title']),
      recordRel('act', 'Акт', m.acts, 'file_name', ['file_name', 'doc_number']),
      recordRel('registry_row', 'Строка реестра', m.registry_rows, 'doc_number', ['doc_number', 'position']),
      int('registry_row_no', 'Строка реестра №'),
      str('position', 'Позиция ВРЦ'),
      sel('check_no', 'Проверка', CHECK_NO),
      str('check_name', 'Название проверки'),
      sel('severity', 'Критичность', SEVERITY),
      str('field', 'Поле'),
      text('expected', 'Ожидалось'),
      text('actual', 'Фактически'),
      text('description', 'Описание'),
      int('page', 'Стр.'),
      str('bbox', 'Место на странице (x,y,w,h)'),
      sel('source', 'Источник', SOURCE),
    ],
  })

  m.position_summary = await ensureModule(api, nsID, {
    name: 'Сводка по позициям (реестр ↔ КС-2)',
    handle: 'position_summary',
    fields: [
      recordRel('package', 'Пакет', m.packages, 'title', ['title']),
      str('position', 'Позиция'),
      text('name', 'Наименование'),
      str('unit', 'Ед.'),
      qty('registry_sum', 'Σ min(12,13) по реестру'),
      qty('registry_ks2', 'Реестр, кол. 11'),
      qty('ks2_qty', 'КС-2'),
      qty('delta', 'Разница'),
      sel('status', 'Статус', POSITION_STATUS),
    ],
  })
  console.log('modules', m)

  // ---- Pages ----
  const listPage = (title, handle, weight, icon, moduleID, fields, extra = {}) => ({
    title, handle, visible: extra.visible ?? true, weight, config: pageIcon(icon),
    blocks: withBlockIDs([recordList(title, [0, 0, 48, 40], moduleID, fields, extra)]),
  })
  const card = (title, handle, moduleID, weight, blocks) => ({
    title, handle, moduleID: String(moduleID), visible: false, weight, config: pageIcon('fas file-lines'),
    blocks: withBlockIDs(blocks),
  })
  // In-place viewer (agent web app, ?view=idcheck) — the scan with the
  // findings boxed, the registry/КС-2 rows highlighted. Relative src goes
  // through the API Gateway route /gateway/agents/stroykontrol/* (see
  // ../compose/create_apigw_proxy.mjs), like the ПД/РД viewer.
  const viewer = (title, xywh, param) => block('IFrame', title, xywh, {
    src: `/gateway/agents/stroykontrol/?view=idcheck&${param}=\${recordID}&namespaceID=\${namespaceID}`,
    srcField: '',
    displayAsImage: false,
  })
  const byPackage = "package = ${recordID}"
  const byAct = "act = ${recordID}"

  const pages = [
    listPage('Пакеты проверки', 'packages', 0, 'fas boxes-stacked', m.packages,
      ['title', 'object', 'ks2_number', 'status', 'acts_total', 'acts_with_issues', 'findings_high', 'findings_medium', 'finished_at']),
    card('Пакет проверки', 'package', m.packages, 1, [
      recordBlock('Пакет', [0, 0, 32, 30], [
        'title', 'object', 'ks2_number', 'period_from', 'period_to', 'status', 'progress',
        'registry_xlsx', 'registry_pdf', 'ks2_xlsx', 'ks2_pdf', 'act_files', 'comment',
      ], { fieldRoles: { title: 'title', status: 'badge', progress: 'body' } }),
      ruleChain('Проверить', [32, 0, 16, 6], {
        chainID: 'stroykontrol-id-run', label: 'Запустить проверку ИД', variant: 'primary', icon: 'clipboard-check',
      }),
      recordBlock('Итоги', [32, 6, 16, 24], [
        'acts_total', 'acts_ok', 'acts_with_issues', 'findings_high', 'findings_medium', 'findings_low', 'started_at', 'finished_at',
      ], { horizontal: false }),
      viewer('Замечания по месту', [0, 30, 48, 64], 'packageID'),
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
    ]),

    listPage('Акты', 'acts', 10, 'fas file-signature', m.acts,
      ['package', 'file_name', 'doc_number', 'doc_date', 'positions', 'ocr_status', 'check_status', 'issues_high', 'issues_medium'], { hideAddButton: true }),
    card('Акт', 'act', m.acts, 11, [
      recordBlock('Акт', [0, 0, 48, 40], [
        'package', 'file_name', 'act_type', 'doc_number', 'doc_date', 'number', 'date', 'positions', 'registry_row',
        'check_status', 'issues_high', 'issues_medium', 'issues_low',
        'p1_work', 'p1_vol_project', 'p1_vol_fact', 'p1_unit', 'p2_rd_code', 'p2_text', 'p3_materials', 'p4_docs',
        'p5_start', 'p5_end', 'p6_norms', 'p6_rd_code', 'appendix', 'attached',
        'scheme_rd_code', 'scheme_vol_project', 'scheme_vol_fact', 'ocr_status', 'ocr_notes',
      ], { fieldRoles: { file_name: 'title', check_status: 'badge' } }),
      viewer('Замечания по месту', [0, 40, 48, 64], 'actID'),
      recordList('Замечания по акту', [0, 104, 48, 24], m.findings,
        ['check_no', 'severity', 'description', 'expected', 'actual', 'page', 'source'],
        { prefilter: byAct, presort: 'check_no ASC', perPage: 50, hideAddButton: true }),
      recordList('Представители', [0, 128, 48, 20], m.participants,
        ['role', 'fio', 'position_title', 'order_no', 'order_date', 'nrs_id'],
        { prefilter: byAct, hideAddButton: true }),
      recordBlock('Извлечённые данные', [0, 148, 48, 20], ['pages_json', 'extracted_json']),
    ]),

    listPage('Замечания', 'findings', 20, 'fas triangle-exclamation', m.findings,
      ['package', 'check_no', 'severity', 'act', 'position', 'description', 'source'], { hideAddButton: true, perPage: 50 }),
    card('Замечание', 'finding', m.findings, 21, [
      viewer('По месту', [0, 24, 48, 56], 'findingID'),
      recordBlock('Замечание', [0, 0, 48, 24], [
        'package', 'act', 'registry_row', 'registry_row_no', 'position', 'check_no', 'check_name', 'severity',
        'field', 'expected', 'actual', 'description', 'page', 'bbox', 'source',
      ], { fieldRoles: { check_name: 'title', severity: 'badge', description: 'body' } }),
    ]),

    listPage('Строки реестра', 'registry_rows', 30, 'fas table-list', m.registry_rows,
      ['package', 'row_no', 'kind', 'position', 'doc_number', 'doc_date', 'qty_ks2', 'qty_project', 'qty_fact'], { hideAddButton: true, perPage: 50 }),
    card('Строка реестра', 'registry_row', m.registry_rows, 31, [
      recordBlock('Строка реестра', [0, 0, 48, 24], [
        'package', 'row_no', 'seq', 'kind', 'position', 'work_name', 'doc_name', 'doc_number', 'doc_number_norm',
        'doc_date', 'doc_date_raw', 'unit', 'qty_vrc', 'qty_total', 'qty_ks2', 'qty_project', 'qty_fact',
      ]),
    ]),

    listPage('Сводка по позициям', 'position_summary', 35, 'fas scale-balanced', m.position_summary,
      ['package', 'position', 'name', 'unit', 'registry_sum', 'registry_ks2', 'ks2_qty', 'delta', 'status'], { hideAddButton: true, perPage: 100 }),

    listPage('Юрлица (ЕГРЮЛ)', 'organizations', 40, 'fas building', m.organizations,
      ['name', 'ogrn', 'inn', 'egrul_status', 'egrul_short_name', 'egrul_region', 'checked_at']),
    card('Юрлицо', 'organization', m.organizations, 41, [
      recordBlock('Юрлицо', [0, 0, 48, 28], [
        'name', 'ogrn', 'inn', 'address', 'sro_name', 'egrul_status', 'egrul_name', 'egrul_short_name', 'egrul_inn',
        'egrul_kpp', 'egrul_region', 'egrul_reg_date', 'egrul_end_date', 'egrul_director', 'checked_at',
      ], { fieldRoles: { name: 'title', egrul_status: 'badge' } }),
    ]),
    listPage('Специалисты НРС', 'nrs_specialists', 42, 'fas id-badge', m.nrs_specialists,
      ['nrs_id', 'fio', 'status', 'status_text', 'checked_at']),
    listPage('Нормативные документы', 'norm_docs', 44, 'fas book', m.norm_docs,
      ['code', 'title', 'valid_from', 'valid_to', 'replaced_by', 'work_tags'], { perPage: 100, presort: 'code ASC' }),
    card('Нормативный документ', 'norm_doc', m.norm_docs, 45, [
      recordBlock('Нормативный документ', [0, 0, 48, 20], ['code', 'title', 'valid_from', 'valid_to', 'replaced_by', 'work_tags'],
        { fieldRoles: { code: 'title' } }),
    ]),
  ]

  const pageIDs = {}
  for (const def of pages) {
    pageIDs[def.handle] = await ensurePage(api, nsID, def)
  }
  await parentPages(api, nsID, pageIDs, { package: 'packages', act: 'acts', finding: 'findings', registry_row: 'registry_rows', organization: 'organizations', norm_doc: 'norm_docs' })

  // ---- Rule chain: кнопка «Запустить проверку ИД» → агент ----
  // Без триггера на afterUpdate: агент сам пишет прогресс в запись пакета,
  // автозапуск по обновлению зациклился бы.
  await ensureRuleChain(api, {
    id: 'stroykontrol-id-run',
    name: 'Проверка ИД: запустить проверку пакета',
    description: 'Отправляет пакет (реестр, КС-2, акты) агенту stroykontrol (idcheck). Агент выполняет проверки 1–8 и записывает акты, замечания и сводку в пространство.',
    entryNode: 'run',
    namespaceID: nsID,
    config: { triggers: [] },
    nodes: [{
      id: 'run',
      type: 'http',
      label: 'Агент: поставить пакет в очередь',
      config: {
        method: 'POST',
        url: `${AGENT_URL}/api/idcheck/run?packageID={{recordID}}&namespaceID={{namespaceID}}`,
        timeout: 30,
      },
    }],
    edges: [],
  })

  await seedNorms(api, nsID, m.norm_docs)

  writeFileSync(join(HERE, 'applied_id.json'), JSON.stringify({ namespaceID: nsID, modules: m, pages: pageIDs }, null, 2))
  console.log('done. open /ns/stroykontrol-id/pages')
}

// Seeds «Нормативные документы» from the agent's built-in catalog
// (idcheck.DefaultNorms, via `stroykontrol-web --dump-norms`) — only codes
// not yet present, so user edits are never overwritten.
async function seedNorms (api, nsID, moduleID) {
  const bin = join(HERE, '..', 'bin', 'stroykontrol-web')
  let norms
  try {
    if (!existsSync(bin)) execFileSync('go', ['build', '-o', 'bin/stroykontrol-web', '.'], { cwd: join(HERE, '..'), stdio: 'inherit' })
    norms = JSON.parse(execFileSync(bin, ['--dump-norms'], { encoding: 'utf8' }))
  } catch (e) {
    console.warn('norm seed skipped:', e.message)
    return
  }
  const existing = new Set(setOf(await api('GET', `/namespace/${nsID}/module/${moduleID}/record/?limit=1000`))
    .map(r => (r.values || []).find(v => v.name === 'code')?.value).filter(Boolean))
  const day = s => (s && !s.startsWith('0001-') ? s.slice(0, 10) : '')
  let n = 0
  for (const x of norms) {
    if (existing.has(x.code)) continue
    const values = Object.fromEntries(Object.entries({
      code: x.code,
      title: x.title,
      valid_from: day(x.valid_from),
      valid_to: day(x.valid_to),
      replaced_by: x.replaced_by,
      work_tags: (x.work_tags || []).join(', '),
    }).filter(([, v]) => v))
    await createRecord(api, nsID, moduleID, values)
    n++
  }
  console.log('norm_docs seeded:', n)
}

main().catch(e => { console.error(e); process.exit(1) })
