import { PageBlock, PageBlockInput, Registry } from './base'
import lodash from 'lodash'
const { merge } = lodash
import { Apply } from '../../../cast'
import { Options as PageBlockRecordListOptions } from './record-list'
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
   */
  async fetch ({ m }: { m: Metric }, reporter: Reporter): Promise<object> {
    const w = await reporter(this.formatParams(m))
    const datasets = w.map((r: any) => r.rp !== undefined ? r.rp : r.count)

    let rtr: number
    if (m.operation === 'max') {
      rtr = datasets.sort((a: number, b: number) => b - a)[0]
    } else if (m.operation === 'min') {
      rtr = datasets.sort((a: number, b: number) => a - b)[0]
    } else if (m.operation === 'avg') {
      rtr = datasets.reduce((acc: number, cur: number) => acc + cur, 0) / datasets.length
    } else {
      rtr = datasets.reduce((acc: number, cur: number) => acc + cur, 0)
    }

    if (m.transformFx) {
      // eslint-disable-next-line no-new-func
      rtr = (new Function('v', `return ${m.transformFx}`))(rtr)
    }

    return [{ value: rtr }]
  }

  /**
   * Helper to construct reporter's params
   */
  private formatParams ({ moduleID, filter, metricField, operation = '', expression = '' }: Metric): ReporterParams {
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
      filter,
      metrics,
      // Scalar aggregate: empty group. deletedAt/createdAt are omitted on
      // external DAL tables (receipt_positions, traffic, stores, …) and the
      // server skip only matches the bare ident — not YEAR()/DATE_FORMAT().
      dimensions: '',
    }
  }

  makeMetric () {
    return merge({}, defaultMetric)
  }
}

Registry.set(kind, PageBlockMetric)
