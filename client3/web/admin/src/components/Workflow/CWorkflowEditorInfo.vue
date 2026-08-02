<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit', workflow)">
        <div class="row g-3 p-3">
          <div class="col-12 col-lg-6">
            <div v-if="workflow.meta" class="mb-3">
              <label class="form-label text-primary">{{ $t('name') }}</label>
              <input
                v-model="workflow.meta.name"
                required
                :class="['form-control', { 'is-invalid': nameState === false }]"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('handle') }}</label>
              <input
                v-model="workflow.handle"
                :class="['form-control', { 'is-invalid': handleState === false }]"
              >
              <div v-if="handleState === false" class="invalid-feedback">
                {{ $t('automation.workflows.editor.info.invalid-handle-characters') }}
              </div>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" :class="{ 'mb-0': !workflow.workflowID }">
              <label class="form-label text-primary">{{ $t('enabled') }}</label>
              <c-input-checkbox
                v-model="workflow.enabled"
                switch
                :labels="checkboxLabel"
              />
            </div>
          </div>
        </div>
  
        <c-system-fields
          :id="workflow.workflowID"
          :resource="workflow"
        />
  
        <input
          type="submit"
          class="d-none"
          :disabled="saveDisabled"
        >
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="workflow && workflow.workflowID && workflow.canDeleteWorkflow"
        :text="getDeleteStatus"
        :disabled="deleteDisabled"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <button
        v-if="workflow.workflowID"
        class="btn btn-outline-secondary"
        @click="openWorkflowBuilder()"
      >
        {{ $t('automation.workflows.editor.info.openBuilder') }}
      </button>

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', workflow)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { handle } from 'corteza-lib/vue/dist'
import { NoID } from 'corteza-lib/js/dist'

const { t } = useI18n()

const props = defineProps({
  workflow: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit', 'delete'])

const checkboxLabel = { on: t('label.general.yes'), off: t('label.general.no') }

const editable = computed(() => (!props.workflow.workflowID || props.workflow.workflowID === NoID) || props.workflow.canUpdateWorkflow)
const nameState = computed(() => props.workflow.meta.name ? null : false)
const handleState = computed(() => handle.handleState(props.workflow.handle))
const saveDisabled = computed(() => !editable.value || [nameState.value, handleState.value].includes(false))
const deleteDisabled = computed(() => !(props.workflow.deletedAt ? props.workflow.canUndeleteWorkflow : props.workflow.canDeleteWorkflow))
const getDeleteStatus = computed(() => props.workflow.deletedAt ? t('undelete') : t('delete'))

function openWorkflowBuilder() {
  window.open(`/workflow/${props.workflow.workflowID}/edit`, '_blank')
}
</script>
