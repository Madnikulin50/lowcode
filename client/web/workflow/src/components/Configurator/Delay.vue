<template>
  <div class="card flex-grow-1 border-bottom border-light rounded-0">
    <div class="card-header p-0 mb-3">
      <h5 class="mb-0">{{ t('configurator.configuration') }}</h5>
    </div>
    <div class="card-body p-0">
      <div class="mb-0">
        <label class="text-primary form-label">{{ t('configurator.delay.duration.label') }}</label>
        <div class="text-muted small mb-2">{{ t('configurator.delay.duration.description') }}</div>
        <expression-editor
          v-model="item.config.arguments[0].expr"
          font-size="18px"
          show-line-numbers
          :show-popout="false"
          @input="valueChanged"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { watch, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import ExpressionEditor from '../ExpressionEditor.vue'

const { t } = useI18n()

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value'])

watch(() => props.item.config.stepID, () => {
  let args = [{ target: 'offset', type: 'Duration', expr: '' }]
  if (props.item.config.arguments && props.item.config.arguments.length) {
    args = props.item.config.arguments.map(({ target, type, value, expr }) => ({
      target, type, expr: expr || (value ? `"${value}"` : ''),
    }))
  }
  props.item.config.arguments = args
}, { immediate: true })

function valueChanged(value) {
  emit('update-default-value', {
    value: `Delay workflow execution for ${value}`,
    force: !props.item.node.value,
  })
  window.dispatchEvent(new CustomEvent('change-detected'))
}
</script>