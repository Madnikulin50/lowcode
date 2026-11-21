export function useListHelpers(route, router) {
  function handleQueryParams(filter, sorting, pagination, initial = false) {
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

    let tempQuery
    if (initial && pageCursor) {
      tempQuery = route.query
      router.replace({ query: { ...route.query, limit: 1, pageCursor: undefined } })
      return { filter, sorting, pagination, tempQuery }
    }

    const refresh = route.query.pageCursor !== pagination.pageCursor
    pagination = { limit, pageCursor, prevPage, nextPage, total, page }

    let { sortBy = sorting.sortBy, sortDesc = sorting.sortDesc, ...r2 } = r1
    sortDesc = sortDesc === true || sortDesc === 'true'

    if (!initial && (sortBy !== sorting.sortBy || sortDesc !== sorting.sortDesc)) {
      pagination.pageCursor = ''
      pagination.page = 1
    }
    sorting = { sortBy, sortDesc }

    for (const key in r2) {
      if (typeof filter[key] === 'boolean') {
        r2[key] = r2[key] === 'true'
      }
    }
    filter = { ...filter, ...r2 }

    return { filter, sorting, pagination, refresh, tempQuery }
  }

  function encodeListParams(sorting, pagination, filter) {
    let { sortBy, sortDesc } = sorting
    const { limit, pageCursor } = pagination

    if (sortBy === 'changedAt') {
      sortBy = 'coalesce(deletedAt, updatedAt, createdAt)'
    }

    const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined

    return {
      limit,
      sort: pageCursor ? undefined : sort,
      ...filter,
      pageCursor,
      incTotal: !pageCursor || undefined,
    }
  }

  return { handleQueryParams, encodeListParams }
}
