<template>
  <div v-if="resource && connection">
    <div class="mb-3">
      <label class="form-label text-primary">{{ translations.sensitivity.label }}</label>
      <div class="form-text mb-2">{{ translations.sensitivity.description }}</div>
      <c-sensitivity-level-picker
        v-model="resource.config.privacy.sensitivityLevelID"
        :options="sensitivityLevels"
        :placeholder="translations.sensitivity.placeholder"
        :max-level="maxLevel"
        :disabled="processing"
      />
    </div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ translations.usage.label }}</label>
      <textarea
        v-model="resource.config.privacy.usageDisclosure"
        class="form-control"
      />
    </div>
  </div>
</template>

<script setup lang="js">
import { ref } from 'vue'
import { components } from 'corteza-lib/vue/dist'
const { CSensitivityLevelPicker } = components

defineProps({
  resource: {
    type: Object,
    required: true,
  },
  connection: {
    type: Object,
    required: true,
  },
  maxLevel: {
    type: String,
    default: undefined,
  },
  sensitivityLevels: {
    type: Array,
    required: true,
  },
  translations: {
    type: Object,
    required: true,
  },
})

const processing = ref(false)
</script>
