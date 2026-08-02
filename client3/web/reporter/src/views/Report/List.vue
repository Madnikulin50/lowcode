<template>
  <div class="container-fluid d-flex flex-column py-3">
    <Teleport to="#topbar-title">{{ t('report.list') }}</Teleport>

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="reportList"
      :translations="{
        searchPlaceholder: t('searchPlaceholder'),
        notFound: t('resourceList.notFound'),
        noItems: t('resourceList.noItems'),
        loading: t('label.loading'),
        showingPagination: t('resourceList.pagination.showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: t('resourceList.pagination.prev'),
        nextPagination: t('resourceList.pagination.next'),
        resourceSingle: t('label.report.single'),
        resourcePlural: t('label.report.plural'),
      }"
      sticky-header
      clickable
      class="h-100 flex-fill"
      @search="filterList"
      @row-clicked="viewReport"
    >
      <template #header>
        <button
          v-if="canCreate"
          class="btn btn-primary btn-lg"
          data-test-id="button-create-report"
          @click="$router.push({ name: 'report.create' })"
        >
          {{ t('report.new') }}
        </button>

        <c-permissions-button
          v-if="canGrant"
          resource="corteza::system:report/*"
          :button-label="t('permissions')"
          size="lg"
        />
      </template>

      <template #name="{ item: r }">{{ r.meta.name }}</template>

      <template #changedAt="{ item }">{{ $locFullDateTime(item.deletedAt || item.updatedAt || item.createdAt) }}</template>

      <template #actions="{ item: r }">
        <div v-if="r.canUpdateReport" class="btn-group btn-group-sm">
          <button
            class="btn btn-primary btn-sm d-flex align-items-center"
            data-test-id="button-report-builder"
            @click="$router.push({ name: 'report.builder', params: { reportID: r.reportID } })"
          >
            {{ t('report.builder') }}
            <font-awesome-icon :icon="['fas', 'tools']" class="ms-2 ml-1" />
          </button>

          <button
            class="btn btn-primary d-flex align-items-center"
            style="margin-left:2px"
            :title="t('report.edit')"
            data-test-id="button-report-edit"
            @click="$router.push({ name: 'report.edit', params: { reportID: r.reportID } })"
          >
            <font-awesome-icon :icon="['far', 'edit']" />
          </button>
        </div>

        <div class="dropdown d-inline-block ms-2 ml-1">
          <button
            v-if="r.canUpdateReport || r.canGrant || r.canDeleteReport"
            class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2"
            data-bs-toggle="dropdown"
          >
            <font-awesome-icon :icon="['fas', 'ellipsis-v']" />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <button class="dropdown-item" @click="handleReportCloning(r)">
                <font-awesome-icon :icon="['fa', 'clone']" /> {{ t('resourceList.clone') }}
              </button>
            </li>
            <li v-if="r.canGrant">
              <c-permissions-button
                :tooltip="t('permissions.resources.system.report.tooltip')"
                :title="r.meta.name || r.handle || r.reportID"
                :target="r.meta.name || r.handle || r.reportID"
                :resource="`corteza::system:report/${r.reportID}`"
                :button-label="t('permissions.ui.label')"
                class="dropdown-item"
              />
            </li>
            <li v-if="r.canDeleteReport">
              <c-input-confirm
                :processing="processingDelete"
                :text="t('report.delete')"
                borderless
                variant="link"
                size="md"
                show-icon
                text-class="p-1"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(r)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, inject } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useToast, components } from 'corteza-lib/vue/dist'
import { useListHelpers } from '../../mixins/listHelpers'
import { useReportHelpers } from '../../mixins/report'

const { CResourceList } = components
const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const toast = useToast()
const toastSuccess = toast.success
const toastErrorHandler = toast.toastErrorHandler
const can = inject('can', () => false)
const { handleClone } = useReportHelpers()
const { handleQueryParams, encodeListParams } = useListHelpers(route, router)

import { system } from 'corteza-lib/js/dist'

const primaryKey = 'reportID'
const processingDelete = ref(false)
const filter = ref({ query: '' })
const sorting = ref({ sortBy: 'handle', sortDesc: false })
const pagination = ref({
  limit: 100,
  pageCursor: undefined,
  prevPage: '',
  nextPage: '',
  total: 0,
  page: 1,
})
const abortableRequests = ref([])
const tempQuery = ref(undefined)

const canGrant = computed(() => can('system/', 'grant'))
const canCreate = computed(() => can('system/', 'report.create'))

const tableFields = computed(() => [
  { key: 'name', label: t('columns.name'), sortable: false, tdClass: 'text-nowrap' },
  { key: 'handle', label: t('columns.handle'), sortable: true },
  { key: 'changedAt', label: t('columns.changedAt'), sortable: true, class: 'text-right text-nowrap' },
  { key: 'actions', label: '', tdClass: 'text-right text-nowrap actions gap-1' },
])

function viewReport({ reportID, canReadReport = false }) {
  if (reportID) {
    router.push({ name: 'report.view', params: { reportID } })
  }
}

async function reportList() {
  abortRequests()
  const params = encodeListParams(sorting.value, pagination.value, filter.value)
  const { response, cancel } = window.__systemAPI.reportListCancellable(params)
  abortableRequests.value.push(cancel)

  if (!tempQuery.value) {
    router.replace({ query: { ...sorting.value, ...filter.value, ...pagination.value } })
  }

  try {
    const [set, filterResult] = await Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    if (filterResult.incTotal) {
      pagination.value.total = filterResult.total
    }
    if (tempQuery.value) {
      const query = tempQuery.value
      tempQuery.value = undefined
      router.replace({ query })
      return []
    }
    pagination.value.pageCursor = undefined
    pagination.value.nextPage = filterResult.nextPage
    pagination.value.prevPage = filterResult.prevPage
    return set
  } catch (error) {
    toastErrorHandler(t('notification.list.load.error'))
    return []
  }
}

function filterList() {
  pagination.value.pageCursor = ''
  pagination.value.page = 1
  abortRequests()
}

function abortRequests() {
  abortableRequests.value.forEach(cancel => cancel())
  abortableRequests.value = []
}

async function handleDelete(reportItem) {
  processingDelete.value = true
  try {
    await window.__systemAPI.reportDelete(reportItem)
    toastSuccess(t('notification.report.delete'))
    filterList()
  } catch (e) {
    toastErrorHandler(t('notification.report.deleteFailed'))(e)
  } finally {
    processingDelete.value = false
  }
}

async function handleReportCloning(reportItem) {
  try {
    const { reportID } = await handleClone(reportItem, toastSuccess, toastErrorHandler, t)
    if (reportID) viewReport({ reportID })
  } catch (e) {}
}

onMounted(() => {
  const query = route.query
  if (query.pageCursor) {
    tempQuery.value = query
    router.replace({ query: { ...query, limit: 1, pageCursor: undefined } })
  }
})

onBeforeUnmount(() => {
  abortRequests()
})
</script>
