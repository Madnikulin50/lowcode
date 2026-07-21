<template>
  <div class="container-fluid d-flex flex-column p-3">
    <Teleport to="#topbar-title-target">{{ t('application.title') }}</Teleport>

    <div class="card shadow-sm mb-4">
      <div class="card-body">
        <div class="mb-0">
          <label class="text-primary form-label">{{ t('application.connection.label') }}</label>
          <c-input-select
            v-model="connectionID"
            :disabled="processing.connections"
            :options="connections"
            :clearable="false"
            :reduce="o => o.connectionID"
            :placeholder="t('application.connection.placeholder')"
            :get-option-label="({ handle, meta }) => meta.name || handle"
            :get-option-key="getOptionKey"
          />
        </div>
      </div>
    </div>

    <div v-if="processing.sensitiveData" class="d-flex align-items-center justify-content-center h-100">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">{{ t('resourceList.loading') }}</span>
      </div>
    </div>

    <h5 v-else-if="!(connectionID && modules[connectionID])" class="text-center mt-5">
      {{ t('application.no-data-available') }}
    </h5>

    <module-records
      v-else
      v-slot="{ value }"
      :modules="modules[connectionID]"
    >
      <p
        v-for="(v, vi) in value.value"
        :key="vi"
        class="mb-0"
        :class="{ 'mt-1': vi > 0 }"
      >
        {{ v }}
      </p>
    </module-records>

    <Teleport to="#editor-toolbar-target">
      <editor-toolbar
        :processing="processing.connections || processing.sensitiveData"
        :back-link="{ name: 'data-overview' }"
      >
        <button
          data-test-id="button-request-deletion"
          :disabled="processing.connections || processing.sensitiveData"
          class="btn btn-outline-secondary btn-lg ms-1"
          @click="router.push({ name: 'request.create', params: { kind: 'delete', connection } })"
        >
          {{ t('application.request-deletion') }}
        </button>

        <button
          data-test-id="button-request-correction"
          :disabled="processing.connections || processing.sensitiveData"
          class="btn btn-primary btn-lg ms-1"
          @click="router.push({ name: 'request.create', params: { kind: 'correct', connection } })"
        >
          {{ t('application.request-correction') }}
        </button>
      </editor-toolbar>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import EditorToolbar from '../../../components/Common/EditorToolbar.vue'
import ModuleRecords from '../../../components/Common/ModuleRecords.vue'

const { t } = useI18n()
const router = useRouter()

const processing = reactive({
  connections: true,
  sensitiveData: true,
})
const connectionID = ref(undefined)
const connections = ref([])
const modules = reactive({})

const connection = computed(() => connections.value.find(({ connectionID: id }) => id === connectionID.value) || {})

watch(connectionID, (val = '') => fetchSensitiveData(val))

function fetchConnections () {
  processing.connections = true
  window.__systemAPI.dataPrivacyConnectionList()
    .then(({ set = [] }) => {
      connections.value = set
      const { connectionID: id } = set[0] || {}
      connectionID.value = id
    })
    .catch((err) => {
      window.__toastErrorHandler(t('notification.connection-load-failed'))(err)
    })
    .finally(() => {
      processing.connections = false
    })
}

function fetchSensitiveData (id) {
  if (id) {
    processing.sensitiveData = true
    window.__composeAPI.dataPrivacyRecordList({ connectionID: [id] })
      .then(({ set = [] }) => {
        if (set.length) {
          modules[id] = set
        }
      })
      .catch((err) => {
        window.__toastErrorHandler(t('notification.sensitive-data-fetch-failed'))(err)
      })
      .finally(() => {
        processing.sensitiveData = false
      })
  }
}

function getOptionKey ({ connectionID: id }) {
  return id
}

onMounted(() => fetchConnections())
</script>