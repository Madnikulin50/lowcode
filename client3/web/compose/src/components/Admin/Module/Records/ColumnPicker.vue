<template>
  <div :class="hideTrigger ? '' : 'd-flex'" :style="hideTrigger ? { display: 'contents' } : undefined">
    <button
      v-if="!hideTrigger"
      class="btn"
      :class="[variant ? `btn-${variant}` : 'btn-light', buttonClass]"
      :title="buttonTooltip.title"
      :disabled="disabled"
      @click="showModal = true"
    >
      <slot />
    </button>

    <div
      id="columns-modal"
      ref="modal"
      class="modal fade"
      :class="{ show: showModal, 'd-block': showModal }"
      tabindex="-1"
    >
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title d-flex align-items-center p-0">
              {{ $t('allRecords.columns.title') }}
              <c-hint
                :tooltip="$t('allRecords.tooltip.configureColumns')"
                icon-class="text-warning"
              />
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              @click="showModal = false"
            />
          </div>
          <div class="modal-body">
            <field-picker
              :module="module"
              v-model:fields="filteredFields"
              :field-subset="fieldSubset"
              style="height: 71vh;"
            />
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-outline-secondary"
              data-bs-dismiss="modal"
              @click="showModal = false"
            >
              {{ $t('label.cancel') }}
            </button>
            <button
              class="btn btn-primary"
              @click="onSave"
            >
              {{ $t('label.saveAndClose') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="showModal"
      class="modal-backdrop fade show"
      @click="showModal = false"
    />
  </div>
</template>

<script setup lang="js">
import { ref, watch, onBeforeUnmount } from 'vue'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'

defineOptions({
  i18nOptions: {
    namespaces: 'module',
  },
})

const props = defineProps({
  module: {
    type: Object,
    required: true,
    default: () => ({}),
  },
  fields: {
    type: Array,
    required: true,
    default: () => [],
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  size: {
    type: String,
    default: 'lg',
  },
  variant: {
    type: String,
    default: 'light',
  },
  buttonClass: {
    type: String,
    default: 'flex-fill',
  },
  fieldSubset: {
    type: Array,
    required: false,
    default: () => null,
  },
  buttonTooltip: {
    type: Object,
    default: () => ({}),
  },
  hideTrigger: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['updateFields'])

const showModal = ref(false)
const filteredFields = ref([])

watch(() => props.fields, (fields) => {
  if (fields) {
    filteredFields.value = props.module.filterFields(fields)
  }
}, { immediate: true })

onBeforeUnmount(() => {
  setDefaultValues()
})

function onSave () {
  emit('updateFields', filteredFields.value)
}

function open () {
  showModal.value = true
}

defineExpose({ open })

function setDefaultValues () {
  filteredFields.value = []
}
</script>

<style lang="scss" scoped>
.fit-modal {
  max-height: calc(100% - 3.5rem);
}
</style>
