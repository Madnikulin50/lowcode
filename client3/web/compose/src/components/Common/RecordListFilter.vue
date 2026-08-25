<template>
  <div>
    <BPopover
      v-model="popoverOpen"
      class="record-list-filter shadow-sm"
      teleport-to="body"
      strategy="fixed"
      manual
      no-auto-close
      no-hide
      placement="bottom"
      :delay="0"
      boundary="viewport"
      :boundary-padding="2"
      @show="onOpen"
    >
      <template #target>
        <button
          ref="filterBtnRef"
          :id="popoverTarget"
          :title="$t('recordList.filter.title')"
          type="button"
          :class="['btn', `btn-${variant}`, 'd-flex align-items-center d-print-none border-0 px-1 h-100', buttonClass]"
          :style="buttonStyle"
          @click.stop="popoverOpen = !popoverOpen"
        >
        <font-awesome-icon
          :icon="['fas', 'filter']"
          :class="[inFilter ? 'text-primary' : inactiveIconClass]"
        />
      </button>
    </template>

    <div class="card position-static w-100 border-0">
      <div class="card-body px-3 pb-0 overflow-y-auto overflow-x-hidden">
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
    </BPopover>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { BPopover } from 'bootstrap-vue-next'
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
    type: [String, Object],
    default: '',
  },
  allowFilterPresetSave: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['filter', 'reset', 'filter-preset'])

const componentFilter = ref([])
const popoverOpen = ref(false)
const filterBtnRef = ref(null)

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

function isFilterUiClick (target) {
  if (!(target instanceof Element)) return false
  if (filterBtnRef.value?.contains(target)) return true
  if (target.closest('.record-list-filter')) return true
  if (target.closest('.vs__dropdown-menu')) return true
  return false
}

function onDocumentPointerDown (e) {
  if (!popoverOpen.value) return
  if (isFilterUiClick(e.target)) return
  popoverOpen.value = false
}

onMounted(() => {
  document.addEventListener('pointerdown', onDocumentPointerDown, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocumentPointerDown, true)
})

function onOpen () {
  componentFilter.value = [...props.recordListFilter]
}

function preventClose () {
  // Keep the popover open while value editors (vue-select, etc.) teleport to body.
  popoverOpen.value = true
}

function savableFilter () {
  const groups = Array.isArray(componentFilter.value) ? componentFilter.value : []
  return groups
    .map(group => ({
      ...group,
      filter: (group.filter || []).filter(f => f && f.name),
    }))
    .filter(({ filter }) => filter.length > 0)
}

function resetFilter () {
  componentFilter.value = undefined
  emit('reset')
}

function onSave (close = true, type = 'filter') {
  const next = savableFilter()
  if (close) {
    popoverOpen.value = false
  }
  emit(type, next)
}
</script>

<style lang="scss">
.record-list-filter {
  z-index: 1070;
  max-width: 800px !important;
  opacity: 1 !important;
  border-color: transparent;

  .popover-body {
    display: flex;
    flex-direction: column;
    width: 100%;
    min-width: min(99vw, 350px);
    max-width: 100%;
    max-height: 25rem;
    padding: 0;
    overflow-x: hidden;
    box-sizing: border-box;
    color: var(--black);
    background: var(--white);
    border: 1px solid var(--bs-border-color, #dee2e6);
    border-radius: 0.25rem;
    opacity: 1 !important;
    box-shadow: 0 3px 48px #00000026;
    font-size: 0.9rem;
  }

  .card {
    min-width: 0;
    max-width: 100%;
  }

  .v-select,
  .field-operator,
  .field-editor {
    min-width: 0;
    max-width: 100%;
  }

  .popover-arrow {
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
