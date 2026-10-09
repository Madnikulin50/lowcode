<script setup lang="ts">
const props = defineProps<{
  editor: Record<string, any>
  format: Record<string, any>
  isActive?: Record<string, any>
  getMarkAttrs?: Function
  currentValue?: string
}>()

const emit = defineEmits<{
  (e: 'click', value: { type: string; attrs?: Record<string, any> }): void
}>()

function onClick(type: string, attrs?: Record<string, any>) {
  emit('click', { type, attrs })
}

function activeClasses(attrs?: Record<string, any>) {
  const isActive = props.editor.isActive(props.format.type, attrs)
  if (isActive) {
    return ['text-primary']
  }
  return undefined
}
</script>
