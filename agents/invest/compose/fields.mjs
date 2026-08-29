import {
  selectOptions, field, recordRel, userField, fileField, dateField,
  moneyField, numberField, boolSwitch, geoField,
} from './helpers.mjs'

export const PHASES = [
  ['concept', '1. Концепция'],
  ['survey', '2. Изыскания'],
  ['design', '3. Проектирование'],
  ['construction', '4. Строительство'],
  ['commissioning', '5. Ввод в эксплуатацию'],
  ['operations', '6. Эксплуатация'],
]

export const PROJECT_STATUS = [
  ['draft', 'Черновик', { backgroundColor: 'light', textColor: 'dark' }],
  ['active', 'Активен', { backgroundColor: 'success', textColor: 'white' }],
  ['on_hold', 'Приостановлен', { backgroundColor: 'warning', textColor: 'dark' }],
  ['completed', 'Завершён', { backgroundColor: 'info', textColor: 'white' }],
  ['cancelled', 'Отменён', { backgroundColor: 'secondary', textColor: 'white' }],
]

export const MEMBER_ROLES = [
  ['investor', 'Инвестор / Заказчик'],
  ['bank', 'Банк / Кредитор'],
  ['contractor', 'Генподрядчик'],
  ['designer', 'ГИП / Проектировщик'],
  ['government', 'Гос. органы'],
  ['pmo', 'PMO'],
]

export const WBS_LEVEL = [
  ['stage', 'Стадия'],
  ['substage', 'Подстадия'],
  ['work', 'Работа'],
]

export const DOC_STATUS = [
  ['draft', 'Черновик', { backgroundColor: 'light', textColor: 'dark' }],
  ['in_review', 'На согласовании', { backgroundColor: 'warning', textColor: 'dark' }],
  ['approved', 'Утверждён', { backgroundColor: 'success', textColor: 'white' }],
  ['rejected', 'Отклонён', { backgroundColor: 'danger', textColor: 'white' }],
  ['archived', 'Архив', { backgroundColor: 'secondary', textColor: 'white' }],
]

export const SIGN_STATUS = [
  ['unsigned', 'Не подписан', { backgroundColor: 'light', textColor: 'dark' }],
  ['pending', 'Ожидает УКЭП', { backgroundColor: 'warning', textColor: 'dark' }],
  ['signed', 'Подписан', { backgroundColor: 'success', textColor: 'white' }],
  ['failed', 'Ошибка подписи', { backgroundColor: 'danger', textColor: 'white' }],
]

export const APPROVAL_DECISION = [
  ['pending', 'Ожидает', { backgroundColor: 'warning', textColor: 'dark' }],
  ['approved', 'Согласовано', { backgroundColor: 'success', textColor: 'white' }],
  ['rejected', 'Отклонено', { backgroundColor: 'danger', textColor: 'white' }],
  ['escalated', 'Эскалация', { backgroundColor: 'info', textColor: 'white' }],
]

export const CONTRACT_STATUS = [
  ['draft', 'Черновик', { backgroundColor: 'light', textColor: 'dark' }],
  ['active', 'Действует', { backgroundColor: 'success', textColor: 'white' }],
  ['completed', 'Исполнен', { backgroundColor: 'info', textColor: 'white' }],
  ['terminated', 'Расторгнут', { backgroundColor: 'danger', textColor: 'white' }],
]

export const RISK_LEVEL = [
  ['low', 'Низкий', { backgroundColor: 'success', textColor: 'white' }],
  ['medium', 'Средний', { backgroundColor: 'info', textColor: 'white' }],
  ['high', 'Высокий', { backgroundColor: 'warning', textColor: 'dark' }],
  ['critical', 'Критический', { backgroundColor: 'danger', textColor: 'white' }],
]

export const RISK_STATUS = [
  ['open', 'Открыт', { backgroundColor: 'danger', textColor: 'white' }],
  ['mitigating', 'Митигация', { backgroundColor: 'warning', textColor: 'dark' }],
  ['closed', 'Закрыт', { backgroundColor: 'success', textColor: 'white' }],
  ['accepted', 'Принят', { backgroundColor: 'secondary', textColor: 'white' }],
]

export const RFC_TYPE = [
  ['budget', 'Смета / бюджет'],
  ['schedule', 'Срок'],
  ['scope', 'Объём работ'],
]

export const RFC_STATUS = [
  ['draft', 'Черновик', { backgroundColor: 'light', textColor: 'dark' }],
  ['in_review', 'На согласовании', { backgroundColor: 'warning', textColor: 'dark' }],
  ['approved', 'Утверждён', { backgroundColor: 'success', textColor: 'white' }],
  ['rejected', 'Отклонён', { backgroundColor: 'danger', textColor: 'white' }],
]

export const CASHFLOW_DIR = [
  ['inflow', 'Приход', { backgroundColor: 'success', textColor: 'white' }],
  ['outflow', 'Расход', { backgroundColor: 'danger', textColor: 'white' }],
]

export const ADVISOR_ROLE = [
  ['lawyer', 'Юрист'],
  ['fincontroller', 'Финконтролёр'],
]

const spiThresholds = [
  { value: 0.9, variant: 'danger' },
  { value: 1, variant: 'warning' },
  { value: 1.05, variant: 'success' },
]

export function documentTypeFields () {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('code', 'Код', 'String', { required: true }),
    field('description', 'Описание', 'String'),
  ]
}

export function constructionTypeFields () {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('code', 'Код', 'String', { required: true }),
    field('description', 'Описание', 'String'),
  ]
}

export function wbsTemplateFields (typesID) {
  return [
    recordRel('construction_type', 'Тип конструкции', typesID, 'name', ['name', 'code'], true),
    field('code', 'Код WBS', 'String', { required: true }),
    field('name', 'Название', 'String', { required: true }),
    field('level', 'Уровень', 'Select', { options: selectOptions(WBS_LEVEL) }),
    field('parent_code', 'Код родителя', 'String'),
    field('predecessor_code', 'Код предшественника', 'String'),
    moneyField('budget_planned', 'BAC шаблон'),
    numberField('duration_days', 'Длительность, дн', { precision: 0, suffix: ' д' }),
  ]
}

export function phaseRequirementFields (typesID) {
  return [
    field('phase', 'Фаза', 'Select', { options: selectOptions(PHASES), required: true }),
    recordRel('doc_type', 'Тип документа', typesID, 'name', ['name', 'code'], true),
    boolSwitch('required', 'Обязателен'),
  ]
}

export function counterpartyFields () {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('inn', 'ИНН', 'String'),
    field('kpp', 'КПП', 'String'),
    field('role', 'Роль', 'Select', { options: selectOptions(MEMBER_ROLES) }),
    field('notes', 'Заметки', 'String'),
  ]
}

export function materialFields () {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('unit', 'Ед. изм.', 'String'),
    moneyField('unit_price', 'Цена за ед.'),
    field('gost', 'ГОСТ / ТУ', 'String'),
  ]
}

export function laborNormFields () {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('unit', 'Ед. изм.', 'String'),
    numberField('hours', 'Норма, ч', { precision: 2, suffix: ' ч' }),
    field('notes', 'Заметки', 'String'),
  ]
}

export function projectFields (constructionTypesID) {
  return [
    field('name', 'Название', 'String', { required: true }),
    field('code', 'Код', 'String'),
    field('phase', 'Фаза', 'Select', { options: selectOptions(PHASES) }),
    field('status', 'Статус', 'Select', { options: selectOptions(PROJECT_STATUS) }),
    recordRel('construction_type', 'Тип конструкции', constructionTypesID, 'name', ['name', 'code']),
    field('investor', 'Инвестор', 'String'),
    dateField('start_planned', 'Старт (план)'),
    dateField('end_planned', 'Финиш (план)'),
    moneyField('budget_planned', 'Бюджет план'),
    moneyField('budget_actual', 'Бюджет факт'),
    moneyField('eac', 'EAC'),
    numberField('spi', 'SPI', { precision: 3, thresholds: spiThresholds }),
    numberField('cpi', 'CPI', { precision: 3, thresholds: spiThresholds }),
    field('description', 'Описание', 'String'),
  ]
}

export function projectMemberFields (projectsID) {
  return [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    userField('user', 'Пользователь', { required: true }),
    field('role', 'Роль', 'Select', { options: selectOptions(MEMBER_ROLES), required: true }),
  ]
}

export function wbsFields (projectsID, wbsID) {
  const fields = [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    field('code', 'Код WBS', 'String', { required: true }),
    field('name', 'Название', 'String', { required: true }),
    field('level', 'Уровень', 'Select', { options: selectOptions(WBS_LEVEL) }),
    dateField('start_planned', 'Старт план'),
    dateField('end_planned', 'Финиш план'),
    dateField('start_actual', 'Старт факт'),
    dateField('end_actual', 'Финиш факт'),
    moneyField('budget_planned', 'BAC / план'),
    moneyField('actual_cost', 'AC факт'),
    numberField('percent_complete', '% выполнения', { precision: 1, suffix: ' %', min: 0, max: 100, display: 'progress' }),
    moneyField('pv', 'PV'),
    moneyField('ev', 'EV'),
    numberField('spi', 'SPI', { precision: 3, thresholds: spiThresholds }),
    numberField('cpi', 'CPI', { precision: 3, thresholds: spiThresholds }),
    moneyField('eac', 'EAC'),
    boolSwitch('is_critical', 'Критический путь'),
    numberField('total_float', 'Полный резерв, дн', { precision: 1, suffix: ' д' }),
    field('notes', 'Заметки', 'String'),
  ]
  if (wbsID && wbsID !== '0') {
    fields.splice(1, 0, recordRel('parent', 'Родитель', wbsID, 'name', ['code', 'name']))
    fields.splice(2, 0, recordRel('predecessor', 'Предшественник', wbsID, 'name', ['code', 'name'], false, { multi: true }))
  }
  return fields
}

export function contractFields (projectsID, counterpartiesID) {
  return [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    field('number', 'Номер', 'String', { required: true }),
    field('title', 'Название', 'String', { required: true }),
    recordRel('counterparty', 'Контрагент', counterpartiesID, 'name', ['name', 'inn']),
    moneyField('amount', 'Сумма'),
    dateField('start_date', 'Начало'),
    dateField('end_date', 'Окончание'),
    field('status', 'Статус', 'Select', { options: selectOptions(CONTRACT_STATUS) }),
    fileField('file', 'Файл договора', { documents: true, images: false }),
    field('terms', 'Условия', 'String'),
  ]
}

export function documentFields (projectsID, wbsID, contractsID, typesID) {
  return [
    field('title', 'Название', 'String', { required: true }),
    field('number', 'Номер', 'String'),
    recordRel('doc_type', 'Тип', typesID, 'name', ['name', 'code']),
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    recordRel('wbs', 'WBS', wbsID, 'name', ['code', 'name']),
    recordRel('contract', 'Договор', contractsID, 'title', ['number', 'title']),
    field('status', 'Статус', 'Select', { options: selectOptions(DOC_STATUS) }),
    field('sign_status', 'УКЭП', 'Select', { options: selectOptions(SIGN_STATUS) }),
    userField('author', 'Автор'),
    userField('assignee', 'Согласующий'),
    dateField('due_date', 'Срок согласования'),
    fileField('file', 'Текущий файл', { documents: true, images: true }),
    field('notes', 'Заметки', 'String'),
  ]
}

export function documentVersionFields (documentsID) {
  return [
    recordRel('document', 'Документ', documentsID, 'title', ['title', 'number'], true),
    numberField('version', 'Версия', { precision: 0, min: 1 }),
    fileField('file', 'Файл', { documents: true, images: true }),
    userField('author', 'Автор'),
    field('comment', 'Комментарий', 'String'),
    dateField('created_on', 'Дата', { onlyDate: false }),
  ]
}

export function approvalFields (documentsID) {
  return [
    recordRel('document', 'Документ', documentsID, 'title', ['title', 'number'], true),
    userField('approver', 'Согласующий', { required: true }),
    field('decision', 'Решение', 'Select', { options: selectOptions(APPROVAL_DECISION) }),
    numberField('step', 'Шаг', { precision: 0, min: 1 }),
    field('role', 'Роль шага', 'Select', { options: selectOptions(MEMBER_ROLES) }),
    dateField('due_date', 'Срок'),
    dateField('decided_at', 'Дата решения', { onlyDate: false }),
    field('comment', 'Комментарий', 'String'),
  ]
}

export function commentFields (documentsID) {
  return [
    recordRel('document', 'Документ', documentsID, 'title', ['title', 'number'], true),
    field('title', 'Заголовок', 'String'),
    field('content', 'Текст', 'String', { required: true }),
    userField('author', 'Автор'),
  ]
}

export function riskFields (projectsID, wbsID) {
  return [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    recordRel('wbs', 'WBS', wbsID, 'name', ['code', 'name']),
    field('title', 'Название', 'String', { required: true }),
    field('probability', 'Вероятность', 'Select', { options: selectOptions(RISK_LEVEL) }),
    field('impact', 'Влияние', 'Select', { options: selectOptions(RISK_LEVEL) }),
    numberField('score', 'Балл', { precision: 0, min: 0, max: 16 }),
    field('status', 'Статус', 'Select', { options: selectOptions(RISK_STATUS) }),
    userField('owner', 'Владелец'),
    field('mitigation', 'План митигации', 'String'),
    field('description', 'Описание', 'String'),
  ]
}

export function rfcFields (projectsID, wbsID) {
  return [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    recordRel('wbs', 'WBS', wbsID, 'name', ['code', 'name']),
    field('title', 'Название', 'String', { required: true }),
    field('rfc_type', 'Тип', 'Select', { options: selectOptions(RFC_TYPE) }),
    field('status', 'Статус', 'Select', { options: selectOptions(RFC_STATUS) }),
    moneyField('delta_budget', 'Влияние на бюджет'),
    numberField('delta_days', 'Влияние на срок, дн', { precision: 0, suffix: ' д' }),
    moneyField('eac_before', 'EAC до'),
    moneyField('eac_after', 'EAC после'),
    dateField('end_after', 'Финиш после RFC'),
    boolSwitch('simulated', 'Симуляция выполнена'),
    userField('author', 'Автор'),
    field('justification', 'Обоснование', 'String'),
  ]
}

export function changeLogFields (rfcID, projectsID) {
  return [
    recordRel('rfc', 'RFC', rfcID, 'title', ['title']),
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code']),
    field('summary', 'Сводка', 'String', { required: true }),
    moneyField('old_budget', 'Бюджет было'),
    moneyField('new_budget', 'Бюджет стало'),
    dateField('old_end', 'Финиш было'),
    dateField('new_end', 'Финиш стало'),
    userField('actor', 'Кто изменил'),
    dateField('changed_at', 'Когда', { onlyDate: false }),
  ]
}

export function budgetLineFields (projectsID, wbsID) {
  return [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    recordRel('wbs', 'WBS', wbsID, 'name', ['code', 'name']),
    field('article', 'Статья', 'String', { required: true }),
    moneyField('planned', 'План'),
    moneyField('actual', 'Факт'),
    moneyField('reserve', 'Резерв'),
    field('notes', 'Заметки', 'String'),
  ]
}

export function cashflowFields (projectsID, budgetID) {
  return [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    recordRel('budget_line', 'Статья бюджета', budgetID, 'article', ['article']),
    dateField('date', 'Дата', { required: true }),
    moneyField('amount', 'Сумма'),
    field('direction', 'Направление', 'Select', { options: selectOptions(CASHFLOW_DIR) }),
    field('description', 'Описание', 'String'),
  ]
}

export function progressFactFields (projectsID, wbsID) {
  return [
    recordRel('project', 'Проект', projectsID, 'name', ['name', 'code'], true),
    recordRel('wbs', 'Работа', wbsID, 'name', ['code', 'name'], true),
    numberField('quantity', 'Объём', { precision: 3 }),
    field('unit', 'Ед. изм.', 'String'),
    numberField('percent', '% к работе', { precision: 1, suffix: ' %', min: 0, max: 100 }),
    moneyField('cost', 'Стоимость факта'),
    fileField('photo', 'Фотофиксация', { images: true, documents: false, webcam: true, multi: true }),
    geoField('geo', 'Геолокация'),
    userField('author', 'Автор'),
    dateField('recorded_at', 'Дата фиксации', { onlyDate: false }),
    boolSwitch('offline', 'Создано офлайн'),
    field('notes', 'Заметки', 'String'),
  ]
}

export function aiAdvisorFields () {
  return [
    field('name', 'Имя', 'String', { required: true }),
    field('role', 'Роль', 'Select', { options: selectOptions(ADVISOR_ROLE) }),
    field('prompt', 'Системный промпт', 'String'),
    field('modules', 'Модули (handles)', 'String'),
    boolSwitch('enabled', 'Включён'),
  ]
}
