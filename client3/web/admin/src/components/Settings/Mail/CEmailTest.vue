<template>
  <div class="card shadow-sm">
    <div class="card-body">
      <div v-if="errors.length" class="alert alert-danger">
        <div v-for="(err, i) in errors" :key="i">{{ err }}</div>
      </div>

      <c-button-submit
        :text="$t('editor.server.testSmtpConfigs.button')"
        variant="outline-secondary"
        :processing="processing"
        @submit="$emit('test')"
      />
    </div>
  </div>
</template>

<script setup>
defineProps({
  modelValue: {
    type: Object,
    default: () => ({}),
  },
  settings: {
    type: Array,
    default: () => [],
  },
  processing: {
    type: Boolean,
    default: false,
  },
  errors: {
    type: Array,
    default: () => [],
  },
})

defineEmits(['test'])
</script>
