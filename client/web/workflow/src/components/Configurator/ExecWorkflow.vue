<template>
  <div>
    <div class="mb-3">
      <label class="text-primary form-label">{{ $t('label.label') }}</label>
      <input class="form-control" v-model="label" @input="emit('update-value', $event.target.value)" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const emit = defineEmits(['update-value'])
const props = defineProps({ item: { type: Object, required: true } })

const label = computed(() => props.item.label || '')
</script>