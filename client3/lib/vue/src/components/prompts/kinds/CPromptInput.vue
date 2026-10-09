<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />
    <label class="text-primary">{{ label }}</label>
    <c-input-date-time
      v-if="type === 'date' || type === 'time' || type === 'datetime'"
      v-model="value"
      :no-date="type === 'time'"
      :no-time="type === 'date'"
      :disabled="loading"
      :labels="labels"
    />
    <input
      v-else
      v-model="value"
      class="form-control"
      :type="type"
      :disabled="loading"
    >
    <button
      :disabled="loading"
      class="btn btn-primary ms-auto"
      @click="emit('submit', { value: { '@value': value, '@type': 'String' }})"
    >
      {{ pVal('buttonLabel', 'Submit') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeMount } from 'vue'
import { getCurrentInstance } from 'vue'
import { pVal as _pVal, pType as _pType } from '../utils'
import { CInputDateTime } from '../../input'

const { $t } = getCurrentInstance()!.appContext.config.globalProperties as any

const props = withDefaults(defineProps<{
  loading?: boolean
  payload?: Record<string, any>
}>(), {
  loading: false,
  payload: () => ({}),
})

const emit = defineEmits<{
  (e: 'submit', value: Record<string, any>): void
}>()

const validTypes = [
  'text',
  'number',
  'email',
  'password',
  'search',
  'date',
  'time',
  'datetime',
]

const value = ref<any>()

const message = computed(() => _pVal(props.payload, 'message', ''))
const label = computed(() => _pVal(props.payload, 'label', ''))

const type = computed(() => {
  const t = pVal('type', 'text')
  if (validTypes.indexOf(t) === -1) {
    return 'text'
  }
  return t
})

const labels = computed(() => ({
  clear: $t('label.clear'),
  none: $t('label.none'),
  now: $t('label.now'),
  today: $t('label.today'),
}))

function pVal(k: string, def?: any) {
  return _pVal(props.payload, k, def)
}

function pType(k: string, def?: any) {
  return _pType(props.payload, k, def)
}

onBeforeMount(() => {
  value.value = pVal('inputValue')
})
</script>
