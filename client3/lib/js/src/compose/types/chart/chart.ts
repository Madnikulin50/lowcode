import { BaseChart } from './base'
import {
  Dimension,
  Metric,
  TemporalDataPoint,
  formatChartValue,
  formatChartTooltip,
  TooltipParams,
  detectAnomalies,
} from './util'
import { getColorschemeColors } from '../../../shared'

function lightenColor (color: string, percent: number): string {
  if (!color) return color
  const hex = color.replace('#', '')
  if (/^[0-9a-fA-F]{6}$/.test(hex)) {
    const r = Math.min(255, parseInt(hex.slice(0, 2), 16) + Math.round(2.55 * percent))
    const g = Math.min(255, parseInt(hex.slice(2, 4), 16) + Math.round(2.55 * percent))
    const b = Math.min(255, parseInt(hex.slice(4, 6), 16) + Math.round(2.55 * percent))
    return `rgb(${r},${g},${b})`
  }
  const m = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/)
  if (m) {
    const r = Math.min(255, parseInt(m[1]) + Math.round(2.55 * percent))
    const g = Math.min(255, parseInt(m[2]) + Math.round(2.55 * percent))
    const b = Math.min(255, parseInt(m[3]) + Math.round(2.55 * percent))
    return `rgb(${r},${g},${b})`
  }
  return color
}

/**
 * Chart represents a generic chart, such as a bar chart, line chart, ...
 */
export default class Chart extends BaseChart {
  // Generic charts (at the moment) support only 1 report per chart
  async fetchReports (a: any) {
    return super.fetchReports(a).then((rr: any) => {
      return rr[0]
    })
  }

  makeDataset (m: Metric, d: Dimension, data: Array<number|TemporalDataPoint>, alias: string) {
    data = this.datasetPostProc(data, m)

    return {
      type: m.type,
      label: m.label || m.field,
      data,
      alias,
      mapType: m.mapType,
      fill: m.fill,
      smooth: m.smooth,
      step: m.step ? 'middle' : undefined,
      roseType: m.rose ? 'radius' : undefined,
      symbol: m.symbol,
      stack: m.stack,
      tooltip: {
        fixed: m.fixTooltips,
        relative: m.relativeValue && !['bar', 'line'].includes(m.type as string),
      },
      formatting: m.formatting,
    }
  }

  makeOptions (data: any): any {
    const { reports = [], colorScheme, noAnimation = false, toolbox, gradient = '' } = this.config
    const { saveAsImage, timeline = '' } = toolbox || {}
    const schemeColors = getColorschemeColors(colorScheme, data.customColorSchemes)

    function gradientItemStyle (seriesColor: string, type: string): object | undefined {
      if (!gradient) return undefined
      if (['pie', 'doughnut'].includes(type)) {
        const innerColor = gradient === 'darkToLight' ? seriesColor : lightenColor(seriesColor, 50)
        const outerColor = gradient === 'darkToLight' ? lightenColor(seriesColor, 50) : seriesColor
        return {
          color: {
            type: 'radial',
            x: 0.5, y: 0.5, r: 0.5,
            colorStops: [
              { offset: 0, color: innerColor },
              { offset: 1, color: outerColor },
            ],
          },
        }
      }
      const topColor = gradient === 'darkToLight' ? seriesColor : lightenColor(seriesColor, 50)
      const bottomColor = gradient === 'darkToLight' ? lightenColor(seriesColor, 50) : seriesColor
      return {
        color: {
          type: 'linear',
          x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: topColor },
            { offset: 1, color: bottomColor },
          ],
        },
      }
    }

    const options: any = {
      animation: !noAnimation,
      series: [],
      xAxis: [],
      yAxis: [],
      tooltip: {
        show: true,
        appendToBody: true,
        position: 'inside',
      },
    }

    const getVal = (v: any): any => Array.isArray(v) ? v[1] : v

    const { labels, datasets = [], themeVariables = {} } = data
    const report = reports[0] || {}
    const {
      dimensions: [dimension] = [],
      yAxis,
      offset,
      tooltip: t,
      legend: l,
    } = report

    const hasAxis = datasets.some(({ type }: any) => ['bar', 'line', 'scatter'].includes(type))
    let horizontal = false

    // ------------------------------------------------------------------
    // Advanced chart types (sankey, graph, heatmap, waterfall, boxplot,
    // candlestick, map, sunburst, parallel, calendar)
    // They need either multiple dimensions (sankey, graph, heatmap,
    // sunburst) or a fixed number of metrics (boxplot, candlestick) and
    // therefore bypass the generic dataset mapper below.
    // ------------------------------------------------------------------
    const chartType = datasets[0]?.type
    if (['sankey', 'graph', 'heatmap', 'waterfall', 'boxplot', 'candlestick', 'map', 'sunburst', 'parallel', 'calendar'].includes(chartType)) {
      return this.makeAdvancedOptions(chartType, {
        labels,
        datasets,
        dimension,
        rows: data.rows || [],
        report,
        t,
        themeVariables,
        colorScheme,
        noAnimation,
        saveAsImage,
        timeline,
        gradient,
        schemeColors,
        offset,
        l,
      })
    }

    if (hasAxis) {
      if (yAxis) {
        const {
          label: yLabel,
          axisType: yType = 'linear',
          axisPosition: position = 'left',
          labelPosition = 'end',
          beginAtZero,
          min,
          max,
        } = yAxis

        horizontal = !!yAxis.horizontal

        const xAxis = {
          nameLocation: 'center',
          type: dimension.timeLabels ? 'time' : 'category',
          axisLabel: {
            interval: 0,
            overflow: 'break',
            hideOverlap: true,
            rotate: dimension.rotateLabel,
          },
          axisTick: {
            show: false,
          },
          axisLine: {
            show: false,
          },
        }

        const tempYAxis = {
          name: yLabel,
          type: yType === 'linear' ? 'value' : 'log',
          position,
          nameLocation: labelPosition,
          min: beginAtZero ? 0 : Number(min) || undefined,
          max: Number(max) || undefined,
          axisLabel: {
            interval: 0,
            overflow: 'break',
            hideOverlap: true,
            rotate: yAxis.rotateLabel,
            formatter: (value: string | number): string => formatChartValue(value, yAxis.formatting),
          },
          axisLine: {
            show: false,
            onZero: false,
          },
          splitLine: {
            lineStyle: {
              color: [themeVariables['extra-light']],
            },
          },
          nameTextStyle: {
            align: labelPosition === 'center' ? 'center' : position,
          },
        }

        // If we provide undefined, log scale breaks
        if (tempYAxis.type === 'log') {
          delete tempYAxis.min
          delete tempYAxis.max
        }

        if (horizontal) {
          options.xAxis = [tempYAxis]
          options.yAxis = [xAxis]
        } else {
          options.xAxis = [xAxis]
          options.yAxis = [tempYAxis]
        }
      }
    }

    options.series = datasets.map(({ formatting, type, label, data, stack, tooltip, fill, smooth, step, roseType, symbol }: any, index: number) => {
      const { fixed, relative } = tooltip

      // We should render the first metric in the dataset as the last
      const z = (datasets.length - 1) - index

      if (['pie', 'doughnut'].includes(type)) {
        const startRadius = type === 'doughnut' ? 40 : 0
        const endRadius = 80
        const radiusLength = (endRadius - startRadius) / (datasets.length || 1)

        const sr = startRadius + (index * radiusLength)
        const er = startRadius + ((index + 1) * radiusLength)

        options.tooltip.trigger = 'item'

        let lbl :any = {
          rotate: dimension.rotateLabel ? +dimension.rotateLabel : 0,
        }

        if (t?.labelsNextToPartition) {
          lbl = {
            ...lbl,
            show: true,
            overflow: 'truncate',
          }
        } else {
          lbl = {
            ...lbl,
            show: fixed,
            position: 'inside',
            align: 'center',
            verticalAlign: 'middle',
          }
        }

        return {
          z,
          stack,
          name: label,
          type: 'pie',
          roseType,
          radius: [`${sr}%`, `${er}%`],
          center: ['50%', '55%'],
          tooltip: {
            trigger: 'item',
            appendToBody: true,
            formatter: (params: TooltipParams): string => {
              const v = formatChartValue(params.value || '', formatting)

              if (t?.formatting) {
                return formatChartTooltip(t?.formatting, params)
              }

              return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${v}${relative ? ' (' + params.percent + '%)' : ''}</span>`
            },
          },
          label: {
            ...lbl,
            formatter: (params: TooltipParams): string => formatChartValue(params.value || '', formatting),
          },
          itemStyle: {
            borderRadius: 5,
            borderColor: themeVariables.white,
            borderWidth: 1,
          },
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
          data: labels.map((name: string, i: number) => {
            const item: any = { name, value: data[i] }
            if (gradient) {
              item.itemStyle = gradientItemStyle(schemeColors[i % schemeColors.length], type)
            }
            return item
          }),
          top: offset?.isDefault ? undefined : offset?.top,
          right: offset?.isDefault ? undefined : offset?.right,
          bottom: offset?.isDefault ? undefined : offset?.bottom,
          left: offset?.isDefault ? undefined : offset?.left,
        }
      } else if (['bar', 'line', 'scatter'].includes(type)) {
        options.tooltip.trigger = 'axis'

        const defaultOffset = {
          top: 65,
          right: timeline.includes('x') ? 40 : 30,
          bottom: timeline.includes('x') ? 60 : 20,
          left: 30,
        }

        options.grid = {
          top: offset?.isDefault ? defaultOffset.top : offset?.top,
          right: offset?.isDefault ? defaultOffset.right : offset?.right,
          bottom: offset?.isDefault ? defaultOffset.bottom : offset?.bottom,
          left: offset?.isDefault ? defaultOffset.left : offset?.left,
          containLabel: true,
        }

        const values = data.map((v: any) => v?.y != null ? v.y : v)
        const anomalyCfg = report.anomaly || (this.config.reports?.[0]?.anomaly)
        const anomalyFlags = anomalyCfg?.enabled ? detectAnomalies(values, anomalyCfg) : []

        if (anomalyFlags.length) {
          const anomalyColor = anomalyCfg!.color || themeVariables?.danger || '#ff4444'
          if (horizontal) {
            data = labels.map((name: string, i: number) => {
              return anomalyFlags[i] ? { value: [values[i], name], itemStyle: { color: anomalyColor } } : [values[i], name]
            })
          } else {
            data = labels.map((name: string, i: number) => {
              return anomalyFlags[i] ? { value: [name, values[i]], itemStyle: { color: anomalyColor } } : [name, values[i]]
            })
          }
        } else {
          if (horizontal) {
            data = labels.map((name: string, i: number) => {
              return [values[i], name]
            })
          } else {
            data = labels.map((name: string, i: number) => {
              return [name, values[i]]
            })
          }
        }

        return {
          z,
          stack,
          name: label,
          type: type,
          smooth,
          step,
          areaStyle: {
            opacity: fill ? 0.7 : 0,
          },
          ...(gradient && type === 'bar' ? {
            itemStyle: gradientItemStyle(schemeColors[index % schemeColors.length], type),
          } : {}),
          symbol,
          symbolSize: type === 'scatter' ? 16 : 10,
          tooltip: {
            appendToBody: true,
            // pass trigger type to determine if valueFormatter or formatter will be used
            trigger: t?.formatting ? 'item' : 'axis',
            // we can either
            // add formatting to the value and apply tooltip if trigger: 'item'
            // display the same tooltip format name <br/> seriesName value if trigger: 'axis'

            // works when trigger is set to axis
            valueFormatter: (value: string | number): string => formatChartValue(getVal(value), formatting),
            // works when trigger is set to item
            formatter: (params: { seriesName?: string; name?: string;value: Array<any>, percent: string | number, marker: string;}): string => {
              const { value: rawValue, percent = '' } = params
              const value = getVal(rawValue)

              const formattedValue = formatChartValue(value, formatting)

              if (t?.formatting) {
                return formatChartTooltip(t?.formatting, { ...params, value, percent })
              }

              return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${formattedValue}${relative ? ' (' + params.percent + '%)' : ''}</span>`
            },
          },
          label: {
            show: fixed,
            position: 'inside',
            align: 'center',
            verticalAlign: 'middle',
            tooltip: {
              trigger: 'axis',
            },
            formatter: (params: { seriesName: string, name: string, value: Array<any>, percent: string | number }): string => {
              const { value: rawValue, percent = '' } = params
              const value = getVal(rawValue)

              return `${formatChartValue(value, formatting)}${relative ? ` (${percent}%)` : ''}`
            },
          },
          data,
        }
      }
    })

    const dataZoom = timeline ? [
      {
        show: timeline.includes('x'),
        type: 'slider',
        height: 30,
      },
      {
        show: timeline.includes('y'),
        type: 'slider',
        width: 15,
        yAxisIndex: 0,
      },
    ] : undefined

    return {
      color: getColorschemeColors(colorScheme, data.customColorSchemes),
      textStyle: {
        fontFamily: themeVariables['font-regular'],
        overflow: 'break',
        color: themeVariables.black,
      },
      toolbox: {
        feature: {
          saveAsImage: saveAsImage ? {
            name: this.name,
          } : undefined,
        },
        top: 23,
        right: 2,
      },
      dataZoom,
      legend: {
        show: !l?.isHidden,
        type: l?.isScrollable ? 'scroll' : 'plain',
        top: (l?.position?.isDefault ? undefined : l?.position?.top) || undefined,
        right: (l?.position?.isDefault ? undefined : l?.position?.right) || undefined,
        bottom: (l?.position?.isDefault ? undefined : l?.position?.bottom) || undefined,
        left: (l?.position?.isDefault ? l?.align || 'center' : l?.position?.left) || 'auto',
        orient: l?.orientation || 'horizontal',
        textStyle: {
          color: themeVariables.black,
        },
        pageTextStyle: {
          color: themeVariables.black,
        },
        pageIconColor: themeVariables.black,
        pageIconInactiveColor: themeVariables.light,
      },
      animation: !noAnimation,
      ...options,
    }
  }

  /**
   * Renders one of the advanced chart types. These need either multiple
   * dimensions (sankey, graph, heatmap, sunburst) or a fixed number of
   * metrics (boxplot, candlestick) and therefore get their own option
   * builders instead of the generic dataset mapper.
   */
  makeAdvancedOptions (chartType: string, a: any): any {
    const {
      labels,
      datasets,
      dimension,
      rows = [] as Array<any>,
      report,
      t,
      themeVariables,
      colorScheme,
      noAnimation = false,
      saveAsImage,
      timeline = '',
      gradient,
      schemeColors,
      offset,
      l,
    } = a

    const getVal = (v: any): any => Array.isArray(v) ? v[1] : v
    const num = (v: any): number => {
      v = getVal(v)
      const n = Number(v)
      return isNaN(n) ? 0 : n
    }

    const dataZoom = timeline ? [
      {
        show: timeline.includes('x'),
        type: 'slider',
        height: 30,
      },
      {
        show: timeline.includes('y'),
        type: 'slider',
        width: 15,
        yAxisIndex: 0,
      },
    ] : undefined

    const common = {
      color: schemeColors,
      textStyle: {
        fontFamily: themeVariables['font-regular'],
        overflow: 'break',
        color: themeVariables.black,
      },
      toolbox: {
        feature: {
          saveAsImage: saveAsImage ? {
            name: this.name,
          } : undefined,
        },
        top: 23,
        right: 2,
      },
      dataZoom,
      legend: {
        show: !l?.isHidden,
        type: l?.isScrollable ? 'scroll' : 'plain',
        top: (l?.position?.isDefault ? undefined : l?.position?.top) || undefined,
        right: (l?.position?.isDefault ? undefined : l?.position?.right) || undefined,
        bottom: (l?.position?.isDefault ? undefined : l?.position?.bottom) || undefined,
        left: (l?.position?.isDefault ? l?.align || 'center' : l?.position?.left) || 'auto',
        orient: l?.orientation || 'horizontal',
        textStyle: {
          color: themeVariables.black,
        },
        pageTextStyle: {
          color: themeVariables.black,
        },
        pageIconColor: themeVariables.black,
        pageIconInactiveColor: themeVariables.light,
      },
      animation: !noAnimation,
    }

    const fmt = (v: any, formatting: any): string => formatChartValue(v, formatting)
    const baseTooltip = (formatting: any, trigger = 'item') => ({
      trigger,
      appendToBody: true,
      formatter: (params: TooltipParams): string => {
        const v = formatChartValue(getVal(params.value), formatting)
        if (t?.formatting) {
          return formatChartTooltip(t?.formatting, params)
        }
        return `${params.seriesName}<br>${params.marker}${params.name}<span style="float: right; margin-left: 20px">${v}</span>`
      },
    })

    // Minimal inline gradient builder (bar series only); full version
    // lives in makeOptions and needs colorscheme-aware series colors.
    const barGradient = (seriesColor: string): any => {
      if (!gradient) return undefined
      const topColor = gradient === 'darkToLight' ? seriesColor : lightenColor(seriesColor, 50)
      const bottomColor = gradient === 'darkToLight' ? lightenColor(seriesColor, 50) : seriesColor
      return {
        color: {
          type: 'linear',
          x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: topColor },
            { offset: 1, color: bottomColor },
          ],
        },
      }
    }

    const defaultOffset = {
      top: 65,
      right: 30,
      bottom: 20,
      left: 30,
    }

    // Build a set of unique nodes and a set of links out of the raw report
    // rows. Each row must contain dimension_0 and dimension_1.
    const buildGraph = (valueOf: (r: any) => number) => {
      const nodes: Array<any> = []
      const links: Array<any> = []
      const nodeMap = new Map<string, any>()
      const ensure = (name: string) => {
        if (!nodeMap.has(name)) {
          nodeMap.set(name, { name })
          nodes.push(nodeMap.get(name))
        }
        return nodeMap.get(name)
      }

      for (const r of rows) {
        if (r.dimension_1 === undefined || r.dimension_1 === null) continue
        const source = String(r.dimension_0 === undefined || r.dimension_0 === null ? 'undefined' : r.dimension_0)
        const target = String(r.dimension_1)
        ensure(source)
        ensure(target)
        links.push({ source, target, value: valueOf(r) })
      }

      return { nodes, links }
    }

    const grid = {
      top: offset?.isDefault ? defaultOffset.top : offset?.top,
      right: offset?.isDefault ? defaultOffset.right : offset?.right,
      bottom: offset?.isDefault ? defaultOffset.bottom : offset?.bottom,
      left: offset?.isDefault ? defaultOffset.left : offset?.left,
      containLabel: true,
    }

    switch (chartType) {
      case 'sankey': {
        const metric = datasets[0]
        const { nodes, links } = buildGraph(r => num(r[metric.alias]))
        return {
          ...common,
          tooltip: baseTooltip(metric.formatting),
          series: [{
            type: 'sankey',
            data: nodes,
            links,
            left: 10,
            right: 10,
            top: 40,
            bottom: 10,
            itemStyle: {
              borderColor: themeVariables.white,
              borderWidth: 1,
            },
            label: {
              color: themeVariables.black,
              fontSize: 12,
            },
            lineStyle: {
              color: 'gradient',
              curveness: 0.5,
            },
            emphasis: {
              focus: 'adjacency',
            },
          }],
        }
      }

      case 'graph': {
        const metric = datasets[0]
        const { nodes, links } = buildGraph(r => num(r[metric.alias]))
        return {
          ...common,
          tooltip: baseTooltip(metric.formatting),
          series: [{
            type: 'graph',
            layout: 'force',
            data: nodes,
            links,
            roam: true,
            force: {
              repulsion: 100,
              edgeLength: [50, 150],
            },
            label: {
              show: true,
              color: themeVariables.black,
            },
            itemStyle: {
              borderColor: themeVariables.white,
              borderWidth: 1,
            },
            lineStyle: {
              color: 'source',
              curveness: 0.2,
            },
            emphasis: {
              focus: 'adjacency',
              lineStyle: {
                width: 3,
              },
            },
          }],
        }
      }

      case 'heatmap': {
        const metric = datasets[0]
        const values: Array<Array<any>> = []
        let yLabels: Array<string>
        let xLabels: Array<string>

        if (rows.some((r: any) => r.dimension_1 !== undefined && r.dimension_1 !== null)) {
          // 2 dimensions: x = dimension_0, y = dimension_1
          const xSet = new Set<string>()
          const ySet = new Set<string>()
          for (const r of rows) {
            xSet.add(String(r.dimension_0 === undefined || r.dimension_0 === null ? 'undefined' : r.dimension_0))
            ySet.add(String(r.dimension_1))
          }
          xLabels = [...xSet]
          yLabels = [...ySet]
          for (const r of rows) {
            const xi = xLabels.indexOf(String(r.dimension_0))
            const yi = yLabels.indexOf(String(r.dimension_1))
            if (xi === -1 || yi === -1) continue
            values.push([xi, yi, num(r[metric.alias])])
          }
        } else {
          // 1 dimension: x = labels, y = metrics
          xLabels = labels
          yLabels = datasets.map((d: any) => d.label)
          datasets.forEach((d: any, yi: number) => {
            d.data.forEach((v: any, xi: number) => {
              values.push([xi, yi, num(v)])
            })
          })
        }

        const all = values.map(v => v[2])
        const min = Math.min(0, ...all)
        const max = Math.max(1, ...all)

        return {
          ...common,
          tooltip: baseTooltip(metric.formatting),
          grid,
          xAxis: {
            type: 'category',
            data: xLabels,
            splitArea: { show: true },
            axisLabel: {
              interval: 0,
              overflow: 'break',
              hideOverlap: true,
              rotate: dimension.rotateLabel,
            },
          },
          yAxis: {
            type: 'category',
            data: yLabels,
            splitArea: { show: true },
          },
          visualMap: {
            min,
            max,
            calculable: true,
            orient: 'horizontal',
            left: 'center',
            bottom: 0,
            inRange: {
              color: [themeVariables['extra-light'] || '#e8e8e8', ...(schemeColors.length ? [schemeColors[0]] : [])],
            },
          },
          series: [{
            type: 'heatmap',
            data: values,
            label: { show: false },
          }],
        }
      }

      case 'waterfall': {
        const metric = datasets[0]
        // Metric values are a running/cumulative series; compute the
        // per-bucket increase/decrease like the official bar-waterfall2
        // example (Income = positive deltas, Expenses = negative deltas).
        // The invisible base series holds the running total so the bars
        // stack onto each other.
        const raw = metric.data.map(num)
        const values = raw.map((v: number, i: number) => (i === 0 ? v : v - raw[i - 1]))

        let base = []
        let prev = 0
        base.push(...raw.map((v: number, i : number) => {
            prev = v
          if (i === 0) return 0
          if (prev > v)
              return v
          return prev
        }))



        return {
          ...common,
          tooltip: {
            trigger: 'axis',
            appendToBody: true,
              // works when trigger is set to axis
              valueFormatter: (value: string | number): string => formatChartValue(getVal(value), metric.formatting),
              // works when trigger is set to item
              formatter: (paramsList:TooltipParams[]): string => {
                  let value = getVal(paramsList[0].value)
                  let percent = 0
                  let marker
                  if (paramsList[1].value !== "-") {
                      value += Number(paramsList[1].value)
                      percent = Number(paramsList[1].value) * 100 / value
                      marker = paramsList[1].marker
                  } else if (paramsList[2].value !== "-") {
                      //value -= Number(paramsList[2].value)
                      percent = -Number(paramsList[2].value) * 100 / value
                      marker = paramsList[2].marker
                  }


                  const formattedValue = formatChartValue(value, metric.formatting)

                  return `${paramsList[0].seriesName}<br>${marker}${paramsList[0].name}<span style="float: right; margin-left: 20px">${formattedValue}${true ? ' (' + percent.toFixed(1) + '%)' : ''}</span>`
              }
          },
          grid,
          xAxis: {
            type: 'category',
            data: labels,
            axisLabel: {
              interval: 0,
              overflow: 'break',
              hideOverlap: true,
              rotate: dimension.rotateLabel,
            },
          },
          yAxis: {
            type: 'value',
            axisLabel: {
              formatter: (value: string | number): string => fmt(value, report.yAxis?.formatting),
            },
          },
          series: [
            {
              type: 'bar',
              stack: 'waterfall',
              name: metric.label,
              silent: true,
              data: base,
              itemStyle: { color: 'transparent' },
            },
            {
              type: 'bar',
              stack: 'waterfall',
              name: "increase",
              data: values.map((v: number, i: number) => ({
                value: v >= 0 ? v : '-' ,
                itemStyle: {
                  color: themeVariables.success || '#28a745'
                },
                label: {
                  show: true,
                  position: 'top',
                  formatter: (p: TooltipParams): string => fmt(p.value as number, metric.formatting),
                },
              })),
              ...(gradient ? { itemStyle: { color: barGradient(schemeColors[0]) } } : {}),
            },
              {
                  type: 'bar',
                  stack: 'waterfall',
                  name: "decrease",
                  data: values.map((v: number, i: number) => ({
                      value: v >= 0 ? '-' : -v,
                      itemStyle: {
                          color: themeVariables.danger || '#dc3545',
                      },
                      label: {
                          show: true,
                          position: 'bottom',
                          formatter: (p: TooltipParams): string => fmt(p.value as number, metric.formatting),
                      },
                  })),
                  ...(gradient ? { itemStyle: { color: barGradient(schemeColors[0]) } } : {}),
              },

          ],
        }
      }

      case 'boxplot': {
        const formatting = datasets[0]?.formatting
        // Boxplot expects [min, Q1, median, Q3, max] per category;
        // we take them from up to 5 metrics (padded with the last value)
        const data = labels.map((_: any, i: number) => {
          const box = datasets.map((d: any) => num(d.data[i]))
          while (box.length < 5) box.push(box[box.length - 1] || 0)
          return box.slice(0, 5)
        })

        return {
          ...common,
          tooltip: baseTooltip(formatting, 'item'),
          grid,
          xAxis: {
            type: 'category',
            data: labels,
            axisLabel: {
              interval: 0,
              overflow: 'break',
              hideOverlap: true,
              rotate: dimension.rotateLabel,
            },
          },
          yAxis: {
            type: 'value',
            axisLabel: {
              formatter: (value: string | number): string => fmt(value, report.yAxis?.formatting),
            },
          },
          series: [{
            type: 'boxplot',
            data,
            itemStyle: {
              borderColor: themeVariables.black,
              borderWidth: 1,
            },
          }],
        }
      }

      case 'candlestick': {
        const formatting = datasets[0]?.formatting
        // Candlestick expects [open, close, low, high] per category;
        // we take them from up to 4 metrics (padded with the last value)
        const data = labels.map((_: any, i: number) => {
          const candle = datasets.map((d: any) => num(d.data[i]))
          while (candle.length < 4) candle.push(candle[candle.length - 1] || 0)
          return candle.slice(0, 4)
        })

        return {
          ...common,
          tooltip: baseTooltip(formatting, 'item'),
          grid,
          xAxis: {
            type: 'category',
            data: labels,
            axisLabel: {
              interval: 0,
              overflow: 'break',
              hideOverlap: true,
              rotate: dimension.rotateLabel,
            },
          },
          yAxis: {
            type: 'value',
            axisLabel: {
              formatter: (value: string | number): string => fmt(value, report.yAxis?.formatting),
            },
          },
          series: [{
            type: 'candlestick',
            data,
            itemStyle: {
              color: themeVariables.success || '#28a745',
              color0: themeVariables.danger || '#dc3545',
              borderColor: themeVariables.success || '#28a745',
              borderColor0: themeVariables.danger || '#dc3545',
            },
          }],
        }
      }

      case 'map': {
        const metric = datasets[0]
        const mapType = metric.mapType || 'world'
        return {
          ...common,
          tooltip: baseTooltip(metric.formatting),
          visualMap: {
            min: Math.min(0, ...metric.data.map(num)),
            max: Math.max(1, ...metric.data.map(num)),
            left: 20,
            bottom: 20,
            calculable: true,
            inRange: {
              color: [themeVariables['extra-light'] || '#e8e8e8', ...(schemeColors.length ? [schemeColors[0]] : [])],
            },
          },
          series: [{
            type: 'map',
            map: mapType,
            roam: true,
            label: {
              show: false,
              color: themeVariables.black,
            },
            emphasis: {
              label: { show: true },
            },
            data: labels.map((l: string, i: number) => ({ name: l, value: num(metric.data[i]) })),
          }],
        }
      }

      case 'sunburst': {
        const metric = datasets[0]
        const data: Array<any> = []

        if (rows.some((r: any) => r.dimension_1 !== undefined && r.dimension_1 !== null)) {
          // 2 dimensions: group dimension_1 values under dimension_0 parents
          const groups = new Map<string, Array<any>>()
          for (const r of rows) {
            const parent = String(r.dimension_0 === undefined || r.dimension_0 === null ? 'undefined' : r.dimension_0)
            const child = String(r.dimension_1)
            const list = groups.get(parent)
            if (list) list.push({ name: child, value: num(r[metric.alias]) })
            else groups.set(parent, [{ name: child, value: num(r[metric.alias]) }])
          }
          for (const [parent, children] of groups.entries()) {
            data.push({ name: parent, children })
          }
        } else {
          // 1 dimension: a single ring of values
          labels.forEach((l: string, i: number) => {
            data.push({ name: l, value: num(metric.data[i]) })
          })
        }

        return {
          ...common,
          tooltip: baseTooltip(metric.formatting),
          series: [{
            type: 'sunburst',
            data,
            radius: [0, '90%'],
            label: {
              color: themeVariables.black,
            },
            itemStyle: {
              borderRadius: 4,
              borderColor: themeVariables.white,
              borderWidth: 1,
            },
            emphasis: {
              focus: 'ancestor',
            },
          }],
        }
      }

      case 'parallel': {
        const formatting = datasets[0]?.formatting
        return {
          ...common,
          tooltip: baseTooltip(formatting, 'item'),
          parallelAxis: datasets.map((d: any, i: number) => ({
            dim: i,
            name: d.label,
            nameTextStyle: {
              color: themeVariables.black,
            },
          })),
          parallel: {
            left: 60,
            right: 60,
            top: 40,
            bottom: 20,
            axisLabel: {
              color: themeVariables.black,
            },
          },
          series: [{
            type: 'parallel',
            data: labels.map((_: any, i: number) => datasets.map((d: any) => num(d.data[i]))),
          }],
        }
      }

      case 'calendar': {
        const metric = datasets[0]
        const data = labels.map((l: string, i: number) => [l, num(metric.data[i])])
        const all = metric.data.map(num)
        const [start, end] = [labels[0], labels[labels.length - 1]]
        const calendarType = metric.calendarType || 'heatmap'
        const max = Math.max(1, ...all)

        return {
          ...common,
          tooltip: baseTooltip(metric.formatting),
          visualMap: {
            min: Math.min(0, ...all),
            max: Math.max(1, ...all),
            left: 20,
            bottom: 20,
            calculable: true,
            inRange: {
              color: [themeVariables['extra-light'] || '#e8e8e8', ...(schemeColors.length ? [schemeColors[0]] : [])],
            },
          },
          calendar: {
            range: [start, end].filter(Boolean),
            left: 40,
            right: 20,
            top: 40,
            bottom: 60,
            cellSize: ['auto', 16],
            itemStyle: {
              color: themeVariables['extra-light'] || '#e8e8e8',
            },
            dayLabel: {
              color: themeVariables.black,
            },
            monthLabel: {
              color: themeVariables.black,
            },
            yearLabel: {
              color: themeVariables.black,
            },
          },
          series: [{
            type: calendarType,
            coordinateSystem: 'calendar',
            data,
            ...(calendarType !== 'heatmap' ? {
              symbolSize: (val: any): number => Math.max(4, Math.sqrt(num(val[1]) / max) * 24),
              rippleEffect: {
                period: 3,
                scale: 2,
              },
              itemStyle: {
                color: schemeColors[0],
              },
            } : {}),
          }],
        }
      }
    }

    return common
  }

  defMetric (): Metric {
    return Object.assign(super.defMetric(), {
      smooth: true,
      fill: false,
      rose: false,
      symbol: 'circle',
    })
  }

  baseChartType (datasets: Array<any>): string {
    return datasets[0].type
  }
}
