<template>
  <div v-if="namespace && module" class="p-3">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <div>
        <h5 class="mb-1">{{ $t('etl.settings', { default: 'ETL settings' }) }}</h5>
        <span class="text-muted small">
          {{ $t('etl.settingsDescription', { default: 'Synchronize records from external sources into this module.' }) }}
        </span>
      </div>
      <button class="btn btn-primary btn-sm" @click="openCreate">
        <font-awesome-icon :icon="['fas', 'plus']" class="me-1" />
        {{ $t('etl.create', { default: 'New job' }) }}
      </button>
    </div>

    <div v-if="loading" class="text-center py-4">
      <span class="spinner-border spinner-border-sm" />
    </div>

    <div v-else-if="jobs.length === 0 && !editing" class="text-center py-4 text-muted">
      {{ $t('etl.noJobs', { default: 'No ETL jobs for this module yet.' }) }}
    </div>

    <table v-if="jobs.length > 0" class="table table-sm table-striped align-middle">
      <thead>
        <tr>
          <th>{{ $t('etl.name', { default: 'Name' }) }}</th>
          <th>{{ $t('etl.source', { default: 'Source' }) }}</th>
          <th>{{ $t('etl.enabled', { default: 'Enabled' }) }}</th>
          <th>{{ $t('etl.lastRun', { default: 'Last run' }) }}</th>
          <th class="text-end">{{ $t('etl.actions', { default: 'Actions' }) }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="job in jobs" :key="job.etlJobID">
          <td>{{ job.name || job.etlJobID }}</td>
          <td>{{ job.source?.type || '-' }}</td>
          <td>{{ $t(job.enabled ? 'label.yes' : 'label.no') }}</td>
          <td>
            <template v-if="job.lastRunAt">
              {{ new Date(job.lastRunAt).toLocaleString() }}
              <span v-if="job.lastStatus" class="badge ms-1" :class="job.lastStatus === 'ok' ? 'bg-success' : 'bg-danger'">
                {{ job.lastStatus }}
              </span>
            </template>
            <span v-else>-</span>
          </td>
          <td class="text-end">
            <div class="d-inline-flex gap-1">
              <button class="btn btn-outline-secondary btn-sm" title="Edit" @click="openEdit(job)">
                <font-awesome-icon :icon="['fas', 'pen']" />
              </button>
              <button class="btn btn-outline-success btn-sm" title="Run now" @click="runJob(job)">
                <font-awesome-icon :icon="['fas', 'play']" />
              </button>
              <button class="btn btn-outline-danger btn-sm" title="Delete" @click="deleteJob(job)">
                <font-awesome-icon :icon="['fas', 'trash']" />
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <form v-if="editing" class="border rounded p-3 mt-2 bg-light" @submit.prevent="save">
      <h6 class="mb-3">
        {{ isNew ? $t('etl.create', { default: 'New job' }) : $t('etl.edit', { default: 'Edit job' }) }}
      </h6>

      <div class="row g-3">
        <div class="col-md-6">
          <label class="form-label">{{ $t('etl.name', { default: 'Name' }) }}</label>
          <input v-model="form.name" class="form-control form-control-sm" required />
        </div>
        <div class="col-md-6">
          <label class="form-label">{{ $t('etl.schedule', { default: 'Schedule (cron)' }) }}</label>
          <input v-model="form.schedule" class="form-control form-control-sm" placeholder="0 * * * *" />
        </div>

        <div class="col-md-4">
          <label class="form-label">{{ $t('etl.source', { default: 'Source' }) }}</label>
          <select v-model="form.source.type" class="form-select form-select-sm">
            <option value="rest">REST</option>
            <option value="mcp">MCP</option>
            <option value="smb">SMB</option>
          </select>
        </div>
        <div class="col-md-4">
          <label class="form-label">{{ $t('etl.format', { default: 'Format' }) }}</label>
          <select v-model="form.source.format" class="form-select form-select-sm">
            <option value="json">JSON</option>
            <option value="csv">CSV</option>
            <option value="xml">XML</option>
          </select>
        </div>
        <div class="col-md-4 d-flex align-items-end">
          <div class="form-check form-switch">
            <input id="etlSettingsEnabled" v-model="form.enabled" class="form-check-input" type="checkbox" />
            <label class="form-check-label" for="etlSettingsEnabled">{{ $t('etl.enabled', { default: 'Enabled' }) }}</label>
          </div>
        </div>

        <template v-if="form.source.type === 'rest'">
          <div class="col-md-6">
            <label class="form-label">{{ $t('etl.rest.url', { default: 'URL' }) }}</label>
            <input v-model="form.source.restUrl" class="form-control form-control-sm" />
          </div>
          <div class="col-md-6">
            <label class="form-label">{{ $t('etl.rest.method', { default: 'Method' }) }}</label>
            <select v-model="form.source.restMethod" class="form-select form-select-sm">
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
            </select>
          </div>
          <div class="col-12">
            <label class="form-label">{{ $t('etl.rest.body', { default: 'Body (JSON)' }) }}</label>
            <textarea v-model="form.source.restBody" class="form-control form-control-sm" rows="2" />
          </div>
        </template>

        <template v-else-if="form.source.type === 'mcp'">
          <div class="col-md-6">
            <label class="form-label">{{ $t('etl.mcp.server', { default: 'Server' }) }}</label>
            <input v-model="form.source.mcpServerId" class="form-control form-control-sm" />
          </div>
          <div class="col-md-6">
            <label class="form-label">{{ $t('etl.mcp.tool', { default: 'Tool' }) }}</label>
            <input v-model="form.source.mcpTool" class="form-control form-control-sm" />
          </div>
        </template>

        <template v-else-if="form.source.type === 'smb'">
          <div class="col-md-4">
            <label class="form-label">{{ $t('etl.smb.host', { default: 'Host' }) }}</label>
            <input v-model="form.source.smbHost" class="form-control form-control-sm" />
          </div>
          <div class="col-md-4">
            <label class="form-label">{{ $t('etl.smb.share', { default: 'Share' }) }}</label>
            <input v-model="form.source.smbShare" class="form-control form-control-sm" />
          </div>
          <div class="col-md-4">
            <label class="form-label">{{ $t('etl.smb.filePath', { default: 'Path' }) }}</label>
            <input v-model="form.source.smbPath" class="form-control form-control-sm" />
          </div>
          <div class="col-md-6">
            <label class="form-label">{{ $t('etl.smb.username', { default: 'Username' }) }}</label>
            <input v-model="form.source.smbUser" class="form-control form-control-sm" />
          </div>
          <div class="col-md-6">
            <label class="form-label">{{ $t('etl.smb.password', { default: 'Password' }) }}</label>
            <input v-model="form.source.smbPass" class="form-control form-control-sm" type="password" />
          </div>
        </template>
      </div>

      <div class="d-flex gap-2 mt-3">
        <button class="btn btn-primary btn-sm" :disabled="saving">
          <span v-if="saving" class="spinner-border spinner-border-sm me-1" />
          {{ $t('label.save') }}
        </button>
        <button class="btn btn-outline-secondary btn-sm" type="button" @click="cancel">
          {{ $t('label.cancel') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'module' } })
import { ref, onMounted, watch } from 'vue'
import { useNsI18n } from 'corteza-lib/vue/dist'

const $t = useNsI18n()
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
})

const jobs = ref([])
const loading = ref(false)
const saving = ref(false)
const editing = ref(false)
const isNew = ref(true)
const etlJobID = ref(null)

const emptyForm = () => ({
  name: '',
  enabled: true,
  schedule: '',
  source: {
    type: 'rest',
    format: 'json',
    restUrl: '',
    restMethod: 'GET',
    restBody: '',
    mcpServerId: '',
    mcpTool: '',
    smbHost: '',
    smbShare: '',
    smbPath: '',
    smbUser: '',
    smbPass: '',
  },
})

const form = ref(emptyForm())

onMounted(() => { fetchJobs() })

watch(() => props.module?.moduleID, () => { fetchJobs() })

function fetchJobs () {
  const namespaceID = props.namespace?.namespaceID
  const moduleID = props.module?.moduleID
  if (!namespaceID || !moduleID) return
  loading.value = true
  $ComposeAPI.etlList({ namespaceID, moduleID })
    .then(res => {
      if (Array.isArray(res)) jobs.value = res
      else if (res?.set) jobs.value = res.set
      else if (res?.response?.set) jobs.value = res.response.set
      else jobs.value = []
    })
    .catch(() => { jobs.value = [] })
    .finally(() => { loading.value = false })
}

function openCreate () {
  isNew.value = true
  etlJobID.value = null
  form.value = emptyForm()
  editing.value = true
}

function openEdit (job) {
  isNew.value = false
  etlJobID.value = job.etlJobID
  const src = job.source || {}
  form.value = {
    name: job.name || '',
    enabled: !!job.enabled,
    schedule: job.schedule || '',
    source: {
      type: src.type || 'rest',
      format: src.format || 'json',
      restUrl: src.restUrl || '',
      restMethod: src.restMethod || 'GET',
      restBody: src.restBody || '',
      mcpServerId: src.mcpServerId || '',
      mcpTool: src.mcpTool || '',
      smbHost: src.smbHost || '',
      smbShare: src.smbShare || '',
      smbPath: src.smbPath || '',
      smbUser: src.smbUser || '',
      smbPass: src.smbPass || '',
    },
  }
  editing.value = true
}

function cancel () {
  editing.value = false
  isNew.value = true
  etlJobID.value = null
}

function save () {
  saving.value = true
  const namespaceID = props.namespace.namespaceID
  const moduleID = props.module.moduleID
  const payload = { ...form.value, namespaceID, moduleID }
  const done = () => {
    saving.value = false
    editing.value = false
    isNew.value = true
    etlJobID.value = null
    fetchJobs()
  }
  const fail = () => { saving.value = false }

  if (isNew.value) {
    $ComposeAPI.etlCreate(payload).then(done).catch(fail)
  } else {
    $ComposeAPI.etlUpdate({ namespaceID, etlID: etlJobID.value, ...payload }).then(done).catch(fail)
  }
}

function runJob (job) {
  $ComposeAPI.etlRun({ namespaceID: props.namespace.namespaceID, etlID: job.etlJobID }).then(() => fetchJobs())
}

function deleteJob (job) {
  if (!confirm($t('etl.confirmDelete', { default: 'Delete this ETL job?' }))) return
  $ComposeAPI.etlDelete({ namespaceID: props.namespace.namespaceID, etlID: job.etlJobID }).then(() => fetchJobs())
}
</script>
