<template>
  <div class="d-flex flex-column h-100">
    <Teleport to="#topbar-title-target">{{ t(`request:kind.${kind}`) }}</Teleport>

    <component :is="kind" @submit="$emit('submit', $event)" />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'request', keyPrefix: 'edit' } })
import { useI18n } from 'vue-i18n'
import Correct from './Correct.vue'
import Delete from './Delete.vue'
import Export from './Export.vue'

defineProps({
  kind: { type: String, required: true },
})

defineEmits(['submit'])

const { t } = useI18n()
</script>