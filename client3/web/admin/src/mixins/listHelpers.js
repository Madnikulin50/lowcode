import axios from 'axios'
import { useUiStore } from '../store/ui'
import { useRouter, useRoute } from 'vue-router'
import { useNsI18n, useToast } from 'corteza-lib/vue/dist'

export function useListHelpers ({ editRoute, primaryKey } = {}) {
  const ui = useUiStore()
  const router = useRouter()
  const route = useRoute()
  const t = useNsI18n()
  const { toastErrorHandler } = useToast()

  const pagination = {
    limit: 100,
    pageCursor: undefined,
    prevPage: '',
    nextPage: '',
    total: 0,
    page: 1,
    incTotal: true,
  }

  const sorting = {}
  const abortableRequests = []

  function incLoader() {
    ui.incLoader()
  }

  function decLoader() {
    ui.decLoader()
  }

  function handleQueryParams(initial = false, filter, paginationState, sortingState, tempQuery) {
    let {
      limit = paginationState.limit,
      pageCursor = paginationState.pageCursor,
      prevPage = paginationState.prevCursor,
      nextPage = paginationState.nextCursor,
      total = paginationState.total,
      page = paginationState.page,
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

    const refresh = route.query.pageCursor !== paginationState.pageCursor
    Object.assign(paginationState, { limit, pageCursor, prevPage, nextPage, total, page })

    let { sortBy = sortingState.sortBy, sortDesc = sortingState.sortDesc, ...r2 } = r1
    sortDesc = sortDesc === true || sortDesc === 'true'

    if (!initial && (sortBy !== sortingState.sortBy || sortDesc !== sortingState.sortDesc)) {
      paginationState.pageCursor = ''
      paginationState.page = 1
    }
    Object.assign(sortingState, { sortBy, sortDesc })

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

  function filterList(paginationState = pagination) {
    paginationState.pageCursor = ''
    paginationState.page = 1
    abortRequests()
    window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
  }

  function encodeListParams(filter, sortingState, paginationState, tempQuery) {
    const { sortBy, sortDesc } = sortingState
    const { limit, pageCursor, incTotal } = paginationState
    const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined
    return {
      limit,
      sort: pageCursor ? undefined : sort,
      ...filter,
      pageCursor,
      incTotal: incTotal && (!pageCursor || tempQuery.value),
    }
  }

  function encodeRouteParams(paginationState, sortingState, filter) {
    const { limit, pageCursor, page } = paginationState
    return {
      query: {
        limit,
        ...sortingState,
        ...filter,
        page,
        pageCursor,
      },
    }
  }

  function procListResults(p, updateQuery = true, paginationState, sortingState, filter, tempQuery) {
    abortRequests()
    const { response, cancel } = p
    abortableRequests.push(cancel)
    incLoader()

    if (updateQuery && !tempQuery.value) {
      router.replace(encodeRouteParams(paginationState, sortingState, filter))
    }

    return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
      .then(async ([{ set, filter: f }]) => {
        if (f.incTotal) {
          paginationState.total = f.total
        }

        if (tempQuery.value) {
          const query = tempQuery.value
          tempQuery.value = undefined
          router.replace({ query })
          return []
        }

        paginationState.pageCursor = undefined
        paginationState.nextPage = f.nextPage
        paginationState.prevPage = f.prevPage

        decLoader()
        return set
      }).catch(error => {
        if (!axios.isCancel(error)) {
          toastErrorHandler({ title: 'Failed to load list' })(error)
        }
        decLoader()
        return []
      })
  }

  function abortRequests() {
    abortableRequests.forEach(cancel => cancel())
    abortableRequests.length = 0
  }

  function genericRowClass(item) {
    return { 'text-secondary': item && !!item.deletedAt }
  }

  function handleRowClicked (item, routeName = editRoute, key = primaryKey) {
    if (!item || !routeName || !key) return
    router.push({ name: routeName, params: { [key]: item[key] } })
  }

  function handleItemDelete({ resource, resourceName, locale, api = 'system' }, filterState, paginationState) {
    incLoader()
    const { deletedAt = '' } = resource
    const method = deletedAt ? `${resourceName}Undelete` : `${resourceName}Delete`
    const API = api === 'system' ? window.__systemAPI : window.__automationAPI

    API[method](resource)
      .then(() => {
        filterList(paginationState)
      })
      .finally(() => {
        decLoader()
      })
  }

  function areActionsVisible({ resource, conditions = [] }) {
    return conditions.some(c => resource[c])
  }

  function getActionText (r) {
    return r.deletedAt ? t('undelete') : t('delete')
  }

  function getActionIcon(r) {
    return r.deletedAt ? ['fas', 'trash-restore'] : ['far', 'trash-alt']
  }

  return {
    pagination,
    sorting,
    incLoader,
    decLoader,
    handleQueryParams,
    filterList,
    encodeListParams,
    encodeRouteParams,
    procListResults,
    abortRequests,
    genericRowClass,
    handleRowClicked,
    handleItemDelete,
    areActionsVisible,
    getActionText,
    getActionIcon,
  }
}
