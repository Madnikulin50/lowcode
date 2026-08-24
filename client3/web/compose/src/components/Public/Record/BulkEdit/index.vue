<template>
  <div>
    <button
      v-if="!openOnSelect"
      class="btn btn-outline-extra-light btn-sm inline-button d-flex align-items-center justify-content-center text-secondary border-0"
      style="width: 2rem; height: 2rem;"
      title="Bulk edit"
      @click="showModal = true"
    >
      <font-awesome-icon
        :icon="['fas', 'pen']"
      />
    </button>

    <div
      v-if="showModal"
      class="modal d-block"
      tabindex="-1"
      role="dialog"
      @click.self="onModalHide"
    >
      <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ modalTitle || $t('recordList.bulkRecord.title') }}</h5>
          </div>
          <div class="modal-body p-0">
            <div v-if="fields.length" class="card pt-0">
              <div class="card-body d-flex flex-column gap-2">
                <div
                  v-for="(field, index) in fields"
                  :key="field"
                  class="position-relative"
                >
                  <FieldEditor
                    :namespace="namespace"
                    :module="module"
                    :field="getField(field)"
                    :errors="fieldErrors(field)"
                    :record="record"
                  >
                    <template v-if="allowAddField" #tools>
                      <c-input-confirm
                        :tooltip="$t('recordList.bulkRecord.field.remove')"
                        show-icon
                        class="ms-1"
                        @confirmed="fields.splice(index, 1)"
                      />
                    </template>
                  </FieldEditor>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer d-flex flex-column align-items-stretch">
            <template v-if="allowAddField">
              <c-input-select
                v-model="selectedField"
                :placeholder="getFieldSelectorPlaceholder"
                :get-option-label="getFieldLabel"
                :get-option-key="getOptionKey"
                :options="moduleFields"
                :selectable="option => !fields.includes(option.name)"
                :reduce="f => f.name"
                @input="addField"
              />
              <hr class="my-3">
            </template>
            <div class="d-flex justify-content-between align-items-center">
              <button
                class="btn btn-outline-secondary"
                :disabled="processing"
                @click="onReset"
              >
                {{ $t('label.reset') }}
              </button>
              <div class="d-flex gap-1">
                <button
                  class="btn btn-outline-secondary rounded"
                  @click="showModal = false"
                >
                  {{ $t('label.cancel') }}
                </button>
                <button
                  class="btn btn-primary"
                  :disabled="!fields.length || processing"
                  @click="handleBulkUpdateSelectedRecords(query)"
                >
                  {{ $t('label.save') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, validator } from 'corteza-lib/js/dist'
import { composables } from 'corteza-lib/vue/dist'
import FieldEditor from 'corteza-webapp-compose/src/components/ModuleFields/Editor'
import { isUserWritableField } from 'corteza-webapp-compose/src/lib/field-editable'

const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = inject('$ComposeAPI')
const { toastSuccess, toastErrorHandler } = composables.useToast()

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  module: { type: compose.Module, required: true },
  selectedFields: { type: Array, default: () => ([]) },
  initialRecord: { type: Object, default: () => ({}) },
  openOnSelect: { type: Boolean, default: false },
  modalTitle: { type: String, default: '' },
  query: { type: String, default: '' },
  allowAddField: { type: Boolean, default: false },
})

const emit = defineEmits(['close', 'save'])

const showModal = ref(false)
const selectedField = ref(undefined)
const fields = ref([])
const processing = ref(false)
const errors = ref(new validator.Validated())

const record = ref(new compose.Record(props.module, {}))

function fieldErrors (name) {
  if (!errors.value || typeof errors.value.filterByMeta !== 'function') {
    return new validator.Validated()
  }
  return errors.value.filterByMeta('field', name)
}

function isFieldEditable(field) {
  if (!isUserWritableField(field)) return false
  const { canCreateOwnedRecord } = props.module || {}
  const { createdAt, canManageOwnerOnRecord } = record.value || {}
  const { name, isSystem } = field || {}
  if (isSystem) {
    if (name === 'ownedBy') {
      return createdAt ? canManageOwnerOnRecord : canCreateOwnedRecord
    }
    return false
  }
  return true
}

const moduleFields = computed(() => {
  return [
    ...[...props.module.fields].sort((a, b) =>
      (a.label || a.name).localeCompare(b.label || b.name),
    ),
    ...props.module.systemFields().filter(({ name }) => name === 'ownedBy'),
  ].filter((field) => isFieldEditable(field))
})

const getFieldSelectorPlaceholder = computed(() => {
  return $t(`recordList.bulkRecord.field.add${fields.value.length ? 'Another' : ''}`)
})

watch(() => props.query, (query) => {
  if (!props.openOnSelect || !query.length) return
  record.value = new compose.Record(props.module, props.initialRecord)
  showModal.value = true
})

watch(() => props.selectedFields, (newFields = []) => {
  if (!newFields.length) return
  newFields.forEach(f => {
    if (fields.value.includes(f)) return
    fields.value.push(f)
  })
}, { immediate: true })

onBeforeUnmount(() => {
  setDefaultValues()
})

function onModalHide() {
  showModal.value = false
  if (props.openOnSelect) {
    fields.value = []
    record.value = new compose.Record(props.module, {})
  }
  emit('close')
}

function getFieldLabel({ kind, label, name }) {
  return label || name || kind
}

function addField(field) {
  if (!field) return
  fields.value.push(field)
  selectedField.value = null
}

function onReset() {
  record.value = new compose.Record(props.module, props.initialRecord)
  fields.value = [...props.selectedFields]
}

function getField(fieldName) {
  const field = moduleFields.value.find(({ name }) => name === fieldName)
  return field || {}
}

function getOptionKey({ fieldID }) {
  return fieldID
}

function setDefaultValues() {
  showModal.value = false
  selectedField.value = undefined
  fields.value = []
  errors.value = new validator.Validated()
}

// These come from the record mixin - stub them out
async function handleBulkUpdateSelectedRecords(query) {
  if (!props.module || !props.module.moduleID) return
  processing.value = true

  const values = []
  fields.value.forEach(f => {
    const { name, isMulti, isSystem } = getField(f)
    const value = isSystem ? record.value[name] : record.value.values[name]

    if (!isMulti) {
      values.push({ name, value: value?.toString() ?? '' })
    } else {
      if (!Array.isArray(value) || value.length === 0) {
        values.push({ name })
        return
      }

      const multiValues = value
        .filter(v => v !== undefined)
        .map(v => ({ name, value: v?.toString() ?? '' }))

      values.push(...multiValues)
    }
  })

  const { namespaceID, moduleID } = props.module
  try {
    await $ComposeAPI.recordPatch({ moduleID, namespaceID, values, query })
    toastSuccess($t('notification.record.bulkRecordUpdateSuccess'))
    showModal.value = false
    emit('save')
    window.dispatchEvent(new CustomEvent('refetch-records', { detail: { stayOnPage: true } }))
  } catch (e) {
    toastErrorHandler($t('notification.record.deleteBulkRecordUpdateFailed'))(e)
  } finally {
    processing.value = false
  }
}
</script>

<style lang="scss">
.position-initial {
  position: initial;
}
</style>

<style lang="scss" scoped>
.inline-button {
  &:hover {
    color: var(--primary) !important;
  }
}
</style>
