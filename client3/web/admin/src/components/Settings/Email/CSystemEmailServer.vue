<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('title') }}
      </h4>
    </div>

    <form
      @submit.prevent="submit()"
    >
      <div class="card-body">
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('host.label') }}</label>
              <div class="form-text mb-2">{{ t('host.description') }}</div>
              <div class="input-group">
                <input
                  v-model="server.host"
                  class="form-control"
                  data-test-id="input-server"
                  :disabled="disabled"
                  placeholder="host.domain.tld"
                  autocomplete="off"
                  required
                >
                <span class="input-group-text">:</span>
                <input
                  v-model="server.port"
                  class="form-control"
                  data-test-id="input-server-port"
                  type="number"
                  :disabled="disabled"
                  step="1"
                  required
                >
              </div>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('user.label') }}</label>
              <div class="form-text mb-2">{{ t('user.description') }}</div>
              <input
                v-model="server.user"
                class="form-control"
                data-test-id="input-user"
                :disabled="disabled"
                autocomplete="off"
              >
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('password.label') }}</label>
              <div class="form-text mb-2">{{ t('password.description') }}</div>
              <input
                v-model="server.pass"
                class="form-control"
                data-test-id="input-password"
                type="password"
                :disabled="disabled"
                autocomplete="off"
              >
            </div>
          </div>
        </div>

        <hr>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('from.label') }}</label>
              <div class="form-text mb-2">{{ t('from.description') }}</div>
              <input
                v-model="server.from"
                class="form-control"
                data-test-id="input-sender-address"
                type="email"
                :disabled="disabled"
                autocomplete="off"
              >
            </div>
          </div>
        </div>

        <hr>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ t('tlsServerName.label') }}</label>
              <div class="form-text mb-2">{{ t('tlsServerName.description') }}</div>
              <input
                v-model="server.tlsServerName"
                class="form-control"
                data-test-id="input-tls-server-name"
                :disabled="disabled"
              >
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3 mt-lg-3">
              <div class="form-text mb-2">{{ t('tlsInsecure.description') }}</div>
              <div class="form-check mt-lg-4 mb-2">
                <input
                  id="tls-insecure"
                  v-model="server.tlsInsecure"
                  class="form-check-input-v3"
                  data-test-id="checkbox-allow-invalid-certificates"
                  type="checkbox"
                  :disabled="disabled"
                >
                <label
                  class="form-check-label"
                  for="tls-insecure"
                >{{ t('tlsInsecure.label') }}</label>
              </div>
            </div>
          </div>
        </div>
      </div>
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        data-test-id="button-smtp"
        :disabled="disabled"
        :processing="processingSMTPTest"
        :success="successSMTPTest"
        :text="t('testSmtpConfigs.button')"
        variant="outline-secondary"
        @submit="smtpConnectionCheck()"
      />

      <c-button-submit
        :disabled="disabled"
        :processing="processing"
        :success="success"
        :text="t('admin.general.label.submit')"
        class="ms-auto"
        @submit="submit()"
      />
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  value: {
    type: Object,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  processingSMTPTest: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
  successSMTPTest: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['submit', 'smtpConnectionCheck'])

const defaults = {
  host: '',
  port: 25,
  user: '',
  pass: '',
  from: '',
  tlsInsecure: false,
  tlsServerName: '',
}

const server = reactive({ ...defaults, ...props.value })

function submit () {
  emit('submit', server)
}

function smtpConnectionCheck () {
  emit('smtpConnectionCheck', server)
}
</script>
