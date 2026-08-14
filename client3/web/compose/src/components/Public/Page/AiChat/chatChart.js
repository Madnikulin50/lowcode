const MAX_CATEGORIES = 20
const MAX_SERIES = 8
const FENCE_OPEN = /```(chart|echarts)\b[^\n]*\n/i

export function splitChartParts (text) {
  const src = String(text || '')
  const parts = []
  let i = 0

  while (i < src.length) {
    const rest = src.slice(i)
    const open = rest.match(FENCE_OPEN)
    if (!open) {
      if (rest) parts.push({ kind: 'md', text: rest })
      break
    }

    const openIdx = i + open.index
    if (openIdx > i) {
      parts.push({ kind: 'md', text: src.slice(i, openIdx) })
    }

    const bodyStart = openIdx + open[0].length
    const closeIdx = src.indexOf('```', bodyStart)
    if (closeIdx === -1) {
      const spec = tryParseSpec(src.slice(bodyStart))
      if (spec) {
        parts.push(chartPart(spec))
      } else {
        parts.push({ kind: 'md', text: src.slice(openIdx) })
      }
      break
    }

    const raw = src.slice(bodyStart, closeIdx)
    const spec = tryParseSpec(raw)
    if (spec) {
      parts.push(chartPart(spec))
    } else {
      parts.push({ kind: 'chart-error' })
    }
    i = closeIdx + 3
    if (src[i] === '\n') i++
  }

  return mergeMd(parts)
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

  if (colors.length && pie) {
    option.color = colors
  } else if (colors.length && !pie) {
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
  const closed = src.replace(/```(?:chart|echarts)\b[^\n]*\n([\s\S]*?)```/gi, (_, body) => {
    return `[chart: ${fenceTitle(body)}]`
  })
  return closed.replace(/```(?:chart|echarts)\b[^\n]*\n[\s\S]*$/gi, '[chart]')
}

function chartPart (spec) {
  return { kind: 'chart', spec, option: specToEchartsOption(spec) }
}

function tryParseSpec (raw) {
  const s = String(raw || '').trim()
  if (!s) return null
  const parsed = parseJSONObject(s) || parseJSONObject(extractObject(s))
  return validateSpec(parsed)
}

function parseJSONObject (s) {
  if (!s) return null
  try {
    const v = JSON.parse(s)
    if (v && typeof v === 'object' && !Array.isArray(v)) return v
  } catch {}
  return null
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
  return v.map(x => (x == null ? '' : String(x)))
}

function asNumList (v) {
  if (!Array.isArray(v)) return []
  return v.map(toNum)
}

function toNum (v) {
  if (typeof v === 'number' && Number.isFinite(v)) return v
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}
