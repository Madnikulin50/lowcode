export const STATUS_LABEL = {
  awaiting_answer: 'Ответ не предоставлен',
  being_fixed: 'Устраняется заявителем',
  answer_review: 'Требуется рассмотрение ответа',
  rework: 'Возвращено на доработку',
  resolved: 'Устранено',
}
export const STATUS_ORDER = ['awaiting_answer', 'being_fixed', 'answer_review', 'rework', 'resolved']
// эти смены статуса нужно объяснить (ТЗ 3.7.4): комментарий обязателен
export const STATUS_NEEDS_REASON = ['resolved', 'rework']
export const CRIT_LABEL = { critical: 'Критичное', noncritical: 'Некритичное' }
export const VERDICT_LABEL = {
  still_differs: 'расхождение сохраняется',
  resolved: 'устранено в этой итерации',
  no_match: 'связанных расхождений не найдено',
  no_documents: 'тома не загружены',
}

export const AI_LABEL = {
  fixed: 'ИИ: устранено',
  partial: 'ИИ: частично',
  not_fixed: 'ИИ: не устранено',
  unclear: 'ИИ: судить нельзя',
}

// Результат проверки устранения (ТЗ 3.4.6): четыре значения; вывод системы и решение специалиста хранятся раздельно
export const OUTCOME_LABEL = {
  fixed: 'Устранено',
  partial: 'Частично устранено',
  not_fixed: 'Не устранено',
  check: 'Требует проверки',
}
// материалы, заявленные в ответе проектировщика (ТЗ 3.4.9): только наличие
export const MAT_STATE_LABEL = { found: 'есть', missing: 'нет в документе', not_found: 'не найдено', no_document: 'том не загружен' }
export const MATERIALS_LABEL = { ok: 'все найдены', missing: 'не найдены', unloaded: 'тома не загружены' }
export const MAT_KIND_LABEL = { volume: 'том / лист', appendix: 'приложение', document: 'документ' }
export const OUTCOME_ORDER = ['fixed', 'partial', 'not_fixed', 'check']
// итоговое значение: решение специалиста, иначе вывод системы
export const finalOutcome = (r) => r.outcomeFinal || r.outcome

export function fmtDate (s) {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d)) return s
  return d.toLocaleDateString('ru-RU') + (s.length > 10 && s.indexOf('T00:00:00') < 0 ? ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) : '')
}

export function isOverdue (r) {
  return !!r.dueDate && r.status !== 'resolved' && r.dueDate.slice(0, 10) < new Date().toISOString().slice(0, 10)
}
