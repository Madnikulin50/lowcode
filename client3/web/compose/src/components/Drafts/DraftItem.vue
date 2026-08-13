<template>
  <div class="draft-item-container px-3 pt-3 pb-2 border-bottom">
    <div
      :id="`draft-item-${draft.revision.changeID}`"
      :class="{ 'border-primary': active }"
      class="list-group-item draft-item border rounded bg-white p-3 position-relative"
      @click="$emit('click', draft)"
    >
      <div
        class="action-menu bg-white pb-2 ps-2"
        style="margin-left: -1rem;"
      >
        <div class="dropdown">
          <button
            class="btn btn-outline-extra-light text-decoration-none border-0 dropdown-toggle-no-caret"
            data-bs-toggle="dropdown"
          >
            <font-awesome-icon
              :icon="['fas', 'ellipsis-v']"
              class="text-secondary"
              style="margin-top: 0.3rem;"
            />
          </button>
          <ul class="dropdown-menu dropdown-menu-end">
            <li v-if="recordID">
              <button class="dropdown-item" @click.stop="$emit('view', draft)">
                <font-awesome-icon :icon="['far', 'file-alt']" class="text-primary" />
                {{ $t('viewRecord') }}
              </button>
            </li>
            <li>
              <c-input-confirm
                :text="$t('label.delete')"
                show-icon
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="$emit('delete', draft)"
              />
            </li>
          </ul>
        </div>
      </div>

      <div class="draft-item-content">
        <h5
          v-if="recordLabel && !useFieldViewer"
          class="fw-bold text-break"
        >
          {{ recordLabel }}
        </h5>

        <div v-if="useFieldViewer" class="changed-field mb-2">
          <div class="text-primary fw-bold">
            {{ firstChangedField.label || firstChangedField.name }}
          </div>
          <FieldViewer
            :namespace="namespace"
            :field="firstChangedField"
            :record="record"
            value-only
            class="text-break"
          />
        </div>

        <button
          :id="`draft-description-${draft.revision.changeID}`"
          class="btn btn-link small text-secondary mb-1 text-break cursor-pointer p-0"
          @click.prevent.stop
        >
          {{ description }}
        </button>
        <div class="d-flex align-items-center justify-content-end flex-wrap gap-1">
          <span
            class="badge bg-primary"
            style="font-size: 85%;"
            title="Namespace"
          >
            {{ namespaceLabel }}
          </span>
          <span
            class="badge bg-extra-light"
            style="font-size: 85%;"
            title="Module"
          >
            {{ moduleLabel }}
          </span>
        </div>
      </div>
    </div>

    <div class="d-flex justify-content-end mt-2">
      <div
        :title="draft.revision.timestamp"
        class="text-muted small"
      >
        {{ $locFullDateTime(draft.revision.timestamp) }}
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: ['drafts', 'general'] } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'

const { CInputConfirm } = components
const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  draft: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  namespace: { type: Object, required: false, default: undefined },
  active: { type: Boolean, default: false },
})

defineEmits(['click', 'view', 'delete'])

const namespaceLabel = computed(() => {
  const { revision } = props.draft
  const parts = revision.resource.split('/')
  const namespaceID = parts[1]
  if (props.namespace) return props.namespace.name || props.namespace.slug || namespaceID
  return namespaceID
})

const moduleLabel = computed(() => {
  const { revision } = props.draft
  const parts = revision.resource.split('/')
  const moduleID = parts[2]
  if (props.module) return props.module.name || props.module.handle || moduleID
  return moduleID
})

const recordLabel = computed(() => {
  const { revision } = props.draft
  if (props.module && revision.record) {
    const rec = new compose.Record(props.module, revision.record)
    const firstField = (props.module.fields || [])[0]
    if (firstField) {
      const value = rec.values[firstField.name]
      if (value) return Array.isArray(value) ? value[0] : value
    }
  }
  const isNew = revision.resource.endsWith('/0') || revision.operation === 'created'
  if (isNew) return $t('label.newRecord') || 'New Record'
  const parts = revision.resource.split('/')
  return parts[3]
})

const description = computed(() => {
  const { revision } = props.draft
  const isNew = revision.resource.endsWith('/0') || revision.operation === 'created'
  if (isNew) return $t('label.newRecord') || 'New Record'
  const changesCount = revision.changes.length
  return $t('changes', { count: changesCount })
})

const record = computed(() => {
  const { revision } = props.draft
  if (props.module && revision.record) return new compose.Record(props.module, revision.record)
  return null
})

const recordID = computed(() => {
  const parts = props.draft.revision.resource.split('/')
  return parts[3] === '0' ? undefined : parts[3]
})

const firstChangedField = computed(() => {
  if (!props.module || !props.draft.revision.changes || props.draft.revision.changes.length === 0) return null
  const changedNames = props.draft.revision.changes.map(c => c.key)
  return (props.module.fields || []).find(f => changedNames.includes(f.name))
})

const useFieldViewer = computed(() => {
  return !!props.module && !!record.value && !!firstChangedField.value
})

const changedFields = computed(() => {
  if (!props.module || !props.draft.revision.changes) return []
  const changedNames = props.draft.revision.changes.map(c => c.key)
  return (props.module.fields || []).filter(f => changedNames.includes(f.name))
})
</script>

<style lang="scss" scoped>
.draft-item-container {
  &:hover {
    background-color: var(--light) !important;
  }

  .draft-item {
    transition: background-color 0.2s ease;
    cursor: pointer;
  }

  &:hover {
    .action-menu {
      opacity: 1 !important;
      pointer-events: auto;
    }
  }
  .action-menu {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    z-index: 2;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.2s ease;
  }
}
</style>
