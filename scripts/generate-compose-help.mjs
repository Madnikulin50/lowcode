#!/usr/bin/env node
/**
 * Generate v1 subtitle/description/help for every compose namespace, page, and chart.
 *
 *   node scripts/generate-compose-help.mjs
 *   FORCE=1 node scripts/generate-compose-help.mjs   # overwrite existing texts
 */
import { execFileSync } from 'node:child_process'
import {
  mintToken, detectBase, apiFactory, setOf,
} from '../agents/invest/compose/helpers.mjs'

const HELP_MARK = '<!-- compose-help:v1 -->'
const FORCE = process.env.FORCE === '1' || process.argv.includes('--force')

const BLOCK_KIND = {
  RecordList: 'список записей',
  Record: 'карточка записи',
  Chart: 'график',
  Content: 'текст',
  Metric: 'показатель',
  Calendar: 'календарь',
  File: 'файлы',
  IFrame: 'встроенная страница',
  Automation: 'кнопки автоматизации',
  Comment: 'комментарии',
  SocialFeed: 'лента',
  Progress: 'прогресс',
  Navigation: 'навигация',
  Tabs: 'вкладки',
  Geometry: 'карта',
  Report: 'отчёт',
  RecordRevisions: 'история изменений',
  RecordOrganizer: 'органайзер',
  AiChat: 'чат с ИИ',
}

function blank (v) {
  return !String(v || '').trim()
}

function generatedHelp (v) {
  return String(v || '').includes(HELP_MARK)
}

function shouldWrite (current, kind) {
  if (FORCE) return true
  if (kind === 'help') return blank(current) || generatedHelp(current)
  return blank(current)
}

function joinList (items, limit = 12) {
  const clean = items.filter(Boolean)
  if (!clean.length) return ''
  const shown = clean.slice(0, limit)
  const extra = clean.length - shown.length
  const text = shown.map(s => `«${s}»`).join(', ')
  return extra > 0 ? `${text} и ещё ${extra}` : text
}

function mdList (items) {
  return items.filter(Boolean).map(s => `- ${s}`).join('\n')
}

function blockSummary (blocks = []) {
  const kinds = []
  const seen = new Set()
  for (const b of blocks) {
    const kind = b.kind || 'блок'
    if (seen.has(kind)) continue
    seen.add(kind)
    kinds.push(BLOCK_KIND[kind] || kind)
  }
  return kinds
}

function pageRole (page, modulesByID) {
  if (page.moduleID && page.moduleID !== '0') {
    const mod = modulesByID.get(String(page.moduleID))
    return mod ? `карточка записи модуля «${mod.name}»` : 'карточка записи'
  }
  return page.visible === false ? 'служебная страница' : 'экран пространства'
}

function generateNamespace (ns, pages, charts, modules) {
  const navPages = pages.filter(p => p.visible !== false && (!p.moduleID || p.moduleID === '0'))
  const recordPages = pages.filter(p => p.moduleID && p.moduleID !== '0')
  const pageNames = navPages.map(p => p.title).filter(Boolean)
  const chartNames = charts.map(c => c.name).filter(Boolean)
  const moduleNames = modules.map(m => m.name).filter(Boolean)

  const subtitle = ns.name
    ? `Пространство «${ns.name}»`
    : (ns.slug || 'Пространство')

  const bits = []
  if (pageNames.length) bits.push(`${pageNames.length} ${plural(pageNames.length, 'страница', 'страницы', 'страниц')}`)
  if (moduleNames.length) bits.push(`${moduleNames.length} ${plural(moduleNames.length, 'модуль', 'модуля', 'модулей')}`)
  if (chartNames.length) bits.push(`${chartNames.length} ${plural(chartNames.length, 'график', 'графика', 'графиков')}`)
  const description = bits.length
    ? `${ns.name}: ${bits.join(', ')}. ${pageNames.length ? `Основные экраны: ${joinList(pageNames, 6)}.` : ''}`
    : `${ns.name || 'Пространство'} — рабочая область Compose.`

  const help = `${HELP_MARK}
# ${ns.name || ns.slug || 'Пространство'}

Это рабочее пространство Compose. Здесь живут страницы, модули и графики этого приложения.

## Как пользоваться

- Откройте страницу в боковом меню
- Иконка **?** рядом с названием показывает справку экрана
- Списки и карточки ведут к записям модулей
- Графики читают те же записи и показывают сводку

${navPages.length ? `## Страницы\n\n${mdList(navPages.map(p => `**${p.title}**${p.description ? ` — ${oneLine(p.description)}` : ''}`))}\n` : ''}${recordPages.length ? `## Карточки записей\n\nСкрыты из меню, открываются из списков:\n\n${mdList(recordPages.map(p => {
    const mod = modules.find(m => String(m.moduleID) === String(p.moduleID))
    return `**${p.title}**${mod ? ` (модуль «${mod.name}»)` : ''}`
  }))}\n` : ''}${chartNames.length ? `## Графики\n\n${mdList(charts.map(c => `**${c.name}**${c.config?.description ? ` — ${oneLine(c.config.description)}` : ''}`))}\n` : ''}${moduleNames.length ? `## Модули\n\n${mdList(moduleNames.map(n => `**${n}**`))}\n` : ''}`.trim()

  return { subtitle, description: oneLine(description), help }
}

function generatePage (page, ns, modulesByID, chartsByID) {
  const mod = page.moduleID && page.moduleID !== '0' ? modulesByID.get(String(page.moduleID)) : null
  const kinds = blockSummary(page.blocks)
  const chartRefs = (page.blocks || [])
    .filter(b => b.kind === 'Chart' && b.options?.chartID)
    .map(b => chartsByID.get(String(b.options.chartID)))
    .filter(Boolean)
  const listMods = (page.blocks || [])
    .filter(b => b.kind === 'RecordList' && b.options?.moduleID)
    .map(b => modulesByID.get(String(b.options.moduleID)))
    .filter(Boolean)

  const role = pageRole(page, modulesByID)
  const description = mod
    ? `Карточка записи «${mod.name}»: просмотр и редактирование.`
    : (kinds.length
      ? `${page.title}: ${kinds.join(', ')}.`
      : `${page.title} — страница пространства «${ns.name}».`)

  const help = `${HELP_MARK}
# ${page.title}

${role} в пространстве **${ns.name}**.

${mod ? `Привязана к модулю **${mod.name}**. Открывается из списков и ссылок на записи.\n` : ''}${kinds.length ? `## Что на странице\n\n${mdList(kinds.map(k => k.charAt(0).toUpperCase() + k.slice(1)))}\n` : ''}${listMods.length ? `## Списки\n\n${mdList(listMods.map(m => `Записи модуля **${m.name}**`))}\n` : ''}${chartRefs.length ? `## Графики\n\n${mdList(chartRefs.map(c => `**${c.name}**`))}\n` : ''}## Подсказка

Иконка **?** в шапке открывает эту справку. Описание страницы видно сразу, полный текст — в панели.
`.trim()

  return { description: oneLine(description), help }
}

function generateChart (chart, ns, modulesByID) {
  const reports = chart.config?.reports || []
  const report = reports[0] || {}
  const mod = report.moduleID ? modulesByID.get(String(report.moduleID)) : null
  const metrics = (report.metrics || []).map(m => m.label || m.field).filter(Boolean)
  const dims = (report.dimensions || []).map(d => d.label || d.field).filter(Boolean)

  const description = [
    chart.name,
    mod ? `по модулю «${mod.name}»` : null,
    metrics.length ? `(${metrics.slice(0, 3).join(', ')})` : null,
  ].filter(Boolean).join(' ') + '.'

  const help = `${HELP_MARK}
# ${chart.name}

График пространства **${ns.name}**.
${mod ? `\nИсточник данных — модуль **${mod.name}**.\n` : ''}${dims.length ? `\n## Измерения\n\n${mdList(dims)}\n` : ''}${metrics.length ? `\n## Метрики\n\n${mdList(metrics)}\n` : ''}
Откройте **?** на блоке графика, чтобы снова увидеть эту справку. Описание показывается под заголовком блока.
`.trim()

  return { description: oneLine(description), help }
}

function oneLine (s) {
  return String(s || '').replace(/\s+/g, ' ').trim()
}

function plural (n, one, few, many) {
  const m10 = n % 10
  const m100 = n % 100
  if (m10 === 1 && m100 !== 11) return one
  if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) return few
  return many
}

const DSN_CANDIDATES = [...new Set([
  process.env.COMPOSE_DSN,
  'postgres://postgres:Zse45rdx@127.0.0.1:5432/test9?sslmode=disable',
  'postgres://postgres:Zse45rdx@127.0.0.1:5432/test10?sslmode=disable',
  'postgres://postgres:Zse45rdx@127.0.0.1:5432/test3?sslmode=disable',
].filter(Boolean))]

function psql (dsn, sql) {
  return execFileSync('psql', [dsn, '-v', 'ON_ERROR_STOP=1', '-tA', '-c', sql], { encoding: 'utf8' }).trim()
}

function sqlLiteral (value) {
  return `E'${String(value).replace(/\\/g, '\\\\').replace(/'/g, "\\'")}'`
}

function detectDsn (namespaceID) {
  for (const dsn of DSN_CANDIDATES) {
    try {
      const hit = psql(dsn, `SELECT 1 FROM compose_namespace WHERE id = ${namespaceID} LIMIT 1`)
      if (hit === '1') return dsn
    } catch {}
  }
  throw new Error('Cannot find compose_namespace in known Postgres DSNs. Set COMPOSE_DSN.')
}

function mergeJsonb (dsn, table, id, column, patch) {
  const pairs = Object.entries(patch)
    .map(([k, v]) => `${sqlLiteral(k)}, to_jsonb(${sqlLiteral(v)}::text)`)
    .join(', ')
  psql(dsn, `
    UPDATE ${table}
    SET ${column} = COALESCE(${column}, '{}'::jsonb) || jsonb_build_object(${pairs})
    WHERE id = ${id}
  `)
}

function mergeChartConfigText (dsn, id, patch) {
  const pairs = Object.entries(patch)
    .map(([k, v]) => `${sqlLiteral(k)}, to_jsonb(${sqlLiteral(v)}::text)`)
    .join(', ')
  psql(dsn, `
    UPDATE compose_chart
    SET config = (
      CASE
        WHEN config IS NULL OR btrim(config) = '' THEN '{}'::jsonb
        ELSE config::jsonb
      END || jsonb_build_object(${pairs})
    )::text
    WHERE id = ${id}
  `)
}

async function main () {
  const token = await mintToken()
  const base = await detectBase(token)
  const api = apiFactory(base, token)
  console.log('API', base, FORCE ? '(force overwrite)' : '(fill empty / refresh v1 help)')

  const namespaces = setOf(await api('GET', '/namespace/?limit=500')).filter(n => !n.deletedAt)
  console.log('namespaces', namespaces.length)
  const dsn = detectDsn(namespaces[0].namespaceID)
  console.log('DSN', dsn.replace(/:[^:@/]+@/, ':***@'))

  let nsN = 0
  let pageN = 0
  let chartN = 0

  for (const ns of namespaces) {
    const nsID = ns.namespaceID
    const [pages, charts, modules] = await Promise.all([
      api('GET', `/namespace/${nsID}/page/?limit=500`).then(setOf).then(s => s.filter(p => !p.deletedAt)),
      api('GET', `/namespace/${nsID}/chart/?limit=500`).then(setOf).then(s => s.filter(c => !c.deletedAt)),
      api('GET', `/namespace/${nsID}/module/?limit=500`).then(setOf).then(s => s.filter(m => !m.deletedAt)),
    ])

    const modulesByID = new Map(modules.map(m => [String(m.moduleID), m]))
    const chartsByID = new Map(charts.map(c => [String(c.chartID), c]))

    // Pages and charts first so namespace help can mention their descriptions after we write them.
    for (const page of pages) {
      const gen = generatePage(page, ns, modulesByID, chartsByID)
      const nextDesc = shouldWrite(page.description, 'description') ? gen.description : page.description
      const nextHelp = shouldWrite(page.config?.help, 'help') ? gen.help : (page.config?.help || '')
      if (nextDesc === page.description && nextHelp === (page.config?.help || '')) continue

      if (nextDesc !== page.description) {
        const full = await api('GET', `/namespace/${nsID}/page/${page.pageID}`)
        await api('POST', `/namespace/${nsID}/page/${page.pageID}`, {
          ...full,
          description: nextDesc,
          updatedAt: full.updatedAt,
        })
      }
      if (nextHelp !== (page.config?.help || '')) {
        mergeJsonb(dsn, 'compose_page', page.pageID, 'config', { help: nextHelp })
      }
      page.description = nextDesc
      page.config = { ...(page.config || {}), help: nextHelp }
      pageN++
      console.log('  page', ns.slug || nsID, page.handle || page.pageID)
    }

    for (const chart of charts) {
      const gen = generateChart(chart, ns, modulesByID)
      const nextDesc = shouldWrite(chart.config?.description, 'description') ? gen.description : chart.config?.description
      const nextHelp = shouldWrite(chart.config?.help, 'help') ? gen.help : (chart.config?.help || '')
      if (nextDesc === (chart.config?.description || '') && nextHelp === (chart.config?.help || '')) continue

      mergeChartConfigText(dsn, chart.chartID, { description: nextDesc, help: nextHelp })
      chart.config = { ...(chart.config || {}), description: nextDesc, help: nextHelp }
      chartN++
      console.log('  chart', ns.slug || nsID, chart.handle || chart.chartID)
    }

    const genNs = generateNamespace(ns, pages, charts, modules)
    const meta = { ...(ns.meta || {}) }
    const nextSub = shouldWrite(meta.subtitle, 'hint') ? genNs.subtitle : meta.subtitle
    const nextDesc = shouldWrite(meta.description, 'description') ? genNs.description : meta.description
    const nextHelp = shouldWrite(meta.help, 'help') ? genNs.help : meta.help
    if (nextSub !== meta.subtitle || nextDesc !== meta.description || nextHelp !== meta.help) {
      if (nextSub !== meta.subtitle || nextDesc !== meta.description) {
        const full = await api('GET', `/namespace/${nsID}`)
        await api('POST', `/namespace/${nsID}`, {
          ...full,
          meta: { ...(full.meta || {}), subtitle: nextSub, description: nextDesc },
          updatedAt: full.updatedAt,
        })
      }
      if (nextHelp !== meta.help) {
        mergeJsonb(dsn, 'compose_namespace', nsID, 'meta', { help: nextHelp })
      }
      nsN++
      console.log('namespace', ns.slug || ns.name, nsID)
    }
  }

  console.log(`done: namespaces ${nsN}, pages ${pageN}, charts ${chartN}`)
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
