<template>
  <div class="d-flex flex-column h-100">
    <div class="card shadow-sm mb-3">
      <div class="card-body">
        <div class="mb-0">
          <label class="text-primary form-label">{{ t('request.edit.delete.connection.label') }}</label>
          <c-input-select
            v-model="connection"
            :disabled="processing.connections"
            :options="connections"
            :clearable="false"
            :get-option-label="({ handle, meta }) => meta.name || handle"
            :get-option-key="getOptionKey"
            :placeholder="t('request.edit.delete.connection.placeholder')"
          />
        </div>
      </div>
    </div>

    <div v-if="processing.sensitiveData" class="d-flex align-items-center justify-content-center h-100">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">{{ t('resourceList.loading') }}</span>
      </div>
    </div>

    <h5 v-else-if="!(connection && modules[connection.connectionID])" class="text-center mt-5">
      {{ t('request.edit.delete.no-data-available') }}
    </h5>

    <module-records
      v-else
      v-slot="{ namespace, module, record, value }"
      :modules="modules[connection.connectionID]"
    >
      <div v-if="value.value.length" class="form-check">
        <input
          :id="`del-${record.recordID}-${value.name}`"
          type="checkbox"
          class="form-check-input-v3"
          :checked="value.selected"
          @change="updateValue({ namespace, module, recordID: record.recordID, field: value.name, value: $event.target.checked, orgValue: value.value })"
        />
        <label class="form-check-label" :for="`del-${record.recordID}-${value.name}`">
          <span
            v-for="(v, vi) in value.value"
            :key="vi"
            class="py-2 mb-0"
          >
            <del v-if="value.selected">{{ v }}</del>
            <span v-else>{{ v }}</span>
          </span>
        </label>
      </div>
    </module-records>

    <Teleport to="#editor-toolbar-target">
      <editor-toolbar
        :processing="processing.connections || processing.sensitiveData"
        :back-link="{ name: 'data-overview.application' }"
        submit-show
        :submit-label="t('request.edit.delete.submit')"
        :submit-disabled="!valid || !connection"
        @submit="emit('submit', { kind: 'delete', payload })"
      />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import EditorToolbar from '../../Common/EditorToolbar.vue'
import ModuleRecords from '../../Common/ModuleRecords.vue'

const { t } = useI18n()
const route = useRoute()
const emit = defineEmits(['submit'])

const processing = reactive({
  connections: true,
  sensitiveData: true,
})
const connection = ref(undefined)
const connections = ref([])
const modules = reactive({})
const payload = ref({})
const valid = ref(false)

watch(connection, (val = {}) => fetchSensitiveData(val.connectionID))

function fetchConnections () {
  processing.connections = true
  window.__systemAPI.dataPrivacyConnectionList()
    .then(({ set = [] }) => {
      connections.value = set
      if (!route.params.connection) {
        connection.value = set[0]
      } else {
        connection.value = route.params.connection
      }
    })
    .catch((err) => { window.__toastErrorHandler(t('notification.connection-load-failed'))(err) })
    .finally(() => { processing.connections = false })
}

function fetchSensitiveData (connectionID) {
  if (connectionID) {
    processing.sensitiveData = true
    window.__composeAPI.dataPrivacyRecordList({ connectionID: [connectionID] })
      .then(({ set = [] }) => {
        if (set.length) {
          modules[connectionID] = set
        }
        payload.value = {
          connectionID,
          modules: {},
        }
      })
      .catch((err) => { window.__toastErrorHandler(t('notification.sensitive-data-fetch-failed'))(err) })
      .finally(() => { processing.sensitiveData = false })
  }
}

function updateValue ({ namespace, module, recordID, field, value, orgValue = [] }) {
  const { connectionID: cid } = connection.value
  const { namespaceID, name: namespaceName } = namespace
  const { moduleID, name: moduleName } = module

  if (!value) {
    delete payload.value.modules[moduleID].records[recordID].values[field]
    if (Object.keys(payload.value.modules[moduleID].records[recordID].values).length === 0) {
      delete payload.value.modules[moduleID].records[recordID]
      if (Object.keys(payload.value.modules[moduleID].records).length === 0) {
        delete payload.value.modules[moduleID]
      }
    }
    valid.value = Object.keys(payload.value.modules).length > 0
    return
  }

  if (cid) {
    if (!payload.value.modules[moduleID]) {
      payload.value.modules[moduleID] = {
        namespace: namespaceName,
        namespaceID,
        module: moduleName,
        moduleID,
        records: {},
      }
    }

    if (!payload.value.modules[moduleID].records[recordID]) {
      payload.value.modules[moduleID].records[recordID] = { values: {} }
    }

    payload.value.modules[moduleID].records[recordID].values[field] = orgValue
    valid.value = true
  }
}

function getOptionKey ({ connectionID }) {
  return connectionID
}

onMounted(() => fetchConnections())
</script>
