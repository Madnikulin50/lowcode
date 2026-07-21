<template>
  <div>
    <div class="mb-3 w-25">
      <label class="form-label text-primary">{{ t('kind.dateTime.type.label') }}</label>
      <div class="btn-group" data-bs-toggle="buttons">
        <label
          v-for="opt in typeOptions"
          :key="opt.value"
          class="btn btn-outline-primary btn-sm"
          :class="{ active: inputType === opt.value }"
          :title="hasData ? t('not-configurable') : ''"
        >
          <input
            type="radio"
            class="btn-check"
            :checked="inputType === opt.value"
            :disabled="hasData"
            autocomplete="off"
            @change="onTypeChange(opt.value)"
          />
          {{ opt.text }}
        </label>
      </div>
    </div>

    <div v-if="!f.options.onlyTime" class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t('kind.dateTime.constraints.label') }}</label>
      <div class="btn-group" data-bs-toggle="buttons">
        <label
          v-for="opt in constraintOptions"
          :key="opt.value"
          class="btn btn-outline-primary btn-sm"
          :class="{ active: constraintType === opt.value }"
        >
          <input
            type="radio"
            class="btn-check"
            :checked="constraintType === opt.value"
            autocomplete="off"
            @change="onConstraintChange(opt.value)"
          />
          {{ opt.text }}
        </label>
      </div>
    </div>

    <div class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t('kind.dateTime.outputFormat') }}</label>
      <div class="form-check mb-2">
        <input
          id="outputRelative"
          v-model="f.options.outputRelative"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label" for="outputRelative">{{ t('kind.dateTime.relativeOutput') }}</label>
      </div>
      <template v-if="!f.options.outputRelative">
        <input
          v-model="f.options.format"
          type="text"
          class="form-control form-control-sm"
          placeholder="YYYY-MM-DD HH:ii"
        />
        <div class="small text-muted">
          <label>{{ t('kind.dateTime.outputFormatFootnote', { interpolation: { escapeValue: false } }, { tag: 'label' }) }}</label>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const typeOptions = computed(() => [
  { value: 'dateTime', text: t('kind.dateTime.type.options.dateTime') },
  { value: 'date', text: t('kind.dateTime.type.options.date') },
  { value: 'time', text: t('kind.dateTime.type.options.time') },
])

const constraintOptions = computed(() => [
  { value: 'all', text: t('kind.dateTime.constraints.options.all') },
  { value: 'pastValuesOnly', text: t('kind.dateTime.constraints.options.pastValuesOnly') },
  { value: 'futureValuesOnly', text: t('kind.dateTime.constraints.options.futureValuesOnly') },
])

const inputType = computed(() => {
  if (f.value.options.onlyDate) return 'date'
  if (f.value.options.onlyTime) return 'time'
  return 'dateTime'
})

const constraintType = computed(() => {
  if (f.value.options.onlyPastValues) return 'pastValuesOnly'
  if (f.value.options.onlyFutureValues) return 'futureValuesOnly'
  return 'all'
})

function onTypeChange (v) {
  f.value.options.onlyDate = v === 'date'
  f.value.options.onlyTime = v === 'time'
}

function onConstraintChange (v) {
  f.value.options.onlyPastValues = v === 'pastValuesOnly'
  f.value.options.onlyFutureValues = v === 'futureValuesOnly'
}
</script>
