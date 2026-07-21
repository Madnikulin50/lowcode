<template>
  <div class="p-3">
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('request.view.export.data-type.label') }}</label>
      <span data-test-id="span-data-type" class="ms-2">{{ dataType }}</span>
    </div>

    <div class="mb-3">
      <label class="text-primary form-label">{{ t('request.view.export.date-range.label') }}</label>
      <span data-test-id="span-date-range" class="ms-2">{{ t(`request:view.export.date-range.${payload.range}`) }}</span>
    </div>

    <div class="mb-3">
      <label class="text-primary form-label">{{ t('request.view.export.file-format.label') }}</label>
      <span data-test-id="span-file-format" class="ms-2">{{ t(`request:view.export.file-format.${payload.format}`) }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  request: { type: Object, required: true },
})

const { t } = useI18n()

const payload = computed(() => {
  const { payload = [] } = props.request || {}
  return payload[0]
})

const dataType = computed(() => {
  const { profile = false, application = false } = payload.value
  return [
    { label: t('request.view.export.data-type.profile-information'), include: profile },
    { label: t('request.view.export.data-type.application-data'), include: application },
  ].filter(({ include }) => include)
    .map(({ label }) => label)
    .join(', ')
})
</script>