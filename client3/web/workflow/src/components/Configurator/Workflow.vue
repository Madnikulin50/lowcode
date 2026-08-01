<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="modal fade show d-block"
      tabindex="-1"
      style="background: rgba(0,0,0,0.5);"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('editor.workflow-configuration') }}</h5>
            <button
              v-if="workflow.workflowID !== '0'"
              type="button"
              class="btn-close"
              @click="handleClose"
            />
          </div>

          <div class="modal-body p-0">
            <div
              v-if="workflow.workflowID && workflow.workflowID !== '0'"
              class="d-flex p-3"
            >
              <import
                data-test-id="button-import-workflow"
                :disabled="importProcessing"
                @import="emit('import', $event)"
              />

              <export
                data-test-id="button-export-workflow"
                :workflows="[workflow.workflowID]"
                :file-name="workflow.meta.name || workflow.handle"
                size="lg"
                class="ms-1"
              />

              <c-permissions-button
                v-if="workflow.canGrant"
                :title="workflow.meta.name || workflow.handle || workflow.workflowID"
                :target="workflow.meta.name || workflow.handle || workflow.workflowID"
                :resource="`corteza::automation:workflow/${workflow.workflowID}`"
                :button-label="t('permissions')"
                class="btn-lg ms-1"
              />
            </div>

            <div v-if="localWorkflow">
              <ul class="nav nav-tabs" role="tablist">
                <li class="nav-item" role="presentation">
                  <button
                    class="nav-link active"
                    type="button"
                    role="tab"
                    @click="activeTab = 'general'"
                  >
                    {{ t('label') }}
                  </button>
                </li>
                <li class="nav-item" role="presentation">
                  <button
                    class="nav-link"
                    type="button"
                    role="tab"
                    @click="activeTab = 'labels'"
                  >
                    {{ t('labels.label') }}
                  </button>
                </li>
              </ul>

              <div class="tab-content  card p-3">
                <div
                  v-show="activeTab === 'general'"
                  class="tab-pane active"
                  role="tabpanel"
                >
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ t('name.label') }}</label>
                    <input
                      v-model="localWorkflow.meta.name"
                      class="form-control"
                      data-test-id="input-label"
                      :placeholder="t('name.placeholder')"
                      :class="{ 'is-invalid': nameState === false }"
                    >
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ t('handle.label') }}</label>
                    <input
                      v-model="localWorkflow.handle"
                      class="form-control"
                      data-test-id="input-handle"
                      :class="{ 'is-invalid': handleState === false }"
                      :placeholder="t('handle.placeholder')"
                    >
                    <div
                      v-if="handleState === false"
                      class="invalid-feedback"
                      data-test-id="input-handle-invalid-state"
                    >
                      {{ t('handle.invalid-handle-characters') }}
                    </div>
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ t('description.label') }}</label>
                    <textarea
                      v-model="localWorkflow.meta.description"
                      class="form-control"
                      data-test-id="input-description"
                      :placeholder="t('description.placeholder')"
                      rows="3"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ t('run-as.label') }}</label>
                    <div class="form-text mb-2">{{ t('run-as.description') }}</div>
                    <c-input-user
                      v-model="localWorkflow.runAs"
                      data-test-id="select-run-as"
                      :placeholder="t('run-as.placeholder')"
                      clearable
                    />
                  </div>

                  <div class="mb-3">
                    <div class="form-check">
                      <input
                        id="wf-enabled"
                        v-model="localWorkflow.enabled"
                        class="form-check-input-v3"
                        data-test-id="checkbox-enable-workflow"
                        type="checkbox"
                      >
                      <label
                        class="form-check-label"
                        for="wf-enabled"
                      >{{ t('enabled') }}</label>
                    </div>
                  </div>

                  <div class="mb-3">
                    <div class="form-text mb-2">{{ t('sub-workflow.description') }}</div>
                    <div class="form-check">
                      <input
                        id="wf-sub"
                        v-model="localWorkflow.meta.subWorkflow"
                        class="form-check-input-v3"
                        data-test-id="checkbox-sub-workflow"
                        type="checkbox"
                      >
                      <label
                        class="form-check-label"
                        for="wf-sub"
                      >{{ t('sub-workflow.label') }}</label>
                    </div>
                  </div>
                </div>

                <div
                  v-show="activeTab === 'labels'"
                  class="tab-pane"
                  role="tabpanel"
                >
                  <namespace-module-selector
                    :namespace-labels="localWorkflow?.labels?.ref_namespace || []"
                    :module-labels="localWorkflow?.labels?.ref_module || []"
                    @change="handleLabelsChange"
                  />
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <div class="d-flex w-100">
              <c-input-confirm
                v-if="workflow.canDeleteWorkflow && !isDeleted"
                size="md"
                size-confirm="md"
                variant="danger"
                :processing="processingDelete"
                :text="t('editor.delete')"
                :borderless="false"
                @confirmed="emit('delete')"
              />

              <c-input-confirm
                v-else-if="isDeleted"
                size="md"
                size-confirm="md"
                variant="outline-secondary"
                :processing="processingDelete"
                :text="t('editor.undelete')"
                :borderless="false"
                @confirmed="emit('undelete')"
              />

              <button
                v-if="workflow.workflowID === '0'"
                class="btn btn-outline-secondary"
                @click="router.back()"
              >
                {{ t('editor.back') }}
              </button>

              <div class="d-flex ms-auto">
                <button
                  v-if="workflow.workflowID !== '0'"
                  class="btn btn-outline-secondary ms-auto me-1"
                  @click="handleCancel"
                >
                  {{ t('cancel') }}
                </button>

                <c-button-submit
                  data-test-id="button-save-workflow"
                  :disabled="isSaveDisabled"
                  :processing="processingSave"
                  :text="t('editor.save')"
                  class="ms-1"
                  @submit="handleSave"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { handle, components } from 'corteza-lib/vue/dist'
import { automation } from 'corteza-lib/js/dist'
import Import from '../Import'
import Export from '../Export'
import NamespaceModuleSelector from '../NamespaceModuleSelector'

const { CInputUser } = components

const { t } = useI18n()
const router = useRouter()

const props = defineProps({
  workflow: {
    type: Object,
    default: () => {},
  },
  canCreate: {
    type: Boolean,
    default: false,
  },
  processingSave: {
    type: Boolean,
    default: false,
  },
  processingDelete: {
    type: Boolean,
    default: false,
  },
  importProcessing: {
    type: Boolean,
    default: false,
  },
  show: { type: Boolean, default: false },
})

const emit = defineEmits(['save', 'delete', 'undelete', 'import', 'update:show'])

const visible = computed(() => props.show)
const activeTab = ref('general')
const localWorkflow = ref(null)

const nameState = computed(() => {
  return localWorkflow.value?.meta?.name ? null : false
})

const handleState = computed(() => {
  return localWorkflow.value ? handle.handleState(localWorkflow.value.handle) : null
})

const canUpdateWorkflow = computed(() => {
  return props.workflow.workflowID === '0' ? props.canCreate : props.workflow.canUpdateWorkflow
})

const isSaveDisabled = computed(() => {
  return !canUpdateWorkflow.value || [nameState.value, handleState.value].includes(false)
})

const isDeleted = computed(() => {
  return props.workflow.deletedAt
})

watch(() => props.workflow, (newWorkflow) => {
  if (!newWorkflow) {
    localWorkflow.value = null
    return
  }
  localWorkflow.value = new automation.Workflow(newWorkflow)
}, { immediate: true })

function handleLabelsChange ({ namespaceLabels, moduleLabels }) {
  if (!localWorkflow.value.labels) {
    localWorkflow.value.labels = {}
  }

  if (namespaceLabels.length > 0) {
    localWorkflow.value.labels.ref_namespace = namespaceLabels
  } else {
    delete localWorkflow.value.labels.ref_namespace
  }

  if (moduleLabels.length > 0) {
    localWorkflow.value.labels.ref_module = moduleLabels
  } else {
    delete localWorkflow.value.labels.ref_module
  }
}

function handleSave () {
  emit('save', localWorkflow.value)
  emit('update:show', false)
}

function handleCancel () {
  localWorkflow.value = new automation.Workflow(props.workflow)
  emit('update:show', false)
}

function handleClose () {
  emit('update:show', false)
}
</script>
