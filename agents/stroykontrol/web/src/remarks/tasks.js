// Задачи (ТЗ 3.7.1): статусы, переходы и правила — те же, что в агенте (cmp/task.go); агент проверяет всё сам,
// здесь — чтобы предлагать только допустимое.
export const TASK_STATUS_LABEL = {
  open: 'Открыта', in_progress: 'В работе', review: 'На проверке', returned: 'Возвращена на доработку', done: 'Выполнена', cancelled: 'Отменена',
}
export const TASK_STATUS_ORDER = ['open', 'in_progress', 'review', 'returned', 'done', 'cancelled']
const NEXT = {
  open: ['in_progress', 'cancelled'],
  in_progress: ['review', 'open', 'cancelled'],
  review: ['done', 'returned', 'cancelled'],
  returned: ['in_progress', 'cancelled'],
  done: ['open'],
  cancelled: ['open'],
}
export const nextStatuses = (s) => NEXT[s] || []
// решение специалиста: принять, вернуть, отменить, открыть заново — нужна причина
export const needsReason = (from, to) => ['done', 'returned', 'cancelled'].includes(to) || ['done', 'cancelled'].includes(from)
export const MOVE_LABEL = {
  open: 'Открыть заново', in_progress: 'Взять в работу', review: 'Передать на проверку', returned: 'Вернуть на доработку', done: 'Принять', cancelled: 'Отменить',
}
