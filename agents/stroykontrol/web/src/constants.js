export const STATUS_LABEL = { new: 'Новое', processing: 'В обработке', done: 'Готово', failed: 'Ошибка' }
export const STATUS_CLASS = { new: '', processing: 'warn', done: 'ok', failed: 'bad' }
export const STATUS_ICON = { new: '•', processing: '⏳', done: '✓', failed: '⚠' }
export const SEV_ORDER = { high: 0, medium: 1, low: 2 }
export const SEV_LABEL = { low: 'низкая', medium: 'средняя', high: 'высокая' }
export const TYPE_LABEL = {
  content: 'содержание', numbering: 'нумерация', drawing: 'чертёж',
  missing_page: 'нет страницы', extra_page: 'лишняя страница',
}

// Small currentColor icons for the mode switcher (no icon font/CDN needed
// for four glyphs) — viewBox 0 0 16 16, stroke-based so they follow badge
// colors including the active/inverted state.
export const ICONS = {
  side: '<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4"><rect x="1" y="2" width="6" height="12" rx="1"/><rect x="9" y="2" width="6" height="12" rx="1"/></svg>',
  overlay: '<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4"><rect x="2" y="3" width="9" height="10" rx="1"/><rect x="5" y="3" width="9" height="10" rx="1" opacity="0.55"/></svg>',
  diff: '<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4"><circle cx="8" cy="8" r="6"/><path d="M8 2a6 6 0 0 1 0 12z" fill="currentColor" stroke="none"/></svg>',
  text: '<svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><path d="M2 4h12M2 8h12M2 12h8"/></svg>',
}

// One-line explanation shown under the toolbar for whichever mode is
// selected — the button labels alone don't say what each mode is *for*.
export const MODE_META = {
  side: { label: 'Рядом', hint: 'Страницы рядом друг с другом — для сопоставления вручную.' },
  overlay: { label: 'Наложение', hint: 'РД поверх ПД с регулируемой прозрачностью — заметны сдвиги и смещения.' },
  diff: { label: 'Авто-разница', hint: 'Совпадающие области гасятся в чёрный, различия остаются светлыми пятнами.' },
  diffColorized: { label: 'Авто-разница', hint: 'ПД красным, РД зелёным — совпадающее гасится в чёрный, различия остаются цветными.' },
  text: { label: 'Текст', hint: 'Построчное сравнение, как git diff: «−» красным — строка только в ПД, «+» зелёным — только в РД.' },
}

// How the two sides are called: ПД/РД in a stage comparison, До/После when two versions
// of one stage are compared (comparison.kind === 'versions'). Reactive: set once the comparison is loaded.
import { reactive } from 'vue'
export const names = reactive({ ref: 'ПД', target: 'РД', versions: false })
export function setMode (kind) {
  names.versions = kind === 'versions'
  names.ref = names.versions ? 'До' : 'ПД'
  names.target = names.versions ? 'После' : 'РД'
}
// nm rewrites the standalone words «ПД» / «РД» of a text for the current mode
export function nm (s) {
  if (!names.versions || !s) return s
  return s.replace(/(^|[^\p{L}\p{N}])ПД(?![\p{L}\p{N}])/gu, `$1${names.ref}`).replace(/(^|[^\p{L}\p{N}])РД(?![\p{L}\p{N}])/gu, `$1${names.target}`)
}

export const ZOOM_LEVELS = [50, 75, 100, 125, 150, 200]

// DXF ACI (AutoCAD Color Index) palette — only the indices filegen.mjs's
// dxfFile() actually emits, plus the common low numbers real DXF exports use.
export const ACI_COLORS = { 1: '#e5352b', 2: '#e0c020', 3: '#37b52b', 4: '#2bb5c0', 5: '#3050d0', 6: '#c02bc0', 7: '#2a2a2a', 8: '#808080', 9: '#a8a8a8' }
export const DXF_DEFAULT_STROKE = '#2a2a2a'

// Unchanged lines kept visible around each change, like `git diff -U3`.
export const DIFF_CONTEXT = 3

export const KIND_LABEL = {
  text: 'текст', table_row: 'таблица', table_cell: 'ячейка таблицы', param: 'ключевой параметр', drawing: 'чертёж', vor: 'ВОР / спецификация',
  not_released: 'не выпущено', scope_diff: 'вне объёма', format: 'оформление', cross_volume: 'между томами', element: 'марки элементов',
  change_reg: 'регистрация изменений',
}
export const REVIEW_LABEL = { new: 'новое', confirmed: 'подтверждено', rejected: 'отклонено', clarify: 'требует уточнения' }
export const REMARK_STATUS_LABEL = { draft: 'Черновик', issued: 'Выдано', closed: 'Закрыто' }
export const ITER_LABEL = { new: 'новое', persisting: 'не устранено', resolved: 'устранено' }
export const VERDICT_LABEL = {
  still_differs: 'расхождение сохраняется',
  resolved: 'устранено в этой итерации',
  no_match: 'связанных расхождений не найдено',
}
export const EXPERT_STATUS_LABEL = { open: 'Открыто', closed: 'Закрыто', disputed: 'Оспорено' }
