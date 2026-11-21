<template>
  <p>{{ $t('system.templates.editor.content.editor.unsupported') }}</p>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.templates', keyPrefix: 'editor.content.editor' } })
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>