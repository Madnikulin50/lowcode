<template>
  <div class="card shadow-sm" data-test-id="card-user-group-edit-roles">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('system.user-groups.editor.roles.title') }}</h4>
    </div>

    <div class="card-body">
      <form @submit.prevent="$emit('submit')">
        <c-role-picker
          :value="modelValue"
          @input="$emit('update:modelValue', $event)"
        />
      </form>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :processing="processing"
        :success="success"
        :text="$t('admin.general.label.submit')"
        class="ms-auto"
        @submit="$emit('submit')"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.user-groups', keyPrefix: 'editor.roles' } })
import { useI18n } from 'vue-i18n'
import CRolePicker from 'corteza-webapp-admin/src/components/CRolePicker'

const { t } = useI18n()

defineProps({
  modelValue: { type: Array, required: true, default: () => [] },
  processing: { type: Boolean, default: false },
  success: { type: Boolean, default: false },
})

defineEmits(['update:modelValue', 'submit'])
</script>
