<template>
  <div>
    <div class="mb-3">
      <div class="form-check">
        <input
          id="oidc-enabled"
          v-model="value.enabled"
          class="form-check-input-v3"
          type="checkbox"
          :true-value="true"
          :false-value="false"
        >
        <label
          class="form-check-label"
          for="oidc-enabled"
        >{{ t('enabled') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('handle') }}</label>
      <div class="input-group">
        <input
          :value="value.handle"
          class="form-control"
          @input="onHandleInput"
          :disabled="!fresh"
        >
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('issuer') }}</label>
      <div class="input-group">
        <input
          v-model.trim="value.issuer"
          class="form-control"
          placeholder="https://issuer.tld"
        >
      </div>
      <div class="form-text" v-html="t('issuerHint')" />
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('clientKey') }}</label>
      <div class="input-group">
        <input
          v-model.trim="value.key"
          class="form-control"
        >
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('clientSecret') }}</label>
      <div class="input-group">
        <input
          v-model.trim="value.secret"
          class="form-control"
        >
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('scope') }}</label>
      <div class="input-group">
        <input
          v-model.trim="value.scope"
          class="form-control"
          :placeholder="t('scopePlaceholder')"
        >
      </div>
      <div class="form-text" v-html="t('scopeHint')" />
    </div>

    <security
      v-model="value.security"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Security from './ExternalSecurity'

const { t } = useI18n()

const props = defineProps({
  value: {
    type: Object,
    required: true,
  },
})

const fresh = computed(() => {
  return Object.prototype.hasOwnProperty.call(props.value, 'fresh') && props.value.fresh
})

function onHandleInput (e) {
  props.value.handle = e.target.value.replace(/[^a-zA-Z0-9\-_]+/, '')
}
</script>
<style scoped lang="scss">
</style>
