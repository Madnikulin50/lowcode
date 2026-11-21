<template>
  <div class="p-4">
    <h4 class="mb-4">{{ $t('app.configuration') }}</h4>
    <div class="alert alert-info">
      {{ $t('configuration.comingSoon') }}
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })
</script>
