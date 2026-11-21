<script setup>
/**
 * Provides base for each feed configurator.
 * Define common props, methods, ... in here.
 */
const props = defineProps({
  feed: { type: Object, required: true, default: () => ({}) },
  modules: { type: Array, required: false, default: () => [] },
})
</script>
