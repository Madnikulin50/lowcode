<template>
  <div
    class="card shadow-sm"
    data-test-id="card-edit-authentication"
  >
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('title') }}
      </h4>
    </div>

    <form
      @submit.prevent="emit('submit', authSettings)"
    >
      <div class="card-body">
        <h5>
          {{ t('internal.title') }}
        </h5>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('internal.enabled') }}</label>
              <c-input-checkbox
                v-model="authSettings['auth.internal.enabled']"
                switch
                :value="true"
                :unchecked-value="false"
                :labels="checkboxLabel"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('internal.password-reset.enabled') }}</label>
              <c-input-checkbox
                v-model="authSettings['auth.internal.password-reset.enabled']"
                switch
                data-test-id="checkbox-password-reset"
                :value="true"
                :labels="checkboxLabel"
                :unchecked-value="false"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('internal.signup.email-confirmation-required') }}</label>
              <c-input-checkbox
                v-model="authSettings['auth.internal.signup.email-confirmation-required']"
                switch
                :value="true"
                :labels="checkboxLabel"
                :unchecked-value="false"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('internal.signup.enabled') }}</label>
              <c-input-checkbox
                v-model="authSettings['auth.internal.signup.enabled']"
                :value="true"
                :unchecked-value="false"
                :labels="checkboxLabel"
                switch
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('internal.profile-avatar.enabled') }}</label>
              <c-input-checkbox
                v-model="authSettings['auth.internal.profile-avatar.enabled']"
                :value="true"
                :unchecked-value="false"
                :labels="checkboxLabel"
                switch
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('internal.signup.split-credentials-check.label') }}</label>
              <c-input-checkbox
                v-model="authSettings['auth.internal.split-credentials-check']"
                :value="true"
                :unchecked-value="false"
                :labels="checkboxLabel"
                switch
              />
            </div>
          </div>
        </div>

        <hr>

        <div>
          <h5>
            {{ t('internal.password-constraints.title') }}
          </h5>

          <div
            v-if="!$Settings.get('auth.internal.passwordConstraints.passwordSecurity')"
            class="alert alert-warning"
            role="alert"
          >
            {{ t('internal.password-constraints.ignored-security') }}
          </div>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('internal.password-constraints.min-upper-case-length') }}</label>
                <div class="form-text mb-2">{{ t('internal.password-constraints.min-upper-case-description') }}</div>
                <input
                  v-model.number="authSettings['auth.internal.password-constraints.min-upper-case']"
                  class="form-control"
                  type="number"
                  :placeholder="`${defaultMinUppCaseChrs}`"
                  :min="defaultMinUppCaseChrs"
                >
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('internal.password-constraints.min-lower-case-length') }}</label>
                <div class="form-text mb-2">{{ t('internal.password-constraints.min-lower-case-description') }}</div>
                <input
                  v-model.number="authSettings['auth.internal.password-constraints.min-lower-case']"
                  class="form-control"
                  type="number"
                  :placeholder="`${defaultMinLowCaseChrs}`"
                  :min="defaultMinLowCaseChrs"
                >
              </div>
            </div>
          </div>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('internal.password-constraints.min-length') }}</label>
                <div class="form-text mb-2">{{ t('internal.password-constraints.min-length-description') }}</div>
                <input
                  v-model.number="authSettings['auth.internal.password-constraints.min-length']"
                  class="form-control"
                  :placeholder="`${defaultMinPwd}`"
                  :min="defaultMinPwd"
                  type="number"
                >
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('internal.password-constraints.min-num-count') }}</label>
                <div class="form-text mb-2">{{ t('internal.password-constraints.min-num-count-description') }}</div>
                <input
                  v-model.number="authSettings['auth.internal.password-constraints.min-num-count']"
                  class="form-control"
                  placeholder="0"
                  min="0"
                  type="number"
                >
              </div>
            </div>
          </div>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('internal.password-constraints.min-special-count') }}</label>
                <div class="form-text mb-2">{{ t('internal.password-constraints.min-special-count-description') }}</div>
                <input
                  v-model.number="authSettings['auth.internal.password-constraints.min-special-count']"
                  class="form-control"
                  placeholder="0"
                  min="0"
                  type="number"
                >
              </div>
            </div>
          </div>
        </div>

        <hr>

        <div>
          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mfa.emailOTP.enabled') }}</label>
                <c-input-checkbox
                  v-model="authSettings['auth.multi-factor.email-otp.enabled']"
                  data-test-id="checkbox-enable-emailOTP"
                  :value="true"
                  :unchecked-value="false"
                  :labels="checkboxLabel"
                  switch
                  @change="handleEmailOTPEnabledChange"
                />
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mfa.emailOTP.expires.label') }}</label>
                <div class="form-text mb-2">{{ t('mfa.emailOTP.expires.description') }}</div>
                <div class="input-group">
                  <input
                    v-model="authSettings['auth.multi-factor.email-otp.expires']"
                    class="form-control"
                    type="number"
                    placeholder="60"
                  >
                  <span class="input-group-text">seconds</span>
                </div>
              </div>
            </div>

            <div
              v-if="authSettings['auth.multi-factor.email-otp.enabled']"
              class="col-12 col-lg-6"
            >
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mfa.emailOTP.enforced') }}</label>
                <c-input-checkbox
                  v-model="authSettings['auth.multi-factor.email-otp.enforced']"
                  :value="true"
                  :unchecked-value="false"
                  :labels="checkboxLabel"
                  :disabled="!authSettings['auth.multi-factor.email-otp.enabled']"
                  switch
                  @change="handleEmailOTPEnforcedChange"
                />
              </div>
            </div>
          </div>
        </div>

        <hr>

        <div>
          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mfa.TOTP.enabled') }}</label>
                <c-input-checkbox
                  v-model="authSettings['auth.multi-factor.totp.enabled']"
                  data-test-id="checkbox-enable-TOTP"
                  :value="true"
                  :unchecked-value="false"
                  :labels="checkboxLabel"
                  switch
                  @change="handleTOTPEnabledChange"
                />
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mfa.TOTP.issuer.label') }}</label>
                <div class="form-text mb-2">{{ t('mfa.TOTP.issuer.description') }}</div>
                <div class="input-group">
                  <input
                    v-model="authSettings['auth.multi-factor.totp.issuer']"
                    class="form-control"
                    placeholder="Lowcoooode"
                  >
                </div>
              </div>
            </div>

            <div
              v-if="authSettings['auth.multi-factor.totp.enabled']"
              class="col-12 col-lg-6"
            >
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mfa.TOTP.enforced') }}</label>
                <c-input-checkbox
                  v-model="authSettings['auth.multi-factor.totp.enforced']"
                  :value="true"
                  :unchecked-value="false"
                  :labels="checkboxLabel"
                  :disabled="!authSettings['auth.multi-factor.totp.enabled']"
                  switch
                  @change="handleTOTPEnforcedChange"
                />
              </div>
            </div>
          </div>
        </div>

        <hr>

        <div>
          <h5>
            {{ t('mail.title') }}
          </h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mail.from-address') }}</label>
                <div class="form-text mb-2">{{ t('mail.validate-email') }}</div>
                <div class="input-group">
                  <input
                    v-model="authSettings['auth.mail.from-address']"
                    class="form-control"
                    type="email"
                  >
                </div>
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('mail.from-name') }}</label>
                <div class="input-group">
                  <input
                    v-model="authSettings['auth.mail.from-name']"
                    class="form-control"
                  >
                </div>
              </div>
            </div>
          </div>
        </div>

        <hr>

        <div>
          <h5>
            {{ t('internal.send-user-invite-email.title') }}
          </h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('internal.send-user-invite-email.enabled') }}</label>
                <div class="form-text mb-2">{{ t('internal.send-user-invite-email.description') }}</div>
                <c-input-checkbox
                  v-model="authSettings['auth.internal.send-user-invite-email.enabled']"
                  :value="true"
                  :unchecked-value="false"
                  :labels="checkboxLabel"
                  switch
                />
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('internal.send-user-invite-email.expires.label') }}</label>
                <div class="form-text mb-2">{{ t('internal.send-user-invite-email.expires.description') }}</div>
                <div class="input-group">
                  <input
                    v-model="authSettings['auth.internal.send-user-invite-email.expires']"
                    class="form-control"
                    type="number"
                  >
                  <span class="input-group-text">hours</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <hr>

        <div>
          <h5>
            {{ t('auto-logout.title') }}
          </h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('auto-logout.enabled.label') }}</label>
                <div class="form-text mb-2">{{ t('auto-logout.enabled.description') }}</div>
                <c-input-checkbox
                  v-model="authSettings['auth.auto-logout.enabled']"
                  :value="true"
                  :unchecked-value="false"
                  :labels="checkboxLabel"
                  switch
                />
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('auto-logout.timeout.label') }}</label>
                <div class="form-text mb-2">{{ t('auto-logout.timeout.description') }}</div>
                <div class="input-group">
                  <input
                    v-model="authSettings['auth.auto-logout.timeout']"
                    class="form-control"
                    type="number"
                  >
                  <span class="input-group-text">seconds</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        v-if="canManage"
        :processing="processing"
        :success="success"
        :text="t('admin.general.label.submit')"
        class="ms-auto"
        @submit="emit('submit', authSettings)"
      />
    </div>
  </div>
</template>

<script setup>
import { useNsI18n } from 'corteza-lib/vue/dist'
defineOptions({ i18nOptions: { namespaces: 'system.settings', keyPrefix: 'editor.auth' } })
import { ref, watch, inject } from 'vue'


const t = useNsI18n()
const $Settings = inject('$Settings', {})

const props = defineProps({
  settings: {
    type: Object,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
  canManage: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['submit'])

const defaultMinPwd = 8
const defaultMinUppCaseChrs = 0
const defaultMinLowCaseChrs = 0
const checkboxLabel = {
  on: t('label.general.yes'),
  off: t('label.general.no'),
}

const authSettings = ref({})

watch(() => props.settings, (settings) => {
  authSettings.value = settings
}, { immediate: true, deep: true })

function handleEmailOTPEnabledChange (value) {
  if (!value) {
    authSettings.value['auth.multi-factor.email-otp.enforced'] = false
  }
}

function handleTOTPEnabledChange (value) {
  if (!value) {
    authSettings.value['auth.multi-factor.totp.enforced'] = false
  }
}

function handleEmailOTPEnforcedChange (value) {
  if (value) {
    authSettings.value['auth.multi-factor.email-otp.enforced'] = true
  }
}

function handleTOTPEnforcedChange (value) {
  if (value) {
    authSettings.value['auth.multi-factor.totp.enforced'] = true
  }
}
</script>
