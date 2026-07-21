<template>
  <div>
    <hr>

    <h5>
      {{ t('title') }}
    </h5>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('permitted-roles.label') }}</label>
      <div class="form-text mb-2">{{ t('permitted-roles.description') }}</div>
      <c-role-picker
        v-model="value.permittedRoles"
        class="mb-3"
      />
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('prohibited-roles.label') }}</label>
      <div class="form-text mb-2">{{ t('prohibited-roles.description') }}</div>
      <c-role-picker
        v-model="value.prohibitedRoles"
        class="mb-3"
      />
    </div>

    <div class="mb-0">
      <label class="form-label text-primary">{{ t('forced-roles.label') }}</label>
      <div class="form-text mb-2">{{ t('forced-roles.description') }}</div>
      <c-role-picker
        v-model="value.forcedRoles"
        class="mb-3"
      />
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import CRolePicker from 'corteza-webapp-admin/src/components/CRolePicker'

const { t } = useI18n()

defineProps({
  value: {
    type: Object,
    required: true,
    default: () => ({}),
  },
})
</script>
