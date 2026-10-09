<template>
  <div class="p-4">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">{{ $t('etl.title') }}</h4>
      <router-link
        :to="{ name: 'admin.etl.create' }"
        class="btn btn-primary btn-sm"
      >
        <font-awesome-icon :icon="['fas', 'plus']" class="me-1" />
        {{ $t('etl.create') }}
      </router-link>
    </div>

    <div v-if="loading" class="text-center py-5">
      <span class="spinner-border" />
    </div>

    <div v-else-if="jobs.length === 0" class="text-center py-5 text-muted">
      {{ $t('etl.noJobs') }}
    </div>

    <div v-else class="table-responsive">
      <table class="table table-striped">
        <thead>
          <tr>
            <th>{{ $t('etl.name') }}</th>
            <th>{{ $t('etl.source') }}</th>
            <th>{{ $t('etl.enabled') }}</th>
            <th>{{ $t('etl.lastRun') }}</th>
            <th>{{ $t('etl.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="job in jobs" :key="job.ETLID">
            <td>{{ job.name || job.ETLID }}</td>
            <td>{{ job.sourceType || '-' }}</td>
            <td>{{ $t(job.enabled ? 'label.yes' : 'label.no') }}</td>
            <td>{{ job.lastRun ? new Date(job.lastRun).toLocaleString() : '-' }}</td>
            <td>
              <div class="d-flex gap-1">
                <router-link
                  :to="{ name: 'admin.etl.edit', params: { etlID: job.ETLID } }"
                  class="btn btn-outline-secondary btn-sm"
                >
                  <font-awesome-icon :icon="['fas', 'pen']" />
                </router-link>
                <button
                  class="btn btn-outline-success btn-sm"
                  :disabled="job.status === 'running'"
                  @click="runJob(job)"
                >
                  <font-awesome-icon :icon="['fas', 'play']" />
                </button>
                <button
                  class="btn btn-outline-danger btn-sm"
                  @click="deleteJob(job)"
                >
                  <font-awesome-icon :icon="['fas', 'trash']" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: { type: Object, required: false, default: undefined },
})

const jobs = ref([])
const loading = ref(false)

onMounted(() => { fetchJobs() })

watch(() => props.namespace?.namespaceID, () => { fetchJobs() })

function fetchJobs() {
  const namespaceID = props.namespace?.namespaceID
  if (!namespaceID) return
  loading.value = true
  $ComposeAPI.etlList({ namespaceID })
    .then(res => {
      if (Array.isArray(res)) jobs.value = res
      else if (res?.set) jobs.value = res.set
      else if (res?.response?.set) jobs.value = res.response.set
      else jobs.value = []
    })
    .catch(() => { jobs.value = [] })
    .finally(() => { loading.value = false })
}

function runJob(job) {
  if (!confirm($t('etl.confirmRun'))) return
  $ComposeAPI.etlRun({ etlID: job.ETLID }).then(() => fetchJobs())
}

function deleteJob(job) {
  if (!confirm($t('etl.confirmDelete'))) return
  $ComposeAPI.etlDelete({ etlID: job.ETLID }).then(() => fetchJobs())
}
</script>
