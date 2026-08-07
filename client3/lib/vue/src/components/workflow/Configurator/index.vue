<template>
  <div class="d-flex flex-column">
    <div
      v-if="kind !== 'Content'"
      class="card flex-grow-1 border-bottom border-light rounded-0"
    >
      <div class="card-header p-0 mb-3">
        <h5 class="mb-0">
          {{ t('general') }}
        </h5>
      </div>
      <div class="card-body p-0">
        <basic
          :item="item"
          @update-value="emit('update-value', $event)"
        />
      </div>
    </div>

    <component
      :is="stepComponent"
      v-if="stepComponent"
      :item="item"
      @update:item="emit('update:item', $event)"
      :edges="edges"
      @update:edges="emit('update:edges', $event)"
      :out-edges="outEdges"
      :is-subworkflow="isSubworkflow"
      @update-value="emit('update-value', $event)"
      @update-default-value="updateDefaultName"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import basic from './basic.vue'
import * as Configurators from './loader.js'

const { t } = useI18n()

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value', 'update:item', 'update:edges'])

const stepComponent = computed(() => Configurators[kind.value])

const kind = computed(() => {
  const { kind, ref } = props.item.config

  if (kind === 'exec-workflow') {
    return 'ExecWorkflow'
  }

  if (kind === 'error-handler') {
    return 'ErrorHandler'
  }

  if (kind === 'visual' && ref === 'content') {
    return 'Content'
  }

  if (kind) {
    return kind.charAt(0).toUpperCase() + kind.slice(1)
  }

  return undefined
})

function updateDefaultName ({ value, force = false }) {
  if (force || props.item.config.defaultName || props.item.config.defaultName === undefined) {
    emit('update-default-value', value)
  }
}
</script>
