<template>
  <c-toolbar class="bg-white shadow border-top">
    <template #start>
      <button
        v-if="backLink"
        data-test-id="button-back"
        class="btn btn-link d-flex align-items-center text-dark back text-nowrap gap-1 text-decoration-none"
        @click="$router.push(backLink)"
      >
        <font-awesome-icon :icon="['fas', 'chevron-left']" class="back-icon" />
        {{ t('label.back') }}
      </button>
    </template>
    <template #center><slot /></template>
    <template #end>
      <c-input-confirm
        v-if="!hideDelete"
        size="lg"
        size-confirm="lg"
        variant="danger"
        :disabled="deleteDisabled || processingDelete || processing"
        :processing="processingDelete"
        :text="t('label.delete')"
        :borderless="false"
        @confirmed="$emit('delete')"
      />
      <c-button-submit
        data-test-id="button-clone"
        :disabled="cloneDisabled || processingClone || processing"
        :processing="processingClone"
        variant="outline-secondary"
        :text="t('label.clone')"
        class="text-nowrap"
        size="lg"
        @submit="$emit('clone')"
      />
      <c-button-submit
        data-test-id="button-save"
        :disabled="saveDisabled || processingSave || processing"
        :processing="processingSave"
        :text="t('label.save')"
        size="lg"
        @submit="$emit('save')"
      />
    </template>
  </c-toolbar>
</template>
<script setup>
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { components } from 'corteza-lib/vue/dist'
const { CToolbar } = components

defineProps({
  backLink: { type: Object, required: false, default: () => ({ name: 'root' }) },
  hideDelete: Boolean,
  deleteDisabled: Boolean,
  hideSave: Boolean,
  saveDisabled: Boolean,
  cloneDisabled: Boolean,
  processing: Boolean,
  processingDelete: Boolean,
  processingSave: Boolean,
  processingClone: Boolean,
})
defineEmits(['delete', 'clone', 'save'])

const { t } = useI18n()
const router = useRouter()
</script>