<template>
  <div>
    <div class="mb-3">
      <div class="form-check">
        <input
          id="saml-enabled"
          v-model="value.enabled"
          class="form-check-input-v3"
          type="checkbox"
          :true-value="true"
          :false-value="false"
        >
        <label
          class="form-check-label"
          for="saml-enabled"
        >{{ t('enabled') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('name') }}</label>
      <div class="form-text mb-2">{{ t('desc.name') }}</div>
      <div class="input-group">
        <input
          v-model.trim="value.name"
          class="form-control"
        >
      </div>
    </div>

    <hr>

    <h5>
      {{ t('certificate') }}
    </h5>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('cert.public') }}</label>
      <div class="form-text mb-2">{{ t('desc.cert.public') }}</div>
      <div class="input-group">
        <textarea
          v-model.trim="value.cert"
          class="form-control"
          rows="3"
        />
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('cert.private') }}</label>
      <div class="form-text mb-2">{{ t('desc.cert.private') }}</div>
      <div class="input-group">
        <textarea
          v-model.trim="value.key"
          class="form-control"
          rows="3"
        />
      </div>
    </div>

    <hr>

    <h5>
      {{ t('requests.title') }}
    </h5>

    <div class="mb-3">
      <div class="form-text mb-2">{{ t('desc.requests.sign-requests') }}</div>
      <div class="form-check">
        <input
          id="sign-requests"
          v-model="value['sign-requests']"
          class="form-check-input-v3"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="sign-requests"
        >{{ t('requests.sign-requests') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('requests.sign-method') }}</label>
      <div class="form-text mb-2">{{ t('desc.requests.sign-method') }}</div>
      <div class="input-group">
        <select
          v-model.trim="value['sign-method']"
          class="form-select"
        >
          <option
            value=""
            disabled
          >
            {{ t('admin.general.label.selectOption') }}
          </option>
          <option
            v-for="opt in signMethods"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.text }}
          </option>
        </select>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('requests.binding') }}</label>
      <div class="form-text mb-2">{{ t('desc.requests.binding') }}</div>
      <div class="input-group">
        <select
          v-model.trim="value['binding']"
          class="form-select"
        >
          <option
            value=""
            disabled
          >
            {{ t('admin.general.label.selectOption') }}
          </option>
          <option
            v-for="opt in httpBindings"
            :key="opt.value"
            :value="opt.value"
          >
            {{ opt.text }}
          </option>
        </select>
      </div>
    </div>

    <hr>

    <h5>
      {{ t('idp.title') }}
    </h5>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('idp.url') }}</label>
      <div class="form-text mb-2">{{ t('desc.idp.url') }}</div>
      <div class="input-group">
        <input
          v-model.trim="value.idp.url"
          class="form-control"
        >
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('idp.ident-name') }}</label>
      <div class="form-text mb-2">{{ t('desc.idp.ident-name') }}</div>
      <div class="input-group">
        <input
          v-model.trim="value.idp['ident-name']"
          class="form-control"
        >
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('idp.ident-handle') }}</label>
      <div class="form-text mb-2">{{ t('desc.idp.ident-handle') }}</div>
      <div class="input-group">
        <input
          v-model.trim="value.idp['ident-handle']"
          class="form-control"
        >
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('idp.ident-identifier') }}</label>
      <div class="form-text mb-2">{{ t('desc.idp.ident-identifier') }}</div>
      <div class="input-group">
        <input
          v-model.trim="value.idp['ident-identifier']"
          class="form-control"
        >
      </div>
    </div>

    <security
      v-model="value.security"
    />
  </div>
</template>

<script setup>
import { useNsI18n } from 'corteza-lib/vue/dist'
defineOptions({ i18nOptions: { namespaces: 'system.settings', keyPrefix: 'editor.external.saml' } })

import Security from './ExternalSecurity'

const t = useNsI18n()

defineProps({
  value: {
    type: Object,
    required: true,
    default: () => ({}),
  },
})

const signMethods = [
  { value: 'http://www.w3.org/2000/09/xmldsig#rsa-sha1', text: 'SHA1' },
  { value: 'http://www.w3.org/2001/04/xmldsig-more#rsa-sha256', text: 'SHA256' },
  { value: 'http://www.w3.org/2001/04/xmldsig-more#rsa-sha512', text: 'SHA512' },
]

const httpBindings = [
  { value: 'urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST', text: t('requests.binding-post') },
  { value: 'urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect', text: t('requests.binding-redirect') },
]
</script>
<style scoped lang="scss">
</style>
