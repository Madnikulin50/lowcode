<template>
  <div>
    <h5>{{ $t('anomaly.label') }}</h5>

    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('anomaly.configurator.module') }}</label>
      <c-input-select
        v-model="options.moduleID"
        :options="moduleOptions"
        label="label"
        :reduce="o => o.value"
        :clearable="false"
      />
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('anomaly.severity') }}</label>
      <c-input-select
        v-model="options.severity"
        :options="severityOptions"
        label="label"
        :reduce="o => o.value"
        :clearable="false"
      />
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('anomaly.status') }}</label>
      <c-input-select
        v-model="options.status"
        :options="statusOptions"
        label="label"
        :reduce="o => o.value"
        :clearable="false"
      />
    </div>

    <div class="form-check mb-2">
      <input
        id="anomaly-show-summary"
        v-model="options.showSummary"
        type="checkbox"
        class="form-check-input"
      >
      <label
        class="form-check-label"
        for="anomaly-show-summary"
      >{{ $t('anomaly.configurator.showSummary') }}</label>
    </div>

    <div class="form-check mb-3">
      <input
        id="anomaly-show-list"
        v-model="options.showList"
        type="checkbox"
        class="form-check-input"
      >
      <label
        class="form-check-label"
        for="anomaly-show-list"
      >{{ $t('anomaly.configurator.showList') }}</label>
    </div>

    <div
      v-if="options.showList"
      class="mb-3"
    >
      <label class="form-label text-primary">{{ $t('anomaly.configurator.limit') }}</label>
      <input
        v-model.number="options.limit"
        type="number"
        min="1"
        max="100"
        class="form-control"
      >
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('configurator.refreshRate', 'Refresh rate (s)') }}</label>
      <input
        v-model.number="options.refreshRate"
        type="number"
        min="0"
        class="form-control"
      >
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useStore } from '../../store'

const { t: $t } = useI18n({ useScope: 'global' })
const store = useStore()

const props = defineProps({
  block: { type: Object, required: true },
  namespace: { type: Object, required: false, default: undefined },
  page: { type: Object, required: true },
})

const options = computed(() => props.block.options)

const modules = computed(() => store.module.set)

const moduleOptions = computed(() => [
  { value: '0', label: $t('anomaly.allModules') },
  ...modules.value.map((m) => ({ value: m.moduleID, label: m.name || m.handle })),
])

const severityOptions = computed(() => [
  { value: '', label: $t('anomaly.allSeverities') },
  { value: 'low', label: $t('anomaly.severityLow') },
  { value: 'medium', label: $t('anomaly.severityMedium') },
  { value: 'high', label: $t('anomaly.severityHigh') },
])

const statusOptions = computed(() => [
  { value: '', label: $t('anomaly.allStatuses') },
  { value: 'new', label: $t('anomaly.statusNew') },
  { value: 'acknowledged', label: $t('anomaly.statusAcknowledged') },
  { value: 'resolved', label: $t('anomaly.statusResolved') },
  { value: 'false_positive', label: $t('anomaly.statusFalsePositive') },
])
</script>
