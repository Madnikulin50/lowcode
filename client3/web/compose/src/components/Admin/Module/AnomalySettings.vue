<template>
  <div
    v-if="module"
    class="p-3"
  >
    <h5 class="mb-1">{{ $t('edit.anomalySettings.title') }}</h5>
    <p class="text-muted small mb-3">
      {{ $t('edit.anomalySettings.description') }}
    </p>

    <div
      v-if="loading"
      class="text-center py-4"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>

    <p
      v-else-if="!rows.length"
      class="text-muted fst-italic"
    >
      {{ $t('edit.anomalySettings.noNumericFields') }}
    </p>

    <table
      v-else
      class="table align-middle"
    >
      <thead>
        <tr>
          <th>{{ $t('edit.anomalySettings.field') }}</th>
          <th>{{ $t('edit.anomalySettings.enabled') }}</th>
          <th>{{ $t('edit.anomalySettings.detector') }}</th>
          <th style="min-width: 12rem;">{{ $t('edit.anomalySettings.threshold') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="row in rows"
          :key="row.field.name"
        >
          <td>{{ row.field.label || row.field.name }}</td>
          <td>
            <div class="form-check form-switch">
              <input
                :id="`anomaly-enabled-${row.field.name}`"
                v-model="row.enabled"
                class="form-check-input"
                type="checkbox"
              >
            </div>
          </td>
          <td>
            <select
              v-model="row.detector"
              class="form-select form-control"
              :disabled="!row.enabled"
              @change="onDetectorChange(row)"
            >
              <option
                v-for="opt in detectorOptions"
                :key="opt.value"
                :value="opt.value"
              >{{ opt.label }}</option>
            </select>
          </td>
          <td>
            <template v-if="row.detector === 'range'">
              <div class="d-flex gap-1 align-items-center">
                <input
                  v-model.number="row.min"
                  type="number"
                  class="form-control form-control-sm"
                  :placeholder="$t('edit.anomalySettings.min')"
                  :disabled="!row.enabled"
                >
                <input
                  v-model.number="row.max"
                  type="number"
                  class="form-control form-control-sm"
                  :placeholder="$t('edit.anomalySettings.max')"
                  :disabled="!row.enabled"
                >
              </div>
            </template>
            <template v-else-if="row.detector === 'new_category'">
              <span class="text-muted small">{{ $t('edit.anomalySettings.noThreshold') }}</span>
            </template>
            <template v-else>
              <input
                v-model.number="row.threshold"
                type="number"
                step="0.01"
                min="0"
                class="form-control"
                :disabled="!row.enabled"
              >
            </template>
            <small class="text-muted d-block mt-1">{{ detectorHint(row.detector) }}</small>
          </td>
        </tr>
      </tbody>
    </table>

    <div
      v-if="rows.length"
      class="d-flex justify-content-end"
    >
      <span
        v-if="processing"
        class="spinner-border spinner-border-sm me-2"
      />
      <button
        class="btn btn-primary"
        :disabled="processing"
        @click="onSave"
      >
        {{ $t('label.save') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'module',
  },
})

const props = defineProps({
  namespace: {
    type: compose.Namespace,
    required: true,
  },
  module: {
    type: compose.Module,
    required: true,
  },
})

// Registered in main.js via CortezaAPI('anomaly', ...) - see that file for
// why its baseURL is overridden (anomaly's REST routes are mounted without
// a namespace prefix, unlike every other CortezaAPI(...) service).
const $AnomalyAPI = window.__anomalyAPI

const loading = ref(false)
const processing = ref(false)
const rows = ref([])

// Default threshold per detector - substituted in whenever the detector for
// a row changes, since each one's threshold means something different
// (stddev multiplier vs. accumulated-sum limit vs. a bare fraction).
const defaultThreshold = {
  zscore: 3,
  ewma: 3,
  cusum: 5,
  mad: 3.5,
  rare_category: 0.01,
  new_category: 0,
  range: 0,
}

const detectorOptions = computed(() => [
  { value: 'zscore', label: t('edit.anomalySettings.detectorZscore') },
  { value: 'ewma', label: t('edit.anomalySettings.detectorEwma') },
  { value: 'cusum', label: t('edit.anomalySettings.detectorCusum') },
  { value: 'mad', label: t('edit.anomalySettings.detectorMad') },
  { value: 'range', label: t('edit.anomalySettings.detectorRange') },
  { value: 'rare_category', label: t('edit.anomalySettings.detectorRareCategory') },
  { value: 'new_category', label: t('edit.anomalySettings.detectorNewCategory') },
])

function detectorHint (detector) {
  return t(`edit.anomalySettings.hint.${detector}`)
}

function onDetectorChange (row) {
  row.threshold = defaultThreshold[row.detector] ?? 3
}

onMounted(() => {
  load()
})

async function load () {
  rows.value = (props.module.fields || []).map((field) => ({
    field,
    ruleID: undefined,
    enabled: false,
    detector: 'zscore',
    threshold: defaultThreshold.zscore,
    min: undefined,
    max: undefined,
  }))

  if (!$AnomalyAPI || !rows.value.length) return

  loading.value = true
  await $AnomalyAPI.ruleSearch({ namespaceID: props.module.namespaceID, moduleID: props.module.moduleID })
    .then(({ set = [] }) => {
      set.forEach((rule) => {
        const row = rows.value.find((r) => r.field.name === rule.field)
        if (!row) return

        row.ruleID = rule.ruleID
        row.enabled = rule.enabled
        row.detector = rule.detector
        row.threshold = rule.threshold

        if (rule.detector === 'range' && rule.params) {
          row.min = rule.params.min
          row.max = rule.params.max
        }
      })
    })
    .catch((err) => { console.error('failed to load anomaly rules', err) })
    .finally(() => { loading.value = false })
}

async function onSave () {
  processing.value = true

  for (const row of rows.value) {
    try {
      const params = row.detector === 'range' ? JSON.stringify({ min: row.min, max: row.max }) : undefined

      if (row.ruleID) {
        await $AnomalyAPI.ruleUpdate({
          namespaceID: props.module.namespaceID,
          ruleID: row.ruleID,
          detector: row.detector,
          threshold: row.threshold,
          enabled: row.enabled,
          params,
        })
      } else if (row.enabled) {
        const created = await $AnomalyAPI.ruleCreate({
          namespaceID: props.module.namespaceID,
          moduleID: props.module.moduleID,
          field: row.field.name,
          detector: row.detector,
          threshold: row.threshold,
          enabled: true,
          params,
        })
        row.ruleID = created.ruleID
      }
    } catch (err) {
      console.error('failed to save anomaly rule', row.field.name, err)
    }
  }

  processing.value = false
}
</script>
