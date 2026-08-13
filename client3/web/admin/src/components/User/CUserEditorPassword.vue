<template>
  <div class="card shadow-sm" data-test-id="card-user-password">
    <div class="card-header border-bottom">
      <h4 data-test-id="card-title" class="m-0">{{ $t('system.users.editor.password.title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="onPasswordSubmit">
        <div class="row g-3 p-3">
          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.users.editor.password.new') }}</label>
              <div class="form-text">{{ getPasswordWarning }}</div>
              <input
                v-model="password"
                data-test-id="input-new-password"
                :class="['form-control', { 'is-invalid': passwordState === false }]"
                autocomplete="new-password"
                required
                type="password"
              >
            </div>
          </div>
  
          <div class="col-12">
            <div class="mb-0">
              <label class="form-label text-primary">{{ $t('system.users.editor.password.confirm') }}</label>
              <div class="form-text">{{ getConfirmPasswordWarning }}</div>
              <input
                v-model="confirmPassword"
                data-test-id="input-confirm-password"
                type="password"
                autocomplete="new-password"
                required
                :disabled="!passwordState"
                :class="['form-control', { 'is-invalid': confirmPasswordState === false }]"
              >
            </div>
          </div>
        </div>
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        data-test-id="button-remove-password"
        :text="$t('system.users.editor.password.removePassword')"
        variant="outline-secondary"
        size="md"
        @confirmed="$emit('submit')"
      />

      <c-corredor-manual-buttons
        ui-page="user/editor"
        ui-slot="passwordFooter"
        resource-type="system:user"
        default-variant="outline-secondary"
        @click="dispatchCortezaSystemEvent($event)"
      />

      <c-button-submit
        :disabled="!passwordState || !confirmPasswordState"
        :processing="processing"
        :success="success"
        :text="$t('admin.general.label.submit')"
        class="ms-auto"
        @submit="onPasswordSubmit"
      />
    </div>
  </div>
</template>

<script setup>
import { useNsI18n } from 'corteza-lib/vue/dist'
defineOptions({ i18nOptions: { namespaces: 'system.users', keyPrefix: 'editor.password' } })
import { ref, computed, inject } from 'vue'


const t = useNsI18n()
const $Settings = inject('$Settings', {})

const props = defineProps({
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  userID: { type: String, required: false, default: undefined },
})

const emit = defineEmits(['submit'])

const password = ref('')
const confirmPassword = ref('')
const minPasswordLength = ref($Settings.get('auth.internal.passwordConstraints.minLength', 8))

const passwordState = computed(() => {
  if (password.value.length > 0) {
    return password.value.length >= minPasswordLength.value
  }
  return null
})

const confirmPasswordState = computed(() => {
  if (passwordState.value && confirmPassword.value.length > 0) {
    return password.value === confirmPassword.value
  }
  return null
})

const getPasswordWarning = computed(() => {
  if (passwordState.value === false) {
    return t('length', { length: minPasswordLength.value })
  }
  return null
})

const getConfirmPasswordWarning = computed(() => {
  if (confirmPasswordState.value === false) {
    return t('missmatch')
  }
  return null
})

function onPasswordSubmit() {
  emit('submit', password.value)
  password.value = ''
  confirmPassword.value = ''
}

function dispatchCortezaSystemEvent($event) {}
</script>

<style scoped>
</style>
