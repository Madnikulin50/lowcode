<template>
  <div class="card shadow-sm" data-test-id="card-role-edit-members">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit')">
        <c-member-picker
          :value="value"
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
defineOptions({ i18nOptions: { namespaces: 'system.roles', keyPrefix: 'editor.members' } })
import CMemberPicker from 'corteza-webapp-admin/src/components/CMemberPicker'

defineProps({
  value: { type: Array, required: true, default: () => [] },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
})

defineEmits(['submit', 'input'])
</script>