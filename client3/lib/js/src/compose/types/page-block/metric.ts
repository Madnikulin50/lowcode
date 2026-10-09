import { PageBlock, PageBlockInput, Registry } from './base'
import lodash from 'lodash'
const { merge } = lodash
import { Apply } from '../../../cast'
import { Options as PageBlockRecordListOptions } from './record-list'
import { dimensionFunctions } from '../chart/util'
import { buildPeriodFilters, buildEqualityFilter, PeriodGranularity, PeriodCompareMode } from '../period'
const kind = 'Metric'

type Reporter = (p: ReporterParams) => Promise<any>

export type MetricRole = 'default' | 'title' | 'badge' | 'meta' | 'hero' | 'balloon' | 'topK'

export interface MetricSection {
  title: string;
  /** Indices into options.metrics */
  metrics: number[];
}

interface DrillDown {
  enabled: boolean;
  blockID: string;
  recordListOptions: Partial<PageBlockRecordListOptions>;
}

interface ReporterParams {
  moduleID: string;
  filter?: string;
  metrics?: string;
  dimensions: string;
}

interface Threshold {
  value: number;
  variant: string;
  /** Optional icon when value matches this threshold: arrow-up|arrow-down|arrow-right|alert|alert-circle|'' */
  icon?: string;
}

interface Style {
  color: string;
  colorThresholds: Threshold[];
  backgroundColor: string;
  fontSize?: string;
}

interface Metric {
  label: string;
  moduleID: string;
  dimensionField: string;
  dateFormat?: string;
  filter?: string;
  bucketSize?: string;
  metricField: string;
  operation: string;
  expression?: string;
  numberFormat?: string;
  prefix?: string;
  suffix?: string;
  transformFx?: string;
  showLabel?: boolean;
  /** Visual role — aligns with Record field roles */
  role?: MetricRole;
  /** When role is balloon — stretch pill to full block width */
  balloonFullWidth?: boolean;

  /** Breakdown: max rows to keep after sorting (0/undefined = unlimited) */
  topN?: number;
  /** Breakdown: how to order rows before topN is applied */
  sortDirection?: 'value-desc' | 'value-asc' | 'label-asc' | 'label-desc';
  /**
   * Pairwise comparison: exactly two dimensionField values to compare
   * head-to-head instead of a full breakdown list.
   */
  compareValues?: [string, string];

  /** Period-over-period comparison */
  periodCompareEnabled?: boolean;
  /** DateTime-kind field the comparison windows are evaluated against */
  periodDateField?: string;
  periodGranularity?: PeriodGranularity;
  periodCompareMode?: PeriodCompareMode;
  /** Which direction of change should read as "good" (controls trend color) */
  trendPositive?: 'increase' | 'decrease';

  valueStyle?: Style;
  drillDown: DrillDown;
}

const defaultMetric: Readonly<Metric> = Object.freeze({
  label: '',
  moduleID: '',
  dimensionField: '',
  dateFormat: '',
  filter: '',
  bucketSize: '',
  metricField: '',
  operation: '',
  expression: '',
  numberFormat: '',
  prefix: '',
  suffix: '',
  transformFx: '',
  showLabel: false,
  role: 'default',
  balloonFullWidth: false,

  topN: 0,
  sortDirection: 'value-desc',
  compareValues: undefined,

  periodCompareEnabled: false,
  periodDateField: '',
  periodGranularity: 'month',
  periodCompareMode: 'previous-period',
  trendPositive: 'increase',

  valueStyle: {
    backgroundColor: '#FFFFFF00',
    colorThresholds: [],
    color: '#000000',
    fontSize: undefined,
  },

  drillDown: {
    enabled: false,
    blockID: '',
    recordListOptions: {
      fields: [],
    },
  },
})

interface Options {
  metrics: Array<Metric>;
  refreshRate: number;
  showRefresh: boolean;
  /** Render metrics like Record fields (shared visual language) */
  likeRecordList: boolean;
  /** @deprecated use itemsPerRow */
  recordFieldLayoutOption: string;
  horizontalFieldLayoutEnabled: boolean;
  /**
   * How many MetricItems per row in the default/body grid:
   * '1' | '2' | '4' | 'auto'
   */
  itemsPerRow: '1' | '2' | '4' | 'auto';
  magnifyOption: string;
  density: 'comfortable' | 'compact';
  hideEmptyMetrics: boolean;
  showEmptyPlaceholder: boolean;
  sections: MetricSection[];
}

const defaults: Readonly<Options> = Object.freeze({
  metrics: [],
  refreshRate: 0,
  showRefresh: false,
  likeRecordList: true,
  recordFieldLayoutOption: 'default',
  horizontalFieldLayoutEnabled: false,
  itemsPerRow: '1',
  magnifyOption: '',
  density: 'comfortable',
  hideEmptyMetrics: false,
  showEmptyPlaceholder: true,
  sections: [],
})

function normalizeItemsPerRow (v: unknown, legacy?: string): Options['itemsPerRow'] {
  // Select/JSON may yield numbers (2) as well as strings ('2')
  const s = v == null ? '' : String(v)
  if (s === '1' || s === '2' || s === '4' || s === 'auto') return s
  // Migrate old Record-style layout keys
  if (legacy === 'wrap') return '2'
  if (legacy === 'noWrap') return 'auto'
  return '1'
}

export class PageBlockMetric extends PageBlock {
  readonly kind = kind

  options: Options = {
    ...defaults,
    metrics: [],
    sections: [],
  }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return
    Apply(this.options, o, Number, 'refreshRate')
    Apply(this.options, o, Boolean, 'showRefresh', 'horizontalFieldLayoutEnabled', 'likeRecordList', 'hideEmptyMetrics', 'showEmptyPlaceholder')
    Apply(this.options, o, String, 'recordFieldLayoutOption', 'magnifyOption', 'density')

    this.options.itemsPerRow = normalizeItemsPerRow(
      (o as any).itemsPerRow,
      o.recordFieldLayoutOption || this.options.recordFieldLayoutOption,
    )

    if (o.metrics) {
      this.options.metrics = o.metrics.map((m) => merge({}, defaultMetric, m))
    }

    if (Array.isArray(o.sections)) {
      this.options.sections = o.sections.map(s => ({
        title: s?.title || '',
        metrics: Array.isArray(s?.metrics) ? s.metrics.map(Number).filter(n => Number.isFinite(n)) : [],
      }))
    }
  }

  /**
   * Helper function to fetch and parse reporter's reports.
   *
   * Returns an array of rows. In the plain scalar case (no dimension, no
   * comparison) this is always exactly one `{value}` row, unchanged from
   * before — every previously-saved metric renders identically. The three
   * "dynamics" modes below are mutually exclusive per metric, checked in
   * priority order:
   *   1. period-over-period comparison (periodCompareEnabled)
   *   2. pairwise dimension comparison (dimensionField + 2 compareValues)
   *   3. dimension breakdown (dimensionField alone)
   */
  async fetch ({ m }: { m: Metric }, reporter: Reporter): Promise<object> {
    if (m.periodCompareEnabled && m.periodDateField) {
      return this.fetchPeriodCompare(m, reporter)
    }

    const compareValues = (m.compareValues || []).filter(v => v !== undefined && v !== null && v !== '')
    if (m.dimensionField && compareValues.length === 2) {
      return this.fetchPairwiseCompare(m, compareValues as [string, string], reporter)
    }

    if (m.dimensionField) {
      return this.fetchBreakdown(m, reporter)
    }

    const w = await reporter(this.formatParams(m))
    return [{ value: this.collapseToScalar(w, m.operation, m.transformFx) }]
  }

  private async fetchBreakdown (m: Metric, reporter: Reporter): Promise<object> {
    const w = await reporter(this.formatParams(m, { dimensions: this.dimensionExpr(m) }))
    let rows = (w || []).map((r: any) => ({
      label: r.dimension_0,
      value: this.applyTransform(r.rp !== undefined ? r.rp : r.count, m.transformFx),
    }))

    rows = this.sortRows(rows, m.sortDirection)
    if (m.topN) rows = rows.slice(0, m.topN)

    return rows
  }

  private async fetchPairwiseCompare (m: Metric, [a, b]: [string, string], reporter: Reporter): Promise<object> {
    const eqFilter = (v: string) => this.andFilters(m.filter, buildEqualityFilter(m.dimensionField, v))
    const [wa, wb] = await Promise.all([
      reporter(this.formatParams(m, { filter: eqFilter(a) })),
      reporter(this.formatParams(m, { filter: eqFilter(b) })),
    ])
    const va = this.collapseToScalar(wa, m.operation, m.transformFx)
    const vb = this.collapseToScalar(wb, m.operation, m.transformFx)

    return [
      { label: a, value: va },
      { label: b, value: vb, previousValue: va, delta: vb - va, deltaPct: va ? (vb - va) / Math.abs(va) : null },
    ]
  }

  private async fetchPeriodCompare (m: Metric, reporter: Reporter): Promise<object> {
    const { currentFilter, previousFilter } = buildPeriodFilters({
      field: m.periodDateField as string,
      granularity: m.periodGranularity || 'month',
      mode: m.periodCompareMode || 'previous-period',
    })

    const [wCur, wPrev] = await Promise.all([
      reporter(this.formatParams(m, { filter: this.andFilters(m.filter, currentFilter) })),
      reporter(this.formatParams(m, { filter: this.andFilters(m.filter, previousFilter) })),
    ])
    const current = this.collapseToScalar(wCur, m.operation, m.transformFx)
    const previous = this.collapseToScalar(wPrev, m.operation, m.transformFx)

    return [{
      value: current,
      previousValue: previous,
      delta: current - previous,
      deltaPct: previous ? (current - previous) / Math.abs(previous) : null,
    }]
  }

  /** Collapses a reporter response into one scalar, per the metric's operation. */
  private collapseToScalar (w: any[], operation: string, transformFx?: string): number {
    const datasets = (w || []).map((r: any) => r.rp !== undefined ? r.rp : r.count)

    let rtr: number
    if (operation === 'max') {
      rtr = datasets.sort((a: number, b: number) => b - a)[0]
    } else if (operation === 'min') {
      rtr = datasets.sort((a: number, b: number) => a - b)[0]
    } else if (operation === 'avg') {
      rtr = datasets.reduce((acc: number, cur: number) => acc + cur, 0) / datasets.length
    } else {
      rtr = datasets.reduce((acc: number, cur: number) => acc + cur, 0)
    }

    return this.applyTransform(rtr, transformFx)
  }

  private applyTransform (value: number, transformFx?: string): number {
    if (!transformFx) return value
    // eslint-disable-next-line no-new-func
    return (new Function('v', `return ${transformFx}`))(value)
  }

  private sortRows (rows: Array<{ label: string; value: number }>, sortDirection?: Metric['sortDirection']): Array<{ label: string; value: number }> {
    const sorted = [...rows]
    switch (sortDirection) {
      case 'value-asc':
        return sorted.sort((a, b) => a.value - b.value)
      case 'label-asc':
        return sorted.sort((a, b) => String(a.label).localeCompare(String(b.label)))
      case 'label-desc':
        return sorted.sort((a, b) => String(b.label).localeCompare(String(a.label)))
      case 'value-desc':
      default:
        return sorted.sort((a, b) => b.value - a.value)
    }
  }

  /** ANDs two QL filter fragments, omitting either side if empty. */
  private andFilters (base: string | undefined, extra: string): string {
    const b = (base || '').trim()
    if (!b) return extra
    return `(${b}) AND (${extra})`
  }

  /** Converts a dimensionField + bucketSize pair into a raw dimension expression. */
  private dimensionExpr (m: Metric): string {
    if (!m.dimensionField) return ''
    return dimensionFunctions.convert({ field: m.dimensionField, modifier: m.bucketSize || 'none' })
  }

  /**
   * Helper to construct reporter's params. `overrides` lets the
   * comparison/breakdown fetch paths above reuse this for their own
   * filter/dimension variants without duplicating the metrics-string logic.
   */
  private formatParams ({ moduleID, filter, metricField, operation = '', expression = '' }: Metric, overrides: { filter?: string; dimensions?: string } = {}): ReporterParams {
    let metrics = ''

    if (operation && metricField && metricField !== 'count') {
      if (metricField !== 'number_expression') {
        metrics = `${operation}(${metricField}) AS rp`
      } else {
        metrics = `(${expression}) AS rp`
      }
    }

    return {
      moduleID,
      filter: overrides.filter !== undefined ? overrides.filter : filter,
      metrics,
      // Scalar aggregate by default: empty group. deletedAt/createdAt are
      // omitted on external DAL tables (receipt_positions, traffic,
      // stores, …) and the server skip only matches the bare ident — not
      // YEAR()/DATE_FORMAT() — so dimension breakdown deliberately opts in
      // per-metric via `overrides.dimensions` rather than being default-on.
      dimensions: overrides.dimensions !== undefined ? overrides.dimensions : '',
    }
  }

  makeMetric () {
    return merge({}, defaultMetric)
  }
}

Registry.set(kind, PageBlockMetric)
