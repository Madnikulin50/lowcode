<template>
  <div>
    <CPromptToast :hide-toasts="hideToasts" />
    <CPromptModal />
  </div>
</template>

<script setup lang="ts">
import CPromptToast from './CPromptToast.vue'
import CPromptModal from './CPromptModal.vue'

defineProps<{
  hideToasts?: boolean
}>()
</script>
