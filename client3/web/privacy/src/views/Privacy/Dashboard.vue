<template>
  <div v-if="isDC !== null" class="container-fluid d-flex flex-column p-3">
    <Teleport to="#topbar-title-target">{{ t('dashboard.title') }}</Teleport>

    <div class="flex-shrink-1">
      <p>
        {{ t('dashboard.description.first') }}<br>
        {{ t('dashboard.description.second') }}
      </p>

      <div class="row">
        <div
          v-for="option in options"
          :key="option.title"
          class="col-12 col-md-6 col-xl-3 mb-3"
        >
          <div class="card card-hover-popup shadow-sm h-100">
            <div class="card-body d-flex flex-column">
              <h5 class="card-title">{{ option.title }}</h5>
              <p class="card-text flex-grow-1">{{ option.description }}</p>
              <router-link
                :to="option.button.to"
                class="btn"
                :class="option.button.variant ? `btn-${option.button.variant}` : 'btn-light'"
              >
                {{ option.button.label }}
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="d-flex flex-column h-100">
      <h6 class="text-primary">{{ t('dashboard.connection-location') }}</h6>

      <connection-map
        :connections="connections"
        style="min-height: 400px;"
        class="rounded-3 shadow-lg"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'dashboard' } })
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import ConnectionMap from '../../components/ConnectionMap.vue'

const { t } = useI18n()

const processing = ref(false)
const isDC = ref(null)
const connections = ref([])

const userOptions = [
  {
    title: t('dashboard.user-options.data-overview.title'),
    description: t('dashboard.user-options.data-overview.description'),
    button: { label: t('dashboard.user-options.data-overview.button-label'), to: { name: 'data-overview' } },
  },
  {
    title: t('dashboard.user-options.privacy-requests.title'),
    description: t('dashboard.user-options.privacy-requests.description'),
    button: { label: t('dashboard.user-options.privacy-requests.button-label'), to: { name: 'request.list' } },
  },
  {
    title: t('dashboard.user-options.export.title'),
    description: t('dashboard.user-options.export.description'),
    button: { label: t('dashboard.user-options.export.button-label'), to: { name: 'request.create', params: { kind: 'export' } } },
  },
  {
    title: t('dashboard.user-options.delete.title'),
    description: t('dashboard.user-options.delete.description'),
    button: { label: t('dashboard.user-options.delete.button-label'), variant: 'danger', to: { name: 'request.create', params: { kind: 'delete' } } },
  },
]

const dcOptions = [
  {
    title: t('dashboard.dc-options.privacy-requests.title'),
    description: t('dashboard.dc-options.privacy-requests.description'),
    button: { label: t('dashboard.dc-options.privacy-requests.button-label'), to: { name: 'request.list' } },
  },
  {
    title: t('dashboard.dc-options.sensitive-data.title'),
    description: t('dashboard.dc-options.sensitive-data.description'),
    button: { label: t('dashboard.dc-options.sensitive-data.button-label'), to: { name: 'sensitive-data' } },
  },
]

const options = computed(() => isDC.value ? dcOptions : userOptions)

function fetchConnections () {
  processing.value = true
  window.__systemAPI.dataPrivacyConnectionList()
    .then(({ set = [] }) => {
      connections.value = set
    })
    .catch((err) => {
      window.__toastErrorHandler(t('notification.connection-load-failed'))(err)
    })
    .finally(() => {
      processing.value = false
    })
}

function checkIsDC () {
  window.__systemAPI.roleList({ query: 'data-privacy-officer', memberID: window.__auth.user.userID })
    .then(({ set = [] }) => {
      isDC.value = !!set.length
    })
}

onMounted(() => {
  fetchConnections()
  checkIsDC()
})
</script>