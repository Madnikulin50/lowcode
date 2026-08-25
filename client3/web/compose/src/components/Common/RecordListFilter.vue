<template>
  <div>
    <button
      :id="popoverTarget"
      :title="$t('recordList.filter.title')"
      type="button"
      :class="['btn', `btn-${variant}`, 'd-flex align-items-center d-print-none border-0 px-1 h-100', buttonClass]"
      :style="buttonStyle"
      data-bs-toggle="popover"
      data-bs-trigger="click"
      data-bs-placement="bottom"
      data-bs-boundary="window"
      :data-bs-content="popoverContent"
    >
      <font-awesome-icon
        :icon="['fas', 'filter']"
        :class="[inFilter ? 'text-primary' : inactiveIconClass]"
      />
    </button>

    <div
      ref="popoverEl"
      class="d-none"
    >
      <div class="card position-static w-100 border-0">
        <div class="card-body px-3 pb-0 overflow-auto">
          <filter-toolbox
            v-model="componentFilter"
            :module="module"
            :namespace="namespace"
            :selected-field="selectedField"
            @value-change="preventClose"
          />
        </div>

        <div class="card-footer d-flex justify-content-between shadow-sm rounded">
          <button
            type="button"
            class="btn btn-outline-secondary"
            @click="resetFilter"
          >
            {{ $t('label.reset') }}
          </button>

          <div class="d-flex">
            <button
              v-if="allowFilterPresetSave"
              type="button"
              class="btn btn-outline-primary me-2"
              @click="onSave(true, 'filter-preset')"
            >
              {{ $t('recordList.filter.addFilterToPreset') }}
            </button>
            <button
              ref="btnSave"
              type="button"
              class="btn btn-primary"
              @click="onSave"
            >
              {{ $t('label.save') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import FilterToolbox from 'corteza-webapp-compose/src/components/Common/FilterToolbox.vue'

const { t } = useI18n()

const props = defineProps({
  target: {
    type: String,
    default: '',
  },
  selectedField: {
    type: Object,
    default: undefined,
  },
  namespace: {
    type: Object,
    required: true,
  },
  module: {
    type: Object,
    required: true,
  },
  recordListFilter: {
    type: Array,
    required: true,
  },
  variant: {
    type: String,
    default: 'outline-light',
  },
  inactiveIconClass: {
    type: String,
    default: 'text-secondary',
  },
  buttonClass: {
    type: String,
    default: '',
  },
  buttonStyle: {
    type: String,
    default: '',
  },
  allowFilterPresetSave: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['filter', 'reset', 'filter-preset'])

const componentFilter = ref([])
const preventPopoverClose = ref(false)
const popoverEl = ref(null)
const popoverContent = computed(() => popoverEl.value?.innerHTML || '')

const inFilter = computed(() => {
  return props.recordListFilter.some(({ filter }) => {
    return filter.some(({ name }) => name === (props.selectedField || {}).name)
  })
})

const popoverTarget = computed(() => {
  return `${props.target || '0'}-${(props.selectedField || {}).name}`
})

watch(() => props.recordListFilter, (val) => {
  componentFilter.value = [...val]
}, { immediate: true, deep: true })

function onHide (e) {
  if (preventPopoverClose.value) {
    e.preventDefault()
  }
}

function onOpen () {
  componentFilter.value = [...props.recordListFilter]
}

function preventClose () {
  preventPopoverClose.value = true
  setTimeout(() => {
    preventPopoverClose.value = false
  }, 100)
}

function resetFilter () {
  componentFilter.value = undefined
  emit('reset')
}

function onSave (close = true, type = 'filter') {
  if (close) {
    const popover = bootstrap.Popover.getInstance(document.getElementById(popoverTarget.value))
    if (popover) popover.hide()
  }
  setTimeout(() => {
    emit(type, componentFilter.value.filter(({ filter = [] }) => filter.filter((f = {}) => !!f.name).length > 0))
  }, 100)
}
</script>

<style lang="scss">
.record-list-filter {
  z-index: 1040;
  max-width: 800px !important;
  opacity: 1 !important;
  border-color: transparent;

  .popover-body {
    display: flex;
    width: 800px;
    min-width: min(99vw, 350px);
    max-width: 99vw;
    max-height: 25rem;
    padding: 0;
    color: var(--black);
    background: var(--white);
    border-radius: 0.25rem;
    opacity: 1 !important;
    box-shadow: 0 3px 48px #00000026;
    font-size: 0.9rem;
  }

  .v-select,
  .field-operator,
  .field-editor {
    min-width: 120px;
  }

  .arrow {
    &::before {
      border-bottom-color: var(--white);
      border-top-color: var(--white);
    }

    &::after {
      border-top-color: var(--white);
    }
  }
}
</style>
