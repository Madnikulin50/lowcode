<template>
  <div class="card shadow-sm auth-clients" data-test-id="card-auth-client-info" v-if="resource">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="submit">
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('name') }}</label>
              <input
                v-model="resource.meta.name"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': nameState === false }"
                data-test-id="input-name"
                required
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.handle.label') }}</label>
              <input
                v-model="resource.handle"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': handleState === false }"
                data-test-id="input-handle"
                :disabled="resource.isDefault"
                :placeholder="$t('system.authclients.editor.info.handle.placeholder-handle')"
              >
              <div v-if="handleState === false" class="invalid-feedback" data-test-id="feedback-invalid-handle">
                {{ $t('system.authclients.editor.info.handle.invalid-handle-characters') }}
              </div>
              <small v-if="resource.isDefault" class="form-text text-muted">{{ $t('system.authclients.editor.info.handle.disabledFootnote') }}</small>
            </div>
          </div>
  
          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.redirectURI') }}</label>
              <c-form-table-wrapper
                :labels="{ addButton: $t('label.add') }"
                test-id="button-add-redirect-uris"
                @add-item="redirectURI.push('')"
              >
                <div v-if="redirectURI.length">
                  <div v-for="(value, index) in redirectURI" :key="index" class="input-group mb-2">
                    <input
                      v-model="redirectURI[index]"
                      type="text"
                      class="form-control"
                      data-test-id="input-uri"
                      :placeholder="$t('system.authclients.editor.info.uri')"
                    >
                    <button
                      type="button"
                      class="btn btn-link text-danger ms-1"
                      data-test-id="button-remove-uri"
                      @click="redirectURI.splice(index, 1)"
                    >
                      <font-awesome-icon :icon="['fas', 'times']" />
                    </button>
                  </div>
                </div>
              </c-form-table-wrapper>
            </div>
          </div>
  
          <div class="col-12">
            <div v-if="!fresh" class="mb-3">
              <label class="form-label text-primary d-flex align-items-center gap-1">
                {{ $t('system.authclients.editor.info.secret') }}
                <button
                  v-if="!secretVisible"
                  type="button"
                  class="btn btn-outline-light text-secondary border-0"
                  data-test-id="button-show-client-secret"
                  :title="$t('system.authclients.editor.info.tooltip.show-client-secret')"
                  @click="showSecret()"
                >
                  <font-awesome-icon :icon="['fas', 'eye']" />
                </button>
                <button
                  v-else
                  type="button"
                  class="btn btn-outline-light text-secondary border-0"
                  data-test-id="button-hide-client-secret"
                  :title="$t('system.authclients.editor.info.tooltip.hide-client-secret')"
                  @click="hideSecret()"
                >
                  <font-awesome-icon :icon="['fas', 'eye-slash']" />
                </button>
              </label>
              <div class="input-group">
                <input
                  v-model="secret"
                  type="text"
                  class="form-control"
                  data-test-id="input-client-secret"
                  disabled
                  placeholder="****************************************************************"
                >
                <button
                  type="button"
                  class="btn btn-outline-light text-secondary border-0"
                  data-test-id="button-regenerate-client-secret"
                  :title="$t('system.authclients.editor.info.tooltip.regenerate-secret')"
                  @click="regenerateSecret()"
                >
                  <font-awesome-icon :icon="['fas', 'sync']" />
                </button>
              </div>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <div class="btn-group" role="group">
                <template v-for="opt in grantOptions" :key="opt.value">
                  <input
                    :id="'grant-' + opt.value"
                    v-model="resource.validGrant"
                    type="radio"
                    class="btn-check"
                    name="grant-options"
                    :value="opt.value"
                    autocomplete="off"
                    @change="onGrantChange"
                  >
                  <label
                    :for="'grant-' + opt.value"
                    class="btn btn-outline-primary"
                  >
                    {{ opt.text }}
                  </label>
                </template>
              </div>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <div class="form-check">
                <input
                  id="scope-profile"
                  type="checkbox"
                  class="form-check-input-v3"
                  data-test-id="checkbox-allow-access-to-user-profile"
                  :checked="(resource.scope || []).includes('profile')"
                  @change="setScope($event, 'profile')"
                >
                <label class="form-check-label" for="scope-profile">{{ $t('system.authclients.editor.info.profile') }}</label>
              </div>
              <div class="form-check">
                <input
                  id="scope-api"
                  type="checkbox"
                  class="form-check-input-v3"
                  data-test-id="checkbox-allow-access-to-corteza-api"
                  :checked="(resource.scope || []).includes('api')"
                  @change="setScope($event, 'api')"
                >
                <label class="form-check-label" for="scope-api">{{ $t('system.authclients.editor.info.api') }}</label>
              </div>
              <div class="form-check">
                <input
                  id="scope-openid"
                  type="checkbox"
                  class="form-check-input-v3"
                  data-test-id="checkbox-allow-client-to-use-oidc"
                  :checked="(resource.scope || []).includes('openid')"
                  @change="setScope($event, 'openid')"
                >
                <label class="form-check-label" for="scope-openid">{{ $t('system.authclients.editor.info.openid') }}</label>
              </div>
              <div v-if="discoveryEnabled" class="form-check">
                <input
                  id="scope-discovery"
                  type="checkbox"
                  class="form-check-input-v3"
                  data-test-id="checkbox-allow-client-access-to-discovery"
                  :checked="(resource.scope || []).includes('discovery')"
                  @change="setScope($event, 'discovery')"
                >
                <label class="form-check-label" for="scope-discovery">{{ $t('system.authclients.editor.info.discovery') }}</label>
              </div>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" data-test-id="valid-from">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.validFrom.label') }}</label>
              <small class="form-text text-muted">{{ $t('system.authclients.editor.info.validFrom.description') }}</small>
              <c-input-date-time
                v-model="resource.validFrom"
                data-test-id="input-valid-from"
                :labels="{
                  clear: $t('label.clear'),
                  none: $t('label.none'),
                  now: $t('label.now'),
                  today: $t('label.today'),
                }"
              />
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" data-test-id="expires-at">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.expiresAt.label') }}</label>
              <small class="form-text text-muted">{{ $t('system.authclients.editor.info.expiresAt.description') }}</small>
              <c-input-date-time
                v-model="resource.expiresAt"
                data-test-id="input-expires-at"
                :labels="{
                  clear: $t('label.clear'),
                  none: $t('label.none'),
                  now: $t('label.now'),
                  today: $t('label.today'),
                }"
              />
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <div class="form-check">
                <input
                  id="checkbox-enabled"
                  v-model="resource.enabled"
                  type="checkbox"
                  class="form-check-input-v3"
                  data-test-id="checkbox-is-client-enabled"
                  :disabled="resource.isDefault"
                >
                <label class="form-check-label" for="checkbox-enabled">{{ $t('system.authclients.editor.info.enabled.label') }}</label>
              </div>
              <small v-if="resource.isDefault" class="form-text text-muted">{{ $t('system.authclients.editor.info.enabled.disabledFootnote') }}</small>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <div class="form-check">
                <input
                  id="checkbox-trusted"
                  v-model="resource.trusted"
                  type="checkbox"
                  class="form-check-input-v3"
                  data-test-id="checkbox-is-client-trusted"
                >
                <label class="form-check-label" for="checkbox-trusted">{{ $t('system.authclients.editor.info.trusted.label') }}</label>
              </div>
              <small class="form-text text-muted">{{ $t('system.authclients.editor.info.trusted.description') }}</small>
            </div>
          </div>
        </div>
  
        <div class="row">
          <div v-show="isClientCredentialsGrant" class="col-12 col-lg-6">
            <div class="mb-3" data-test-id="impersonate-user">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.security.impersonateUser.label') }}</label>
              <small class="form-text text-muted">{{ $t('system.authclients.editor.info.security.impersonateUser.description') }}</small>
              <c-input-user
                v-model="resource.security.impersonateUser"
                :placeholder="$t('system.authclients.editor.info.security.impersonateUser.placeholder')"
                :clearable="true"
              />
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" data-test-id="permitted-roles">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.security.permittedRoles.label') }}</label>
              <c-role-picker v-model="resource.security.permittedRoles">
                <template #description>{{ $t('system.authclients.editor.info.security.permittedRoles.description') }}</template>
              </c-role-picker>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" data-test-id="prohibited-roles">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.security.prohibitedRoles.label') }}</label>
              <c-role-picker v-model="resource.security.prohibitedRoles">
                <template #description>{{ $t('system.authclients.editor.info.security.prohibitedRoles.description') }}</template>
              </c-role-picker>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" data-test-id="forced-roles">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.security.forcedRoles.label') }}</label>
              <c-role-picker v-model="resource.security.forcedRoles" class="mb-3">
                <template #description>{{ $t('system.authclients.editor.info.security.forcedRoles.description') }}</template>
              </c-role-picker>
            </div>
          </div>
  
          <div v-if="!fresh && isClientCredentialsGrant" class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary d-flex align-items-center gap-1">
                {{ $t('system.authclients.editor.info.cUrl') }}
                <button
                  type="button"
                  class="btn btn-outline-light text-secondary border-0"
                  data-test-id="button-copy-cURL"
                  :title="$t('system.authclients.editor.info.tooltip.copy-cURL')"
                  @click="copyToClipboard(exampleCurl)"
                >
                  <font-awesome-icon :icon="['far', 'copy']" />
                </button>
              </label>
              <textarea
                :value="exampleCurl"
                class="form-control"
                data-test-id="cURL-string"
                disabled
                rows="3"
              />
            </div>
  
            <div class="mb-3">
              <label class="form-label text-primary d-flex align-items-center gap-1">
                {{ $t('system.authclients.editor.info.accessToken') }}
                <button
                  v-if="tokenRequest.token"
                  type="button"
                  class="btn btn-outline-light text-secondary border-0"
                  data-test-id="copy-token-from-request"
                  :title="$t('system.authclients.editor.info.tooltip.copy-access-token')"
                  @click="copyToClipboard(tokenRequest.token)"
                >
                  <font-awesome-icon :icon="['far', 'copy']" />
                </button>
              </label>
              <textarea
                v-if="tokenRequest.token"
                :value="tokenRequest.token"
                class="form-control"
                data-test-id="cURL-string"
                disabled
                rows="5"
              />
              <button
                v-else
                type="button"
                class="btn btn-outline-secondary"
                data-test-id="button-test-cURL"
                @click="getAccessTokenAPI()"
              >
                {{ $t('system.authclients.editor.info.generateAccessToken') }}
              </button>
              <p v-if="tokenRequest.error" class="text-danger mt-2">{{ tokenRequest.error }}</p>
            </div>
          </div>
        </div>
  
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.authclients.editor.info.security.defaultUserGroup.label') }}</label>
              <c-input-user-group
                v-model="resource.security.userGroup"
                :placeholder="$t('system.authclients.editor.info.security.defaultUserGroup.placeholder')"
              />
            </div>
          </div>
        </div>
  
        <c-system-fields :resource="resource" />
  
        <input
          type="submit"
          class="d-none"
          data-test-id="button-submit"
          :disabled="saveDisabled"
        >
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <template v-if="canDelete">
        <c-input-confirm
          :data-test-id="isDeleted ? 'button-undelete' : 'button-delete'"
          :disabled="processing"
          :text="isDeleted ? $t('undelete') : $t('delete')"
          variant="danger"
          size="md"
          @confirmed="$emit(isDeleted ? 'undelete' : 'delete', resource.authClientID)"
        />
      </template>

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="submit"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.authclients', keyPrefix: 'editor.info' } })
import { ref, computed, watch, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { handle, components } from 'corteza-lib/vue/dist'
import CRolePicker from 'corteza-webapp-admin/src/components/CRolePicker'
import copy from 'copy-to-clipboard'
import axios from 'axios'

const { CInputDateTime, CInputUser, CInputUserGroup } = components
const { t } = useI18n()
const $auth = inject('auth', {})
const $Settings = inject('$Settings', {})

const props = defineProps({
  resource: {
    type: Object,
    required: true,
  },
  canDelete: {
    type: Boolean,
    default: () => false,
  },
  processing: {
    type: Boolean,
    value: false,
  },
  success: {
    type: Boolean,
    value: false,
  },
  canCreate: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['submit', 'delete', 'undelete'])

const requestedSecret = ref('')
const secret = ref('')
const redirectURI = ref(props.resource.redirectURI ? props.resource.redirectURI.split(' ') : [])
const curlVisible = ref(false)
const tokenRequest = ref({
  token: '',
  error: '',
})

const grantOptions = [
  { value: 'authorization_code', text: t('grant.authorization_code') },
  { value: 'client_credentials', text: t('grant.client_credentials') },
]

const fresh = computed(() => {
  return !props.resource.authClientID || props.resource.authClientID === NoID
})

const editable = computed(() => {
  return fresh.value ? props.canCreate : props.resource.canUpdateAuthClient
})

const isDeleted = computed(() => {
  return props.resource.deletedAt && props.resource.canDeleteAuthClient
})

const secretVisible = computed(() => {
  return secret.value.length > 0
})

const nameState = computed(() => {
  return props.resource.meta.name ? null : false
})

const handleState = computed(() => {
  return handle.handleState(props.resource.handle)
})

const isClientCredentialsGrant = computed(() => {
  return props.resource.validGrant === 'client_credentials'
})

const discoveryEnabled = computed(() => {
  return $Settings.get('discovery.enabled', false)
})

const saveDisabled = computed(() => {
  return !editable.value || [nameState.value, handleState.value].includes(false)
})

const curlURL = computed(() => {
  return $auth.cortezaAuthURL + '/oauth2/token'
})

const exampleCurl = computed(() => {
  return `curl -X POST ${curlURL.value} -d grant_type=${props.resource.validGrant} -d scope='${props.resource.scope}' -u ${props.resource.authClientID}:${secret.value || 'PLACE-YOUR-CLIENT-SECRET-HERE'}`
})

watch(redirectURI, (val) => {
  props.resource.redirectURI = val.filter(ru => ru).join(' ')
}, { deep: true })

function onGrantChange(grant) {
  if (grant === 'client_credentials' && (!props.resource.security.impersonateUser || props.resource.security.impersonateUser === NoID)) {
    props.resource.security.impersonateUser = $auth.user.userID
  }
}

function copyToClipboard(value) {
  copy(value)
}

function toggleCurlSnippet() {
  curlVisible.value = !curlVisible.value
}

function submit() {
  if (!isClientCredentialsGrant.value || !props.resource.security.impersonateUser) {
    props.resource.security.impersonateUser = NoID
  }

  emit('submit', props.resource)
}

function setScope(event, target) {
  const value = event.target.checked
  let items = props.resource.scope ? props.resource.scope.split(' ') : []

  if (value) {
    items.push(target)
  } else {
    items = items.filter(i => i !== target)
  }

  props.resource.scope = items.join(' ')
}

function requestSecret() {
  const clientID = props.resource.authClientID
  return window.__SystemAPI.authClientExposeSecret({ clientID }).then(s => {
    requestedSecret.value = s
  })
}

async function showSecret() {
  if (!requestedSecret.value) {
    await requestSecret()
  }
  secret.value = requestedSecret.value
}

function hideSecret() {
  secret.value = ''
}

async function regenerateSecret() {
  const clientID = props.resource.authClientID
  window.__SystemAPI.authClientRegenerateSecret({ clientID }).then(s => {
    requestedSecret.value = s
  })
}

async function getAccessTokenAPI() {
  const clientID = props.resource.authClientID

  if (!requestedSecret.value) {
    await requestSecret()
  }

  const params = new URLSearchParams()
  params.append('grant_type', props.resource.validGrant)
  params.append('scope', props.resource.scope)

  axios.post(curlURL.value, params, { auth: { username: clientID, password: requestedSecret.value } })
    .then(response => {
      tokenRequest.value.token = (response.data || {}).access_token
      tokenRequest.value.error = ''
    })
    .catch(e => {
      const { error } = e.response.data || {}
      tokenRequest.value.error = error
      tokenRequest.value.token = ''
    })
}
</script>
