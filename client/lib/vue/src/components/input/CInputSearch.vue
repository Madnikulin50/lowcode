<template>
  <div
    style="min-width: 150px;"
    :class="{ submittable: isSubmittable }"
    class="c-input-search d-flex position-relative"
  >
    <input
      ref="searchInput"
      data-test-id="input-search"
      :type="inputType"
      name="search"
      :value="localValue"
      :disabled="disabled"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :class="['form-control', inputSize, 'text-truncate']"
      @input="onInput"
      @keyup.enter="submitQuery"
    />

    <div
      v-if="loading"
      class="spinner d-inline-flex align-items-center p-3"
    >
      <span class="spinner-border spinner-border-sm text-secondary" />
    </div>

    <button
      v-else-if="clearable && localValue && !disabled"
      class="clear-button d-inline-flex align-items-center rounded-0 p-3 btn btn-outline-extra-light"
      @click="onClear"
    >
      <font-awesome-icon
        :icon="['fas', 'times']"
        class="text-secondary"
      />
    </button>

    <button
      v-if="showSubmittable"
      :class="[isSubmittable ? 'btn-outline-light' : 'btn-link', isSubmittable ? '' : 'border-0 cursor-default', 'btn search-button d-inline-flex align-items-center rounded-0 border-light']"
      :disabled="disabled"
      @[isSubmittable]="submitQuery"
    >
      <font-awesome-icon
        :icon="['fas', 'search']"
        class="align-middle text-primary"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { debounce } from 'lodash'

interface Props {
  modelValue?: string
  placeholder?: string
  size?: string
  disabled?: boolean
  clearable?: boolean
  submittable?: boolean
  autocomplete?: string
  debounceMs?: number
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '',
  size: 'md',
  disabled: false,
  clearable: true,
  submittable: false,
  autocomplete: 'on',
  debounceMs: 0,
  loading: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  search: [value: string]
}>()

const searchInput = ref<HTMLInputElement | null>(null)
const localValue = ref(props.modelValue)

const inputSize = computed(() => props.size === 'sm' ? 'form-control-sm' : props.size === 'lg' ? 'form-control-lg' : '')
const inputType = computed(() => props.clearable ? 'search' : 'text')
const showSubmittable = computed(() => !localValue.value || (props.clearable && props.submittable))
const isSubmittable = computed(() => props.submittable && !props.disabled ? 'click' : null)

const debouncedEmit = computed(() => {
  if (props.debounceMs > 0) {
    return debounce((value: string) => {
      emit('update:modelValue', value)
    }, props.debounceMs)
  }
  return null
})

function onInput (e: Event) {
  const value = (e.target as HTMLInputElement).value
  localValue.value = value
  if (props.debounceMs > 0 && debouncedEmit.value) {
    debouncedEmit.value(value)
  } else {
    emit('update:modelValue', value)
  }
}

function submitQuery () {
  if (props.submittable) {
    emit('search', localValue.value)
  }
}

function onClear () {
  localValue.value = ''
  emit('update:modelValue', '')
  if (debouncedEmit.value) {
    debouncedEmit.value.cancel()
  }
  requestAnimationFrame(() => {
    searchInput.value?.focus()
  })
}

watch(() => props.modelValue, (value) => {
  localValue.value = value
})
</script>

<style lang="scss" scoped>
input:focus::placeholder {
  color: transparent;
}

.c-input-search {
  .search-button {
    position: absolute;
    right: 2px;
    top: 2px;
    bottom: 2px;
    z-index: 4;
    border-left-width: 2px;
  }

  .clear-button,
  .spinner {
    position: absolute;
    right: 1px;
    top: 1px;
    bottom: 1px;
    z-index: 5;
    outline: none !important;
    box-shadow: none !important;
    border: none !important;
    background-color: transparent !important;

    &:hover {
      text-decoration: none;
    }
  }

  &.submittable {
    .clear-button,
    .spinner {
      right: 48px;
    }
  }

  .form-control {
    padding-right: 40px;
  }

  &.submittable .form-control {
    padding-right: 85px;
  }

  ::-webkit-search-cancel-button {
    -webkit-appearance: none;
    display: none;
  }
}

.cursor-default {
  cursor: default !important;
}
</style>
