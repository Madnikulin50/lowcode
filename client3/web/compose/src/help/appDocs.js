/**
 * App-authored docs shown on public pages.
 * Used when the running Compose API does not yet return config.help / meta.help.
 */
import { GENERATED_NAMESPACE_DOCS, GENERATED_PAGE_DOCS, GENERATED_CHART_DOCS, GENERATED_BLOCK_DOCS } from './generatedCatalog.js'

const GENERATED_DESC = /:\s*(список записей|показатель|органайзер|график|чат с ИИ|RuleChain)/i

export const INVEST_NAMESPACE_DOCS = {
  hint: 'Документы, WBS, бюджет, EVM и ИИ-советчики',
  description: 'Единое пространство PMO: документы, WBS, договоры, бюджет, EVM, RFC и ИИ-советчики.',
  help: `Пространство сопровождения инвестиционных проектов.

- **Дашборд** — портфель, EVM/CPM, документы на вас
- **Проекты / WBS** — карточка проекта, иерархия работ, Гант
- **Документы / Договоры** — согласование, версии, пакет
- **Риски / Изменения** — реестр и RFC с влиянием на EAC
- **Бюджет / Прогресс** — план/факт и фиксация объёма
- **НСИ** — справочники и шаблоны
- **ИИ-советчики** — Юрист и Финконтролёр; решение принимает человек

Кнопка **Справка** на каждой странице открывает описание этого экрана.`,
}

export const INVEST_PAGE_DOCS = {
  dashboard: {
    description: 'Сводка портфеля: метрики, пересчёт EVM, критический путь, графики и документы на вас.',
    help: `Сводка инвестора по всему портфелю.

## Что смотреть
- Счётчики: активные проекты, документы на согласовании, открытые RFC и риски
- **Пересчитать EVM** / **Критический путь** — для PMO и инвестора
- **Проверить пороги** — алерты по документам, CPI и резерву
- Графики документов, RFC, рисков и Гант WBS
- **Портфель** — проекты, отсортированные по SPI
- **Документы на мне** — согласования, где вы исполнитель

Откройте проект из списка, чтобы перейти к WBS, бюджету и команде.`,
  },
  projects: {
    description: 'Реестр проектов: фаза, статус, бюджет, SPI/CPI/EAC.',
    help: `Список инвестиционных проектов.

Колонки: имя, код, фаза 1–6, статус, плановый бюджет, SPI, CPI, EAC.

- Откройте строку — карточка проекта, WBS, участники и пересчёт EVM
- Новая запись — кнопка добавления в списке
- SPI < 1 и CPI < 1 — отставание по срокам и перерасход`,
  },
  project: {
    description: 'Карточка проекта: реквизиты, EVM, WBS, команда.',
    help: `Карточка инвестиционного проекта.

- Фаза, статус, тип конструкции, инвестор, даты и бюджет
- SPI / CPI / EAC — после пересчёта EVM
- **Пересчитать EVM** — собирает WBS и факты прогресса
- Связанные списки: участники, WBS, документы`,
  },
  wbs: {
    description: 'Иерархия работ и график Ганта: сроки, SPI/CPI, критический путь.',
    help: `Рабочая структура проекта (WBS).

Сверху — график работ, ниже — список: код, имя, проект, уровень, % выполнения, SPI/CPI, признак критического пути.

- Откройте работу, чтобы увидеть факты прогресса, документы и статьи бюджета
- Критический путь считается кнопкой на дашборде`,
  },
  'wbs-item': {
    description: 'Работа WBS: сроки, EVM, факты прогресса и связанные документы.',
    help: `Позиция иерархии работ.

Уровень: стадия / подстадия / работа. Есть предшественник, плановые даты и поля EVM.

Справа — факты прогресса по этой работе; ниже — документы и статьи бюджета.`,
  },
  documents: {
    description: 'Канбан и реестр документов: черновик → согласование → утверждён / отклонён.',
    help: `Реестр проектных документов.

- Счётчики по статусам
- Канбан: черновик, на согласовании, утверждён, отклонён
- График «по статусу» и полный реестр справа

Откройте карточку, чтобы отправить на согласование, закрыть свой шаг, посмотреть версии и маршрут.`,
  },
  document: {
    description: 'Карточка документа: файл, маршрут согласования, версии и комментарии.',
    help: `Документ проекта.

- **Отправить** — на согласование, создаётся маршрут и версия
- **Согласовать мой шаг** / **Отклонить** / **Эскалировать**
- Версии файла и шаги маршрута
- Комментарии внизу

Утверждение документа — только когда пройдены все шаги.`,
  },
  contracts: {
    description: 'Реестр договоров: номер, контрагент, сумма, срок и статус.',
    help: `Договоры инвестиционных проектов.

Колонки: номер, название, проект, контрагент, сумма, статус, дата окончания.

Откройте карточку, чтобы увидеть пакет документов и спросить ИИ-юриста.`,
  },
  contract: {
    description: 'Карточка договора: условия, пакет документов и проверка юристом.',
    help: `Договор и связанные документы.

- Реквизиты, сумма, сроки, файл и условия
- **Спросить юриста** — рекомендация, решение за человеком
- Список документов пакета`,
  },
  risks: {
    description: 'Реестр рисков: вероятность, влияние, балл и митигация.',
    help: `Риски портфеля и проектов.

Канбан по статусу: открыт, митигация, закрыт, принят. Ниже — полный реестр с баллом и владельцем.

Балл считается из вероятности и влияния. Откройте карточку, чтобы описать митигацию.`,
  },
  risk: {
    description: 'Карточка риска: оценка, владелец и план митигации.',
    help: `Один риск проекта или работы WBS.

Заполните вероятность, влияние, владельца и текст митигации. Статус двигайте по мере работы с риском.`,
  },
  changes: {
    description: 'Запросы на изменение (RFC): влияние на бюджет, сроки и EAC.',
    help: `Изменения scope, бюджета и сроков.

Канбан RFC и таблица с дельтой бюджета, дней и EAC после изменения.

Откройте RFC: сначала **симулируйте EAC**, затем отправьте на согласование. Утверждение двигает baseline.`,
  },
  'change-request': {
    description: 'Карточка RFC: симуляция EAC, согласование и журнал.',
    help: `Запрос на изменение.

- Обоснование, тип, дельта бюджета и дней
- **Симулировать EAC** — оценка до утверждения
- Отправка, утверждение и отклонение
- Журнал — аудиторский след после решения`,
  },
  budget: {
    description: 'План, факт и резерв; денежный поток; вопрос финконтролёру.',
    help: `Финансы портфеля.

- Суммы плана, факта и резерва по статьям
- **Спросить финконтролёра** — рекомендация, бюджет сам не меняется
- Статьи бюджета и платежи (приход/расход)
- Графики потока по направлению и датам`,
  },
  'budget-line': {
    description: 'Статья бюджета: план, факт, резерв и связанные платежи.',
    help: `Одна статья сметы.

Справа — платежи денежного потока по этой статье. Резерв держите отдельно от факта.`,
  },
  progress: {
    description: 'Факты выполнения: объём, процент, фото, геометка и автор.',
    help: `Фиксация прогресса на площадке.

Каждая строка — факт по работе WBS: количество, %, дата, автор. Откройте карточку, чтобы приложить фото/гео и пересчитать EVM.`,
  },
  'progress-fact': {
    description: 'Карточка факта: объём, фото, гео и пересчёт EVM.',
    help: `Одна фиксация объёма.

После сохранения нажмите **Зафиксировать и пересчитать** — SPI/CPI обновятся на работе и проекте.`,
  },
  nsi: {
    description: 'Справочники: типы документов, контрагенты, шаблоны WBS, материалы, требования фаз.',
    help: `Нормативно-справочная информация.

- Типы документов и требования к фазам проекта
- Контрагенты
- Типы конструкций и шаблоны WBS (типовая иерархия)
- Материалы и нормы времени

Шаблоны WBS клонируются на карточке проекта.`,
  },
  advisors: {
    description: 'ИИ-советчики «Юрист» и «Финконтролёр». Рекомендации проверяет человек.',
    help: `Два чата с ИИ.

- **Юрист** — договоры и пакеты документов: сроки, сумма, файл, УКЭП, формулировки
- **Финконтролёр** — план/факт, SPI/CPI/EAC, влияние RFC

ИИ не утверждает документы и не пересчитывает бюджет. Решение всегда за человеком.`,
  },
}

function pickDescription (pageDesc, catalogDesc) {
  if (catalogDesc && GENERATED_DESC.test(pageDesc || '')) return catalogDesc
  return pageDesc || catalogDesc || ''
}

export function lookupPageDocs (namespaceSlug, pageHandle, page) {
  const handle = pageHandle || page?.handle || ''
  const pageID = page?.pageID || ''
  const title = page?.title || ''
  const generated = GENERATED_PAGE_DOCS[`${namespaceSlug}/${handle}`]
    || GENERATED_PAGE_DOCS[`${namespaceSlug}/${pageID}`]
    || GENERATED_PAGE_DOCS[`${namespaceSlug}/${title}`]
  if (generated) return generated
  if (namespaceSlug === 'invest') return INVEST_PAGE_DOCS[handle] || {}
  return {}
}

export function lookupNamespaceDocs (namespaceSlug) {
  return GENERATED_NAMESPACE_DOCS[namespaceSlug] || (namespaceSlug === 'invest' ? INVEST_NAMESPACE_DOCS : {})
}

export function pageHelpDocs (namespace, page) {
  const slug = namespace?.slug || namespace?.handle || ''
  const catalog = lookupPageDocs(slug, page?.handle, page)
  return {
    title: page?.title || '',
    description: pickDescription(page?.description, catalog.description),
    help: page?.config?.help || catalog.help || '',
  }
}

/** Copy options.help from the API payload after JS page-block types drop unknown keys. */
export function restoreBlockHelp (page, raw) {
  if (!page?.blocks) return page
  const rawBlocks = Array.isArray(raw?.blocks) ? raw.blocks : []
  const byID = new Map(rawBlocks.map(b => [String(b.blockID || ''), b]))
  page.blocks.forEach((block, i) => {
    const src = byID.get(String(block.blockID || '')) || rawBlocks[i]
    const help = src?.options?.help
    if (!block.options) block.options = {}
    if (!String(block.options.help || '').trim() && help) {
      block.options.help = help
    }
    if (src?.options?.hideHelpButton != null) {
      block.options.hideHelpButton = src.options.hideHelpButton
    }
  })
  return page
}

export function hydrateBlockDocs (namespace, page) {
  if (!page?.blocks) return page
  const slug = namespace?.slug || namespace?.handle || ''
  const handle = page.handle || page.pageID || ''
  page.blocks.forEach(block => {
    if (!block) return
    if (!block.options) block.options = {}
    const current = String(block.options.help || '').trim()
    const genericMetric = current.includes('Каждое число — агрегат по модулю')
    if (current && !genericMetric) return
    const docs = GENERATED_BLOCK_DOCS?.[`${slug}/${handle}/${block.blockID}`]
      || GENERATED_BLOCK_DOCS?.[`${slug}/${page.pageID}/${block.blockID}`]
    if (docs?.help) block.options.help = docs.help
  })
  return page
}

/** Fill empty page description/help from the catalog so the editor matches the Help button. */
export function hydratePageDocs (namespace, page) {
  if (!page) return page
  const docs = pageHelpDocs(namespace, page)
  if (!page.config) page.config = {}
  if (!String(page.config.help || '').trim() && docs.help) {
    page.config.help = docs.help
  }
  if (docs.description && pickDescription(page.description, docs.description) !== page.description) {
    page.description = docs.description
  }
  hydrateBlockDocs(namespace, page)
  return page
}

export function lookupChartDocs (namespaceSlug, chart) {
  const handle = chart?.handle || ''
  const chartID = chart?.chartID || ''
  const name = chart?.name || ''
  return GENERATED_CHART_DOCS?.[`${namespaceSlug}/${handle}`]
    || GENERATED_CHART_DOCS?.[`${namespaceSlug}/${chartID}`]
    || GENERATED_CHART_DOCS?.[`${namespaceSlug}/${name}`]
    || {}
}

export function chartHelpDocs (namespace, chart) {
  const slug = namespace?.slug || namespace?.handle || ''
  const catalog = lookupChartDocs(slug, chart)
  return {
    title: chart?.name || '',
    description: chart?.config?.description || catalog.description || '',
    help: chart?.config?.help || catalog.help || '',
  }
}

export function hydrateChartDocs (namespace, chart) {
  if (!chart) return chart
  const docs = chartHelpDocs(namespace, chart)
  if (!chart.config) chart.config = {}
  if (!String(chart.config.help || '').trim() && docs.help) {
    chart.config.help = docs.help
  }
  if (!String(chart.config.description || '').trim() && docs.description) {
    chart.config.description = docs.description
  }
  return chart
}

export function namespaceHelpDocs (namespace) {
  const slug = namespace?.slug || namespace?.handle || ''
  const catalog = lookupNamespaceDocs(slug)
  const meta = namespace?.meta || {}
  return {
    title: namespace?.name || '',
    hint: meta.subtitle || catalog.hint || '',
    description: meta.description || catalog.description || '',
    help: meta.help || catalog.help || '',
  }
}
