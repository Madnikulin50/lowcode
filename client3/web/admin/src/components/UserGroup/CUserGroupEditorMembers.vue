<template>
  <div class="card shadow-sm" data-test-id="card-user-group-edit-members">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
      <form @submit.prevent="$emit('submit')">
        <c-member-picker
          :value="value"
          no-remove
          @input="$emit('input', $event)"
        />
      </form>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit')"
      />
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import CMemberPicker from 'corteza-webapp-admin/src/components/CMemberPicker'

const { t } = useI18n()

defineProps({
  value: { type: Array, required: true, default: () => [] },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
})

defineEmits(['input', 'submit'])
</script>
