/**
 * Модули проверки исполнительной документации (АОСР/АООК ↔ реестр ИД ↔ КС-2 ↔ нормативка).
 * Раньше жили в отдельном пространстве «Проверка ИД (АОСР)» (apply_id.mjs); теперь
 * входят в «Стройконтроль ПТО» — ID модулей и записей при переносе не менялись
 * (см. migrate_id_into_main.sql). Сама проверка — в Go-агенте (пакет idcheck).
 */
import { existsSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  selectOptions, field, recordRel, fileField, dateField, numberField, ensureModule, createRecord, setOf,
} from './helpers.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

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

/**
 * Создаёт/обновляет модули ИД в пространстве nsID и кладёт их ID в m.
 * objectsModuleID — модуль «Объекты»: пакет ссылается на объект (object_ref).
 * dataKind — готовое поле «Тип данных» (демо/рабочие).
 */
export async function defineIdModules (api, nsID, m, { objectsModuleID, dataKind }) {
  m.packages = await ensureModule(api, nsID, {
    name: 'Пакеты проверки',
    handle: 'packages',
    fields: [
      str('title', 'Название', { required: true }),
      str('object', 'Объект (как в реестре)'),
      recordRel('object_ref', 'Объект', objectsModuleID, 'name', ['name', 'code']),
      dataKind(),
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
}

// Seeds «Нормативные документы» from the agent's built-in catalog
// (idcheck.DefaultNorms, via `stroykontrol-web --dump-norms`) — only codes
// not yet present, so user edits are never overwritten.
export async function seedNorms (api, nsID, moduleID) {
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

