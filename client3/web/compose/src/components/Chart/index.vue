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

    <c-chart
      v-else-if="renderer"
      :chart="renderer"
      class="flex-fill p-1"
      @click="drillDown"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'notification' } })
import { ref, watch, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { chartConstructor } from 'corteza-webapp-compose/src/lib/charts'
import { ensureMapRegistered } from 'corteza-webapp-compose/src/lib/chart-maps'
import { compose } from 'corteza-lib/js/dist'
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

const instance = getCurrentInstance()

watch(() => props.record?.recordID, () => {
  updateChart()
}, { immediate: true })

onBeforeUnmount(() => {
  processing.value = false
  renderer.value = undefined
  valueMap.value = new Map()
})

async function updateChart () {
  error.value = undefined
  renderer.value = undefined

  const [report = {}] = props.chart.config?.reports || []

  if (!report.moduleID) return

  processing.value = true

  const chart = chartConstructor(props.chart)

  try {
    chart.isValid()

    const data = await chart.fetchReports({ reporter: props.reporter })

    const module = getModuleByID(report.moduleID)
    const fields = [
      ...module.fields,
      ...module.systemFields(),
    ]

    if (!!data.labels && Array.isArray(data.labels)) {
      const [dimension = {}] = report.dimensions
      let { field } = dimension

      if (!module) throw new Error('Module not found')

      field = fields.find(({ name }) => name === field)

      if (!field) throw new Error('Dimension field not found')

      data.labels = await prettyDimensionValues(dimension, field, data.labels)
    }

    // Multi-dimension charts (sankey, graph, heatmap, sunburst) rely on raw
    // report rows; prettify both dimension columns the same way labels are.
    const rowChartType = data.datasets?.[0]?.type
    const usesRows = ['sankey', 'graph', 'heatmap', 'sunburst'].includes(rowChartType)
    if (usesRows && Array.isArray(data.rows) && data.rows.length) {
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
  } catch (e) {
    error.value = e
    processing.value = false
  }

  setTimeout(() => {
    processing.value = false
    emit('updated')
  }, 300)
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
  const mod = store.module.getByID(id)
  return mod
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
