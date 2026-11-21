#!/usr/bin/env node
/**
 * Provision the "Стройконтроль ПТО" Compose namespace (modules + pages)
 * for the три ТЗ из Документ2.docx:
 *   1. Сравнение ПД и РД (PDF-файлы проектной/рабочей документации)
 *   2. Проверка исполнительной документации (ИД) по реестру
 *   3. Проверка формирования ВОИСР против сметной документации
 *
 * Same pattern as agents/invest/compose/apply.mjs: idempotent, resources
 * are matched and updated by handle, existing IDs preserved.
 *
 *   COMPOSE_DSN=postgres://postgres:Zse45rdx@127.0.0.1:5432/test11?sslmode=disable \
 *   node apply.mjs
 */
import { writeFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import {
  field, recordRel, fileField, dateField, moneyField, numberField, boolSwitch,
  selectOptions, mintToken, detectBase, apiFactory, ensureNamespace,
  ensureModule, ensurePage, parentPages, withRevisions, ensureRuleChain, setOf,
} from './helpers.mjs'
import { buildRuleChains } from './chains.mjs'
import { buildPages } from './pages.mjs'
import { defineIdModules, seedNorms } from './modules_id.mjs'

const HERE = dirname(fileURLToPath(import.meta.url))

const OBJECT_STATUS = [
  ['active', 'Строится', { backgroundColor: 'success', textColor: 'white' }],
  ['suspended', 'Приостановлен', { backgroundColor: 'warning', textColor: 'dark' }],
  ['closed', 'Сдан', { backgroundColor: 'info', textColor: 'white' }],
]

const COMPARISON_STATUS = [
  ['new', 'Новое', { backgroundColor: 'light', textColor: 'dark' }],
  ['processing', 'В обработке', { backgroundColor: 'warning', textColor: 'dark' }],
  ['done', 'Готово', { backgroundColor: 'success', textColor: 'white' }],
  ['failed', 'Ошибка', { backgroundColor: 'danger', textColor: 'white' }],
]

const DISCREPANCY_TYPE = [
  ['content', 'Содержание'],
  ['numbering', 'Нумерация листов'],
  ['missing_page', 'Отсутствует лист'],
  ['extra_page', 'Лишний лист'],
  ['drawing', 'Чертёж/схема'],
  ['format', 'Оформление'],
]

const SEVERITY = [
  ['low', 'Низкая', { backgroundColor: 'light', textColor: 'dark' }],
  ['medium', 'Средняя', { backgroundColor: 'warning', textColor: 'dark' }],
  ['high', 'Высокая', { backgroundColor: 'danger', textColor: 'white' }],
]

const CMP_SIDE = [
  ['ref', 'ПД (стадия П, эталон)', { backgroundColor: 'primary', textColor: 'white' }],
  ['target', 'РД (стадия Р/РД)', { backgroundColor: 'info', textColor: 'white' }],
]

const COMPARISON_KIND = [
  ['stages', 'ПД → РД (стадии)'],
  ['versions', 'Версии: до → после'],
]

const VOLUME_SOURCE = [
  ['manual', 'Вручную'],
  ['contents', 'Из содержания документов'],
  ['set', 'Из первого комплекта'],
]

// стадия комплекта: те же коды, что у документов сравнения (ref = ПД, target = РД)
const SET_STAGE = [
  ['ref', 'ПД', { backgroundColor: 'primary', textColor: 'white' }],
  ['target', 'РД', { backgroundColor: 'info', textColor: 'white' }],
]

const SET_STATUS = [
  ['draft', 'Черновик', { backgroundColor: 'light', textColor: 'dark' }],
  ['accepted', 'Принят', { backgroundColor: 'success', textColor: 'white' }],
  ['superseded', 'Заменён новой версией', { backgroundColor: 'secondary', textColor: 'white' }],
]

// рабочие данные / демо: демо скрыты из рабочих списков и метрик (см. pages.mjs, REAL)
const DATA_KIND = [
  ['real', 'Рабочие данные'],
  ['demo', 'Демо'],
]

const DETECTED_BY = [
  ['auto', 'Автоматически'],
  ['manual', 'Вручную'],
]

const DISCREPANCY_KIND = [
  ['text', 'Текст'],
  ['table_row', 'Строка таблицы'],
  ['table_cell', 'Ячейка таблицы'],
  ['param', 'Ключевой параметр'],
  ['drawing', 'Чертёж'],
  ['vor', 'ВОР / спецификация'],
  ['not_released', 'Не выпущено (РД)'],
  ['scope_diff', 'Вне объёма (ПД)'],
  ['format', 'Оформление'],
  ['cross_volume', 'Согласованность между томами'],
  ['element', 'Марки элементов'],
  ['change_reg', 'Регистрация изменений'],
]

const ITER_STATE = [
  ['new', 'Новое', { backgroundColor: 'warning', textColor: 'dark' }],
  ['persisting', 'Не устранено', { backgroundColor: 'danger', textColor: 'white' }],
  ['resolved', 'Устранено', { backgroundColor: 'success', textColor: 'white' }],
]

const EXPERT_VERDICT = [
  ['still_differs', 'Расхождение сохраняется', { backgroundColor: 'danger', textColor: 'white' }],
  ['resolved', 'Устранено в этой итерации', { backgroundColor: 'success', textColor: 'white' }],
  ['no_match', 'Связанных расхождений не найдено', { backgroundColor: 'light', textColor: 'dark' }],
  ['no_documents', 'Тома не загружены', { backgroundColor: 'secondary', textColor: 'white' }],
]

// Единственный рабочий статус замечания: формулировки — как в выгрузке экспертизы.
const NOTIFY_EVENT = [
  ['assigned', 'Назначение'],
  ['status', 'Смена статуса'],
  ['due_soon', 'Срок приближается'],
  ['overdue', 'Срок нарушен'],
]

const VOLUME_UNIT_KIND = [
  ['volume', 'Том'],
  ['section', 'Раздел'],
  ['appendix', 'Приложение'],
  ['reference', 'Ссылочный документ'],
]
const STRUCTURE_KIND = [
  ['structure', 'Сооружение'],
  ['section', 'Участок'],
  ['network', 'Инженерная сеть / переустройство'],
]
const UPLOAD_KIND = [
  ['new', 'Новый том'],
  ['edition', 'Новая редакция тома'],
  ['repeat', 'Повтор уже загруженного файла'],
  ['unparsed', 'Шифр тома не прочитан'],
]
const NOTIFY_ROLE = [
  ['assignee', 'Исполнитель'],
  ['manager', 'Руководитель проекта'],
]

const NOTIFY_DELIVERY = [
  ['none', 'Только лента'],
  ['pending', 'К отправке'],
  ['sent', 'Отправлено'],
  ['failed', 'Ошибка'],
]

const REMARK_WORKFLOW = [
  ['awaiting_answer', 'Ответ не предоставлен', { backgroundColor: 'secondary', textColor: 'white' }],
  ['being_fixed', 'Устраняется заявителем', { backgroundColor: 'warning', textColor: 'dark' }],
  ['answer_review', 'Требуется рассмотрение ответа', { backgroundColor: 'info', textColor: 'white' }],
  ['rework', 'Возвращено на доработку', { backgroundColor: 'danger', textColor: 'white' }],
  ['resolved', 'Устранено', { backgroundColor: 'success', textColor: 'white' }],
]

// результат проверки устранения замечания (ТЗ 3.4.6): итог решения специалиста; вывод системы считается агентом
const OUTCOME = [
  ['fixed', 'Устранено', { backgroundColor: 'success', textColor: 'white' }],
  ['partial', 'Частично устранено', { backgroundColor: 'warning', textColor: 'dark' }],
  ['not_fixed', 'Не устранено', { backgroundColor: 'danger', textColor: 'white' }],
  ['check', 'Требует проверки', { backgroundColor: 'secondary', textColor: 'white' }],
]
const OUTCOME_DECISION = [
  ['confirmed', 'Вывод системы подтверждён'],
  ['changed', 'Результат изменён специалистом'],
]

// статусы задачи (ТЗ 3.7.1); переходы и права проверяет агент (cmp/task.go)
const TASK_STATUS = [
  ['open', 'Открыта', { backgroundColor: 'secondary', textColor: 'white' }],
  ['in_progress', 'В работе', { backgroundColor: 'warning', textColor: 'dark' }],
  ['review', 'На проверке', { backgroundColor: 'info', textColor: 'white' }],
  ['returned', 'Возвращена на доработку', { backgroundColor: 'danger', textColor: 'white' }],
  ['done', 'Выполнена', { backgroundColor: 'success', textColor: 'white' }],
  ['cancelled', 'Отменена', { backgroundColor: 'light', textColor: 'dark' }],
]

const REMARK_ORIGIN = [
  ['expertise', 'Выгрузка экспертизы'],
  ['comparison', 'Сформировано по сравнению'],
]

const AI_VERDICT = [
  ['fixed', 'ИИ: устранено', { backgroundColor: 'success', textColor: 'white' }],
  ['partial', 'ИИ: частично', { backgroundColor: 'warning', textColor: 'dark' }],
  ['not_fixed', 'ИИ: не устранено', { backgroundColor: 'danger', textColor: 'white' }],
  ['unclear', 'ИИ: судить нельзя', { backgroundColor: 'light', textColor: 'dark' }],
]

// наличие материалов, заявленных в ответе проектировщика (ТЗ 3.4.9): только есть / нет, без оценки содержания
const MATERIALS_STATE = [
  ['ok', 'Материалы найдены', { backgroundColor: 'success', textColor: 'white' }],
  ['missing', 'Материалы не найдены', { backgroundColor: 'danger', textColor: 'white' }],
  ['unloaded', 'Тома не загружены', { backgroundColor: 'secondary', textColor: 'white' }],
]

const CRITICALITY = [
  ['critical', 'Критичное', { backgroundColor: 'danger', textColor: 'white' }],
  ['noncritical', 'Некритичное', { backgroundColor: 'light', textColor: 'dark' }],
]

const REVIEW_STATUS = [
  ['new', 'Новое', { backgroundColor: 'light', textColor: 'dark' }],
  ['confirmed', 'Подтверждено', { backgroundColor: 'danger', textColor: 'white' }],
  ['rejected', 'Отклонено', { backgroundColor: 'secondary', textColor: 'white' }],
  ['clarify', 'Требует уточнения', { backgroundColor: 'warning', textColor: 'dark' }],
]

const REMARK_STATUS = [
  ['draft', 'Черновик', { backgroundColor: 'light', textColor: 'dark' }],
  ['issued', 'Выдано', { backgroundColor: 'warning', textColor: 'dark' }],
  ['closed', 'Закрыто', { backgroundColor: 'success', textColor: 'white' }],
]

const DOC_TYPE_ID = [
  ['act', 'Акт'],
  ['journal', 'Журнал работ'],
  ['certificate', 'Сертификат/паспорт'],
  ['scheme', 'Исполнительная схема'],
  ['other', 'Иное'],
]

const MATCH_STATUS = [
  ['matched', 'Совпадает', { backgroundColor: 'success', textColor: 'white' }],
  ['name_mismatch', 'Несовпадение наименования', { backgroundColor: 'warning', textColor: 'dark' }],
  ['number_mismatch', 'Несовпадение номера', { backgroundColor: 'warning', textColor: 'dark' }],
  ['version_mismatch', 'Несовпадение версии', { backgroundColor: 'warning', textColor: 'dark' }],
  ['not_in_registry', 'Не найден в реестре', { backgroundColor: 'danger', textColor: 'white' }],
]

const RUN_STATUS = [
  ['success', 'Успешно', { backgroundColor: 'success', textColor: 'white' }],
  ['fail', 'Не успешно', { backgroundColor: 'danger', textColor: 'white' }],
]

const VOISR_CHECK_STATUS = [
  ['match', 'Совпадает', { backgroundColor: 'success', textColor: 'white' }],
  ['mismatch', 'Не совпадает', { backgroundColor: 'danger', textColor: 'white' }],
  ['not_found', 'Смета не найдена', { backgroundColor: 'secondary', textColor: 'white' }],
]

const dataKind = () => field('data_kind', 'Тип данных', 'Select', { options: selectOptions(DATA_KIND) })

function textArea (name, label, extra = {}) {
  return field(name, label, 'String', { ...extra, options: { multiLine: true, ...(extra.options || {}) } })
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  console.log('API', base)

  const nsID = await ensureNamespace(api, {
    name: 'Стройконтроль ПТО',
    slug: 'stroykontrol',
    meta: {
      subtitle: 'Проверка документации ПТО и сметного отдела',
      description: 'Сравнение ПД/РД, проверка комплектности ИД по реестру, проверка формирования ВОИСР против смет',
    },
  })
  console.log('namespace', nsID)

  const m = {}

  // ---- НСИ: объекты строительства (общий контекст для всех трёх задач) ----
  m.objects = await ensureModule(api, nsID, {
    name: 'Объекты',
    handle: 'objects',
    fields: [
      field('name', 'Наименование', 'String', { required: true }),
      field('address', 'Адрес', 'String'),
      field('code', 'Код объекта', 'String'),
      field('status', 'Статус', 'Select', { options: selectOptions(OBJECT_STATUS) }),
      textArea('notes', 'Примечания'),
      // выгрузки замечаний экспертизы (Excel): агент читает их и ведёт реестр замечаний объекта
      fileField('remarks_file', 'Выгрузки замечаний экспертизы (Excel или PDF)', { multi: true, images: false }),
      dataKind(),
    ],
  })

  // ---- Сооружения и участки объекта (ТЗ 3.1.7): справочник, к которому привязываются документы ----
  m.structures = await ensureModule(api, nsID, {
    name: 'Сооружения и участки',
    handle: 'structures',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('name', 'Наименование', 'String', { required: true }),
      field('code', 'Шифр / обозначение', 'String'),
      field('kind', 'Вид', 'Select', { options: selectOptions(STRUCTURE_KIND) }),
      field('location', 'Привязка (ПК, км)', 'String'),
      textArea('note', 'Примечание'),
    ],
  })

  // ---- Комплекты документов объекта: ПД / РД по версиям + ведомость состава томов ----
  // Документы (cmp_documents) принадлежат комплекту; сравнение ссылается на два комплекта —
  // так всегда видно, что у объекта есть, насколько полно и какая версия с какой сравнивается.
  m.document_sets = await ensureModule(api, nsID, {
    name: 'Комплекты документов',
    handle: 'document_sets',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('stage', 'Стадия', 'Select', { required: true, options: selectOptions(SET_STAGE) }),
      field('version', 'Версия', 'String', { required: true }),
      dateField('issued_at', 'Дата выпуска', { onlyDate: true }),
      field('status', 'Статус', 'Select', { options: selectOptions(SET_STATUS) }),
      textArea('comment', 'Комментарий'),
      // считает агент по ведомости томов («Проверить комплектность»)
      numberField('volumes_total', 'Томов по ведомости', { precision: 0 }),
      numberField('volumes_present', 'Томов в комплекте', { precision: 0 }),
      numberField('completeness_pct', 'Комплектность, %', { precision: 0, max: 100, suffix: '%' }),
      textArea('completeness_note', 'Что не хватает / лишнее'),
      dataKind(),
    ],
  })

  m.object_volumes = await ensureModule(api, nsID, {
    name: 'Ведомость состава томов',
    handle: 'object_volumes',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('stage', 'Стадия', 'Select', { options: selectOptions(SET_STAGE) }),
      field('volume_code', 'Шифр тома', 'String', { required: true }),
      field('title', 'Наименование', 'String'),
      boolSwitch('required', 'Обязателен'),
      // кроме томов: обязательный раздел (марка «КЖ»), приложение («Приложение Б», в каком томе — parent_code)
      // и ссылочный документ; без вида строка — том (ТЗ 3.2.1, 3.2.8)
      field('kind', 'Вид позиции', 'Select', { options: selectOptions(VOLUME_UNIT_KIND) }),
      field('parent_code', 'В томе (для приложения)', 'String'),
      // автоматическая ведомость: агент предлагает строки (proposed), человек принимает или отклоняет;
      // на комплектность влияют только принятые
      field('source', 'Источник', 'Select', { options: selectOptions(VOLUME_SOURCE) }),
      boolSwitch('proposed', 'Предложено агентом (не принято)'),
      field('evidence', 'Откуда взято', 'String'),
    ],
  })

  // ---- ТЗ №1: Сравнение ПД и РД ----
  // the module refers to itself (parent_comparison): its own ID is needed for the field
  const existingCmp = setOf(await api('GET', `/namespace/${nsID}/module/?handle=pd_rd_comparisons&limit=5`)).find(x => x.handle === 'pd_rd_comparisons' && !x.deletedAt)
  const comparisonFields = (selfID) => [
    field('title', 'Название', 'String', { required: true }),
    recordRel('object', 'Объект', m.objects, 'name', ['name', 'code']),
    fileField('pd_file', 'Файл ПД'),
    fileField('rd_file', 'Файл РД'),
    field('status', 'Статус', 'Select', { options: selectOptions(COMPARISON_STATUS) }),
    numberField('similarity_percent', 'Схожесть, %', { precision: 1, max: 100, suffix: '%' }),
    numberField('total_pages_pd', 'Страниц в ПД', { precision: 0 }),
    numberField('total_pages_rd', 'Страниц в РД', { precision: 0 }),
    numberField('matching_pages', 'Совпадающих страниц', { precision: 0 }),
    numberField('differing_pages', 'Различающихся страниц', { precision: 0 }),
    fileField('report_file', 'Отчёт о сравнении'),
    textArea('comment', 'Комментарий'),
    // какая версия комплекта с какой сравнивается
    field('kind', 'Вид сравнения', 'Select', { options: selectOptions(COMPARISON_KIND) }),
    recordRel('set_pd', 'Эталон: комплект ПД (или версия «до»)', m.document_sets, 'version', ['version', 'comment']),
    recordRel('set_rd', 'Проверяемый: комплект РД (или версия «после»)', m.document_sets, 'version', ['version', 'comment']),
    field('sets_label', 'Версии комплектов', 'String'),
    // ручное сопоставление томов и фрагментов (JSON, правится в просмотрщике) и результат последнего сопоставления
    textArea('pairing', 'Ручное сопоставление (JSON)'),
    // области листов, исключённые из сравнения (JSON: сторона, файл, страница, рамка в долях страницы); исходные файлы не меняются
    textArea('excluded_regions', 'Исключаемые области (JSON)'),
    textArea('pairs_json', 'Сопоставление томов (JSON)'),
    dataKind(),
    // итерации: повторная проверка той же пары после исправлений
    numberField('iteration_no', 'Итерация', { precision: 0 }),
    ...(selfID ? [recordRel('parent_comparison', 'Предыдущая итерация', selfID, 'title', ['title'])] : []),
  ]
  m.pd_rd_comparisons = await ensureModule(api, nsID, {
    name: 'Сравнения ПД/РД',
    handle: 'pd_rd_comparisons',
    fields: comparisonFields(existingCmp?.moduleID),
    config: withRevisions(),
  })
  if (!existingCmp) {
    // a brand-new module only now has an ID to point parent_comparison at
    await ensureModule(api, nsID, {
      name: 'Сравнения ПД/РД',
      handle: 'pd_rd_comparisons',
      fields: comparisonFields(m.pd_rd_comparisons),
      config: withRevisions(),
    })
  }

  // Комплект документов сравнения: любое число PDF/DWG/DXF с каждой стороны
  // (один файл на стадию, пачка против пачки, пачка против одного файла).
  // Том/шифр определяются агентом по основной надписи; поля можно поправить
  // вручную (detected_by = manual — агент их больше не перезаписывает).
  // .dwg обязателен в списке: клиентский загрузчик не знает MIME DWG, а сервер
  // сверяет расширение, даже если файл определился как application/octet-stream.
  const drawingMimes = [
    'application/pdf', '.pdf',
    '.dwg', 'image/vnd.dwg', 'application/acad', 'application/x-dwg', 'application/dwg',
    '.dxf', 'image/vnd.dxf', 'application/dxf', 'application/x-dxf',
  ].join(',')
  m.cmp_documents = await ensureModule(api, nsID, {
    name: 'Документы сравнения',
    handle: 'cmp_documents',
    fields: [
      recordRel('set', 'Комплект', m.document_sets, 'version', ['version', 'stage']),
      recordRel('comparison', 'Сравнение (если файл приложен прямо к нему)', m.pd_rd_comparisons, 'title', ['title']),
      field('side', 'Стадия', 'Select', { required: true, options: selectOptions(CMP_SIDE) }),
      fileField('file', 'Файл PDF или DWG', { documents: true, images: false, mimetypes: drawingMimes }),
      field('doc_version', 'Версия (изм.)', 'String'),
      field('volume_code', 'Шифр тома', 'String'),
      field('volume_title', 'Наименование тома', 'String'),
      field('object_name', 'Объект/сооружение', 'String'),
      numberField('pages', 'Страниц', { precision: 0 }),
      // номера листов по основной надписи: «страница:лист,…» (считает агент; по ним замечание «лист N» открывается на нужной странице)
      field('sheet_map', 'Листы по основной надписи', 'String'),
      // приложения и документы (расчёты, записки, письма…) с их страницами: здесь ищутся материалы, заявленные в ответах (JSON, считает агент)
      textArea('headings_json', 'Приложения и документы в файле (JSON)'),
      // ссылки на другие тома проекта из текста (ТЗ 3.2.4) и тома, перечисленные в содержании файла (ТЗ 3.2.5): считает агент
      textArea('volume_refs_json', 'Ссылки на другие тома (JSON)'),
      textArea('contents_json', 'Состав по содержанию файла (JSON)'),
      // редакция и классификация: агент читает номер изменения («Изм. 2») с титульного листа или из имени файла,
      // часть/книгу/подраздел — с титульных листов; значения, введённые вручную, не перезаписываются, пока файл тот же
      numberField('change_no', 'Номер изменения (редакция)', { precision: 0 }),
      field('structure_object', 'Сооружение / участок', 'String'),
      // раздел — марка из шифра тома (пишет агент, ручное значение не перезаписывает); сооружение — запись справочника
      // (агент связывает по наименованию или шифру сооружения в поле выше и в наименовании тома)
      field('section', 'Раздел', 'String'),
      recordRel('structure', 'Сооружение (справочник)', m.structures, 'name', ['name', 'code']),
      field('subsection', 'Подраздел', 'String'),
      field('part', 'Часть', 'String'),
      field('book', 'Книга', 'String'),
      field('file_hash', 'Хэш файла (версия загруженного PDF)', 'String'),
      boolSwitch('released', 'Выпущен'),
      field('detected_by', 'Определено', 'Select', { options: selectOptions(DETECTED_BY) }),
    ],
  })

  m.pd_rd_discrepancies = await ensureModule(api, nsID, {
    name: 'Расхождения ПД/РД',
    handle: 'pd_rd_discrepancies',
    fields: [
      recordRel('comparison', 'Сравнение', m.pd_rd_comparisons, 'title', ['title'], true),
      numberField('page_number', 'Номер страницы', { precision: 0 }),
      field('discrepancy_type', 'Тип расхождения', 'Select', { options: selectOptions(DISCREPANCY_TYPE) }),
      field('severity', 'Критичность', 'Select', { options: selectOptions(SEVERITY) }),
      textArea('description', 'Описание'),
      field('kind', 'Вид', 'Select', { options: selectOptions(DISCREPANCY_KIND) }),
      field('status', 'Решение', 'Select', { options: selectOptions(REVIEW_STATUS) }),
      textArea('reviewer_comment', 'Комментарий к решению'),
      field('doc_ref', 'Документ ПД (ID)', 'String'),
      field('doc_target', 'Документ РД (ID)', 'String'),
      numberField('page_pd', 'Страница в ПД', { precision: 0 }),
      numberField('page_rd', 'Страница в РД', { precision: 0 }),
      field('bbox_pd', 'Область в ПД', 'String'),
      field('bbox_rd', 'Область в РД', 'String'),
      textArea('anchor_text', 'Фрагмент'),
      textArea('evidence', 'Обоснование (JSON)'),
      field('fingerprint', 'Отпечаток', 'String'),
      field('iter_state', 'Итерация', 'Select', { options: selectOptions(ITER_STATE) }),
      dataKind(),
    ],
  })

  m.remark_cards = await ensureModule(api, nsID, {
    name: 'Карточки замечаний',
    handle: 'remark_cards',
    fields: [
      recordRel('comparison', 'Сравнение', m.pd_rd_comparisons, 'title', ['title'], true),
      recordRel('discrepancies', 'Расхождения', m.pd_rd_discrepancies, 'description', ['description'], false, { multi: true }),
      field('document_ref', 'Документы', 'String'),
      field('sheet', 'Листы', 'String'),
      textArea('what_diverges', 'В чём расхождение'),
      textArea('must_align', 'Что требуется'),
      textArea('draft_text', 'Черновик формулировки (ИИ)'),
      textArea('final_text', 'Текст замечания'),
      field('status', 'Статус', 'Select', { options: selectOptions(REMARK_STATUS) }),
    ],
  })

  // Реестр замечаний экспертизы по объекту: агент ведёт его по выгрузкам Excel,
  // хранит историю итераций и проверяет ответы по загруженным документам.
  m.expertise_remarks = await ensureModule(api, nsID, {
    name: 'Замечания экспертизы',
    handle: 'expertise_remarks',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('remark_key', 'Ключ в объекте', 'String'),
      field('number', '№ замечания', 'String'),
      field('batch', 'Письмо', 'String'),
      textArea('text', 'Вывод о несоответствии'),
      textArea('refs_text', 'Ссылка на материалы'),
      textArea('volume_refs', 'Тома (JSON)'),
      // куда ведёт замечание: места по ссылке на материалы и области сопоставленных находок (JSON, считает агент)
      textArea('links_json', 'Места в документах (JSON)'),
      // требования замечания: список, каждое с проверкой по документам, разбором ИИ, результатом и решением специалиста (JSON, ведёт агент)
      textArea('requirements_json', 'Требования замечания (JSON)'),
      // материалы, заявленные в ответе (тома, листы, приложения, расчёты…), и найдены ли они (JSON, считает агент; ТЗ 3.4.9)
      textArea('materials_json', 'Заявленные материалы (JSON)'),
      field('materials_state', 'Заявленные материалы', 'Select', { options: selectOptions(MATERIALS_STATE) }),
      // ответы, расчёты, записки, приложения, пришедшие отдельными файлами: среди них агент тоже ищет заявленное
      fileField('materials_files', 'Обосновывающие материалы', { multi: true, images: false }),
      field('section', 'Раздел документации', 'String'),
      textArea('basis', 'Основание'),
      textArea('answer', 'Ответ на замечание'),
      textArea('reject_reason', 'Причина отклонения'),
      textArea('prev_answer', 'Предыдущий ответ'),
      textArea('prev_reject_reason', 'Предыдущая причина отклонения'),
      field('discipline', 'НД', 'String'),
      field('expert', 'Эксперт-соисполнитель', 'String'),
      field('criticality', 'Критичность', 'Select', { options: selectOptions(CRITICALITY) }),
      textArea('proposals', 'Предложения по отработке'),
      textArea('expected_result', 'Ожидаемый результат'),
      numberField('iterations', 'Итераций', { precision: 0 }),
      dateField('source_changed_at', 'Изменено в экспертизе', { onlyDate: false }),
      field('source_status', 'Статус в выгрузке', 'String'),
      // единственный рабочий статус: правится вручную, выгрузка меняет его, только если решение изменилось на стороне экспертизы
      field('status', 'Статус', 'Select', { options: selectOptions(REMARK_WORKFLOW) }),
      field('assignee', 'Ответственный', 'String'),
      dateField('due_date', 'Срок', { onlyDate: true }),
      textArea('pto_note', 'Заметка ПТО'),
      numberField('estimate_h', 'Оценка трудоёмкости, ч', { precision: 1 }),
      field('verdict', 'Проверка по документам', 'Select', { options: selectOptions(EXPERT_VERDICT) }),
      numberField('match_score', 'Уверенность, %', { precision: 0 }),
      textArea('match_note', 'Пояснение проверки'),
      field('claim_mismatch', 'Ответ ≠ документы', 'String'),
      recordRel('matched_discrepancies', 'Связанные расхождения', m.pd_rd_discrepancies, 'description', ['description'], false, { multi: true }),
      recordRel('linked_documents', 'Связанные тома', m.cmp_documents, 'volume_code', ['volume_code', 'volume_title'], false, { multi: true }),
      // разбор ИИ по фрагментам связанных томов: подсказка, рабочий статус не меняет
      field('ai_verdict', 'Разбор ИИ', 'Select', { options: selectOptions(AI_VERDICT) }),
      textArea('ai_note', 'Пояснение ИИ (с цитатами)'),
      textArea('ai_cites_json', 'Цитаты ИИ: места в томах (JSON)'),
      dateField('ai_at', 'Разбор ИИ: когда', { onlyDate: false }),
      // решение специалиста по результату проверки (вывод системы — ИИ и проверка по документам — рабочий статус не меняет)
      field('outcome_final', 'Результат проверки (решение специалиста)', 'Select', { options: selectOptions(OUTCOME) }),
      field('outcome_decision', 'Решение по результату', 'Select', { options: selectOptions(OUTCOME_DECISION) }),
      textArea('outcome_comment', 'Комментарий к решению'),
      field('outcome_by', 'Решение: кто', 'String'),
      dateField('outcome_at', 'Решение: когда', { onlyDate: false }),
      field('archived', 'Нет в последней выгрузке', 'String'),
      field('source_file', 'Файл-источник', 'String'),
      // замечания, сформированные по сравнению ПД/РД (подтверждённые находки одного листа)
      field('origin', 'Происхождение', 'Select', { options: selectOptions(REMARK_ORIGIN) }),
      recordRel('source_comparison', 'Сравнение', m.pd_rd_comparisons, 'title', ['title']),
      field('group_key', 'Ключ листа', 'String'),
      dataKind(),
    ],
  })

  m.expertise_remark_history = await ensureModule(api, nsID, {
    name: 'История замечаний',
    handle: 'expertise_remark_history',
    fields: [
      recordRel('remark', 'Замечание', m.expertise_remarks, 'number', ['number', 'text'], true),
      numberField('iteration', 'Итерация', { precision: 0 }),
      textArea('answer', 'Ответ'),
      textArea('reject_reason', 'Причина отклонения'),
      field('status', 'Статус', 'Select', { options: selectOptions(REMARK_WORKFLOW) }),
      field('source_status', 'Статус в выгрузке', 'String'),
      dateField('at', 'Когда', { onlyDate: false }),
      field('author', 'Кто', 'String'),
      field('source', 'Источник', 'String'),
      textArea('note', 'Комментарий'),
      // вид записи: import | status | assignee | due | note | decision | auto | comparison
      field('kind', 'Вид', 'String'),
    ],
  })

  // срезы реестра замечаний по дням: динамика устранения (ТЗ 3.9.4); пишет агент, одна запись на объект и день
  m.expertise_snapshots = await ensureModule(api, nsID, {
    name: 'Срезы реестра замечаний',
    handle: 'expertise_snapshots',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('day', 'День', 'String'),
      numberField('total', 'Всего', { precision: 0 }),
      numberField('overdue', 'Просрочено', { precision: 0 }),
      numberField('n_awaiting_answer', 'Ответ не предоставлен', { precision: 0 }),
      numberField('n_being_fixed', 'Устраняется заявителем', { precision: 0 }),
      numberField('n_answer_review', 'Требуется рассмотрение ответа', { precision: 0 }),
      numberField('n_rework', 'Возвращено на доработку', { precision: 0 }),
      numberField('n_resolved', 'Устранено', { precision: 0 }),
    ],
  })

  // уведомления участников (ТЗ 3.7.5): справочник получателей и лента; события находит и пишет агент
  // задачи (ТЗ 3.7.1): у замечания их несколько, бывают и без замечания; у каждой свой исполнитель, срок, оценка и статус
  m.tasks = await ensureModule(api, nsID, {
    name: 'Задачи',
    handle: 'tasks',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      recordRel('remark', 'Замечание', m.expertise_remarks, 'number', ['number', 'text']),
      numberField('requirement', 'Требование замечания (№)', { precision: 0 }),
      field('title', 'Задача', 'String', { required: true }),
      textArea('description', 'Описание'),
      field('assignee', 'Исполнитель', 'String'),
      dateField('due_date', 'Срок', { onlyDate: true }),
      numberField('estimate_h', 'Оценка трудоёмкости, ч', { precision: 1 }),
      field('status', 'Статус', 'Select', { options: selectOptions(TASK_STATUS) }),
      textArea('comment', 'Причина последнего решения'),
      field('created_by', 'Поставил', 'String'),
      dateField('created_at', 'Создана', { onlyDate: false }),
      dateField('done_at', 'Выполнена', { onlyDate: false }),
      field('closed_by', 'Закрыл / вернул', 'String'),
      dataKind(),
    ],
  })

  m.task_history = await ensureModule(api, nsID, {
    name: 'История задач',
    handle: 'task_history',
    fields: [
      recordRel('task', 'Задача', m.tasks, 'title', ['title'], true),
      // вид записи: created | status | assignee | due | estimate | edit
      field('kind', 'Вид', 'String'),
      field('source', 'Что изменилось', 'String'),
      textArea('note', 'Комментарий'),
      field('author', 'Кто', 'String'),
      field('status', 'Статус', 'String'),
      dateField('at', 'Когда', { onlyDate: false }),
    ],
  })

  m.notify_recipients = await ensureModule(api, nsID, {
    name: 'Получатели уведомлений',
    handle: 'notify_recipients',
    fields: [
      field('name', 'ФИО (как в поле «Ответственный»)', 'String', { required: true }),
      field('email', 'E-mail', 'String'),
      field('role', 'Роль', 'Select', { options: selectOptions(NOTIFY_ROLE) }),
      recordRel('object', 'Объект (пусто — все)', m.objects, 'name', ['name', 'code']),
      field('events', 'События (пусто — по роли)', 'Select', { multi: true, options: selectOptions(NOTIFY_EVENT) }),
      boolSwitch('active', 'Активен'),
    ],
  })

  m.notifications = await ensureModule(api, nsID, {
    name: 'Уведомления',
    handle: 'notifications',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      recordRel('remark', 'Замечание', m.expertise_remarks, 'number', ['number', 'text']),
      recordRel('task', 'Задача', m.tasks, 'title', ['title']),
      field('number', '№ замечания', 'String'),
      field('event', 'Событие', 'Select', { options: selectOptions(NOTIFY_EVENT) }),
      field('recipient', 'Получатель', 'String'),
      field('email', 'E-mail', 'String'),
      field('title', 'Тема', 'String'),
      textArea('text', 'Текст'),
      field('key', 'Ключ (повтор не отправляется)', 'String'),
      dateField('created_at', 'Создано', { onlyDate: false }),
      boolSwitch('read', 'Прочитано'),
      field('delivery', 'Доставка', 'Select', { options: selectOptions(NOTIFY_DELIVERY) }),
      dateField('delivered_at', 'Отправлено', { onlyDate: false }),
      numberField('tries', 'Попыток', { precision: 0 }),
      field('error', 'Ошибка доставки', 'String'),
    ],
  })

  // журнал загрузок документации (ТЗ 3.1.4): запись на каждый файл, когда агент впервые его читает
  m.upload_log = await ensureModule(api, nsID, {
    name: 'Журнал загрузок документации',
    handle: 'upload_log',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      recordRel('set', 'Комплект', m.document_sets, 'version', ['version', 'stage']),
      recordRel('document', 'Документ', m.cmp_documents, 'volume_code', ['volume_code', 'volume_title']),
      field('file_name', 'Файл', 'String'),
      field('volume_code', 'Шифр тома', 'String'),
      numberField('pages', 'Страниц', { precision: 0 }),
      field('file_hash', 'Хэш файла', 'String'),
      field('kind', 'Что загружено', 'Select', { options: selectOptions(UPLOAD_KIND) }),
      field('note', 'Пояснение', 'String'),
      dateField('uploaded_at', 'Загружено', { onlyDate: false }),
      field('uploaded_by', 'Кто загрузил', 'String'),
      dateField('read_at', 'Прочитано агентом', { onlyDate: false }),
    ],
  })

  m.expertise_imports = await ensureModule(api, nsID, {
    name: 'Журнал выгрузок замечаний',
    handle: 'expertise_imports',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('file', 'Файл', 'String'),
      field('attachment', 'Вложение (ID)', 'String'),
      dateField('imported_at', 'Загружено', { onlyDate: false }),
      numberField('total', 'Замечаний', { precision: 0 }),
      numberField('new_count', 'Новых', { precision: 0 }),
      numberField('changed_count', 'Изменено', { precision: 0 }),
      numberField('status_changes', 'Смен статуса', { precision: 0 }),
      numberField('reopened', 'Возвращено', { precision: 0 }),
      numberField('archived_count', 'Исчезло из выгрузки', { precision: 0 }),
      textArea('details', 'Что изменилось'),
    ],
  })

  // пакет исправления: исправленная редакция, загруженная в новую итерацию сравнения, с замечаниями, к которым она относится,
  // и результатом проверки; по записи на каждый прогон хранится в history_json (пишет агент)
  m.correction_packages = await ensureModule(api, nsID, {
    name: 'Пакеты исправления',
    handle: 'correction_packages',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      recordRel('comparison', 'Сравнение (итерация)', m.pd_rd_comparisons, 'title', ['title'], true),
      recordRel('parent_comparison', 'Предыдущая итерация', m.pd_rd_comparisons, 'title', ['title']),
      field('title', 'Название', 'String'),
      numberField('iteration_no', 'Итерация', { precision: 0 }),
      recordRel('documents', 'Исправленные документы', m.cmp_documents, 'volume_code', ['volume_code', 'volume_title'], false, { multi: true }),
      recordRel('remarks', 'Замечания', m.expertise_remarks, 'number', ['number', 'text'], false, { multi: true }),
      numberField('n_documents', 'Документов', { precision: 0 }),
      numberField('n_remarks', 'Замечаний', { precision: 0 }),
      numberField('resolved', 'Устранено', { precision: 0 }),
      numberField('persisting', 'Осталось', { precision: 0 }),
      numberField('new_findings', 'Новых', { precision: 0 }),
      numberField('runs', 'Проверок', { precision: 0 }),
      dateField('last_run', 'Последняя проверка', { onlyDate: false }),
      textArea('history_json', 'История проверок (JSON)'),
      // ответы, расчёты, пояснительные записки, приложения пакета отдельными файлами (ТЗ 3.4.11)
      fileField('materials_files', 'Обосновывающие материалы', { multi: true, images: false }),
      dataKind(),
    ],
  })

  // ---- ТЗ №2: Проверка ИД по реестру ----
  m.id_registry = await ensureModule(api, nsID, {
    name: 'Реестр ИД',
    handle: 'id_registry',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('doc_name', 'Наименование документа', 'String', { required: true }),
      field('doc_number', 'Номер документа', 'String'),
      field('doc_version', 'Версия', 'String'),
      field('doc_type', 'Тип документа', 'Select', { options: selectOptions(DOC_TYPE_ID) }),
      boolSwitch('is_required', 'Обязателен'),
    ],
  })

  m.id_files = await ensureModule(api, nsID, {
    name: 'Файлы ИД',
    handle: 'id_files',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      fileField('file', 'Файл'),
      textArea('extracted_text', 'Извлечённый текст'),
      field('recognized_name', 'Распознанное наименование', 'String'),
      field('recognized_number', 'Распознанный номер', 'String'),
      field('recognized_version', 'Распознанная версия', 'String'),
      recordRel('matched_registry', 'Позиция реестра', m.id_registry, 'doc_name', ['doc_name', 'doc_number']),
      field('match_status', 'Статус сопоставления', 'Select', { options: selectOptions(MATCH_STATUS) }),
      dateField('uploaded_at', 'Загружен', { onlyDate: false }),
    ],
  })

  m.id_check_runs = await ensureModule(api, nsID, {
    name: 'Проверки комплектности ИД',
    handle: 'id_check_runs',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      dateField('run_date', 'Дата проверки', { onlyDate: false }),
      numberField('total_registry_docs', 'Всего в реестре', { precision: 0 }),
      numberField('uploaded_docs', 'Загружено', { precision: 0 }),
      numberField('missing_docs', 'Отсутствует', { precision: 0 }),
      numberField('mismatched_docs', 'С несоответствиями', { precision: 0 }),
      field('overall_status', 'Итоговый статус', 'Select', { options: selectOptions(RUN_STATUS) }),
      fileField('report_file', 'Отчёт'),
    ],
  })

  // ---- ТЗ №3: Проверка ВОИСР ----
  m.smeta_items = await ensureModule(api, nsID, {
    name: 'Позиции смет',
    handle: 'smeta_items',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('smeta_number', 'Номер сметы', 'String', { required: true }),
      field('item_number', 'Пункт сметы', 'String'),
      field('work_name', 'Наименование работы', 'String', { required: true }),
      field('unit', 'Ед. измерения', 'String'),
      numberField('volume', 'Объём', { precision: 3 }),
      moneyField('unit_price', 'Цена за ед.'),
      moneyField('total_amount', 'Сумма'),
    ],
  })

  m.voisr_items = await ensureModule(api, nsID, {
    name: 'Позиции ВОИСР',
    handle: 'voisr_items',
    fields: [
      recordRel('object', 'Объект', m.objects, 'name', ['name', 'code'], true),
      field('position_number', 'Номер позиции ВОИСР', 'String', { required: true }),
      recordRel('smeta_ref', 'Смета/пункт', m.smeta_items, 'work_name', ['smeta_number', 'item_number', 'work_name']),
      field('work_name', 'Наименование работы', 'String', { required: true }),
      numberField('volume', 'Объём по ВОИСР', { precision: 3 }),
      moneyField('amount', 'Стоимость по ВОИСР'),
    ],
  })

  m.voisr_check_results = await ensureModule(api, nsID, {
    name: 'Результаты проверки ВОИСР',
    handle: 'voisr_check_results',
    fields: [
      recordRel('voisr_item', 'Позиция ВОИСР', m.voisr_items, 'position_number', ['position_number', 'work_name'], true),
      numberField('smeta_volume', 'Объём по смете', { precision: 3 }),
      moneyField('smeta_amount', 'Сумма по смете'),
      numberField('volume_diff', 'Разница по объёму', { precision: 3 }),
      moneyField('amount_diff', 'Разница по сумме'),
      numberField('amount_diff_percent', 'Разница, %', { precision: 1, suffix: '%' }),
      field('status', 'Статус', 'Select', { options: selectOptions(VOISR_CHECK_STATUS) }),
      textArea('note', 'Примечание'),
      dataKind(),
    ],
  })

  // ---- Проверка исполнительной документации (бывшее пространство «Проверка ИД (АОСР)») ----
  await defineIdModules(api, nsID, m, { objectsModuleID: m.objects, dataKind })

  console.log('modules', m)

  // ---- Pages: меню «от объекта», карточка объекта — рабочее место с вкладками ----
  const { defs, parentOf } = buildPages(m)
  const pageIDs = {}
  for (const def of defs) {
    // группы создаются раньше детей (порядок в defs), у детей группа — selfID
    const withParent = def.parent && pageIDs[def.parent] ? { ...def, selfID: String(pageIDs[def.parent]) } : def
    pageIDs[def.handle] = await ensurePage(api, nsID, withParent)
  }
  // карточка записи — дочерняя страница своего списка (в меню не видна)
  await parentPages(api, nsID, pageIDs, parentOf)
  await hideStalePages(api, nsID, pageIDs)

  // ---- Rule chains: агент (script det + ai LLM, вес — в узле combine) ----
  for (const chain of buildRuleChains({ nsID, modules: m, agentURL: (process.env.STROYKONTROL_AGENT_URL || 'http://localhost:8092').replace(/\/$/, '') })) {
    await ensureRuleChain(api, chain)
  }

  await seedNorms(api, nsID, m.norm_docs)

  writeFileSync(join(HERE, 'applied.json'), JSON.stringify({ namespaceID: nsID, modules: m, pages: pageIDs }, null, 2))
  console.log('done. open /ns/stroykontrol (Главная)')
}

// Страницы, которых больше нет в схеме (прежнее меню), не удаляем, а прячем из меню и
// переносим в корень: данные и ссылки остаются, но интерфейс не засоряется.
async function hideStalePages (api, nsID, pageIDs) {
  const keep = new Set(Object.values(pageIDs).map(String))
  const all = setOf(await api('GET', `/namespace/${nsID}/page/?limit=500`))
  for (const p of all) {
    if (p.deletedAt || keep.has(String(p.pageID)) || !p.visible) continue
    const rec = await api('GET', `/namespace/${nsID}/page/${p.pageID}`)
    await api('POST', `/namespace/${nsID}/page/${p.pageID}`, { ...rec, visible: false, updatedAt: rec.updatedAt })
    console.log('hidden stale page', p.handle)
  }
}

main().catch(e => { console.error(e); process.exit(1) })
