<template>
  <div v-if="isDC !== null" class="container-fluid d-flex flex-column p-3">
    <Teleport to="#topbar-title-target">{{ t('request.list.title') }}</Teleport>

    <c-resource-list
      primary-key="requestID"
      :items="items"
      :fields="fields"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :translations="{
        notFound: t('resourceList.notFound'),
        noItems: t('resourceList.noItems'),
        loading: t('resourceList.loading'),
        searchPlaceholder: t('resourceList.search.placeholder'),
        showingPagination: t('resourceList.pagination.showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: t('resourceList.pagination.prev'),
        nextPagination: t('resourceList.pagination.next'),
        resourceSingle: t('label.privacy_request.single'),
        resourcePlural: t('label.privacy_request.plural'),
      }"
      :is-item-selectable="isItemSelectable"
      selectable
      clickable
      hide-total
      class="flex-grow-1"
      @search="filterList"
      @row-clicked="rowClicked"
    >
      <template #header="{ selected = [] }">
        <template v-if="isDC">
          <c-input-confirm
            :disabled="processing || !selected.length"
            :processing="processingApprove"
            :text="t('request.list.request.approve')"
            variant="primary"
            size="lg"
            size-confirm="lg"
            @confirmed="handleSelectedRequests(selected, 'approved')"
          />

          <c-input-confirm
            :disabled="processing || !selected.length"
            :processing="processingReject"
            :text="t('request.list.request.reject')"
            variant="danger"
            size="lg"
            size-confirm="lg"
            @confirmed="handleSelectedRequests(selected, 'rejected')"
          />
        </template>

        <template v-else>
          <c-input-confirm
            :disabled="processing || !selected.length"
            :processing="processingCancel"
            :text="t('request.list.request.cancel')"
            variant="outline-secondary"
            size="lg"
            size-confirm="lg"
            @confirmed="handleSelectedRequests(selected, 'canceled')"
          />
        </template>
      </template>

      <template #status="{ item }">
        <div class="d-flex align-items-baseline">
          <span
            class="d-inline-block rounded-circle me-1"
            :class="`bg-${statusVariants[item.status]}`"
            style="width: 0.6rem; height: 0.6rem;"
          />
          {{ t(`request:status.${item.status}`) }}
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'request', keyPrefix: 'list' } })
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { components } from 'corteza-lib/vue/dist'
import moment from 'moment'

const { CResourceList } = components

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const processing = ref(false)
const processingApprove = ref(false)
const processingReject = ref(false)
const processingCancel = ref(false)

const isDC = ref(null)
const users = ref({})

const filter = reactive({
  requestedBy: [],
})
const sorting = reactive({
  sortBy: 'requestedAt',
  sortDesc: true,
})
const pagination = reactive({
  limit: 10,
  pageCursor: undefined,
  prevPage: '',
  nextPage: '',
  total: 0,
  page: 1,
})
const abortableRequests = ref([])
const tempQuery = ref(undefined)

const statusVariants = {
  canceled: 'secondary',
  pending: 'warning',
  rejected: 'danger',
  approved: 'success',
}

const fields = computed(() => {
  const base = [
    {
      key: 'kind',
      sortable: true,
      formatter: kind => t(`request:kind.${kind}`),
    },
    {
      key: 'requestedAt',
      sortable: true,
      formatter: requestedAt => moment(requestedAt).fromNow(),
    },
    {
      key: 'requestedBy',
      sortable: false,
      formatter: requestedBy => formatUser(requestedBy),
    },
    {
      key: 'status',
      sortable: true,
    },
  ]

  return base.map(c => ({
    ...c,
    label: c.label || t(`request:list.columns.${c.key}`),
  }))
})

const visibleFields = computed(() => {
  return fields.value.filter(c => {
    if (c.key === 'requestedBy') return isDC.value
    return true
  })
})

watch(() => route.fullPath, () => handleQueryParams())

function handleQueryParams (initial = false) {
  let {
    limit = pagination.limit,
    pageCursor = pagination.pageCursor,
    prevPage = pagination.prevCursor,
    nextPage = pagination.nextCursor,
    total = pagination.total,
    page = pagination.page,
    ...r1
  } = route.query

  limit = parseInt(limit)
  total = parseInt(total)
  page = parseInt(page)

  if (initial && pageCursor) {
    tempQuery.value = route.query
    router.replace({ query: { ...route.query, limit: 1, pageCursor: undefined } })
    return
  }

  const refresh = route.query.pageCursor !== pagination.pageCursor
  Object.assign(pagination, { limit, pageCursor, prevPage, nextPage, total, page })

  let { sortBy = sorting.sortBy, sortDesc = sorting.sortDesc, ...r2 } = r1
  sortDesc = sortDesc === true || sortDesc === 'true'

  if (!initial && (sortBy !== sorting.sortBy || sortDesc !== sorting.sortDesc)) {
    pagination.pageCursor = ''
    pagination.page = 1
  }
  Object.assign(sorting, { sortBy, sortDesc })

  for (const key in r2) {
    if (typeof filter[key] === 'boolean') {
      r2[key] = r2[key] === 'true'
    }
  }

  Object.assign(filter, r2)

  if (refresh) {
    window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
  }
}

function filterList () {
  pagination.pageCursor = ''
  pagination.page = 1
  abortRequests()
  window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
}

function encodeListParams () {
  const { sortBy, sortDesc } = sorting
  const { limit, pageCursor } = pagination
  const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined
  return {
    limit,
    sort: pageCursor ? undefined : sort,
    ...filter,
    pageCursor,
    incTotal: !pageCursor || tempQuery.value,
  }
}

function encodeRouteParams () {
  const { query } = filter
  const { limit, pageCursor, page } = pagination
  return {
    query: {
      limit,
      ...sorting,
      query,
      page,
      pageCursor,
    },
  }
}

function procListResults (p, updateQuery = true) {
  abortRequests()
  const { response, cancel } = p
  abortableRequests.value.push(cancel)

  if (updateQuery && !tempQuery.value) {
    router.replace(encodeRouteParams())
  }

  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(async ([{ set, filter }]) => {
      if (filter.incTotal) {
        pagination.total = filter.total
      }

      if (tempQuery.value) {
        const query = tempQuery.value
        tempQuery.value = undefined
        router.replace({ query })
        return []
      }

      pagination.pageCursor = undefined
      pagination.nextPage = filter.nextPage
      pagination.prevPage = filter.prevPage

      return set
    }).catch(() => {})
}

function abortRequests () {
  abortableRequests.value.forEach((cancel) => cancel())
  abortableRequests.value = []
}

function checkIsDC () {
  window.__systemAPI.roleList({ query: 'data-privacy-officer', memberID: window.__auth.user.userID })
    .then(({ set = [] }) => {
      if (!set.length) {
        filter.requestedBy = [window.__auth.user.userID]
      }
      isDC.value = !!set.length
    })
}

function items () {
  return procListResults(window.__systemAPI.dataPrivacyRequestListCancellable(encodeListParams()))
    .then(async (set) => {
      if (isDC.value && set && set.length) {
        await fetchUsers(set.map(({ requestedBy }) => requestedBy))
      }
      return set
    })
}

function handleSelectedRequests (selected, status) {
  processing.value = true
  if (status === 'approved') {
    processingApprove.value = true
  } else if (status === 'rejected') {
    processingReject.value = true
  } else {
    processingCancel.value = true
  }

  Promise.all(selected.map(requestID => {
    return window.__systemAPI.dataPrivacyRequestUpdateStatus({ requestID, status })
  }))
    .then(() => {
      window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
    })
    .finally(() => {
      processing.value = false
      if (status === 'approved') {
        processingApprove.value = false
      } else if (status === 'rejected') {
        processingReject.value = false
      } else {
        processingCancel.value = false
      }
    })
}

function fetchUsers (userID = []) {
  userID = [...new Set(userID)]
  return window.__systemAPI.userList({ userID, suspended: 1, deleted: 1 })
    .then(({ set }) => {
      set.forEach(user => {
        users.value[user.userID] = user
      })
    })
}

function isItemSelectable (item) {
  return item.status === 'pending'
}

function formatUser (userID) {
  const { name, username, email, handle } = users.value[userID] || {}
  return name || username || email || handle || userID || ''
}

function rowClicked ({ requestID, kind }) {
  router.push({ name: 'request.view', params: { requestID, kind } })
}

function genericRowClass (item) {
  return { 'text-secondary': item && !!item.deletedAt }
}

onMounted(() => {
  checkIsDC()
  handleQueryParams(true)
})
onBeforeUnmount(() => abortRequests())
</script>