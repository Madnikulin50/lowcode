<template>
  <div>
    <div class="d-flex align-items-center">
      <button
        class="btn p-0 rounded-circle bg-white border-white shadow-none"
        :style="`color: ${modelValue}; fill: ${modelValue};`"
        @click="toggleMenu"
      >
        <svg
          viewBox="0 0 32 32"
          :style="{ width: width, height: height }"
          class="border border-light rounded-circle"
        >
          <pattern
            id="checkerboard"
            width="12"
            height="12"
            patternUnits="userSpaceOnUse"
            fill="FFF"
          >
            <rect
              fill="#7080707f"
              x="0"
              width="6"
              height="6"
              y="0"
            />
            <rect
              fill="#7080707f"
              x="6"
              width="6"
              height="6"
              y="6"
            />
          </pattern>

          <circle
            cx="16"
            cy="16"
            r="16"
            fill="url(#checkerboard)"
          />

          <circle
            cx="16"
            cy="16"
            r="16"
          />
        </svg>
      </button>
      <span
        v-if="showText"
        class="ml-2"
      >
        {{ displayValue }}
      </span>
    </div>

    <Teleport to="body">
      <div
        v-if="showModal"
        class="modal fade show d-block"
        tabindex="-1"
      >
        <div class="modal-dialog modal-dialog-centered modal-md">
          <div class="modal-content">
            <div
              v-if="translations.modalTitle"
              class="modal-header"
            >
              <h5 class="modal-title">{{ translations.modalTitle }}</h5>
              <button
                type="button"
                class="btn-close"
                @click="closeMenu"
              />
            </div>

            <div class="modal-body p-0">
              <chrome
                :model-value="currentColor"
                class="w-100 shadow-none"
                @update:model-value="updateColor"
              />

              <div
                v-if="themes.length > 0"
                class="d-flex flex-column border-top p-3 gap-2"
              >
                <div
                  v-for="theme in themes"
                  :key="theme.id"
                  class="mb-3"
                >
                  <label class="form-label text-primary">
                    {{ translations[theme.id] }}
                  </label>
                  <div class="d-flex flex-wrap border">
                    <button
                      v-for="variable in themeVariables"
                      :key="variable.key"
                      :title="variable.label"
                      class="swatch flex-grow-1 rounded-0 btn"
                      :style="{ backgroundColor: theme.values[variable.key], borderColor: theme.values[variable.key] }"
                      @click="setSystemColor(theme.values[variable.key], variable.key)"
                    />
                  </div>
                </div>
              </div>
            </div>

            <div class="modal-footer">
              <button
                v-if="defaultValue"
                class="btn btn-light"
                @click="setColor()"
              >
                {{ translations.defaultBtnLabel || 'Default' }}
              </button>
              <slot name="footer" />

              <button
                class="btn btn-outline-light ml-auto text-primary border-0"
                @click="closeMenu"
              >
                {{ translations.cancelBtnLabel || 'Cancel' }}
              </button>

              <button
                class="btn btn-primary"
                @click="saveColor"
              >
                {{ translations.saveBtnLabel || 'Save' }}
              </button>
            </div>
          </div>
        </div>
      </div>
      <div
        v-if="showModal"
        class="modal-backdrop fade show"
        @click="closeMenu"
      />
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ChromePicker as Chrome } from 'vue-color'
import { debounce } from 'lodash'

const props = withDefaults(defineProps<{
  modelValue?: string
  defaultValue?: string
  translations?: Record<string, string>
  width?: string
  height?: string
  showText?: boolean
  themeSettings?: any[]
  themeVariables?: { key: string; label: string }[]
}>(), {
  modelValue: '#000000FF',
  defaultValue: '',
  translations: () => ({}),
  width: '32px',
  height: '32px',
  showText: true,
  themeSettings: () => [],
  themeVariables: () => [
    { key: 'white', label: 'White' },
    { key: 'black', label: 'Black' },
    { key: 'primary', label: 'Primary' },
    { key: 'secondary', label: 'Secondary' },
    { key: 'success', label: 'Success' },
    { key: 'warning', label: 'Warning' },
    { key: 'danger', label: 'Danger' },
    { key: 'light', label: 'Light' },
    { key: 'extra-light', label: 'Extra light' },
  ],
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const showModal = ref(false)
const currentColor = ref(props.modelValue)
const systemColorKey = ref('')

const displayValue = computed(() => {
  if (systemColorKey.value) return systemColorKey.value
  if (!currentColor.value || currentColor.value[0] !== '#') return currentColor.value
  for (const theme of themes.value) {
    for (const [k, v] of Object.entries(theme.values)) {
      if (v === currentColor.value) return k
    }
  }
  return currentColor.value
})

const themes = computed(() =>
  props.themeSettings
    .filter((theme: any) => theme.id !== 'general')
    .map((theme: any) => ({
      id: theme.id,
      values: JSON.parse(theme.values),
    })),
)

watch(() => props.modelValue, (value) => {
    if (currentColor.value !== value) {
      systemColorKey.value = ''
    }
    currentColor.value = value
    if (!value && props.defaultValue) {
      emit('update:modelValue', props.defaultValue)
    }
  }, { immediate: true })

const updateColor = debounce(function ({ hex8 = '' }) {
  currentColor.value = hex8
}, 300)

function setColor(defaultColor = props.defaultValue) {
  currentColor.value = defaultColor
}

function setSystemColor(hex: string, key?: string) {
  currentColor.value = hex
  systemColorKey.value = key || ''
  emit('update:modelValue', hex)
  closeMenu()
}

function saveColor() {
  systemColorKey.value = ''
  emit('update:modelValue', currentColor.value)
  closeMenu()
}

function toggleMenu() {
  if (showModal.value) {
    closeMenu()
  } else {
    openMenu()
  }
}

function openMenu() {
  showModal.value = true
}

function closeMenu() {
  showModal.value = false
}
</script>

<style lang="css">
@import 'vue-color/dist/vue-color.css';
</style>

<style lang="scss" scoped>
.swatch {
  height: 58px;
  min-width: 50px;
}
</style>

<style lang="scss">
.vc-chrome {
  font-family: var(--font-medium) !important;

  .vc-chrome-body {
    background: var(--white) !important;

    .vc-input__input {
      color: var(--black) !important;
      background-color: var(--white) !important;
    }

    .vc-input__label {
      color: var(--black) !important;
    }

    .vc-chrome-toggle-btn {
      path {
        fill: var(--black) !important;
      }

      .vc-chrome-toggle-icon-highlight {
        background: var(--light) !important;
      }
    }
  }
}
</style>
