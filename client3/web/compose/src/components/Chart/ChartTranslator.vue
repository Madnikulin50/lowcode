<template>
  <c-translator-button
    v-if="canManageResourceTranslations && resourceTranslationsEnabled"
    v-bind="{ ...$attrs, ...$props }"
    :tooltip="$t('tooltip')"
    :resource="resource"
    :titles="titles"
    :fetcher="fetcher"
    :updater="updater"
  />
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'

const { t } = useI18n()

const $ComposeAPI = window.__composeAPI
const $Settings = window.__settings

const props = defineProps({
  field: {
    type: String,
    default: '',
  },
  chart: {
    type: compose.Chart,
    required: true,
  },
  highlightKey: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: () => false,
  },
})

const emit = defineEmits(['update:chart', 'update:field'])

const resourceTranslationsEnabled = computed(() => $Settings.get('resourceTranslations.enabled', false))
const canManageResourceTranslations = computed(() => {
  try {
    return window.__auth?.can?.('compose/', 'resource-translations.manage')
  } catch {
    return false
  }
})

const resource = computed(() => {
  const { namespaceID, chartID } = props.chart
  return `compose:chart/${namespaceID}/${chartID}`
})

const titles = computed(() => {
  const { chartID, handle } = props.chart
  const tls = {}
  tls[resource.value] = t('title', { handle: handle || chartID })
  return tls
})

const fetcher = computed(() => {
  const { namespaceID, chartID } = props.chart
  return () => $ComposeAPI.chartListTranslations({ namespaceID, chartID })
})

const updater = computed(() => {
  const { namespaceID, chartID } = props.chart
  return (translations) => {
    return $ComposeAPI
      .chartUpdateTranslations({ namespaceID, chartID, translations })
      .then(() => fetcher.value())
      .then((translations) => {
        const find = (key) => translations.find(t => t.key === key && t.resource === resource.value)
        let tr
        const [report = {}] = props.chart.config.reports
        tr = find('yAxis.label')
        if (tr !== undefined) {
          report.yAxis.label = tr.message
        }
        report.metrics.forEach((metric) => {
          tr = find(`metrics.${metric.metricID}.label`)
          if (tr) {
            metric.label = tr.message
          }
        })
        if (Array.isArray(report.dimensions)) {
          report.dimensions.forEach(d => {
            if (!Array.isArray(d.meta.steps)) return
            d.meta.steps.forEach((step) => {
              tr = find(`dimensions.${d.dimensionID}.meta.steps.${step.stepID}.label`)
              if (tr) {
                step.label = tr.message
              }
            })
          })
        }
        return props.chart
      })
      .then(chart => {
        emit('update.chart', chart)
      })
  }
})
</script>
