import axios from 'axios'
import { useUiStore } from '../store/ui'
import { ref, computed, watch, onBeforeUnmount } from 'vue'

function toastErrorHandler (t) {
  return (e) => {
    console.error(t, e)
  }
}

function toastSuccess (msg) {
  console.log(msg)
}

export function setup ({ store, router, route }) {
  const filter = ref({})
  const pagination = ref({
    limit: 100,
    pageCursor: undefined,
    prevPage: '',
    nextPage: '',
    total: 0,
    page: 1,
  })
  const tempQuery = ref(undefined)
  const sorting = ref({})
  const abortableRequests = ref([])
  const cancelled = ref(false)

  function handleQueryParams (initial = false) {
    let {
      limit = pagination.value.limit,
      pageCursor = pagination.value.pageCursor,
      prevPage = pagination.value.prevCursor,
      nextPage = pagination.value.nextCursor,
      total = pagination.value.total,
      page = pagination.value.page,
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

    const refresh = route.query.pageCursor !== pagination.value.pageCursor
    pagination.value = { limit, pageCursor, prevPage, nextPage, total, page }

    let { sortBy = sorting.value.sortBy, sortDesc = sorting.value.sortDesc, ...r2 } = r1
    sortDesc = sortDesc === true || sortDesc === 'true'

    if (!initial && (sortBy !== sorting.value.sortBy || sortDesc !== sorting.value.sortDesc)) {
      pagination.value.pageCursor = ''
      pagination.value.page = 1
    }
    sorting.value = { sortBy, sortDesc }

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
    pagination.value.pageCursor = ''
    pagination.value.page = 1
    abortRequests()
    window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
  }

  function encodeListParams () {
    let { sortBy, sortDesc } = sorting.value
    const { limit, pageCursor } = pagination.value

    if (sortBy === 'changedAt') {
      sortBy = 'coalesce(deletedAt, updatedAt, createdAt)'
    }

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
    const { limit, pageCursor, page } = pagination.value

    return {
      query: {
        limit,
        ...sorting.value,
        ...filter.value,
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
      .then(async ([{ set, filter: f }]) => {
        if (f.incTotal) {
          pagination.value.total = f.total
        }

        if (tempQuery.value) {
          const query = tempQuery.value
          tempQuery.value = undefined
          router.replace({ query })
          return []
        }

        pagination.value.pageCursor = undefined
        pagination.value.nextPage = f.nextPage
        pagination.value.prevPage = f.prevPage

        return set
      }).catch(error => {
        if (!axios.isCancel(error)) {
          toastErrorHandler(this?.$t?.('notification:list.load.error') || 'List load error')(error)
        } else {
          cancelled.value = true
        }
        return []
      }).finally(() => {
        cancelled.value = false
      })
  }

  function genericRowClass (item) {
    return { 'text-secondary': item && !!item.deletedAt }
  }

  function abortRequests () {
    abortableRequests.value.forEach((cancel) => { cancel() })
    abortableRequests.value = []
  }

  const ui = useUiStore()

  watch(() => route.fullPath, () => handleQueryParams())
  handleQueryParams(true)
  onBeforeUnmount(() => abortRequests())

  return {
    filter,
    pagination,
    sorting,
    tempQuery,
    cancelled,
    abortableRequests,
    incLoader: ui.incLoader,
    decLoader: ui.decLoader,
    handleQueryParams,
    filterList,
    encodeListParams,
    encodeRouteParams,
    procListResults,
    genericRowClass,
    abortRequests,
    toastErrorHandler,
    toastSuccess,
  }
}

export default { setup }
