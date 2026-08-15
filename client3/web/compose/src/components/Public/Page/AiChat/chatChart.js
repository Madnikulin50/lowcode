const MAX_CATEGORIES = 24
const MAX_SERIES = 8
// Opening fence: 3+ backticks, optional spaces, language. JSON may start on the same line.
const FENCE_OPEN = /```+\s*(compose-chart|chart|echarts|json)\b[^\n`]*/i

export function splitChartParts (text) {
  const src = String(text || '')
  const wholeCompose = parseComposeChartSpec(src)
  if (wholeCompose && looksLikeBareObject(src)) {
    return [composeChartPart(wholeCompose)]
  }
  const whole = tryParseSpec(src)
  if (whole && looksLikeBareObject(src)) {
    return [chartPart(whole)]
  }

  const parts = []
  let i = 0

  while (i < src.length) {
    const rest = src.slice(i)
    const open = rest.match(FENCE_OPEN)
    if (!open) {
      const embedded = tryEmbeddedAny(rest)
      if (embedded) {
        if (embedded.before) parts.push({ kind: 'md', text: embedded.before })
        parts.push(embedded.part)
        if (embedded.after) parts.push({ kind: 'md', text: embedded.after })
      } else if (rest) {
        parts.push({ kind: 'md', text: rest })
      }
      break
    }

    const openIdx = i + open.index
    if (openIdx > i) {
      parts.push({ kind: 'md', text: src.slice(i, openIdx) })
    }

    const lang = String(open[1] || '').toLowerCase()
    let bodyStart = openIdx + open[0].length
    if (src[bodyStart] === '\r') bodyStart++
    if (src[bodyStart] === '\n') bodyStart++

    const closeIdx = src.indexOf('```', bodyStart)
    const raw = closeIdx === -1 ? src.slice(bodyStart) : src.slice(bodyStart, closeIdx)
    const part = classifyFenceBody(lang, raw, closeIdx !== -1)
    if (part) {
      parts.push(part)
    } else if (closeIdx === -1) {
      parts.push({ kind: 'md', text: src.slice(openIdx) })
      break
    } else {
      parts.push({ kind: 'md', text: src.slice(openIdx, closeIdx + 3) })
    }

    if (closeIdx === -1) break
    i = closeIdx + 3
    while (src[i] === '`') i++
    if (src[i] === '\r') i++
    if (src[i] === '\n') i++
  }

  return mergeMd(parts)
}

function classifyFenceBody (lang, raw, closed) {
  const composeSpec = parseComposeChartSpec(raw)
  if (composeSpec) return composeChartPart(composeSpec)
  const spec = tryParseSpec(raw)
  if (spec) return chartPart(spec)
  if (!closed || lang === 'json') return null
  if (lang === 'compose-chart' || lang === 'chart' || lang === 'echarts') {
    return { kind: 'chart-error' }
  }
  return null
}

function mergeMd (parts) {
  const out = []
  for (const part of parts) {
    const prev = out[out.length - 1]
    if (part.kind === 'md' && prev && prev.kind === 'md') {
      prev.text += part.text
    } else {
      out.push(part)
    }
  }
  return out
}

export function specToEchartsOption (spec) {
  const rawType = String(spec?.type || 'bar').toLowerCase()
  const doughnut = rawType === 'doughnut'
  const pie = rawType === 'pie' || doughnut
  const seriesType = pie ? 'pie' : (rawType === 'line' ? 'line' : 'bar')

  let labels = asStringList(spec?.labels).slice(0, MAX_CATEGORIES)
  const seriesIn = Array.isArray(spec?.series) ? spec.series.slice(0, MAX_SERIES) : []
  const colors = asStringList(spec?.colors)

  if (!labels.length && seriesIn[0] && Array.isArray(seriesIn[0].data)) {
    labels = seriesIn[0].data.map((_, idx) => String(idx + 1)).slice(0, MAX_CATEGORIES)
  }

  const series = seriesIn.map((s, idx) => {
    const name = s && s.name != null ? String(s.name) : (idx === 0 ? 'count' : `series ${idx + 1}`)
    const values = asNumList(s && s.data).slice(0, labels.length)
    while (values.length < labels.length) values.push(0)

    if (pie) {
      return {
        name,
        type: 'pie',
        radius: doughnut ? ['40%', '80%'] : '70%',
        center: ['50%', '55%'],
        data: labels.map((label, i) => ({ name: label, value: values[i] })),
      }
    }

    const item = { name, type: seriesType, data: values }
    if (colors[idx]) item.itemStyle = { color: colors[idx] }
    return item
  }).filter(Boolean)

  const option = {
    tooltip: { trigger: pie ? 'item' : 'axis' },
    series,
  }

  if (spec?.title) {
    option.title = {
      text: String(spec.title),
      left: 'center',
      textStyle: { fontSize: 13, fontWeight: 600 },
    }
  }

  if (colors.length) {
    option.color = colors
  }

  const showLegend = pie || series.length > 1
  if (showLegend) {
    option.legend = { type: 'scroll', bottom: 0 }
  }

  if (!pie) {
    option.grid = {
      left: 44,
      right: 16,
      top: spec?.title ? 44 : 28,
      bottom: showLegend ? 40 : 28,
    }
    option.xAxis = { type: 'category', data: labels, axisLabel: { hideOverlap: true } }
    option.yAxis = { type: 'value', name: spec?.yName ? String(spec.yName) : undefined }
  }

  return option
}

export function replaceChartFences (text) {
  const src = String(text || '')
  if (looksLikeBareObject(src)) {
    const composeSpec = parseComposeChartSpec(src)
    if (composeSpec) return composeSpec.title ? `[chart: ${composeSpec.title}]` : '[chart]'
    const spec = tryParseSpec(src)
    if (spec) return spec?.title ? `[chart: ${spec.title}]` : '[chart]'
  }
  const closed = src.replace(/```\s*(?:compose-chart|chart|echarts|json)\b[^\n]*\r?\n([\s\S]*?)```/gi, (full, body, offset, hay) => {
    const open = hay.slice(offset, offset + 40).toLowerCase()
    if (open.includes('compose-chart')) {
      const spec = parseComposeChartSpec(body)
      if (!spec) return full
      return spec.title ? `[chart: ${spec.title}]` : '[chart]'
    }
    const spec = tryParseSpec(body)
    if (!spec) return full
    return `[chart: ${fenceTitle(body)}]`
  })
  return closed.replace(/```\s*(?:compose-chart|chart|echarts)\b[^\n]*\r?\n[\s\S]*$/gi, (full) => {
    if (/```\s*compose-chart/i.test(full)) {
      const spec = parseComposeChartSpec(full.replace(/^```[^\n]*\n/, ''))
      return spec?.title ? `[chart: ${spec.title}]` : '[chart]'
    }
    return '[chart]'
  })
}

function composeChartPart (spec) {
  return { kind: 'compose-chart', spec }
}

function chartPart (spec) {
  return { kind: 'chart', spec, option: specToEchartsOption(spec) }
}

function parseComposeChartSpec (raw) {
  const parsed = parseJSONObject(String(raw || '').trim()) || parseJSONObject(extractObject(String(raw || '')))
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return null
  if (Array.isArray(parsed.series)) return null
  const chartID = firstDefined(parsed.chartID, parsed.chartId, parsed.chart_id)
  if (!chartID || chartID === '0') return null
  const namespaceID = firstDefined(parsed.namespaceID, parsed.namespaceId, parsed.namespace_id)
  return {
    chartID,
    namespaceID,
    title: parsed.title != null ? String(parsed.title) : '',
  }
}

function firstDefined (...vals) {
  for (const v of vals) {
    if (v == null || v === '') continue
    const s = String(v).trim()
    if (s && s !== 'undefined' && s !== 'null') return s
  }
  return ''
}

function looksLikeBareObject (text) {
  const s = String(text || '').trim()
  return s.startsWith('{') && s.endsWith('}')
}

function tryEmbeddedAny (text) {
  const obj = extractObject(text)
  if (!obj) return null
  const start = text.indexOf('{')
  const end = text.lastIndexOf('}')
  if (start < 0 || end <= start) return null
  const composeSpec = parseComposeChartSpec(obj)
  const spec = composeSpec ? null : tryParseSpec(obj)
  if (!composeSpec && !spec) return null
  return {
    part: composeSpec ? composeChartPart(composeSpec) : chartPart(spec),
    before: text.slice(0, start),
    after: text.slice(end + 1),
  }
}

function tryParseSpec (raw) {
  const s = String(raw || '').trim()
  if (!s) return null
  const parsed = parseJSONObject(s) || parseJSONObject(extractObject(s))
  return validateSpec(parsed)
}

function parseJSONObject (s) {
  if (!s) return null
  const candidates = [s, sanitizeJSON(s)]
  for (const cand of candidates) {
    try {
      const v = JSON.parse(cand)
      if (v && typeof v === 'object' && !Array.isArray(v)) return v
    } catch {}
  }
  return null
}

function sanitizeJSON (s) {
  return String(s || '')
    .replace(/[\u201C\u201D\u00AB\u00BB]/g, '"')
    .replace(/[\u2018\u2019]/g, "'")
    .replace(/,\s*([}\]])/g, '$1')
}

function extractObject (s) {
  const start = s.indexOf('{')
  const end = s.lastIndexOf('}')
  if (start < 0 || end <= start) return ''
  return s.slice(start, end + 1)
}

function validateSpec (spec) {
  if (!spec) return null
  const hasLabels = Array.isArray(spec.labels)
  const hasSeries = Array.isArray(spec.series) && spec.series.length > 0
  if (!hasSeries) return null
  if (!hasLabels && !Array.isArray(spec.series[0] && spec.series[0].data)) return null
  return spec
}

function fenceTitle (body) {
  const spec = tryParseSpec(body)
  if (spec && spec.title) return String(spec.title)
  const m = /"title"\s*:\s*"((?:\\.|[^"\\])*)"/.exec(body)
  return (m && m[1]) ? m[1] : 'chart'
}

function asStringList (v) {
  if (!Array.isArray(v)) return []
  return v.map(x => {
    if (x == null) return ''
    if (typeof x === 'object') return String(x.name || x.label || x.value || '')
    return String(x)
  })
}

function asNumList (v) {
  if (!Array.isArray(v)) return []
  return v.map(toNum)
}

function toNum (v) {
  if (typeof v === 'number' && Number.isFinite(v)) return v
  const n = Number(String(v).replace(',', '.'))
  return Number.isFinite(n) ? n : 0
}
