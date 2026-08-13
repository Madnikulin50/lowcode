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
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import FieldEditor from 'corteza-webapp-compose/src/components/ModuleFields/Editor'

const { t: $t } = useI18n({ useScope: 'global' })

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

const emit = defineEmits(['close'])

const showModal = ref(false)
const selectedField = ref(undefined)
const fields = ref([])

// Inline from record mixin
const record = ref(new compose.Record(props.module, {}))

function fieldErrors(field) {
  return {}
}

function isFieldEditable(field) {
  if (!field) return false
  const { canCreateOwnedRecord } = props.module || {}
  const { createdAt, canManageOwnerOnRecord } = record.value || {}
  const { name, canUpdateRecordValue, isSystem, expressions = {} } = field || {}
  if (!canUpdateRecordValue) return false
  if (isSystem) {
    if (name === 'ownedBy') {
      return createdAt ? canManageOwnerOnRecord : canCreateOwnedRecord
    }
    return false
  }
  return !expressions.value
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
}

// These come from the record mixin - stub them out
async function handleBulkUpdateSelectedRecords(query) {
  const $ComposeAPI = window.__composeAPI
  const { namespaceID, moduleID } = props.module
  // This would normally call recordPatch API - simplified stub
  const values = {}
  fields.value.forEach(f => {
    if (record.value.values[f]) {
      values[f] = record.value.values[f]
    }
  })
  if (Object.keys(values).length && query) {
    try {
      await $ComposeAPI.recordPatch({ namespaceID, moduleID, values, query })
      showModal.value = false
    } catch (e) {
      // error handling
    }
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
