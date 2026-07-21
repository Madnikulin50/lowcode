<template>
  <div>
    <div
      v-for="(error, i) in set"
      :key="i"
      class="invalid-feedback d-block mt-1"
    >
      <span
        :class="{ 'text-secondary': error.meta.isWarning }"
      >
        {{ t(error.message, { interpolation: { escapeValue: false }, value: error.meta.value}) }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { validator } from 'corteza-lib/js/dist'

const props = defineProps({
  errors: {
    type: validator.Validated,
    required: true,
    default: undefined,
  },
  index: {
    type: Number,
    required: false,
    default: -1,
  },
})

const { t } = useI18n({ useScope: 'global', messages: {} })

const set = computed(() => {
  return (props.index >= 0 ? props.errors.filterByMeta('index', props.index).get() : props.errors.get()).slice(0, 1)
})
</script>
