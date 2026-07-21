<template>
  <div v-if="reports.length" class="h-100">
        <div class="bg-white sticky-top py-2">
          <button
            data-test-id="button-report-list"
            class="btn btn-outline-secondary w-100 mb-2"
            @click="$router.push({ name: 'report.list' })"
          >
            {{ t('report-list') }}
          </button>
          <input
            v-model.trim="query"
            class="form-control"
            :placeholder="t('search-reports')"
          />
        </div>
        <c-sidebar-nav-items
          :items="filteredReports"
          :start-expanded="!!query"
          default-route-name="report.view"
          class="overflow-auto h-100"
        />
      </div>
      <h5 v-else class="d-flex justify-content-center mt-5">
        {{ t('no-reports') }}
      </h5>
</template>
<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import { components } from 'corteza-lib/vue/dist'
const { CSidebarNavItems } = components

const { t } = useI18n()
const route = useRoute()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const query = ref('')
const reports = ref([])

const filteredReports = computed(() => {
  let r = reports.value
  if (query.value) {
    r = reports.value.filter(({ reportID, handle, meta: { name = '' } }) => {
      const reportString = `${reportID}${handle}${name}`.toLowerCase().trim()
      return reportString.indexOf(query.value.toLowerCase().trim()) > -1
    })
  }
  return r.map(({ reportID, handle, meta: { name = '' } }) => ({
    page: { pageID: reportID, name: route.name, title: name || handle },
    params: { reportID },
  }))
})

function fetchReports() {
  window.__systemAPI.reportList()
    .then(res => { reports.value = (res || {}).set || [] })
    .catch(toastErrorHandler(t('notification.report.listFetchFailed')))
}

watch(() => route.name, (name) => {
  if (!['report.create', 'report.edit'].includes(name)) {
    fetchReports()
  }
}, { immediate: true })

function onRefetch() { fetchReports() }
onMounted(() => { window.addEventListener('refetch.reports', onRefetch) })
onBeforeUnmount(() => { window.removeEventListener('refetch.reports', onRefetch) })
</script>
<style scoped lang="scss">
.pointer-none { pointer-events: none; }
.nav-active { color: var(--primary); -webkit-text-stroke: 0.4px var(--primary); }
</style>