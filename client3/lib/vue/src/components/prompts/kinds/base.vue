<script setup lang="ts">
import { computed } from 'vue'
import { pVal as _pVal, pType as _pType } from '../utils.ts'

const props = defineProps<{
  loading?: boolean
  payload?: Record<string, any>
}>()

const message = computed(() => _pVal(props.payload, 'message', ''))
const label = computed(() => _pVal(props.payload, 'label', ''))

function pVal(k: string, def?: any) {
  return _pVal(props.payload, k, def)
}

function pType(k: string, def?: any) {
  return _pType(props.payload, k, def)
}

function pRaw(k?: string, defValue?: any, defType?: string) {
  if (k && props.payload && props.payload[k] && props.payload[k] !== undefined) {
    return props.payload[k]
  }

  return { '@type': defType, '@value': defValue }
}
</script>
