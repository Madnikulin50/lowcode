<template>
  <vue-select
    v-bind="$attrs"
    ref="vueSelectRef"
    v-model="_value"
    data-test-id="select"
    :clearable="clearable"
    :options="options"
    :searchable="searchable"
    :disabled="disabled"
    :selectable="selectable"
    :multiple="multiple"
    :loading="loading"
    :calculate-position="calculateDropdownPosition"
    :append-to-body="appendToBody"
    class="bg-white rounded"
    :class="sizeClass"
    @search="onSearch"
  >
    <template
      v-for="(_, name) in $slots"
      :key="name"
      #[name]="data"
    >
      <slot
        :name="name"
        v-bind="data"
      />
    </template>

    <template
      v-if="badge"
      #option="option"
    >
      <span
        class="badge rounded-pill"
        :style="getOptionStyle(option)"
      >
        {{ option.text }}
      </span>
    </template>

    <template
      v-if="!multiple && badge"
      #selected-option="option"
    >
      <span
        class="badge rounded-pill"
        :style="getOptionStyle(option)"
      >
        {{ option.text }}
      </span>
    </template>

    <template
      v-if="multiple && badge"
      #selected-option-container="{ option, deselect }"
    >
      <span
        class="d-flex align-items-center badge rounded-pill mx-1 mt-1 w-auto"
        :style="getOptionStyle(option)"
      >
        {{ option.text }}

        <font-awesome-icon
          :icon="['fas', 'times']"
          class="pointer ms-2"
          @click="deselect(option)"
        />
      </span>
    </template>
  </vue-select>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import VueSelect from 'vue-select'
import { createPopper } from '@popperjs/core'
import 'vue-select/dist/vue-select.css'

interface Option {
  text?: string
  disabled?: boolean
  style?: Record<string, string>
  [key: string]: unknown
}

interface DropdownList extends HTMLElement {
  style: CSSStyleDeclaration
}

interface VueSelectInstance {
  $el: HTMLElement
  $refs: {
    toggle: HTMLElement
  }
}

interface Props {
  modelValue?: string | unknown[] | Record<string, unknown>
  options?: Option[]
  clearable?: boolean
  searchable?: boolean
  appendToBody?: boolean
  defaultValue?: string | unknown[] | Record<string, unknown>
  size?: string
  disabled?: boolean
  selectable?: (option: Option) => boolean
  multiple?: boolean
  loading?: boolean
  badge?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => '',
  options: () => [],
  clearable: true,
  searchable: true,
  appendToBody: true,
  defaultValue: () => '',
  size: 'md',
  disabled: false,
  selectable: (o: Option) => !o.disabled,
  multiple: false,
  loading: false,
  badge: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: unknown]
  search: [query: string, loading: (v: boolean) => void]
}>()

const vueSelectRef = ref<VueSelectInstance | null>(null)
VueSelect
const query = ref('')

const _value = computed({
  get () {
    const fallbackValue = props.multiple ? [] : ''
    return props.defaultValue && props.modelValue === props.defaultValue ? fallbackValue : props.modelValue
  },
  set (v: unknown) {
    emit('update:modelValue', !v ? props.defaultValue : v)
  },
})

const sizeClass = computed(() =>
  props.size === 'sm' ? 'c-input-sm' : props.size === 'lg' ? 'c-input-lg' : ''
)

function calculateDropdownPosition (
  dropdownList: DropdownList,
  component: VueSelectInstance,
  { width }: { width: string }
) {
  dropdownList.style.width = width
  const popper = createPopper(component.$refs.toggle, dropdownList, {
    placement: 'bottom',
    modifiers: [
      {
        name: 'offset',
        options: {
          offset: [0, -1],
        },
      },
      {
        name: 'toggleClass',
        enabled: true,
        phase: 'write',
        fn ({ state }) {
          component.$el.classList.toggle('drop-up', state.placement === 'top')
        },
      },
    ],
  })
  return () => popper.destroy()
}

function onSearch (q: string, loading: (v: boolean) => void) {
  if (q !== query.value) {
    query.value = q
  }
  emit('search', q, loading)
}

// Theme tokens (danger, info, …) are stored in option.style; they are not
// valid CSS. Bootstrap `.badge` defaults to white text, so an unresolved
// background makes the selected value invisible on the white select.
const THEME_COLORS: Record<string, string> = {
  primary: '#4e73df',
  secondary: '#858796',
  success: '#1cc88a',
  info: '#36b9cc',
  warning: '#f6c23e',
  danger: '#e74a3b',
  light: '#f8f9fc',
  'extra-light': '#f8f9fc',
  dark: '#5a5c69',
  white: '#FFFFFF',
  black: '#0B344E',
  'body-bg': '#F3F5F7',
}

function resolveColor (value?: string, fallback = ''): string {
  if (!value) return fallback
  const v = String(value).trim()
  if (!v) return fallback
  if (v[0] === '#' || v.startsWith('rgb') || v.startsWith('hsl') || v.startsWith('var(')) {
    return v
  }
  const hex = THEME_COLORS[v]
  if (hex) return `var(--${v}, ${hex})`
  return v
}

function getOptionStyle ({ style = {} }: Option = {}) {
  if (!props.badge) return {}
  return {
    fontSize: '0.9rem',
    color: resolveColor(style.textColor, 'var(--dark)'),
    backgroundColor: resolveColor(style.backgroundColor, 'var(--extra-light)'),
  }
}
</script>

<style lang="scss">
:root {
  --vs-dropdown-bg: var(--white);
  --vs-dropdown-option--active-bg: var(--light);
  --vs-state-disabled-color: var(--secondary);
  --vs-state-disabled-bg: var(--light);
  --vs-colors--light: var(--black);
  --vs-colors--dark: var(--black);
  --vs-dropdown-option-color: var(--black);
  --vs-dropdown-option--active-color: var(--black);
  --vs-selected-bg: var(--extra-light);
  --vs-search-input-color: var(--secondary);
  --vs-search-input-bg: var(--white);
}

.v-select {
  min-width: auto;
  position: relative;
  -ms-flex: 1 1 auto;
  flex: 1 1 auto;
  margin-bottom: 0;
  font-size: .9rem !important;
  font-family: var(--font-regular);

  .vs__selected-options {
    width: 0;
    padding: 0;
  }

  .vs__selected {
    display: flex;
    white-space: nowrap;
    text-overflow: ellipsis;
    max-width: 100%;
    overflow: hidden;
    border: 0;
    color: var(--black);
  }

  .vs__search {
    font-size: .9rem;
    border: 0;
    padding: 0 2px;
    padding-top: 0.375rem;
    margin: 0;
  }

  &:not(.vs--open):not(.vs--loading) .vs__selected + .vs__search {
    width: 0;
    margin: 0;
    border: none;
    padding: 0;
    height: 0;
  }

  .vs__dropdown-toggle {
    min-height: calc(1.5em + 0.75rem + 4px);
    padding: 0.375rem calc(0.75rem - 2px);
    padding-top: 0 !important;
    border-width: 2px;
    border-color: var(--extra-light);

    .vs__selected {
      margin-top: 0.375rem;
    }

    .vs__actions {
      padding-top: 0.375rem;
      padding-right: 0;
    }
  }

  .vs__clear,
  .vs__open-indicator {
    fill: var(--black);
    display: inline-flex;
  }

  .vs__clear {
    padding: 0;
    border: 0;
    background-color: transparent;
    cursor: pointer;
    margin-right: 8px
  }

  &.vs--single {
    .vs__selected {
      margin-left: 0;
      margin-right: 0;
    }
  }
}

.vs--open {
  .vs__dropdown-toggle {
    border-color: var(--primary);
    border-radius: 0.25rem !important;
  }
}

.input-group > .v-select:not(:last-child) {
  .vs__dropdown-toggle {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
  }
}

.vs__spinner {
  border: .7em solid var(--dark);
  border-left-color: var(--white);
}

.vs__spinner, .vs__spinner::after {
  width: 4em;
  height: 4em;
}

.vs__dropdown-menu {
  z-index: 1100;

  .vs__dropdown-option {
    &.vs__dropdown-option--selected {
      background: var(--vs-dropdown-option--active-bg);
      color: var(--vs-dropdown-option--active-color);
    }

    &.vs__dropdown-option--disabled {
      cursor: var(--vs-state-disabled-cursor) !important;
      opacity: 0.5;
    }

    &:active {
      color: var(--white);
      background-color: var(--primary);
    }
  }
}

.c-input-sm {
  font-size: 0.7875rem !important;

  .vs__search {
    font-size: 0.7875rem;
    padding-top: 0.25rem;
  }

  .vs__dropdown-toggle {
    min-height: calc(1.5em + 0.5rem + 4px);
    padding: 0.25rem calc(0.5rem - 2px);
    border-radius: 0.2rem;

    .vs__selected {
      margin-top: 0.25rem;
    }

    .vs__actions {
      padding-top: 0.25rem;
    }
  }
}
.c-input-lg {
  font-size: 1.125rem !important;

  .vs__search {
    font-size: 1.125rem;
    padding-top: .5rem;
  }

  .vs__dropdown-toggle {
    min-height: calc(1.5em + 1rem + 4px);
    padding: .5rem calc(1rem - 2px);
    border-radius: 0.3rem;

    .vs__selected {
      margin-top: .5rem;
    }

    .vs__actions {
      padding-top: .5rem;
    }
  }
}
</style>
