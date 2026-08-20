<template>
  <div class="d-inline-flex">
    <button
      ref="btnRef"
      :id="popoverTarget"
      :title="$t('recordList.filter.title')"
      type="button"
      :class="['btn', `btn-${variant}`, 'd-flex align-items-center d-print-none border-0 px-1 h-100', buttonClass]"
      :style="buttonStyle"
      @click.stop="toggle"
    >
      <font-awesome-icon
        :icon="['fas', 'filter']"
        :class="[inFilter ? 'text-primary' : inactiveIconClass]"
      />
    </button>

    <Teleport to="body">
      <div
        v-if="open"
        ref="panelRef"
        class="record-list-filter card border-0"
        :style="panelStyle"
        @click.stop
      >
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
              type="button"
              class="btn btn-primary"
              @click="onSave(true, 'filter')"
            >
              {{ $t('label.save') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import FilterToolbox from 'corteza-webapp-compose/src/components/Common/FilterToolbox.vue'

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

const open = ref(false)
const btnRef = ref(null)
const panelRef = ref(null)
const panelStyle = ref({})
const componentFilter = ref([])
const preventPopoverClose = ref(false)

const inFilter = computed(() => {
  return (props.recordListFilter || []).some(({ filter }) => {
    return (filter || []).some(({ name }) => name === (props.selectedField || {}).name)
  })
})

const popoverTarget = computed(() => {
  return `${props.target || '0'}-${(props.selectedField || {}).name}`
})

watch(() => props.recordListFilter, (val) => {
  if (!open.value) {
    componentFilter.value = [...(val || [])]
  }
}, { immediate: true, deep: true })

function toggle () {
  if (open.value) {
    closePanel()
    return
  }
  componentFilter.value = [...(props.recordListFilter || [])]
  open.value = true
  nextTick(() => {
    positionPanel()
    document.addEventListener('mousedown', onDocMouseDown)
  })
}

function closePanel () {
  open.value = false
  document.removeEventListener('mousedown', onDocMouseDown)
}

function positionPanel () {
  const btn = btnRef.value
  if (!btn) return
  const r = btn.getBoundingClientRect()
  const width = Math.min(800, Math.max(350, window.innerWidth - 16))
  let left = r.left
  if (left + width > window.innerWidth - 8) left = window.innerWidth - width - 8
  if (left < 8) left = 8
  const maxH = Math.min(400, window.innerHeight - 24)
  let top = r.bottom + 8
  if (top + maxH > window.innerHeight && r.top > maxH + 8) {
    top = Math.max(8, r.top - maxH - 8)
  }
  panelStyle.value = {
    position: 'fixed',
    top: `${top}px`,
    left: `${left}px`,
    width: `${width}px`,
    maxHeight: `${maxH}px`,
    zIndex: 1080,
  }
}

function isExemptTarget (el) {
  if (!el || !el.closest) return false
  return !!el.closest('.vs__dropdown-menu, .vs__dropdown, .flatpickr-calendar, .datepicker, .modal, .record-list-filter')
}

function onDocMouseDown (e) {
  if (!open.value || preventPopoverClose.value) return
  const t = e.target
  if (btnRef.value?.contains(t) || panelRef.value?.contains(t) || isExemptTarget(t)) return
  closePanel()
}

function onWindowChange () {
  if (open.value) positionPanel()
}

function preventClose () {
  preventPopoverClose.value = true
  setTimeout(() => {
    preventPopoverClose.value = false
  }, 150)
}

function activeFilterGroups (raw) {
  return (Array.isArray(raw) ? raw : []).filter(({ filter = [] }) =>
    filter.filter((f = {}) => !!f.name).length > 0,
  )
}

function resetFilter () {
  componentFilter.value = []
  emit('reset')
  emit('filter', [])
  closePanel()
}

function onSave (close = true, type = 'filter') {
  const groups = activeFilterGroups(componentFilter.value)
  if (close) closePanel()
  emit(type, groups)
}

onMounted(() => {
  window.addEventListener('resize', onWindowChange)
  window.addEventListener('scroll', onWindowChange, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocMouseDown)
  window.removeEventListener('resize', onWindowChange)
  window.removeEventListener('scroll', onWindowChange, true)
})
</script>

<style lang="scss">
.record-list-filter {
  display: flex;
  flex-direction: column;
  max-width: 800px !important;
  color: var(--black);
  background: var(--white);
  border-radius: 0.25rem;
  box-shadow: 0 3px 48px #00000026;
  font-size: 0.9rem;

  .card-body {
    max-height: 20rem;
  }

  .v-select,
  .field-operator,
  .field-editor {
    min-width: 120px;
  }
}
</style>
