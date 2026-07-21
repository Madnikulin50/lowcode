<template>
  <div>
    <div class="card shadow-sm" data-test-id="card-general-info">
      <div class="card-header border-bottom">
        <h4 class="m-0">{{ $t('label') }}</h4>
      </div>
      <div class="row g-3 p-3">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('id') }}</label>
            <input
              :value="request.ID"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('route') }}</label>
            <input
              :value="request.route"
              type="text"
              class="form-control-plaintext"
              data-test-id="input-route"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('URL') }}</label>
            <input
              :value="request.url"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('method') }}</label>
            <input
              :value="request.method"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('statusCode') }}</label>
            <div class="d-flex align-items-center h-100">
              <h5 class="mb-0">
                <span :class="'badge bg-' + getStatusCodeVariant(request.statusCode)">{{ request.statusCode }}</span>
              </h5>
            </div>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('remoteAddress') }}</label>
            <input
              :value="request.remoteAddress"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('duration') }}</label>
            <input
              :value="request.duration"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('start') }}</label>
            <input
              :value="request.start"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('end') }}</label>
            <input
              :value="request.end"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <button
            v-if="showOpenRoute"
            class="btn btn-outline-secondary"
            @click="openRoute"
          >
            {{ $t('openRoute') }}
          </button>
        </div>
      </div>
    </div>

    <div class="card shadow-sm mt-3">
      <div class="card-header border-bottom">
        <h4 class="m-0">{{ $t('headers.label') }}</h4>
      </div>
      <div class="row g-3 p-3">
        <div
          v-for="header in request.headers"
          :key="header.label"
          class="col-12 col-lg-6"
        >
          <div class="mb-3">
            <label class="form-label text-primary">{{ header.label }}</label>
            <input
              :value="header.value"
              type="text"
              class="form-control-plaintext"
              disabled
              readonly
            >
          </div>
        </div>
      </div>
    </div>

    <div class="card shadow-sm mt-3 overflow-hidden">
      <div class="card-header border-bottom">
        <h4 class="m-0">{{ $t('body.label') }}</h4>
      </div>
      <div class="card-body p-0">
        <c-ace-editor
          :value="request.body"
          :border="false"
          lang="json"
          min-height="400px"
          show-line-numbers
          read-only
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { components } from 'corteza-lib/vue/dist'
import { fmt, NoID } from 'corteza-lib/js/dist'

const { CAceEditor } = components
const { t } = useI18n()
const router = useRouter()

const props = defineProps({
  hit: {
    type: Object,
    required: true,
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

const request = computed(() => {
  const { request: req = {}, body = '', route = NoID, time_duration: duration = 0, time_start, time_finish, http_status_code: statusCode } = props.hit || {}
  const { URL = {}, RequestURI, Method, RemoteAddr, Header = {} } = req
  const { Path } = URL
  const headers = Object.entries(Header).map(([key, value = []]) => {
    return { label: key, value: value.join('') }
  })

  let jsonBody = atob(body)
  try {
    jsonBody = JSON.stringify(JSON.parse(jsonBody), null, 2)
  } catch (e) {}

  return {
    routeID: route,
    route: Path,
    url: RequestURI,
    method: Method,
    statusCode,
    remoteAddress: RemoteAddr,
    duration: `${duration.toFixed(2)} ms`,
    start: fmt.fullDateTime(time_start),
    end: fmt.fullDateTime(time_finish),
    headers,
    body: jsonBody,
  }
})

const showOpenRoute = computed(() => {
  return request.value.routeID !== NoID
})

function openRoute () {
  router.push({ name: 'system.apigw.edit', params: { routeID: request.value.routeID } })
}

function getStatusCodeVariant (statusCode = '') {
  const codeVariants = {
    2: 'success',
    3: 'info',
    4: 'danger',
    5: 'warning',
  }
  return codeVariants[statusCode[0]]
}
</script>
