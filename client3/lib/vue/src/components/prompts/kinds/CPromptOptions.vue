<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />
    <label class="text-primary">{{ label }}</label>
    <c-input-select
      v-if="inputType === 'select'"
      v-model="value"
      :options="itemOptions"
      :disabled="loading"
      :multiple="multiple"
      append-to-body
      label="text"
      :get-option-key="r => r.value"
      :placeholder="placeholder"
      :reduce="r => r.value"
      class="w-100"
    />
    <template v-else-if="inputType === 'radio'">
      <div
        v-for="opt in itemOptions"
        :key="opt.value"
        class="form-check"
      >
        <input
          :id="`radio-${opt.value}`"
          v-model="value"
          class="form-check-input"
          type="radio"
          :value="opt.value"
          :disabled="loading"
        >
        <label
          class="form-check-label"
          :for="`radio-${opt.value}`"
        >{{ opt.text }}</label>
      </div>
    </template>
    <button
      :disabled="loading"
      class="btn btn-primary ms-auto"
      @click="emit('submit', { value: encodeValue() })"
    >
      {{ pVal('buttonLabel', 'Submit') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeMount } from 'vue'
import { pVal as _pVal, pType as _pType } from '../utils'
import CInputSelect from '../../input/CInputSelect.vue'

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
  'select',
  'radio',
]

const value = ref<any>()

const message = computed(() => _pVal(props.payload, 'message', ''))
const label = computed(() => _pVal(props.payload, 'label', ''))

const itemOptions = computed(() => {
  const out: Array<{ value: string; text: string }> = []
  const options = pVal('options', {})
  for (const val in options) {
    out.push({ value: val, text: options[val] })
  }
  return out
})

const inputType = computed(() => {
  const t = pVal('type', 'text')
  if (validTypes.indexOf(t) === -1) {
    return 'select'
  }
  return t
})

const multiple = computed(() => pVal('multiselect', false))

const placeholder = computed(() => pVal('placeholder', 'Select an option'))

function pVal(k: string, def?: any) {
  return _pVal(props.payload, k, def)
}

function pType(k: string, def?: any) {
  return _pType(props.payload, k, def)
}

function encodeValue() {
  if (Array.isArray(value.value)) {
    return {
      '@type': 'Array',
      '@value': value.value || [],
    }
  } else {
    return { '@type': 'String', '@value': value.value }
  }
}

onBeforeMount(() => {
  let val = pVal('value')
  if (pVal('multiselect', false)) {
    if (Array.isArray(val)) {
      val = val.map((v: any) => v['@value'])
    } else {
      val = val ? [val] : []
    }
  }
  value.value = val
})
</script>
