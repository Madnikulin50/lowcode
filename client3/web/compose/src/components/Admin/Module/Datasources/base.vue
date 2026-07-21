<script>
import { computed } from 'vue'

export function useDataSourceBase (props) {
  const kind = computed(() => Object.keys(props.step))

  const nameState = computed(() => {
    const name = props.step[kind.value].name
    const isDuplicate = props.datasources.some(({ step }, index) => {
      return index !== props.index && step[Object.keys(step)].name === name && name
    })
    return name.length > 0 && !isDuplicate ? null : false
  })

  return { kind, nameState }
}
</script>
