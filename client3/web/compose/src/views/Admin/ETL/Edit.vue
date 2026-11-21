<template>
  <div class="p-4">
    <div class="d-flex align-items-center gap-2 mb-4">
      <router-link :to="{ name: 'admin.etl' }" class="btn btn-outline-secondary btn-sm">
        <font-awesome-icon :icon="['fas', 'arrow-left']" />
      </router-link>
      <h4 class="mb-0">{{ isNew ? $t('etl.create') : $t('etl.edit') }}</h4>
    </div>

    <div v-if="loading" class="text-center py-5">
      <span class="spinner-border" />
    </div>

    <div v-else class="row">
      <div class="col-md-8">
        <div class="mb-3">
          <label class="form-label">{{ $t('etl.name') }}</label>
          <input v-model="job.name" class="form-control" :placeholder="$t('etl.name')" />
        </div>

        <div class="mb-3">
          <label class="form-label">{{ $t('etl.source') }}</label>
          <select v-model="job.sourceType" class="form-select">
            <option value="rest">{{ $t('etl.sources.rest') }}</option>
            <option value="mcp">{{ $t('etl.sources.mcp') }}</option>
            <option value="smb">{{ $t('etl.sources.smb') }}</option>
          </select>
        </div>

        <div class="mb-3">
          <label class="form-label">{{ $t('etl.sourceConfig') }}</label>
          <textarea v-model="job.sourceConfig" class="form-control" rows="4" />
        </div>

        <div class="mb-3">
          <label class="form-label">{{ $t('etl.schedule') }}</label>
          <input v-model="job.schedule" class="form-control" :placeholder="$t('etl.schedule')" />
        </div>

        <div class="form-check form-switch mb-3">
          <input id="etlEnabled" v-model="job.enabled" class="form-check-input" type="checkbox" />
          <label class="form-check-label" for="etlEnabled">{{ $t('etl.enabled') }}</label>
        </div>

        <div class="d-flex gap-2">
          <button class="btn btn-primary" @click="save">{{ $t('label.save') }}</button>
          <router-link :to="{ name: 'admin.etl' }" class="btn btn-outline-secondary">
            {{ $t('label.cancel') }}
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI
const route = useRoute()

const props = defineProps({
  namespace: { type: Object, required: false, default: undefined },
})

const etlID = computed(() => route.params.etlID)
const isNew = computed(() => !etlID.value)

const job = ref({ name: '', sourceType: 'rest', sourceConfig: '', schedule: '', enabled: true })
const loading = ref(false)

onMounted(() => {
  if (!isNew.value) fetchJob()
})

watch(() => props.namespace?.namespaceID, () => {
  if (!isNew.value) fetchJob()
})

function fetchJob() {
  loading.value = true
  $ComposeAPI.etlRead({ etlID: etlID.value })
    .then(res => {
      const data = res?.response || res
      if (data) Object.assign(job.value, data)
    })
    .finally(() => { loading.value = false })
}

function save() {
  const namespaceID = props.namespace?.namespaceID
  const payload = { ...job.value, namespaceID }
  if (isNew.value) {
    $ComposeAPI.etlCreate(payload)
      .then(() => window.history.back())
  } else {
    $ComposeAPI.etlUpdate({ etlID: etlID.value, ...payload })
      .then(() => window.history.back())
  }
}
</script>
