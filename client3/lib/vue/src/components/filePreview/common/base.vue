<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { getComponent } from './index.js'

defineOptions({ inheritAttrs: false })

defineProps<{
  inline?: boolean
}>()

const attrs = useAttrs()

const previewType = computed(() => getComponent({
  type: (attrs as any).mime,
  src: (attrs as any).src,
  name: (attrs as any).name,
}))
</script>
