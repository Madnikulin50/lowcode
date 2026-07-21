<script setup>
import { computed } from 'vue'

const props = defineProps({
  item: { type: Object, required: true },
})

const options = computed({
  get: () => props.item.options,
  set: (opts) => { props.item.options = opts },
})
</script>
