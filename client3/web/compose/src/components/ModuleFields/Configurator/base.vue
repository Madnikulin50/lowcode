<script>
import { computed } from 'vue'
import { compose, NoID } from 'corteza-lib/js/dist'

export function useConfiguratorBase(props, emit) {
  const f = computed({
    get: () => props.field,
    set: (val) => emit('update.field', val),
  })
  const isNew = computed(() => props.module.moduleID === NoID || props.field.fieldID === NoID)
  const hasData = computed(() => !isNew.value && props.hasRecords)
  return { f, isNew, hasData }
}
</script>
