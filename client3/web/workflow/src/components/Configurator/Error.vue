<template>
  <div class="card flex-grow-1 border-bottom border-light rounded-0">
    <div class="card-header p-0 mb-3">
      <h5 class="mb-0">{{ t('configurator.configuration') }}</h5>
    </div>
    <div class="card-body p-0">
      <div class="mb-0">
        <label class="text-primary form-label">{{ t('error-expression') }}</label>
        <expression-editor
          v-model="item.config.arguments[0].expr"
          font-size="18px"
          show-line-numbers
          @open="openInEditor"
          @input="valueChanged"
        />
      </div>
    </div>
  </div>

  <div class="modal fade" id="expression-editor-error" tabindex="-1">
    <div class="modal-dialog modal-xl">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ t('editor.editor') }}</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
        </div>
        <div class="modal-body p-0">
          <expression-editor
            v-model="expressionEditor.currentExpression"
            min-height="80vh"
            font-size="18px"
            show-line-numbers
            :border="false"
            :show-popout="false"
          />
        </div>
        <div class="modal-footer">
          <button class="btn btn-outline-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
          <button class="btn btn-primary" @click="saveExpression">{{ t('save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import ExpressionEditor from '../ExpressionEditor.vue'

const { t } = useI18n()
const $AutomationAPI = inject('automationAPI', {})

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value'])

const expressionEditor = reactive({ currentExpression: undefined })

{
  let args = [{ target: 'message', type: 'String', expr: '' }]
  if (props.item.config.arguments && props.item.config.arguments.length) {
    args = props.item.config.arguments.map(({ target, type, value, expr }) => ({
      target, type, expr: expr || (value ? `"${value}"` : ''),
    }))
  }
  props.item.config.arguments = args
}

function valueChanged(value) {
  emit('update-default-value', {
    value: `Stop workflow with error: ${value}`,
    force: !props.item.node.value,
  })
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function openInEditor() {
  expressionEditor.currentExpression = props.item.config.arguments[0].expr
  const modal = new bootstrap.Modal(document.getElementById('expression-editor-error'))
  modal.show()
}

function saveExpression() {
  props.item.config.arguments[0].expr = expressionEditor.currentExpression
  window.dispatchEvent(new CustomEvent('change-detected'))
  expressionEditor.currentExpression = undefined
  const modal = bootstrap.Modal.getInstance(document.getElementById('expression-editor-error'))
  if (modal) modal.hide()
}
</script>