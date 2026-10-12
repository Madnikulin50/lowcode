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
  selectOptions, block, recordList, recordBlock, metricBlock, metricItem,
  pageIcon, withBlockIDs, mintToken, detectBase, apiFactory, ensureNamespace,
  ensureModule, ensurePage, withRevisions, ensureRuleChain, ruleChain, setOf,
} from './helpers.mjs'
import { buildRuleChains } from './chains.mjs'

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

const DETECTED_BY = [
  ['auto', 'Автоматически'],
  ['manual', 'Вручную'],
]

const DISCREPANCY_KIND = [
  ['text', 'Текст'],
  ['table_row', 'Строка таблицы'],
  ['param', 'Ключевой параметр'],
  ['drawing', 'Чертёж'],
  ['vor', 'ВОР / спецификация'],
  ['not_released', 'Не выпущено (РД)'],
  ['scope_diff', 'Вне объёма (ПД)'],
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
]

const EXPERT_STATUS = [
  ['open', 'Открыто', { backgroundColor: 'warning', textColor: 'dark' }],
  ['closed', 'Закрыто', { backgroundColor: 'success', textColor: 'white' }],
  ['disputed', 'Оспорено', { backgroundColor: 'secondary', textColor: 'white' }],
]

const REVIEW_STATUS = [
  ['new', 'Новое', { backgroundColor: 'light', textColor: 'dark' }],
  ['confirmed', 'Подтверждено', { backgroundColor: 'danger', textColor: 'white' }],
  ['rejected', 'Отклонено', { backgroundColor: 'secondary', textColor: 'white' }],
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
    // итерации: повторная проверка той же пары после исправлений
    numberField('iteration_no', 'Итерация', { precision: 0 }),
    // замечания экспертизы: Excel с замечаниями и (отдельным файлом) с ответами проектировщика
    fileField('remarks_file', 'Замечания экспертизы (Excel)', { multi: true, images: false }),
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

  // Комплект документов сравнения: любое число PDF с каждой стороны
  // (один PDF на стадию, пачка против пачки, пачка против одного файла).
  // Том/шифр определяются агентом по основной надписи; поля можно поправить
  // вручную (detected_by = manual — агент их больше не перезаписывает).
  m.cmp_documents = await ensureModule(api, nsID, {
    name: 'Документы сравнения',
    handle: 'cmp_documents',
    fields: [
      recordRel('comparison', 'Сравнение', m.pd_rd_comparisons, 'title', ['title'], true),
      field('side', 'Стадия', 'Select', { required: true, options: selectOptions(CMP_SIDE) }),
      fileField('file', 'Файл PDF', { documents: true, images: false, mimetypes: 'application/pdf' }),
      field('doc_version', 'Версия (изм.)', 'String'),
      field('volume_code', 'Шифр тома', 'String'),
      field('volume_title', 'Наименование тома', 'String'),
      field('object_name', 'Объект/сооружение', 'String'),
      numberField('pages', 'Страниц', { precision: 0 }),
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

  // Замечания экспертизы, прочитанные агентом из Excel и сопоставленные с расхождениями
  m.expertise_remarks = await ensureModule(api, nsID, {
    name: 'Замечания экспертизы',
    handle: 'expertise_remarks',
    fields: [
      recordRel('comparison', 'Сравнение', m.pd_rd_comparisons, 'title', ['title'], true),
      field('number', 'Номер', 'String'),
      field('location', 'Раздел, лист', 'String'),
      textArea('text', 'Замечание'),
      textArea('answer', 'Ответ проектировщика'),
      field('verdict', 'Итог сопоставления', 'Select', { options: selectOptions(EXPERT_VERDICT) }),
      numberField('match_score', 'Уверенность, %', { precision: 0 }),
      textArea('match_note', 'Пояснение'),
      recordRel('matched_discrepancies', 'Связанные расхождения', m.pd_rd_discrepancies, 'description', ['description'], false, { multi: true }),
      field('claim_mismatch', 'Ответ ≠ документы', 'String'),
      field('status', 'Статус', 'Select', { options: selectOptions(EXPERT_STATUS) }),
      field('source_file', 'Файл-источник', 'String'),
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
    ],
  })

  console.log('modules', m)

  // ---- Pages ----
  function listPage (title, handle, weight, icon, moduleID, fields) {
    return {
      title,
      handle,
      visible: true,
      weight,
      config: pageIcon(icon),
      blocks: withBlockIDs([recordList(title, [0, 0, 48, 40], moduleID, fields)]),
    }
  }

  function card (title, handle, moduleID, weight, fields, fieldRoles = {}, extra = {}) {
    const blocks = [recordBlock(title, [0, 0, 48, 28], fields, { fieldRoles })]
    if (extra.extraBlocks) blocks.push(...extra.extraBlocks)
    return {
      title,
      handle,
      moduleID: String(moduleID),
      visible: false,
      weight,
      config: pageIcon('fas file-lines'),
      blocks: withBlockIDs(blocks),
    }
  }

  const pages = [
    {
      title: 'Дашборд',
      handle: 'dashboard',
      visible: true,
      weight: 0,
      config: pageIcon('fas gauge-high'),
      blocks: withBlockIDs([
        metricBlock('Сводка', [0, 0, 48, 14], [
          metricItem('Объекты', m.objects, "status = 'active'", { color: '#2e59d9' }),
          metricItem('Сравнения ПД/РД в работе', m.pd_rd_comparisons, "status = 'new' OR status = 'processing'", { color: '#f6c23e' }),
          metricItem('Открытые расхождения ПД/РД', m.pd_rd_discrepancies, '', { color: '#e74a3b' }),
          metricItem('Проверки ИД с недостачами', m.id_check_runs, "overall_status = 'fail'", { color: '#e74a3b' }),
          metricItem('Расхождения ВОИСР', m.voisr_check_results, "status = 'mismatch'", { color: '#e74a3b' }),
        ]),
      ]),
    },

    listPage('Объекты', 'objects', 10, 'fas industry', m.objects, ['name', 'address', 'code', 'status']),
    card('Объект', 'object', m.objects, 11, ['name', 'address', 'code', 'status', 'notes'], {
      name: 'title', code: 'subtitle', status: 'badge', notes: 'body',
    }, {
      extraBlocks: [ruleChain('Проверить комплектность ИД', [0, 28, 48, 8], {
        chainID: 'stroykontrol-check-id', label: 'Проверить комплектность ИД', variant: 'primary', icon: 'clipboard-check',
      })],
    }),

    listPage('Сравнения ПД/РД', 'pd_rd_comparisons', 20, 'fas file-invoice', m.pd_rd_comparisons,
      ['title', 'object', 'iteration_no', 'pd_file', 'rd_file', 'status', 'similarity_percent', 'differing_pages']),
    card('Сравнение ПД/РД', 'pd_rd_comparison', m.pd_rd_comparisons, 21, [
      'title', 'object', 'iteration_no', 'parent_comparison', 'status', 'pd_file', 'rd_file', 'remarks_file', 'similarity_percent',
      'total_pages_pd', 'total_pages_rd', 'matching_pages', 'differing_pages', 'report_file', 'comment',
    ], { title: 'title', object: 'meta', status: 'badge', comment: 'body' }, {
      extraBlocks: [
        // pd_file/rd_file above stay for the simple one-file-per-stage case;
        // for packs (several PDFs per stage) add rows here. The agent takes
        // both into account.
        recordList('Комплект документов (ПД — эталон, РД — проверяемая)', [0, 28, 48, 14], m.cmp_documents,
          ['side', 'file', 'doc_version', 'volume_code', 'volume_title', 'pages', 'detected_by'],
          { prefilter: 'comparison = ${recordID}', refField: 'comparison', presort: 'side ASC, createdAt ASC', perPage: 20 }),
        ruleChain('Сравнить документы', [0, 42, 48, 8], {
          chainID: 'stroykontrol-compare-pd-rd', label: 'Сравнить документы (агент)', variant: 'primary', icon: 'code-compare',
        }),
        // Visual comparison viewer — a standalone agent (agents/stroykontrol),
        // not the lowcode platform itself; embedded via the IFrame block.
        // See client3/web/compose/src/components/PageBlocks/IFrameBase.vue.
        //
        // Relative path, not an absolute host:port: IFrameBase.vue already
        // prepends window.CortezaAPI to any src that doesn't start with
        // "http" (IFrameBase.vue:82), and /gateway/agents/stroykontrol/* is
        // an API Gateway route (create_apigw_proxy.mjs) reverse-proxying to
        // wherever the agent actually runs in *this* environment — so the
        // page-layout JSON itself never needs editing again per
        // dev/staging/prod; only the gateway route's `location` filter param
        // does (see create_apigw_proxy.mjs's AGENT_URL).
        block('IFrame', 'Просмотр сравнения (агент)', [0, 50, 48, 56], {
          src: '/gateway/agents/stroykontrol/?recordID=${recordID}&namespaceID=${namespaceID}',
          srcField: '',
          displayAsImage: false,
        }),
        recordList('Карточки замечаний', [0, 106, 48, 20], m.remark_cards,
          ['document_ref', 'sheet', 'what_diverges', 'final_text', 'status'],
          { prefilter: 'comparison = ${recordID}', presort: 'createdAt DESC', perPage: 20, hideAddButton: true }),
        ruleChain('Замечания экспертизы', [0, 126, 48, 6], {
          chainID: 'stroykontrol-import-remarks', label: 'Загрузить и сопоставить замечания экспертизы', variant: 'outline-primary', icon: 'comments',
        }),
        recordList('Замечания экспертизы', [0, 132, 48, 22], m.expertise_remarks,
          ['number', 'location', 'text', 'verdict', 'status'],
          { prefilter: 'comparison = ${recordID}', presort: 'number ASC', perPage: 50, hideAddButton: true }),
      ],
    }),

    listPage('Документы сравнения', 'cmp_documents', 24, 'fas file-pdf', m.cmp_documents,
      ['comparison', 'side', 'file', 'doc_version', 'volume_code', 'volume_title', 'pages', 'detected_by']),
    card('Документ сравнения', 'cmp_document', m.cmp_documents, 25,
      ['comparison', 'side', 'file', 'doc_version', 'volume_code', 'volume_title', 'object_name', 'pages', 'released', 'detected_by'],
      { file: 'title', side: 'badge' }),

    listPage('Расхождения ПД/РД', 'pd_rd_discrepancies', 22, 'fas triangle-exclamation', m.pd_rd_discrepancies,
      ['comparison', 'severity', 'kind', 'status', 'iter_state', 'page_pd', 'page_rd', 'description']),
    card('Расхождение ПД/РД', 'pd_rd_discrepancy', m.pd_rd_discrepancies, 23,
      ['comparison', 'severity', 'kind', 'status', 'page_pd', 'page_rd', 'description', 'anchor_text', 'evidence'],
      { discrepancy_type: 'title', severity: 'badge', status: 'badge', description: 'body' }),

    listPage('Замечания экспертизы', 'expertise_remarks', 28, 'fas comments', m.expertise_remarks,
      ['comparison', 'number', 'location', 'text', 'verdict', 'status']),
    card('Замечание экспертизы', 'expertise_remark', m.expertise_remarks, 29,
      ['comparison', 'number', 'location', 'text', 'answer', 'verdict', 'match_score', 'match_note', 'claim_mismatch', 'status', 'source_file'],
      { number: 'title', verdict: 'badge', status: 'badge', text: 'body' }),

    listPage('Карточки замечаний', 'remark_cards', 26, 'fas clipboard-list', m.remark_cards,
      ['comparison', 'document_ref', 'sheet', 'final_text', 'status']),
    card('Карточка замечания', 'remark_card', m.remark_cards, 27,
      ['comparison', 'document_ref', 'sheet', 'what_diverges', 'must_align', 'draft_text', 'final_text', 'status'],
      { document_ref: 'title', status: 'badge', what_diverges: 'body' }),

    listPage('Реестр ИД', 'id_registry', 30, 'fas list-check', m.id_registry,
      ['object', 'doc_name', 'doc_number', 'doc_version', 'doc_type', 'is_required']),
    card('Позиция реестра ИД', 'id_registry_item', m.id_registry, 31,
      ['object', 'doc_name', 'doc_number', 'doc_version', 'doc_type', 'is_required'],
      { doc_name: 'title', doc_number: 'subtitle', doc_type: 'badge' }),

    listPage('Файлы ИД', 'id_files', 32, 'fas file-arrow-up', m.id_files,
      ['object', 'file', 'recognized_name', 'matched_registry', 'match_status']),
    card('Файл ИД', 'id_file', m.id_files, 33, [
      'object', 'file', 'recognized_name', 'recognized_number', 'recognized_version',
      'matched_registry', 'match_status', 'uploaded_at', 'extracted_text',
    ], { recognized_name: 'title', match_status: 'badge' }),

    listPage('Проверки комплектности ИД', 'id_check_runs', 34, 'fas clipboard-check', m.id_check_runs,
      ['object', 'run_date', 'total_registry_docs', 'uploaded_docs', 'missing_docs', 'mismatched_docs', 'overall_status']),
    card('Проверка комплектности ИД', 'id_check_run', m.id_check_runs, 35, [
      'object', 'run_date', 'total_registry_docs', 'uploaded_docs', 'missing_docs',
      'mismatched_docs', 'overall_status', 'report_file',
    ], { object: 'title', overall_status: 'badge' }),

    listPage('Позиции смет', 'smeta_items', 40, 'fas file-invoice-dollar', m.smeta_items,
      ['object', 'smeta_number', 'item_number', 'work_name', 'volume', 'unit', 'total_amount']),
    card('Позиция сметы', 'smeta_item', m.smeta_items, 41,
      ['object', 'smeta_number', 'item_number', 'work_name', 'unit', 'volume', 'unit_price', 'total_amount'],
      { work_name: 'title', smeta_number: 'subtitle' }),

    listPage('Позиции ВОИСР', 'voisr_items', 42, 'fas table-list', m.voisr_items,
      ['object', 'position_number', 'work_name', 'smeta_ref', 'volume', 'amount']),
    card('Позиция ВОИСР', 'voisr_item', m.voisr_items, 43,
      ['object', 'position_number', 'work_name', 'smeta_ref', 'volume', 'amount'],
      { work_name: 'title', position_number: 'subtitle' }, {
        extraBlocks: [ruleChain('Проверить ВОИСР', [0, 20, 48, 8], {
          chainID: 'stroykontrol-check-voisr', label: 'Проверить со сметой (агент)', variant: 'primary', icon: 'scale-balanced',
        })],
      }),

    listPage('Результаты проверки ВОИСР', 'voisr_check_results', 44, 'fas scale-balanced', m.voisr_check_results,
      ['voisr_item', 'smeta_volume', 'smeta_amount', 'volume_diff', 'amount_diff', 'amount_diff_percent', 'status']),
    card('Результат проверки ВОИСР', 'voisr_check_result', m.voisr_check_results, 45, [
      'voisr_item', 'smeta_volume', 'smeta_amount', 'volume_diff', 'amount_diff', 'amount_diff_percent', 'status', 'note',
    ], { voisr_item: 'title', status: 'badge', note: 'body' }),
  ]

  const pageIDs = {}
  for (const def of pages) {
    pageIDs[def.handle] = await ensurePage(api, nsID, def)
  }

  // ---- Rule chains: агент (script det + ai LLM, вес — в узле combine) ----
  for (const chain of buildRuleChains({ nsID, modules: m, agentURL: (process.env.STROYKONTROL_AGENT_URL || 'http://localhost:8092').replace(/\/$/, '') })) {
    await ensureRuleChain(api, chain)
  }

  writeFileSync(join(HERE, 'applied.json'), JSON.stringify({ namespaceID: nsID, modules: m, pages: pageIDs }, null, 2))
  console.log('done. open /ns/stroykontrol (Дашборд)')
}

main().catch(e => { console.error(e); process.exit(1) })
