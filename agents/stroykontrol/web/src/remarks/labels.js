export const STATUS_LABEL = {
  awaiting_answer: 'Ответ не предоставлен',
  being_fixed: 'Устраняется заявителем',
  answer_review: 'Требуется рассмотрение ответа',
  resolved: 'Устранено',
}
export const STATUS_ORDER = ['awaiting_answer', 'being_fixed', 'answer_review', 'resolved']
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

export function fmtDate (s) {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d)) return s
  return d.toLocaleDateString('ru-RU') + (s.length > 10 && s.indexOf('T00:00:00') < 0 ? ' ' + d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) : '')
}

export function isOverdue (r) {
  return !!r.dueDate && r.status !== 'resolved' && r.dueDate.slice(0, 10) < new Date().toISOString().slice(0, 10)
}
