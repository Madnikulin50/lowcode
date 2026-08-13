<template>
  <div class="container-fluid d-flex flex-column p-3">
    <Teleport to="#topbar-title-target">{{ t('title') }}</Teleport>

    <c-resource-list
      primary-key="moduleID"
      :items="items"
      :fields="fields"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :selectable="false"
      :translations="{
        noItems: t('resourceList.noItems'),
        loading: t('resourceList.loading'),
        searchPlaceholder: t('resourceList.search.placeholder'),
        showingPagination: t('resourceList.pagination.showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: t('resourceList.pagination.prev'),
        nextPagination: t('resourceList.pagination.next'),
        resourceSingle: t('label.sensitive_module.single'),
        resourcePlural: t('label.sensitive_module.plural'),
      }"
      clickable
      hide-total
      class="flex-grow-1"
      @search="filterList"
      @row-clicked="rowClicked"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'sensitive-data', keyPrefix: 'list' } })
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'

import { useRoute, useRouter } from 'vue-router'
import { components, useNsI18n } from 'corteza-lib/vue/dist'

const { CResourceList } = components
const t = useNsI18n()
const route = useRoute()
const router = useRouter()

const filter = ref({})
const pagination = reactive({ limit: 10, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1 })
const sorting = reactive({ sortBy: '', sortDesc: false })
const abortableRequests = ref([])
const tempQuery = ref(undefined)

const fields = computed(() => {
  return [
    { key: 'module', formatter: module => module ? module.name : '' },
    { key: 'namespace', formatter: namespace => namespace ? namespace.name : '' },
    {
      key: 'connection',
      formatter: connection => {
        const { name } = connection.meta || {}
        return name
      },
    },
    {
      key: 'location',
      formatter: (value, key, item) => {
        const { meta = {} } = item.connection || {}
        const { properties = {} } = meta.location || {}
        return properties.name
      },
    },
    {
      key: 'ownership',
      formatter: (value, key, item) => {
        const { meta = {} } = item.connection || {}
        const { ownership } = meta || {}
        return ownership
      },
    },
  ].map(c => ({
    ...c,
    label: c.label || t(`list.columns.${c.key}`),
  }))
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
    if (typeof filter.value[key] === 'boolean') {
      r2[key] = r2[key] === 'true'
    }
  }

  filter.value = { ...filter.value, ...r2 }

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
    ...filter.value,
    pageCursor,
    incTotal: !pageCursor || tempQuery.value,
  }
}

function encodeRouteParams () {
  const { limit, pageCursor, page } = pagination
  return { query: { limit, ...sorting, ...filter.value, page, pageCursor } }
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
  abortableRequests.value.forEach(cancel => cancel())
  abortableRequests.value = []
}

function items () {
  return procListResults(window.__composeAPI.dataPrivacyModuleListCancellable(encodeListParams()))
}

function rowClicked ({ namespace, module }) {
  const { namespaceID, slug } = namespace
  window.open(`/compose/ns/${slug || namespaceID}/admin/modules/${module.moduleID}/edit`, '_blank')
}

function genericRowClass (item) {
  return { 'text-secondary': item && !!item.deletedAt }
}

onMounted(() => handleQueryParams(true))
onBeforeUnmount(() => abortRequests())
</script>