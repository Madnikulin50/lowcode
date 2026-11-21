<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  displayElement: Record<string, any>
  labels?: Record<string, any>
}>()

const dataframes = computed(() => props.displayElement.dataframes || [])
const options = computed(() => props.displayElement.options || undefined)
</script>
