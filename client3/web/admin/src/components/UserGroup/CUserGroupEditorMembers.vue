<template>
  <div class="card shadow-sm" data-test-id="card-user-group-edit-members">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('system.user-groups.editor.members.title') }}</h4>
    </div>

    <div class="card-body">
      <form @submit.prevent="$emit('submit')">
        <c-member-picker
          :value="modelValue"
          no-remove
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
defineOptions({ i18nOptions: { namespaces: 'system.user-groups', keyPrefix: 'editor.members' } })
import { useI18n } from 'vue-i18n'
import CMemberPicker from 'corteza-webapp-admin/src/components/CMemberPicker'

const { t } = useI18n()

defineProps({
  modelValue: { type: Array, required: true, default: () => [] },
  processing: { type: Boolean, default: false },
  success: { type: Boolean, default: false },
})

defineEmits(['update:modelValue', 'submit'])
</script>
