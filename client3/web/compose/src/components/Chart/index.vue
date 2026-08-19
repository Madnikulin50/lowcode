<template>
  <div class="d-flex h-100 w-100 position-relative">
    <div
      v-if="processing"
      class="d-flex flex-column align-items-center justify-content-center flex-fill"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>

    <label
      v-else-if="error"
      class="text-primary p-3"
    >
      {{ error }}
    </label>

    <div
      v-else-if="unknownTotal"
      class="d-flex flex-column align-items-center justify-content-center flex-fill text-secondary p-3 text-center"
      :title="t('chart.countUnavailableHint')"
    >
      <span class="fs-4">—</span>
      <small>{{ t('chart.countUnavailable') }}</small>
    </div>

    <c-chart
      v-else-if="renderer"
      :chart="renderer"
      class="flex-fill p-1"
      @click="drillDown"
    />

    <button
      v-if="renderer && hasDataTableEnabled"
      class="btn btn-outline-light chart-table-button position-absolute d-flex d-print-none border-0 px-1 text-secondary"
      :class="[tableVisible && 'text-primary']"
      :title="t('chart.dataTable.title')"
      @click="tableVisible = !tableVisible"
    >
      <font-awesome-icon :icon="['fas', 'table']" />
    </button>

    <div
      v-if="tableVisible"
      class="modal fade show d-block"
      tabindex="-1"
    >
      <div class="modal-dialog modal-dialog-centered modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('chart.dataTable.modal.title') }}</h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              @click="tableVisible = false"
            />
          </div>
          <div class="modal-body p-0">
            <div
              v-if="tableColumns.length"
              class="table-responsive"
              style="max-height: 60vh; overflow: auto"
            >
              <table class="table table-sm table-striped mb-0">
                <thead class="table-light">
                  <tr>
                    <th
                      v-for="column in tableColumns"
                      :key="column.key"
                    >
                      {{ column.label }}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(row, ri) in tableRows"
                    :key="ri"
                  >
                    <td
                      v-for="column in tableColumns"
                      :key="column.key"
                    >
                      {{ row[column.key] }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div
              v-else
              class="p-3 text-secondary"
            >
              {{ t('chart.dataTable.empty') }}
            </div>
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-outline-secondary"
              @click="tableVisible = false"
            >
              {{ t('label.close') }}
            </button>
          </div>
        </div>
      </div>
    </div>
    <div
      v-if="tableVisible"
      class="modal-backdrop fade show"
      @click="tableVisible = false"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'notification' } })
import { ref, computed, watch, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { chartConstructor } from 'corteza-webapp-compose/src/lib/charts'
import { ensureMapRegistered } from 'corteza-webapp-compose/src/lib/chart-maps'
import { compose, isUnknownReportCount } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import { useStore } from '../../store'
const { CChart } = components

const { t } = useI18n()
const store = useStore()

const $Settings = window.__settings
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  chart: {
    type: Object,
    required: true,
  },
  reporter: {
    type: Function,
    required: true,
  },
  record: {
    type: compose.Record,
    required: false,
    default: undefined,
  },
})

const emit = defineEmits(['updated', 'drill-down'])
defineExpose({ updateChart })

const error = ref(undefined)
const processing = ref(false)
const valueMap = ref(new Map())
const renderer = ref(undefined)
const unknownTotal = ref(false)
const tableData = ref({ columns: [], rows: [] })
const tableVisible = ref(false)

const hasDataTableEnabled = computed(() => {
  const { config = {} } = props.chart || {}
  if (!config.toolbox) return false
  return config.toolbox.showDataTable
})

const tableColumns = computed(() => tableData.value.columns || [])
const tableRows = computed(() => tableData.value.rows || [])

const instance = getCurrentInstance()

watch(() => props.record?.recordID, () => {
  updateChart()
}, { immediate: true })

onBeforeUnmount(() => {
  processing.value = false
  renderer.value = undefined
  unknownTotal.value = false
  valueMap.value = new Map()
})

async function updateChart () {
  error.value = undefined
  renderer.value = undefined
  unknownTotal.value = false

  const [report = {}] = props.chart.config?.reports || []

  if (!report.moduleID) return

  processing.value = true

  const chart = chartConstructor(props.chart)

  // Hoisted so the deferred (setTimeout) chart-data emit can use them
  let fields = []
  let data = null

  try {
    chart.isValid()

    const module = getModuleByID(report.moduleID)
    if (!module) throw new Error('Module not found')
    fields = [
      ...module.fields,
      ...module.systemFields(),
    ]

    const isGantt = (report.metrics || []).some(m => m.type === 'gantt')

    if (isGantt) {
      data = await fetchGanttReport(report)
    } else {
      data = await chart.fetchReports({ reporter: props.reporter })
    }

    if (data?.unknownTotal || (Array.isArray(data?.rows) && data.rows.some(isUnknownReportCount))) {
      unknownTotal.value = true
      renderer.value = undefined
    } else {
    if (!!data.labels && Array.isArray(data.labels)) {
      const [dimension = {}] = report.dimensions
      let { field } = dimension

      if (!module) throw new Error('Module not found')

      field = fields.find(({ name }) => name === field)

      if (field) {
        data.labels = await prettyDimensionValues(dimension, field, data.labels)
      } else if (!isGantt) {
        throw new Error('Dimension field not found')
      }
    }

    // Prettify raw dimension values (Select/User/Record/Bool → labels) in
    // rows so the data-table export shows readable values. Multi-dimension
    // charts (sankey, graph, heatmap, sunburst) rely on these rows as well.
    if (Array.isArray(data.rows) && data.rows.length) {
      const dims = report.dimensions || []
      const dimFields = dims.map(d => fields.find(({ name }) => name === d?.field))

      for (let i = 0; i < 2; i++) {
        const dim = dims[i]
        const field = dimFields[i]
        const rawValues = data.rows.map(r => r[`dimension_${i}`])
        const pretty = await prettyDimensionValues(dim, field, rawValues)
        data.rows.forEach((r, j) => {
          r[`dimension_${i}`] = pretty[j]
        })
      }
    }

    data.datasets = data.datasets.map((dataset = {}) => {
      const { label } = dataset
      if (label === 'count') {
        dataset.label = t('chart.general.label.count')
      } else {
        const field = fields.find(({ name }) => name === label) || {}
        dataset.label = field.label || label
      }
      return dataset
    })

    data.labels = data.labels.map(l => l === 'undefined' ? t('chart.undefined') : l)
    data.customColorSchemes = $Settings.get('ui.charts.colorSchemes', [])
    data.themeVariables = getThemeVariables()

    // Map charts require the GeoJSON to be registered with ECharts first
    const mapTypes = [...new Set(data.datasets
      .filter(({ type }) => type === 'map')
      .map(({ mapType }) => mapType || 'world'))]
    await Promise.all(mapTypes.map(name => ensureMapRegistered(name)))

    renderer.value = chart.makeOptions(data)
    }
  } catch (e) {
    error.value = e
    processing.value = false
  }

  setTimeout(() => {
    processing.value = false
    emit('updated')
    const rows = data?.rows || []
    onChartData(buildTableColumns(report, fields, rows), rows)
  }, 300)
}

function recordFieldValue (rec, name) {
  if (!name || !rec) return undefined
  const values = rec.values
  if (Array.isArray(values)) {
    const hit = values.find(v => v.name === name)
    if (!hit) return undefined
    const v = hit.value
    return Array.isArray(v) ? v[0] : v
  }
  if (values && typeof values === 'object') {
    const v = values[name]
    return Array.isArray(v) ? v[0] : v
  }
  return undefined
}

async function fetchGanttReport (report) {
  const metric = (report.metrics || [])[0] || {}
  const dimension = (report.dimensions || [])[0] || {}
  const startField = metric.startField || 'start_date_planned'
  const endField = metric.endField || 'end_date_planned'
  const labelField = dimension.field
  const params = {
    namespaceID: props.chart.namespaceID,
    moduleID: report.moduleID,
    query: report.filter || '',
    limit: 200,
    sort: `${startField} ASC`,
  }

  let set = []
  try {
    const res = await $ComposeAPI.recordList(params)
    set = res.set || []
  } catch (e) {
    const res = await $ComposeAPI.recordList({ ...params, sort: undefined })
    set = res.set || []
  }

  const rows = []
  for (const raw of set) {
    const start = recordFieldValue(raw, startField)
    const end = recordFieldValue(raw, endField)
    if (!start || !end) continue
    const label = recordFieldValue(raw, labelField)
    rows.push({
      dimension_0: label == null || label === '' ? (raw.recordID || raw.ID) : label,
      gantt_start: start,
      gantt_end: end,
    })
  }

  return {
    labels: rows.map(r => r.dimension_0),
    datasets: [{
      type: 'gantt',
      label: metric.label || 'gantt',
      data: rows.map(() => 1),
      alias: 'gantt',
      startField,
      endField,
      formatting: metric.formatting || {},
    }],
    dimension,
    rows,
  }
}

function onChartData (columns, rows) {
  tableData.value = { columns: columns || [], rows: rows || [] }
}

function buildTableColumns (report, fields, rows) {
  const columns = []
  const isGantt = (report.metrics || []).some(m => m.type === 'gantt')

  if (isGantt) {
    const metric = (report.metrics || [])[0] || {}
    const dimension = (report.dimensions || [])[0] || {}
    const dimField = fields.find(f => f.name === dimension.field)
    const startField = fields.find(f => f.name === (metric.startField || 'start_date_planned'))
    const endField = fields.find(f => f.name === (metric.endField || 'end_date_planned'))
    columns.push({ key: 'dimension_0', label: dimField?.label || dimension.field || 'name' })
    columns.push({ key: 'gantt_start', label: startField?.label || metric.startField || 'start' })
    columns.push({ key: 'gantt_end', label: endField?.label || metric.endField || 'end' })
    return columns
  }

  ;(report.dimensions || []).slice(0, 2).forEach((dimension, i) => {
    if (i > 0 && !rows.some(r => r.dimension_1 !== undefined && r.dimension_1 !== null)) return
    const field = fields.find(f => f.name === dimension?.field)
    columns.push({ key: `dimension_${i}`, label: field?.label || dimension?.field || `dimension_${i}` })
  })

  ;(report.metrics || []).forEach(metric => {
    const alias = metric.field === 'count' ? 'count' : (metric.alias || `${metric.aggregate || metric.modifier || 'none'}_${metric.field}`.toLowerCase())
    const label = metric.field === 'count'
      ? t('chart.general.label.count')
      : (fields.find(f => f.name === metric.field)?.label || metric.label || metric.field)
    columns.push({ key: alias, label })
  })

  return columns
}

function drillDown (e) {
  e.trueName = valueMap.value[e.name] || e.name
  emit('drill-down', e)
}

function getThemeVariables () {
  const getCssVariable = (variableName) => {
    return getComputedStyle(document.documentElement).getPropertyValue(variableName).trim()
  }
  return ['white', 'black', 'primary', 'secondary', 'success', 'warning', 'danger', 'light', 'extra-light', 'dark', 'font-regular'].reduce((acc, variable) => {
    acc[variable] = getCssVariable(`--${variable}`)
    return acc
  }, {})
}

function getModuleByID (id) {
  if (id == null || id === '') return undefined
  return store.module.getByID(id) ||
    store.module.getByID(String(id)) ||
    (store.module.set || []).find(m => String(m.moduleID) === String(id))
}

function getUserByID (id) {
  return store.user.findByID(id)
}

/**
 * Converts raw dimension values into pretty labels, following the same
 * rules for every field kind (Bool, Select, User, Record). Also fills the
 * valueMap used for drill-down.
 */
async function prettyDimensionValues (dimension, field, values) {
  if (!dimension || !field || !Array.isArray(values)) return values || []

  const isValidValue = (value) => value !== dimension.default && value !== 'undefined'

  if (field.kind === 'Bool') {
    const { trueLabel, falseLabel } = field.options
    return values.map(value => {
      return value === '1' ? trueLabel || t('label.yes') : falseLabel || t('label.no')
    })
  } else if (field.kind === 'Select') {
    return values.map(value => {
      const { text } = field.options.options.find(o => o.value === value) || {}
      const label = text || value
      valueMap.value[label] = value
      return label
    })
  } else if (field.kind === 'User') {
    await store.user.resolveUsers(values.filter(userID => isValidValue(userID)))
    return values.map(userID => {
      const label = field.formatter(store.user.findByID(userID)) || userID
      valueMap.value[label] = userID
      return label
    })
  } else if (field.kind === 'Record') {
    const { namespaceID } = props.chart || {}
    const recordModule = getModuleByID(field.options.moduleID)
    if (!recordModule || !values) return values

    return Promise.all(values.map(recordID => {
      if (isValidValue(recordID)) {
        return $ComposeAPI.recordRead({ namespaceID, moduleID: recordModule.moduleID, recordID }).then(record => {
          record = new compose.Record(recordModule, record)
          if (field.options.recordLabelField) {
            const relatedField = recordModule.fields.find(({ name }) => name === field.options.labelField)
            return $ComposeAPI.recordRead({ namespaceID, moduleID: relatedField.options.moduleID, recordID: record.values[field.options.labelField] }).then(labelRecord => {
              record.values[field.options.labelField] = (labelRecord.values.find(({ name }) => name === field.options.recordLabelField) || {}).value
              return record
            })
          } else {
            return record
          }
        })
      } else {
        const record = { values: {} }
        record.values[field.options.labelField] = recordID
        return record
      }
    })).then(records => {
      return records.map(record => {
        const value = field.options.labelField ? record.values[field.options.labelField] : record.recordID
        const label = Array.isArray(value) ? value.join(', ') : value
        valueMap.value[label] = record.recordID
        return value
      })
    })
  }

  return values
}
</script>

<style lang="scss" scoped>
.chart-table-button {
  right: 2.2rem;
  top: 0.7rem;
  z-index: 1;
}
</style>
