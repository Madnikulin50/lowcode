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
      startField: m.startField,
      endField: m.endField,
      fill: m.fill,
      smooth: m.smooth,
      step: m.step ? 'middle' : undefined,
      roseType: m.rose ? 'radius' : undefined,
      symbol: m.symbol,
      showSymbol: m.showSymbol,
      stack: m.stack,
      tooltip: {
        fixed: m.fixTooltips,
        relative: m.relativeValue && !['bar', 'line'].includes(m.type as string),
        // 'top' = above bar (or right for horizontal); 'inside' = inside the bar
        valueLabelPosition: m.valueLabelPosition || 'top',
      },
      formatting: m.formatting,
    }
  }

  makeOptions (data: any): any {
    const { reports = [], colorScheme, noAnimation = false, toolbox, gradient = '' } = this.config
    const { saveAsImage, showDataTable, timeline = '' } = toolbox || {}
    const schemeColors = getColorschemeColors(colorScheme, data.customColorSchemes)

    function linearGradientColor (seriesColor: string): object {
      const topColor = gradient === 'darkToLight' ? seriesColor : lightenColor(seriesColor, 50)
      const bottomColor = gradient === 'darkToLight' ? lightenColor(seriesColor, 50) : seriesColor
      return {
        type: 'linear',
        x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: topColor },
          { offset: 1, color: bottomColor },
        ],
      }
    }

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
      return { color: linearGradientColor(seriesColor) }
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

    // vertical data is [category, value]; horizontal is [value, category]
    const getVal = (v: any): any => {
      if (!Array.isArray(v)) return v
      return horizontal ? v[0] : v[1]
    }

    const gridNum = (v: any, fallback: number): number => {
      const n = Number(v)
      return Number.isFinite(n) ? n : fallback
    }

    const hasOutsideValueLabels = datasets.some(({ type, tooltip: tip }: any) => {
      if (!['bar', 'line', 'scatter'].includes(type)) return false
      const { fixed, valueLabelPosition = 'top' } = tip || {}
      return !!fixed && valueLabelPosition !== 'inside'
    })

    // ------------------------------------------------------------------
    // Advanced chart types (sankey, graph, heatmap, waterfall, boxplot,
    // candlestick, map, sunburst, parallel, calendar, gantt)
    // They need either multiple dimensions (sankey, graph, heatmap,
    // sunburst) or a fixed number of metrics (boxplot, candlestick) and
    // therefore bypass the generic dataset mapper below.
    // ------------------------------------------------------------------
    const chartType = datasets[0]?.type
    if (['sankey', 'graph', 'heatmap', 'waterfall', 'boxplot', 'candlestick', 'map', 'sunburst', 'parallel', 'calendar', 'gantt'].includes(chartType)) {
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

        const axisMuted = themeVariables.secondary || themeVariables.light || '#6c757d'
        const gridLine = themeVariables['extra-light'] || '#e9ecef'

        const xAxis = {
          nameLocation: 'center',
          type: dimension.timeLabels ? 'time' : 'category',
          axisLabel: {
            interval: 0,
            overflow: 'truncate',
            hideOverlap: true,
            rotate: dimension.rotateLabel,
            color: axisMuted,
            fontSize: 11,
          },
          axisTick: {
            show: false,
          },
          axisLine: {
            show: false,
          },
          splitLine: {
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
            overflow: 'truncate',
            hideOverlap: true,
            rotate: yAxis.rotateLabel,
            color: axisMuted,
            fontSize: 11,
            formatter: (value: string | number): string => formatChartValue(value, yAxis.formatting),
          },
          axisLine: {
            show: false,
            onZero: false,
          },
          splitLine: {
            show: true,
            lineStyle: {
              color: gridLine,
              type: 'dashed',
              width: 1,
            },
          },
          nameTextStyle: {
            align: labelPosition === 'center' ? 'center' : position,
            color: axisMuted,
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

    const pieLike = datasets.some(({ type }: any) => ['pie', 'doughnut'].includes(type))
    const legendShown = !l?.isHidden && (pieLike || datasets.length > 1)

    // Last bar series in each stack gets the rounded "cap" end.
    const stackCapIndex = new Map<string, number>()
    datasets.forEach((d: any, i: number) => {
      if (d?.type === 'bar') {
        stackCapIndex.set(d.stack ? `s:${d.stack}` : `i:${i}`, i)
      }
    })

    options.series = datasets.map(({ formatting, type, label, data, stack, tooltip, fill, smooth, step, roseType, symbol, showSymbol }: any, index: number) => {
      const { fixed, relative, valueLabelPosition = 'top' } = tooltip || {}
      const labelOutside = valueLabelPosition !== 'inside'

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
          // Single-series bars hide the legend — less top chrome needed.
          top: legendShown ? 65 : 44,
          right: timeline.includes('x') ? 40 : 30,
          bottom: timeline.includes('x') ? 60 : 20,
          left: 30,
        }

        // Outside value labels sit beyond the bar tip; pad grid so the plot
        // does not look shifted/clipped inside the block.
        const outsidePadTop = !horizontal && hasOutsideValueLabels ? 18 : 0
        const outsidePadRight = horizontal && hasOutsideValueLabels ? 40 : 0

        options.grid = {
          top: (offset?.isDefault ? defaultOffset.top : gridNum(offset?.top, defaultOffset.top)) + outsidePadTop,
          right: (offset?.isDefault ? defaultOffset.right : gridNum(offset?.right, defaultOffset.right)) + outsidePadRight,
          bottom: offset?.isDefault ? defaultOffset.bottom : gridNum(offset?.bottom, defaultOffset.bottom),
          left: offset?.isDefault ? defaultOffset.left : gridNum(offset?.left, defaultOffset.left),
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

        const labelMuted = themeVariables.secondary || themeVariables.light || '#64748b'
        const isBar = type === 'bar'
        const barRadius = 5
        const isStackCap = isBar && stackCapIndex.get(stack ? `s:${stack}` : `i:${index}`) === index
        // ECharts borderRadius: [top-left, top-right, bottom-right, bottom-left]
        const barBorderRadius = !isStackCap
          ? 0
          : (horizontal ? [0, barRadius, barRadius, 0] : [barRadius, barRadius, 0, 0])

        const barItemStyle: any = {
          ...(gradient && isBar ? gradientItemStyle(schemeColors[index % schemeColors.length], type) : {}),
          borderRadius: barBorderRadius,
          ...(stack && isBar ? {
            borderColor: themeVariables.white || '#fff',
            borderWidth: 1,
          } : {}),
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
            ...(gradient && type === 'line' && fill ? {
              color: linearGradientColor(schemeColors[index % schemeColors.length]),
            } : {}),
          },
          ...(isBar ? {
            barMaxWidth: 48,
            barCategoryGap: '35%',
            barGap: '25%',
            itemStyle: barItemStyle,
            emphasis: {
              itemStyle: {
                shadowBlur: 0,
                opacity: 0.88,
              },
            },
          } : {}),
          symbol,
          symbolSize: type === 'scatter' ? 16 : 10,
          showSymbol: type === 'line' && showSymbol === false ? false : undefined,
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
          // Keep ECharts defaults for outside positions (right → align left,
          // top → verticalAlign bottom). Forcing align:'center' on horizontal
          // bars centers the label on the tip and looks like a sideways shift.
          label: {
            show: fixed,
            color: labelOutside ? labelMuted : (themeVariables.white || '#fff'),
            fontSize: 11,
            fontWeight: 500,
            ...(fixed
              ? (labelOutside
                ? {
                    position: horizontal ? 'right' : 'top',
                    distance: 4,
                  }
                : {
                    position: 'inside',
                    align: 'center',
                    verticalAlign: 'middle',
                  })
              : {}),
            tooltip: {
              trigger: 'axis',
            },
            formatter: (params: { seriesName: string, name: string, value: Array<any>, percent: string | number }): string => {
              const { value: rawValue, percent = '' } = params
              const value = getVal(rawValue)

              return `${formatChartValue(value, formatting)}${relative ? ` (${percent}%)` : ''}`
            },
          },
          // Allow outside labels to render past the bar/grid edge
          clip: !(fixed && labelOutside),
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
        top: showDataTable ? 62 : 23,
        right: 2,
      },
      dataZoom,
      legend: {
        show: legendShown,
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
      showDataTable,
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
        top: showDataTable ? 62 : 23,
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
        // Invisible base: sit each delta bar on min(prev, current) so the
        // visible segment spans the change (classic waterfall stacking).
        const base = raw.map((v: number, i: number) => (i === 0 ? 0 : Math.min(raw[i - 1], v)))

        const axisMuted = themeVariables.secondary || themeVariables.light || '#6c757d'
        const gridLine = themeVariables['extra-light'] || '#e9ecef'
        const labelMuted = axisMuted
        const barRadius = 5
        const successColor = themeVariables.success || '#28a745'
        const dangerColor = themeVariables.danger || '#dc3545'
        const softEmphasis = { itemStyle: { shadowBlur: 0, opacity: 0.88 } }

        const increaseData = values.map((v: number) => ({
          value: v >= 0 ? v : '-',
          itemStyle: {
            color: gradient
              ? (barGradient(successColor)?.color || successColor)
              : successColor,
            borderRadius: [barRadius, barRadius, 0, 0],
          },
          label: {
            show: true,
            position: 'top' as const,
            distance: 4,
            color: labelMuted,
            fontSize: 11,
            fontWeight: 500,
            formatter: (p: TooltipParams): string => fmt(p.value as number, metric.formatting),
          },
        }))

        const decreaseData = values.map((v: number) => ({
          value: v >= 0 ? '-' : -v,
          itemStyle: {
            color: gradient
              ? (barGradient(dangerColor)?.color || dangerColor)
              : dangerColor,
            borderRadius: [barRadius, barRadius, 0, 0],
          },
          label: {
            show: true,
            position: 'bottom' as const,
            distance: 4,
            color: labelMuted,
            fontSize: 11,
            fontWeight: 500,
            formatter: (p: TooltipParams): string => fmt(p.value as number, metric.formatting),
          },
        }))

        return {
          ...common,
          legend: { show: false },
          tooltip: {
            trigger: 'axis',
            appendToBody: true,
            valueFormatter: (value: string | number): string => formatChartValue(getVal(value), metric.formatting),
            formatter: (paramsList: TooltipParams[]): string => {
              let value = getVal(paramsList[0].value)
              let percent = 0
              let marker
              if (paramsList[1].value !== '-') {
                value += Number(paramsList[1].value)
                percent = Number(paramsList[1].value) * 100 / value
                marker = paramsList[1].marker
              } else if (paramsList[2].value !== '-') {
                percent = -Number(paramsList[2].value) * 100 / value
                marker = paramsList[2].marker
              }

              const formattedValue = formatChartValue(value, metric.formatting)
              return `${paramsList[0].seriesName}<br>${marker}${paramsList[0].name}<span style="float: right; margin-left: 20px">${formattedValue}${true ? ' (' + percent.toFixed(1) + '%)' : ''}</span>`
            },
          },
          grid: {
            ...grid,
            top: (typeof grid.top === 'number' ? grid.top : Number(grid.top) || 65) + 12,
          },
          xAxis: {
            type: 'category',
            data: labels,
            axisLabel: {
              interval: 0,
              overflow: 'truncate',
              hideOverlap: true,
              rotate: dimension.rotateLabel,
              color: axisMuted,
              fontSize: 11,
            },
            axisTick: { show: false },
            axisLine: { show: false },
            splitLine: { show: false },
          },
          yAxis: {
            type: 'value',
            axisLabel: {
              color: axisMuted,
              fontSize: 11,
              formatter: (value: string | number): string => fmt(value, report.yAxis?.formatting),
            },
            axisTick: { show: false },
            axisLine: { show: false },
            splitLine: {
              show: true,
              lineStyle: { color: gridLine, type: 'dashed', width: 1 },
            },
          },
          series: [
            {
              type: 'bar',
              stack: 'waterfall',
              name: metric.label,
              silent: true,
              data: base,
              itemStyle: { color: 'transparent', borderColor: 'transparent' },
              emphasis: { disabled: true },
            },
            {
              type: 'bar',
              stack: 'waterfall',
              name: 'increase',
              barMaxWidth: 48,
              barCategoryGap: '35%',
              emphasis: softEmphasis,
              data: increaseData,
            },
            {
              type: 'bar',
              stack: 'waterfall',
              name: 'decrease',
              barMaxWidth: 48,
              barCategoryGap: '35%',
              emphasis: softEmphasis,
              data: decreaseData,
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

      case 'gantt': {
        const metric = datasets[0] || {}
        const toTime = (v: any): number => {
          if (v == null || v === '') return NaN
          if (typeof v === 'number') return v < 1e12 ? v * 1000 : v
          const raw = String(v).trim()
          if (/^\d+(\.\d+)?$/.test(raw)) {
            const n = Number(raw)
            return n < 1e12 ? n * 1000 : n
          }
          return Date.parse(raw)
        }
        const fmtDate = (ms: number): string => {
          if (!isFinite(ms)) return ''
          const d = new Date(ms)
          const y = d.getFullYear()
          const m = String(d.getMonth() + 1).padStart(2, '0')
          const day = String(d.getDate()).padStart(2, '0')
          return `${y}-${m}-${day}`
        }

        const axisMuted = themeVariables.secondary || themeVariables.light || '#6c757d'
        const gridLine = themeVariables['extra-light'] || '#e9ecef'
        const labelOnBar = themeVariables.white || '#fff'
        const barRadius = 5

        const tasks: Array<{ name: string; start: number; end: number }> = []
        for (let i = 0; i < rows.length; i++) {
          const r = rows[i]
          const start = toTime(r.gantt_start)
          const end = toTime(r.gantt_end)
          if (!isFinite(start) || !isFinite(end) || end < start) continue
          const name = String(
            (labels && labels[i] != null && labels[i] !== '')
              ? labels[i]
              : (r.dimension_0 == null ? `#${i + 1}` : r.dimension_0)
          )
          tasks.push({ name, start, end })
        }

        const categories = tasks.map(tk => tk.name)
        const colors = Array.isArray(schemeColors) && schemeColors.length ? schemeColors : [themeVariables.primary || '#3366cc']
        const data = tasks.map((tk, i) => ({
          value: [i, tk.start, tk.end, tk.name],
          itemStyle: {
            color: colors[i % colors.length],
            opacity: 0.92,
          },
        }))

        return {
          ...common,
          legend: { show: false },
          tooltip: {
            trigger: 'item',
            appendToBody: true,
            formatter: (params: any): string => {
              const v = params?.value || []
              const start = Number(v[1])
              const end = Number(v[2])
              const days = isFinite(start) && isFinite(end)
                ? Math.max(1, Math.round((end - start) / 86400000))
                : 0
              const title = v[3] || params?.name || ''
              if (t?.formatting) {
                return formatChartTooltip(t.formatting, params)
              }
              return `${params.marker || ''}<b>${title}</b><br/>${fmtDate(start)} → ${fmtDate(end)}<span style="float: right; margin-left: 16px">${days}d</span>`
            },
          },
          grid: {
            ...grid,
            left: offset?.isDefault ? 120 : offset?.left,
            bottom: offset?.isDefault ? 36 : offset?.bottom,
            containLabel: true,
          },
          dataZoom: [
            {
              type: 'slider',
              xAxisIndex: 0,
              height: 16,
              bottom: 6,
              borderColor: 'transparent',
              backgroundColor: gridLine,
              fillerColor: 'rgba(100, 116, 139, 0.18)',
              handleSize: 0,
              showDetail: false,
              filterMode: 'weakFilter',
            },
            { type: 'inside', xAxisIndex: 0, filterMode: 'weakFilter' },
          ],
          xAxis: {
            type: 'time',
            axisLabel: {
              color: axisMuted,
              fontSize: 11,
              hideOverlap: true,
            },
            axisTick: { show: false },
            axisLine: { show: false },
            splitLine: {
              show: true,
              lineStyle: { color: gridLine, type: 'dashed', width: 1 },
            },
          },
          yAxis: {
            type: 'category',
            data: categories,
            inverse: true,
            axisLabel: {
              color: axisMuted,
              fontSize: 11,
              width: 110,
              overflow: 'truncate',
            },
            axisTick: { show: false },
            axisLine: { show: false },
            splitLine: {
              show: true,
              lineStyle: { color: gridLine, type: 'dashed', width: 1 },
            },
          },
          series: [{
            type: 'custom',
            name: metric.label || metric.field || 'gantt',
            renderItem: (params: any, api: any) => {
              const categoryIndex = api.value(0)
              const startMs = api.value(1)
              const endMs = api.value(2)
              const start = api.coord([startMs, categoryIndex])
              const end = api.coord([endMs, categoryIndex])
              const height = api.size([0, 1])[1] * 0.5
              const coordSys = params.coordSys || {}
              const shape = {
                x: start[0],
                y: start[1] - height / 2,
                width: Math.max(end[0] - start[0], 3),
                height,
                r: barRadius,
              }
              if (coordSys.width) {
                const minX = coordSys.x || 0
                const maxX = minX + coordSys.width
                if (shape.x < minX) {
                  shape.width -= (minX - shape.x)
                  shape.x = minX
                }
                if (shape.x + shape.width > maxX) {
                  shape.width = Math.max(0, maxX - shape.x)
                }
              }
              if (shape.width <= 0) return

              const rect = {
                type: 'rect',
                transition: ['shape'],
                shape,
                style: api.style(),
                emphasisDisabled: false,
              }

              const days = Math.max(1, Math.round((Number(endMs) - Number(startMs)) / 86400000))
              if (shape.width < 44) {
                return rect
              }

              return {
                type: 'group',
                children: [
                  rect,
                  {
                    type: 'text',
                    style: {
                      x: shape.x + shape.width / 2,
                      y: shape.y + shape.height / 2,
                      text: `${days}d`,
                      fill: labelOnBar,
                      fontSize: 11,
                      fontWeight: 500,
                      align: 'center',
                      verticalAlign: 'middle',
                    },
                    silent: true,
                  },
                ],
              }
            },
            encode: { x: [1, 2], y: 0 },
            data,
            emphasis: {
              style: {
                opacity: 0.88,
                shadowBlur: 0,
              },
            },
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
