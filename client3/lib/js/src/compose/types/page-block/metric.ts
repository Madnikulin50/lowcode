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
  recordFieldLayoutOption: string;
  horizontalFieldLayoutEnabled: boolean;
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
  magnifyOption: '',
  density: 'comfortable',
  hideEmptyMetrics: false,
  showEmptyPlaceholder: true,
  sections: [],
})

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
      // Since metric produces one value we want one dataset, deletedAt is the same for all existing records
      dimensions: 'deletedAt',
    }
  }

  makeMetric () {
    return merge({}, defaultMetric)
  }
}

Registry.set(kind, PageBlockMetric)
