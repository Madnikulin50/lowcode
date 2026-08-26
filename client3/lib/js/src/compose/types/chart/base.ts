import _ from 'lodash'
import {
  ChartConfig,
  Dimension,
  Metric,
  Report,
  dimensionFunctions,
  makeAlias,
  TemporalDataPoint,
  defFormatData,
} from './util'

import {
  CortezaID,
  NoID,
  ISO8601Date,
  Apply,
} from '../../../cast'

export type PartialChart = Partial<BaseChart>

// The default dataset post processing function to use.
// This one simply returns the current value.
const defaultFx = 'n'

/**
 * Chart types that support stacking (multiple series on shared axis).
 */
export function isStackableType (type?: string): boolean {
  return ['line', 'bar'].includes(type as string)
}

function chartPointValue (v: unknown): number | null {
  if (v === undefined || v === null || v === '') return null
  if (typeof v === 'object' && v !== null && 'y' in (v as Record<string, unknown>)) {
    return chartPointValue((v as TemporalDataPoint).y)
  }
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

/**
 * BaseChart represents a structure that stores any configuration data.
 * Any display and data rendering operations should be handled by any sub classes.
 */
export class BaseChart {
  public chartID = NoID
  public namespaceID = NoID
  public name = ''
  public handle = ''

  public createdAt?: Date = undefined
  public updatedAt?: Date = undefined
  public deletedAt?: Date = undefined

  public canUpdateChart = false
  public canDeleteChart = false
  public canGrant = false

  public config: ChartConfig = {}

  constructor (def: PartialChart = {}) {
    this.merge(def)
  }

  /**
   * The method performs post processing for each value in the given dataset.
   * It works with a simple equation written in javascript (example: n + m).
   * Available variables to use:
   * * n - current value
   * * m - previous value (undefined in case of the first element)
   * * r - entire data array.
   *
   * @param data Array of values in the given data set
   * @param m Metric for the given dataset
   */
  datasetPostProc (data: Array<number|TemporalDataPoint>, m: Metric): Array<number|TemporalDataPoint> {
    // Define a valid function to evaluate
    let fxRaw = (m.fx || defaultFx).trim()
    if (!fxRaw.startsWith('return')) {
      fxRaw = 'return ' + fxRaw
    }
    // eslint-disable-next-line no-new-func
    const fx = new Function('n', 'm', 'r', fxRaw)

    // Define a new array, so we don't alter the original one.
    const r = [...data]

    // Run postprocessing for all data in the given data set
    // There is a slight difference between temporal data points and categorical data points.
    if (data[0] instanceof Object) {
      // Temporal
      for (let i = 0; i < data.length; i++) {
        const a = data[i] as TemporalDataPoint
        const b = data[i - 1] as TemporalDataPoint|undefined

        const n = a.y
        let m: number|undefined
        if (i > 0) {
          m = b?.y
        }

        a.y = fx(n, m, r)
      }
    } else {
      // Categorical
      for (let i = 0; i < data.length; i++) {
        const n = data[i] as number
        let m: number|undefined
        if (i > 0) {
          m = data[i - 1] as number
        }
        data[i] = fx(n, m, r)
      }
    }

    return data
  }

  merge (c: PartialChart) {
    let conf = { ...(c.config || {}) }
    Apply(this, c, CortezaID, 'chartID', 'namespaceID')
    Apply(this, c, String, 'name', 'handle')
    Apply(this, c, ISO8601Date, 'createdAt', 'updatedAt', 'deletedAt')
    Apply(this, c, Boolean, 'canUpdateChart', 'canDeleteChart', 'canGrant')
    Apply(this, c, Object, 'config')

    if (typeof c.config === 'object') {
      // Verify & normalize
      const { reports = [], ...rest } = c.config

      conf = { reports: reports || [], ...rest }
    }

    this.config = (conf ? _.merge(this.defConfig(), conf) : false) || this.config || this.defConfig()

    this.config.reports?.forEach(report => {
      Object.assign(report, _.merge(this.defReport(), report))

      const { dimensions = [], metrics = [] } = report || {}

      report.dimensions = dimensions.map(d => {
        // Legacy support
        if (d.modifier === 'auto') {
          d.timeLabels = true
          d.modifier = '(no grouping / buckets)'
        }

        if (d.field === 'created_at') {
          d.field = 'createdAt'
        }

        return _.merge(this.defDimension(), d)
      })

      report.metrics = metrics.map(m => _.merge(this.defMetric(), m))
    })
  }

  /**
   * Checks reports validity.
   * Validates dimensions and metrics.
   * If invalid it throws an error.
   */
  isValid () {
    if (!this.config.reports || !this.config.reports.length) {
      throw new Error('notification.chart.invalidConfig.missingReports')
    }

    this.config.reports.forEach(({ moduleID, dimensions, metrics }) => {
      if (!moduleID) {
        throw new Error('notification.chart.invalidConfig.missingModuleID')
      }

      // Expecting all dimensions to have defined fields
      dimensions?.forEach((d) => {
        if (metrics?.some((m: Metric) => m.type === 'gantt')) {
          if (!d.field) {
            throw new Error('notification.chart.invalidConfig.missingDimensionsField')
          }
          return
        }
        this.dimCheck(d)
      })

      // Expecting all metrics to have defined fields
      metrics?.forEach(this.mtrCheck)
    })

    return true
  }

  /**
   * Checks validity of dimensions.
   * If invalid it throws an error
   */
  dimCheck ({ field, modifier }: Dimension) {
    if (!field) {
      throw new Error('notification.chart.invalidConfig.missingDimensionsField')
    }
    if (!modifier) {
      throw new Error('notification.chart.invalidConfig.missingDimensionsModifier')
    }
  }

  /**
   * Checks validity of metrics.
   * If invalid it throws an error
   */
  mtrCheck ({ field, aggregate, type }: Metric) {
    if (type === 'gantt') {
      return
    }
    if (!field) {
      throw new Error('notification.chart.invalidConfig.missingMetricsField')
    }
    if (field !== 'count' && !aggregate) {
      throw new Error('notification.chart.invalidConfig.missingMetricsAggregate')
    }
    if (!type) {
      throw new Error('notification.chart.invalidConfig.missingMetricsType')
    }
  }

  /**
   * Prepares params that the reporter can use for querying.
   */
  formatReporterParams ({ moduleID, metrics, dimensions, filter }: Report) {
    let dims = (dimensions || []).slice(0, 2).map((d: Dimension) => {
      const field = typeof d.field === 'string' ? d.field.trim() : ''
      // Vue 3 / empty dimension used to default to createdAt and 500 on
      // external DAL tables that omit system timestamps.
      if (!field || field === 'undefined' || field === 'null') {
        return ''
      }
      const modifier = d.modifier || '(no grouping / buckets)'
      // Ungrouped createdAt/deletedAt is Corteza's dummy scalar dimension
      // (same class as Metric's old deletedAt group). Skip it.
      if ((field === 'createdAt' || field === 'deletedAt') &&
          (modifier === '(no grouping / buckets)' || modifier === 'none' || modifier === 'auto')) {
        return ''
      }
      return dimensionFunctions.convert({ ...d, field })
    }).filter((expr: string) => !!expr)

    // If any metric has stackBy, use it as the 2nd dimension so the server
    // groups records by the stack field and returns dimension_1 values.
    if (dims.length > 0) {
      const stackBy = metrics?.find((m: Metric) => isStackableType(m.type) && m.stackBy)?.stackBy
      if (stackBy) {
        if (dims.length >= 2) dims[1] = stackBy
        else dims.push(stackBy)
      }
    }

    return {
      moduleID,
      filter,

      // Remove count (we'll get it anyway) and construct FUNC(ARG) params
      metrics: metrics?.filter((m: Metric) => m.field !== 'count').map((m: Metric) => `${m.aggregate}(${m.field}) AS ${makeAlias(m)}`).join(','),

      // Construct dimensions \w modifiers...
      // @note SQL expressions may contain commas (eg. DATE_FORMAT(field, '%Y-%m-01'))
      //       so we use ';' to separate multiple dimensions
      dimensions: dims.join(';'),
    }
  }

  /**
   * Fetcher reports defined in the given configuration with the help of the provided
   * reporter.
   */
  async fetchReports ({ reporter }: { reporter(p: any): Promise<any> }) {
    const out: Array<any> = []

    // Prepare params & filter out invalid combos (formatReporterParams will return null on invalid params)
    const reports: any = this.config.reports?.map(this.formatReporterParams)
      // Send requests to reporter (API caller)
      .map(params => reporter(params))
      // Process each result
      .map((p: any, index: number) => p.then((results: any) => {
        results = results || []
        out[index] = this.processReporterResults(results, (this.config.reports || [])[index])
      }))

    // Wait for all requests to finish and return new promise, with results
    return Promise.all(reports).then(() => new Promise(resolve => {
      resolve(out)
    }))
  }

  /**
   * Processes provided report with it's results:
   * * skip missing values, if so requested,
   * * generate labels,
   * * creates dataset for the chart.
   */
  private processReporterResults (results: Array<object> = [], report: Report): object {
    const dLabel = 'dimension_0'
    const { dimensions: [dimension] = [] } = report
    let labels: Array<string> = []

    // helper to choose between eight the provided value, default value or a generic 'undefined'
    const pickValue = (val: unknown, { default: dDft }: Dimension): unknown => {
      return val || val === 0 ? val : dDft || 'undefined'
    }

    // Skip missing values; if so requested
    if (dimension.skipMissing) {
      results = results.filter((r: any) => r[dLabel] || r[dLabel] === 0)
    }

    // Not a time dimensions, build set of labels
    labels = results.map((r: any) => pickValue(r[dLabel], dimension)) as Array<string>

    // Build data sets
    const { metrics } = report
    const stackBy = metrics?.find((m: any) => isStackableType(m.type) && m.stackBy)?.stackBy
    let datasets

    if (stackBy && results.length) {
      // Split rows into one series per group value (dimension_1)
      labels = [...new Set(labels.map(String))]
      if (dimension.timeLabels) {
        labels.sort((a: string, b: string) => (new Date(a)).getTime() - (new Date(b)).getTime())
      }
      const groups = [...new Set(results.map((r: any) => String(r.dimension_1)))]

      datasets = []
      for (const m of (metrics || [])) {
        const alias = makeAlias({ field: m.field, aggregate: m.aggregate })
        const lookup = (m as any).field === 'count' ? (m as any).field : alias
        for (const g of groups) {
          const data = labels.map(l => {
            const row = results.find((r: any) => String(pickValue(r[dLabel], dimension)) === l && String(r.dimension_1) === g)
            return row ? pickValue((row as any)[lookup], dimension) : 0
          })
          const ds = this.makeDataset(m, dimension, data, lookup)
          ds.label = g
          if (!(m as any).stack) ds.stack = 'total'
          datasets.push(ds)
        }
      }
    } else {
      datasets = metrics?.map(m => {
        const alias = makeAlias({ field: m.field, aggregate: m.aggregate })
        // For the count metric the server returns a plain 'count' column
        const lookup = (m as any).field === 'count' ? (m as any).field : alias
        const data = results.map((r: any) => {
          return pickValue(r[lookup], dimension)
        })

        // Any sub class has the ability to define how the dataset looks like.
        // this comes in handy when we want to support charts with different definitions.
        return this.makeDataset(m, dimension, data, lookup)
      })
    }

    const dropped = this.dropEmptyBarCategories(labels, datasets, dimension)
    return {
      labels: this.processLabels(dropped.labels, dimension),
      datasets: dropped.datasets,
      dimension,
      // Raw rows; used by charts that need more than one dimension (sankey, graph, heatmap, ...)
      rows: results,
    }
  }

  /**
   * With skipMissing, bar charts also drop categories whose values are all
   * zero — aggregations often return 0 for empty buckets, which still draw
   * as a blank bar if left in the series.
   */
  protected dropEmptyBarCategories (
    labels: Array<string>,
    datasets: Array<any> | undefined,
    dimension: Dimension,
  ): { labels: Array<string>, datasets: Array<any> | undefined } {
    if (!dimension?.skipMissing || !datasets?.length || !labels.length) {
      return { labels, datasets }
    }
    if (!datasets.some((ds: any) => ds.type === 'bar')) {
      return { labels, datasets }
    }

    const keep = labels.map((_, i) =>
      datasets.some((ds: any) => {
        const n = chartPointValue(ds.data?.[i])
        return n !== null && n !== 0
      }),
    )
    if (keep.every(Boolean)) {
      return { labels, datasets }
    }

    return {
      labels: labels.filter((_, i) => keep[i]),
      datasets: datasets.map((ds: any) => ({
        ...ds,
        data: (ds.data || []).filter((_: any, i: number) => keep[i]),
      })),
    }
  }

  processLabels (ll: Array<string>, d: Dimension) {
    return ll
  }

  makeDataset (m: Metric, d: Dimension, data: Array<number|any>, alias: string): any {
    throw new Error('method.makeDataset.notImplemented')
  }

  makeOptions (data?: any) {
    throw new Error('method.makeOptions.notImplemented')
  }

  plugins (mm: Array<Metric>) {
    throw new Error('method.plugins.notImplemented')
  }

  baseChartType (datasets: Array<any>) {
    throw new Error('method.baseChartType.notImplemented')
  }

  /**
   * Performs chart export; used by exporter feature.
   */
  async export (findModuleByID: ({ namespaceID, moduleID }: { namespaceID: string; moduleID: string }) => Promise<any>) {
    const { namespaceID } = this
    const copy = new BaseChart(this)
    if (copy.config?.reports) {
      await Promise.all(copy.config.reports.map(async (r: any) => {
        const { moduleID } = r
        if (moduleID) {
          const module = await findModuleByID({ namespaceID, moduleID })
          r.moduleID = module.name
          return r
        } else {
          return null
        }
      })).then((a: any) => {
        return a
      })
    }
    return copy
  }

  /**
   * Performs import; used by importer feature
   */
  import (getModuleID: (moduleID: string) => string) {
    const copy = new BaseChart(this)
    copy.config.reports = copy.config?.reports?.map(r => {
      const { moduleID } = r
      if (moduleID) {
        r.moduleID = getModuleID(moduleID)
      }
      return r
    })
    return copy
  }

  defDimension (): Dimension {
    return Object.assign({}, {
      conditions: {},
      meta: {},
      rotateLabel: 0,
    })
  }

  defMetric (): Metric {
    return Object.assign({}, {
      formatting: defFormatData(),
      valueLabelPosition: 'top' as const,
    })
  }

  defReport (): Report {
    return Object.assign({}, {
      moduleID: undefined,
      filter: '',
      dimensions: [this.defDimension()],
      metrics: [this.defMetric()],
      anomaly: {
        enabled: false,
        method: 'zscore' as 'zscore' | 'iqr' | 'fixed' | 'pct_change',
        threshold: 2,
        min: undefined,
        max: undefined,
        color: '',
      },
      compare: {
        enabled: false,
        dateField: '',
        granularity: 'month' as 'week' | 'month' | 'quarter' | 'year',
        mode: 'previous-period' as 'previous-period' | 'year-over-year',
        currentLabel: '',
        previousLabel: '',
      },
      yAxis: {
        axisType: 'linear',
        axisPosition: 'left',
        labelPosition: 'end',
        rotateLabel: 0,
        formatting: defFormatData(),
      },
      tooltip: {},
      legend: {
        isScrollable: true,
        orientation: 'horizontal',
        align: 'center',
        position: {
          top: undefined,
          right: undefined,
          bottom: undefined,
          left: undefined,
          isDefault: true,
        },
      },
      offset: {
        top: '50',
        right: '30',
        bottom: '20',
        left: '30',
        isDefault: true,
      },
    })
  }

  defConfig (): ChartConfig {
    return Object.assign({}, {
      colorScheme: '',
      reports: [this.defReport()],
      noAnimation: false,
      gradient: '' as '' | 'lightToDark' | 'darkToLight',
      toolbox: {
        saveAsImage: false,
        showDataTable: false,
        timeline: '',
      },
    })
  }

  /**
   * Resource type
   */
  get resourceType (): string {
    return 'compose:chart'
  }

  clone (): BaseChart {
    return new BaseChart(JSON.parse(JSON.stringify(this)))
  }
}

export { chartUtil } from './util'
