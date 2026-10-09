<template>
  <div>
    <div class="mb-3">
      <label class="text-primary form-label">{{ $t('label.label') }}</label>
      <input class="form-control" v-model="label" @input="emit('update-value', $event.target.value)" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value'])

const label = computed({
  get () {
    if (getSourceType.value) {
      if (getSourceType.value === 'gatewayExclusive') {
        const [edgeID, ...rest] = (props.item.node.value || '').split(' - ')
        return rest.join(' - ')
      }
    }
    return props.item.node.value
  },
  set (label) {
    if (getSourceType.value) {
      if (getSourceType.value === 'gatewayExclusive') {
        const [edgeID, ...rest] = (props.item.node.value || '').split(' - ')
        const newLabel = [edgeID]
        if (label) {
          newLabel.push(label)
        }
        label = newLabel.join(' - ')
      }
    }
    props.item.node.value = label
  },
})

const getSourceType = computed(() => {
  const { source } = props.item.node
  if (source && source.style) {
    return source.style
  }
  return undefined
})
</script>