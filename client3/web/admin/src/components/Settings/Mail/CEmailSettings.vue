<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('editor.server.title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit', serverData)">
        <div class="row g-3 p-3">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('editor.server.host.label') }}</label>
              <small class="text-muted d-block mb-1">{{ $t('editor.server.host.description') }}</small>
              <div class="input-group">
                <input
                  v-model="serverData.host"
                  type="text"
                  class="form-control"
                  placeholder="host.domain.tld"
                  autocomplete="off"
                  required
                >
                <span class="input-group-text">:</span>
                <input
                  v-model="serverData.port"
                  type="number"
                  class="form-control"
                  step="1"
                  required
                >
              </div>
            </div>
          </div>
        </div>
  
        <div class="row g-3 p-3 pt-0">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('editor.server.user.label') }}</label>
              <small class="text-muted d-block mb-1">{{ $t('editor.server.user.description') }}</small>
              <input
                v-model="serverData.user"
                type="text"
                class="form-control"
                autocomplete="off"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('editor.server.password.label') }}</label>
              <small class="text-muted d-block mb-1">{{ $t('editor.server.password.description') }}</small>
              <input
                v-model="serverData.pass"
                type="password"
                class="form-control"
                autocomplete="off"
              >
            </div>
          </div>
        </div>
  
        <hr class="mx-3">
  
        <div class="row g-3 p-3 pt-0">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('editor.server.from.label') }}</label>
              <small class="text-muted d-block mb-1">{{ $t('editor.server.from.description') }}</small>
              <input
                v-model="serverData.from"
                type="email"
                class="form-control"
                autocomplete="off"
              >
            </div>
          </div>
        </div>
  
        <hr class="mx-3">
  
        <div class="row g-3 p-3 pt-0">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('editor.server.tlsServerName.label') }}</label>
              <small class="text-muted d-block mb-1">{{ $t('editor.server.tlsServerName.description') }}</small>
              <input
                v-model="serverData.tlsServerName"
                type="text"
                class="form-control"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <div class="form-check mt-lg-4 mb-2">
                <input
                  id="tls-insecure"
                  v-model="serverData.tlsInsecure"
                  type="checkbox"
                  class="form-check-input-v3"
                >
                <label class="form-check-label" for="tls-insecure">
                  {{ $t('editor.server.tlsInsecure.label') }}
                </label>
                <small class="text-muted d-block">{{ $t('editor.server.tlsInsecure.description') }}</small>
              </div>
            </div>
          </div>
        </div>
      </form>
  </div>
  </div>
</template>

<script setup>
import { reactive, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
  settings: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['update:modelValue', 'submit'])

const serverData = reactive({
  host: '',
  port: 25,
  user: '',
  pass: '',
  from: '',
  tlsInsecure: false,
  tlsServerName: '',
})

watch(() => props.modelValue?.['smtp.servers'], (val) => {
  if (val && val.length > 0 && typeof val[0] === 'object') {
    Object.assign(serverData, val[0])
  }
}, { immediate: true })

watch(serverData, () => {
  const servers = props.modelValue?.['smtp.servers']
  if (servers && servers.length > 0) {
    servers[0] = { ...serverData }
  }
}, { deep: true })
</script>
