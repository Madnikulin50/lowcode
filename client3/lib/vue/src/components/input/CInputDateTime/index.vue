<template>
  <div class="c-input-date-time d-flex flex-wrap w-100 gap-1">
    <input
      v-if="!noDate"
      type="date"
      :value="date"
      data-test-id="picker-date"
      :placeholder="labels.none"
      :min="minDate"
      :max="maxDate"
      class="form-control h-100 overflow-hidden"
      @input="onDateInput"
    />

    <input
      v-if="!noTime"
      type="time"
      :value="time"
      data-test-id="picker-time"
      :placeholder="labels.none"
      class="form-control h-100 overflow-hidden"
      @input="onTimeInput"
    />

    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { getDate, getTime, setDate, setTime } from './lib/index.ts'
import { shared } from 'corteza-lib/js/dist'

const { getWeekStartDay } = shared

const props = defineProps<{
  modelValue?: string | Date
  noTime?: boolean
  noDate?: boolean
  onlyFuture?: boolean
  onlyPast?: boolean
  size?: string
  labels: { none: string; clear: string; today: string; now: string }
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string | undefined]
}>()

const date = computed({
  get: () => getDate(props.modelValue as string),
  set: (date: string | undefined) => {
    emit('update:modelValue', setDate(date, props.modelValue as string, props.noDate, props.noTime))
  },
})

const time = computed({
  get: () => getTime(props.modelValue as string),
  set: (time: string | undefined) => {
    emit('update:modelValue', setTime(time, props.modelValue as string, props.noDate, props.noTime))
  },
})

const minDate = computed(() => props.onlyFuture ? new Date().toISOString().split('T')[0] : undefined)
const maxDate = computed(() => props.onlyPast ? new Date().toISOString().split('T')[0] : undefined)

function onDateInput(e: Event) {
  date.value = (e.target as HTMLInputElement).value || undefined
}

function onTimeInput(e: Event) {
  time.value = (e.target as HTMLInputElement).value || undefined
}
</script>

<style lang="scss">
.c-input-date-time {
  min-width: 120px;

  .btn {
    padding: 0.25rem 0.5rem;
  }

  label {
    font-family: var(--font-regular);
    color: var(--black) !important;
  }

  input[type="date"],
  input[type="time"] {
    flex: 1 0 130px;
  }
}
</style>
