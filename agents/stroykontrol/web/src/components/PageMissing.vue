<script setup>
// What is shown instead of a page picture that is missing or failed to load.
defineProps({ label: String, reason: String, retry: Boolean, title: String })
const emit = defineEmits(['retry'])
</script>

<template>
  <div class="page-missing" role="alert">
    <div class="pm-title"><span v-if="label">{{ label }}: </span>{{ title || 'страница недоступна' }}</div>
    <div class="pm-reason">{{ reason }}</div>
    <button v-if="retry" class="small" @click="emit('retry')">Повторить</button>
  </div>
</template>

<style scoped>
.page-missing { border: 1px dashed var(--border); border-radius: 8px; background: var(--panel); padding: 16px 18px; margin: 8px 0; text-align: left; max-width: 520px; }
.pm-title { font-weight: 600; margin-bottom: 4px; }
.pm-reason { color: var(--muted); font-size: 12.5px; margin-bottom: 8px; }
button.small { border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: 6px; padding: 2px 10px; font-size: 12px; cursor: pointer; }
</style>
