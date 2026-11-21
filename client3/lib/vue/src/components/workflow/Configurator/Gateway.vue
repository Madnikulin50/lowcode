<template>
  <div>
    <div v-if="['incl', 'excl'].includes(gatewayKind)" class="card flex-grow-1 border-bottom border-light rounded-0">
      <div class="card-header p-0 mb-3">
        <h5 class="mb-0">{{ t('configurator.configuration') }}</h5>
      </div>
      <div class="card-body p-0">
        <var v-if="outEdges < 2">{{ t('steps.gateway.configurator.two-paths') }}</var>
        <div v-else>
          <div v-for="edge in gatewayEdges" :key="edge.id" class="mb-3">
            <label class="text-primary form-label">{{ edge.value }}</label>
            <expression-editor
              v-model="edge.expr"
              show-line-numbers
              :show-popout="false"
              @input="updateEdge(edge.id, $event)"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
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

const gatewayKind = computed(() => props.item.config.ref)

const gatewayEdges = computed(() => {
  const edges = []
  if (['incl', 'excl'].includes(gatewayKind.value)) {
    if (props.outEdges && props.item.node.edges) {
      props.item.node.edges.forEach(({ id, source, target, value = '' }) => {
        if (source.id === props.item.node.id) {
          edges.push({
            id,
            source: source.id,
            target: target.id,
            value,
            expr: props.edges[id]?.config?.expr || '',
          })
        }
      })
    }
  }
  return edges
})

function updateEdge(id, expr) {
  props.edges[id].config.expr = expr
  window.dispatchEvent(new CustomEvent('change-detected'))
}
</script>