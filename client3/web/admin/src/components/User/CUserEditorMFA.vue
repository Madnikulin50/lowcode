<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
      <div class="d-flex align-items-center flex-wrap">
        <div>
          <span v-if="mfa.enforcedEmailOTP" v-html="$t('emailOTP.enabled.text')" />
          <span v-else v-html="$t('emailOTP.disabled.text')" />
        </div>
        <div class="ms-auto">
          <button
            v-if="mfa.enforcedEmailOTP"
            class="btn btn-outline-secondary"
            @click="$emit('patch', '/meta/securityPolicy/mfa/enforcedEmailOTP', false)"
          >
            {{ $t('emailOTP.disable.label') }}
          </button>
          <button
            v-else
            class="btn btn-outline-secondary"
            @click="$emit('patch', '/meta/securityPolicy/mfa/enforcedEmailOTP', true)"
          >
            {{ $t('emailOTP.enable.label') }}
          </button>
        </div>
      </div>

      <div class="d-flex align-items-center justify-content-between flex-wrap mt-2 pt-2 border-top">
        <div>
          <span v-if="mfa.enforcedTOTP" v-html="$t('TOTP.enabled.text')" />
          <span v-else v-html="$t('TOTP.disabled.text')" />
        </div>
        <div class="ms-auto">
          <button
            class="btn btn-outline-secondary"
            :disabled="!mfa.enforcedTOTP"
            @click="$emit('patch', '/meta/securityPolicy/mfa/enforcedTOTP', false)"
          >
            {{ $t('TOTP.remove.label') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  mfa: { type: Object, required: true },
  processing: { type: Boolean, default: false },
  success: { type: Boolean, default: false },
})

defineEmits(['patch'])
</script>

<style scoped>
</style>
