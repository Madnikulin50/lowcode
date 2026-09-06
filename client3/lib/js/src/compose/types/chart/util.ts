import numeral from 'numeral'
import * as fmt from '../../../formatting'

export const rgbaRegex = /^rgba\((\d+),.*?(\d+),.*?(\d+),.*?(\d*\.?\d*)\)$/

const ln = (n: number) => Math.round(n < 0 ? 255 + n : (n > 255) ? n - 255 : n)
export const toRGBA = ([r, g, b, a]: Array<number>) =>
  `rgba(${ln(r)}, ${ln(g)}, ${ln(b)}, ${a})`

export enum ChartType {
  pie = 'pie',
  bar = 'bar',
  line = 'line',
  doughnut = 'doughnut',
  funnel = 'funnel',
  gauge = 'gauge',
  radar = 'radar',
  scatter = 'scatter',
  sankey = 'sankey',
  heatmap = 'heatmap',
  waterfall = 'waterfall',
  boxplot = 'boxplot',
  graph = 'graph',
  candlestick = 'candlestick',
  map = 'map',
  sunburst = 'sunburst',
  parallel = 'parallel',
  calendar = 'calendar',
  gantt = 'gantt',
}

export interface TemporalDataPoint {
  t: Date;
  y: number;
}

export interface KV {
  [_: string]: any;
}

export interface FormatData {
  format?: string,
  prefix?: string,
  suffix?: string,
  presetFormat?: string,
}

export interface Tooltip {
  formatting?: string;
  labelsNextToPartition?: boolean;
}

export interface TooltipParams {
  seriesName?: string;
  name?: string;
  value?: string | number;
  percent?: string | number;
  marker?: string;
}

export interface Dimension {
  meta?: KV;
  conditions: object;
  field?: string;
  modifier?: string;
  default?: string;
  skipMissing?: boolean;
  timeLabels?: boolean;
  autoSkip?: boolean;
  rotateLabel?: number;
}

export interface Metric {
  axisType?: string;
  field?: string;
  fixTooltips?: boolean;
  /** Where to place value labels when fixTooltips is on: 'top' (outside) or 'inside'. Default 'top'. */
  valueLabelPosition?: 'top' | 'inside';
  relativeValue?: boolean;
  cumulative?: boolean;
  type?: ChartType;
  alias?: string;
  aggregate?: string;
  modifier?: string;
  fx?: string;
  backgroundColor?: string;
  symbol?: string;
  formatting: FormatData;
  [_: string]: any;
}

export interface YAxis {
  axisPosition?: string;
  axisType?: string;
  beginAtZero?: boolean;
  label?: string;
  labelPosition?: string;
  min?: string;
  max?: string;
  rotateLabel?: number;
  horizontal?: boolean;
  formatting: FormatData;
}

export interface ChartOffset {
  top?: string;
  right?: string;
  bottom?: string;
  left?: string;
  isDefault?: boolean;
}

export interface Position {
  isDefault?: boolean;
  top?: string;
  right?: string;
  bottom?: string;
  left?: string;
}

export interface Legend {
  isHidden?: boolean;
  orientation?: string;
  align?: string;
  isScrollable?: boolean;
  isDefault?: boolean;
  position?: Position;
}

export interface Report {
  moduleID?: string|null;
  filter?: string|null;
  dimensions?: Array<Dimension>;
  metrics?: Array<Metric>;
  yAxis?: YAxis;
  tooltip?: Tooltip;
  legend?: Legend;
  offset?: ChartOffset;
  anomaly?: AnomalyConfig;
  compare?: CompareConfig;
  forecast?: ForecastConfig;
}

export interface AnomalyConfig {
  enabled: boolean;
  method: 'zscore' | 'iqr' | 'fixed' | 'pct_change';
  threshold: number;
  min?: number;
  max?: number;
  color: string;
}

/**
 * Period-over-period comparison for bar/line reports (Chart's analogue of
 * PageBlockMetric's periodCompare* fields — see page-block/metric.ts and
 * ../period.ts). When enabled, this replaces the report's own `dimensions`
 * configuration: the x-axis becomes day-of-period (or day-of-week for
 * `week` granularity), and two series per metric are rendered — one for
 * the current period, one for the previous — aligned on that shared axis
 * so e.g. day 15 of this month lines up with day 15 of last month.
 */
export interface CompareConfig {
  enabled: boolean;
  dateField: string;
  granularity: 'week' | 'month' | 'quarter' | 'year';
  mode: 'previous-period' | 'year-over-year';
  /** Legend label for the current-period series; falls back to a generic default when empty. */
  currentLabel?: string;
  /** Legend label for the previous-period series; falls back to a generic default when empty. */
  previousLabel?: string;
}

/**
 * Forward projection of a temporal bar/line report (Chart's `forecast`,
 * alongside `anomaly` and `compare` above). Unlike `compare`, it does not
 * replace the report's dimension — it only makes sense on top of a genuinely
 * temporal x-axis, and extends it with `periods` future points continuing
 * the existing series. Optionally adds an optimistic/pessimistic band around
 * that projection (see computeForecast below).
 */
export interface ForecastConfig {
  enabled: boolean;
  method: 'linear' | 'moving-average' | 'exp-smoothing';
  /** Number of future points to project past the last known data point. */
  periods: number;
  /** Show an optimistic/pessimistic scenario around the main projection. */
  scenarios: boolean;
  /** 'lines': three dashed/dotted lines. 'band': shaded area between the two scenarios, main projection dashed on top. */
  scenarioStyle: 'lines' | 'band';
  /** 'auto' derives the scenario spread from historical volatility; 'manual' uses deviationPct. */
  deviation: 'auto' | 'manual';
  deviationPct?: number;
  color?: string;
}

export interface ChartToolbox {
  saveAsImage: boolean;
  showDataTable: boolean;
  timeline: string;
}

export interface ChartConfig {
  reports?: Array<Report>;
  colorScheme?: string;
  noAnimation?: boolean;
  gradient?: 'lightToDark' | 'darkToLight' | '';
  toolbox?: ChartToolbox;
  description?: string;
  help?: string;
}

export const aggregateFunctions = [
  {
    value: 'SUM',
    text: 'sum',
  },
  {
    value: 'MAX',
    text: 'max',
  },
  {
    value: 'MIN',
    text: 'min',
  },
  {
    value: 'AVG',
    text: 'avg',
  },
  {
    value: 'STD',
    text: 'std',
  },
]

interface DimensionFunction {
  text: string;
  value: string;
  convert: (f: string) => string;
}

export class DimensionFunctions<T> extends Array<T> {
  private constructor (items?: Array<T>) {
    super(...(items || []))
  }

  static create<T> (): DimensionFunctions<T> {
    return Object.create(DimensionFunctions.prototype)
  }

  public lookup (d: any): any {
    return this.find((f: any) => d.modifier === f.value)
  }

  public convert (d: any): any {
    return (this.lookup(d) || {}).convert(d.field)
  }
}

export const dimensionFunctions: DimensionFunctions<DimensionFunction> = DimensionFunctions.create<DimensionFunction>()
dimensionFunctions.push(...[
  {
    text: 'none',
    value: '(no grouping / buckets)',
    convert: (f: string) => f,
  },

  {
    text: 'date',
    value: 'DATE',
    convert: (f: string) => `DATE(${f})`,
  },

  {
    text: 'week',
    value: 'WEEK',
    convert: (f: string) => `WEEK(${f})`,
  },

  {
    text: 'month',
    value: 'MONTH',
    convert: (f: string) => `DATE_FORMAT(${f}, '%Y-%m-01')`,
  },

  {
    text: 'quarter',
    value: 'QUARTER',
    convert: (f: string) => `QUARTER(${f})`,
  },

  {
    text: 'year',
    value: 'YEAR',
    convert: (f: string) => `DATE_FORMAT(${f}, '%Y-01-01')`,
  },
])

export const predefinedFilters = [
  {
    value: 'YEAR(createdAt) = YEAR(NOW())',
    text: 'recordsCreatedThisYear',
  },
  {
    value: 'YEAR(createdAt) = YEAR(NOW()) - 1',
    text: 'recordsCreatedLastYear',
  },

  {
    value: 'YEAR(createdAt) = YEAR(NOW()) AND QUARTER(createdAt) = QUARTER(NOW())',
    text: 'recordsCreatedThisQuarter',
  },
  {
    value: 'YEAR(createdAt) = YEAR(NOW()) AND QUARTER(createdAt) = QUARTER(DATE_SUB(NOW(), INTERVAL 3 MONTH)',
    text: 'recordsCreatedLastQuarter',
  },

  {
    value: 'DATE_FORMAT(createdAt, \'%Y-%m\') = DATE_FORMAT(NOW(), \'%Y-%m\')',
    text: 'recordsCreatedThisMonth',
  },
  {
    value: 'DATE_FORMAT(createdAt, \'%Y-%m\') = DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), \'%Y-%m\')',
    text: 'recordsCreatedLastMonth',
  },
]

dimensionFunctions.lookup = d => dimensionFunctions.find(f => d.modifier === f.value) || dimensionFunctions[0]
dimensionFunctions.convert = d => dimensionFunctions.lookup(d).convert(d.field)

export const isRadialChart = ({ type }: KV) => type === 'doughnut' || type === 'pie'
export const hasRelativeDisplay = ({ type }: KV) => isRadialChart({ type })

// Makes a standardized alias from modifier or dimension report option
export const makeAlias = ({ alias, aggregate, modifier, field }: Partial<Metric>) => alias || `${aggregate || modifier || 'none'}_${field}`.toLocaleLowerCase()

export function formatChartValue (value: string | number, formatting?: FormatData): string {
  let n: number | string = 0 || ''
  // if value contains alphabetic chars parseFloat() will return NaN
  // and n will equal 0
  const containsAlphabeticChars = isNaN(Number(value))
  let result = ''

  if (!containsAlphabeticChars) {
    switch (typeof value) {
      case 'string':
        n = parseFloat(value)
        break
      case 'number':
        n = value
        break
      default:
        n = 0
    }

    if (formatting?.format) {
      result = numeral(n).format(formatting.format)
    } else {
      result = fmt.number(n)
    }
  }

  if (formatting?.presetFormat === 'accounting') {
    result = fmt.accountingNumber(Number(n))
  }

  return ` ${formatting?.prefix ?? ''} ${result || value} ${formatting?.suffix ?? ''}`
}

export function formatChartTooltip (tooltip: string, params: TooltipParams): string {
  const { seriesName = '', name = '', value = '', percent = '' } = params

  return tooltip
    .replace('{a}', seriesName)
    .replace('{b}', name)
    .replace('{c}', value.toString())
    .replace('{d}', percent.toString())
}

export function defFormatData (): FormatData {
  return Object.assign({}, {
    presetFormat: 'custom',
    prefix: '',
    suffix: '',
    format: '',
  })
}

export function detectAnomalies(values: number[], cfg: AnomalyConfig): boolean[] {
  if (!cfg.enabled || !values.length) return values.map(() => false)

  const n = values.length
  const flags = new Array(n).fill(false)

  if (cfg.method === 'zscore') {
    const mean = values.reduce((a, b) => a + b, 0) / n
    const std = Math.sqrt(values.reduce((s, v) => s + (v - mean) ** 2, 0) / n) || 1
    values.forEach((v, i) => { if (Math.abs(v - mean) / std > cfg.threshold) flags[i] = true })
  }

  if (cfg.method === 'iqr') {
    const sorted = [...values].sort((a, b) => a - b)
    const q1 = sorted[Math.floor(n * 0.25)]
    const q3 = sorted[Math.floor(n * 0.75)]
    const iqr = q3 - q1
    values.forEach((v, i) => { if (v < q1 - 1.5 * iqr || v > q3 + 1.5 * iqr) flags[i] = true })
  }

  if (cfg.method === 'fixed') {
    values.forEach((v, i) => {
      if ((cfg.min != null && v < cfg.min) || (cfg.max != null && v > cfg.max)) flags[i] = true
    })
  }

  if (cfg.method === 'pct_change') {
    values.forEach((v, i) => {
      if (i === 0) return
      const prev = values[i - 1]
      if (prev === 0) { if (v !== 0) flags[i] = true; return }
      if (Math.abs(v - prev) / prev > cfg.threshold) flags[i] = true
    })
  }

  return flags
}

export interface ForecastResult {
  main: number[];
  positive?: number[];
  negative?: number[];
}

/**
 * Fits one of the supported trend models to `values` and returns a function
 * predicting the value at any point index — including indexes past the end
 * of the input, i.e. the forecast itself. Deliberately simple, explainable
 * math (no ML), same spirit as detectAnomalies() above.
 */
function fitTrend (values: number[], method: ForecastConfig['method']): (x: number) => number {
  const n = values.length

  if (method === 'moving-average') {
    // Project forward from the last value using the average point-to-point
    // change over a short recent window — follows recent momentum without
    // being thrown off by one-off spikes further back.
    const window = Math.min(5, n - 1)
    let sum = 0
    for (let i = n - window; i < n; i++) sum += values[i] - values[i - 1]
    const avgDelta = sum / window
    const last = values[n - 1]
    return (x: number) => last + avgDelta * (x - (n - 1))
  }

  if (method === 'exp-smoothing') {
    // Holt's linear (double exponential smoothing): a level and a trend
    // that both adapt to recent data, so a shifting trend is picked up
    // faster than a single best-fit line over the whole history would.
    const alpha = 0.3
    const beta = 0.3
    let level = values[0]
    let trend = values[1] - values[0]
    for (let i = 1; i < n; i++) {
      const prevLevel = level
      level = alpha * values[i] + (1 - alpha) * (level + trend)
      trend = beta * (level - prevLevel) + (1 - beta) * trend
    }
    const finalTrend = trend
    const finalLevel = level
    return (x: number) => finalLevel + finalTrend * (x - (n - 1))
  }

  // 'linear' (default): ordinary least-squares fit over the point index.
  const xMean = (n - 1) / 2
  const yMean = values.reduce((a, b) => a + b, 0) / n
  let num = 0
  let den = 0
  for (let i = 0; i < n; i++) {
    num += (i - xMean) * (values[i] - yMean)
    den += (i - xMean) ** 2
  }
  const slope = den ? num / den : 0
  const intercept = yMean - slope * xMean
  return (x: number) => intercept + slope * x
}

/**
 * Projects `cfg.periods` future points from a historical numeric series.
 * Returns null when there isn't enough history to fit a trend (fewer than
 * 2 points) or forecasting is off/misconfigured.
 *
 * When `cfg.scenarios` is set, also returns an optimistic (`positive`) and
 * pessimistic (`negative`) variant around the main projection: either a
 * fixed `deviationPct` of the projected value, or (auto) a spread derived
 * from how far actual history deviated from its own fitted trend, widening
 * the further out the forecast reaches — the further ahead, the less sure
 * we can be.
 */
export function computeForecast (values: number[], cfg: ForecastConfig): ForecastResult | null {
  const n = values.length
  if (!cfg?.enabled || n < 2 || !cfg.periods || cfg.periods < 1) return null

  const trend = fitTrend(values, cfg.method)
  const main = Array.from({ length: cfg.periods }, (_, i) => trend(n + i))

  if (!cfg.scenarios) return { main }

  let widths: number[]
  if (cfg.deviation === 'manual') {
    const pct = (cfg.deviationPct ?? 10) / 100
    widths = main.map(v => Math.abs(v) * pct)
  } else {
    const residuals = values.map((v, i) => v - trend(i))
    const mean = residuals.reduce((a, b) => a + b, 0) / n
    const std = Math.sqrt(residuals.reduce((s, r) => s + (r - mean) ** 2, 0) / n)
    widths = main.map((_, i) => std * 1.5 * Math.sqrt(i + 1))
  }

  return {
    main,
    positive: main.map((v, i) => v + widths[i]),
    negative: main.map((v, i) => v - widths[i]),
  }
}

function calendarStepper (modifier?: string): ((d: Date) => Date) | null {
  switch (modifier) {
    case 'DATE': return d => { const r = new Date(d); r.setUTCDate(r.getUTCDate() + 1); return r }
    case 'WEEK': return d => { const r = new Date(d); r.setUTCDate(r.getUTCDate() + 7); return r }
    case 'MONTH': return d => { const r = new Date(d); r.setUTCMonth(r.getUTCMonth() + 1); return r }
    case 'QUARTER': return d => { const r = new Date(d); r.setUTCMonth(r.getUTCMonth() + 3); return r }
    case 'YEAR': return d => { const r = new Date(d); r.setUTCFullYear(r.getUTCFullYear() + 1); return r }
    default: return null
  }
}

/**
 * Generates `count` future labels continuing a temporal x-axis. When
 * `modifier` is one of the calendar buckets DATE/WEEK/MONTH/QUARTER/YEAR
 * (see dimensionFunctions above), steps the calendar exactly that much each
 * time — a fixed millisecond delta would drift on months/quarters/years,
 * whose lengths vary. Otherwise (an ungrouped raw datetime field, whose
 * bucket size isn't known upfront) falls back to repeating the gap between
 * the last two known labels. Returns [] when there are too few labels, or
 * they don't parse as dates, to extrapolate from.
 */
export function nextTemporalLabels (labels: string[], count: number, modifier?: string): string[] {
  const n = labels.length
  if (n < 2 || count < 1) return []

  const lastTime = new Date(labels[n - 1]).getTime()
  if (Number.isNaN(lastTime)) return []

  const stepper = calendarStepper(modifier)
  if (stepper) {
    const out: string[] = []
    let cur = new Date(lastTime)
    for (let i = 0; i < count; i++) {
      cur = stepper(cur)
      out.push(cur.toISOString())
    }
    return out
  }

  const prevTime = new Date(labels[n - 2]).getTime()
  if (Number.isNaN(prevTime)) return []
  const delta = lastTime - prevTime
  if (!delta) return []

  return Array.from({ length: count }, (_, i) => new Date(lastTime + delta * (i + 1)).toISOString())
}

const chartUtil = {
  dimensionFunctions,
  hasRelativeDisplay,
  aggregateFunctions,
  predefinedFilters,
  ChartType,
  detectAnomalies,
  computeForecast,
  nextTemporalLabels,
}

export {
  chartUtil,
}
