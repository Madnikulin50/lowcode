<template>
  <div class="container-fluid d-flex flex-column py-3">
    <Teleport to="#topbar-title">
      {{ $t('pages.rag.title', 'Pages RAG') }}
    </Teleport>

    <c-resource-list
      ref="resourceList"
      data-test-id="table-pages-rag"
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :translations="translations"
      sticky-header
      class="h-100 flex-fill"
      @search="onSearch"
    >
      <template #header>
        <button
          class="btn btn-primary"
          :disabled="reindexing"
          @click="reindex"
        >
          <span
            v-if="reindexing"
            class="spinner-border spinner-border-sm me-1"
          />
          <font-awesome-icon
            :icon="['fas', 'sync']"
            class="me-1"
          />
          {{ $t('pages.rag.reindex', 'Re-index') }}
        </button>
      </template>

      <template #toolbar>
        <div
          v-if="reindexing && progressData.totalPages > 0"
          class="w-100"
        >
          <div class="d-flex justify-content-between text-muted small mb-1">
            <span>
              {{ $t('pages.rag.indexing', 'Indexing page {indexed} of {total}', { indexed: progressData.indexedPages, total: progressData.totalPages }) }}
            </span>
            <span
              v-if="progressData.currentPage"
              class="text-truncate ms-2"
            >
              {{ progressData.currentPage }}
            </span>
          </div>
          <div class="progress" style="height: 6px;">
            <div
              class="progress-bar progress-bar-striped progress-bar-animated"
              :style="{ width: progressPercent + '%' }"
            />
          </div>
        </div>

        <div
          v-if="reindexMessage"
          class="w-100 alert alert-success py-2 mb-0"
        >
          {{ reindexMessage }}
        </div>

        <div
          v-if="error"
          class="w-100 alert alert-danger py-2 mb-0"
        >
          {{ error }}
        </div>
      </template>

      <template #page="{ item }">
        <span class="text-nowrap">{{ item.title || item.pageID }}</span>
      </template>

      <template #block="{ item }">
        <span class="badge bg-secondary me-1">{{ item.blockKind }}</span>
        <span>{{ item.blockTitle }}</span>
      </template>

      <template #text="{ item }">
        <span
          class="d-block text-break"
          style="max-width: 500px;"
        >
          {{ item.text }}
        </span>
      </template>

      <template #index="{ item }">
        {{ item.chunkIndex }}
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const route = useRoute()

const $ComposeAPI = window.__composeAPI

const primaryKey = 'id'

const resourceList = ref(null)
const chunks = ref([])
const filter = ref({ query: '' })
const sorting = ref({
  sortBy: 'page',
  sortDesc: false,
})
const pagination = ref({
  total: 0,
  limit: 25,
  page: 1,
  prevPage: '',
  nextPage: '',
})

const reindexing = ref(false)
const reindexMessage = ref('')
const error = ref('')
const progressData = ref({ totalPages: 0, indexedPages: 0, currentPage: '', totalBlocks: 0, indexedBlocks: 0, complete: false })

let progressTimer = null
let fetchedOnce = false

const translations = computed(() => ({
  searchPlaceholder: t('pages.rag.searchPlaceholder', 'Search chunks'),
  notFound: t('resourceList.notFound', 'Not found'),
  noItems: t('pages.rag.noItems', 'No indexed pages found. Click "Re-index" to start indexing.'),
  loading: t('label.loading', 'Loading'),
  showingPagination: t('resourceList.pagination.showing', 'Showing'),
  singlePluralPagination: t('resourceList.pagination.single', 'chunk'),
  prevPagination: t('resourceList.pagination.prev', 'Previous'),
  nextPagination: t('resourceList.pagination.next', 'Next'),
  resourceSingle: t('pages.rag.chunk.single', 'chunk'),
  resourcePlural: t('pages.rag.chunk.plural', 'chunks'),
  recordsPerPage: t('resourceList.pagination.recordsPerPage', 'per page'),
}))

const fields = computed(() => [
  {
    key: 'page',
    label: t('pages.rag.columns.page', 'Page'),
    sortable: true,
    tdClass: 'text-nowrap',
  },
  {
    key: 'block',
    label: t('pages.rag.columns.block', 'Block'),
    sortable: true,
  },
  {
    key: 'text',
    label: t('pages.rag.columns.text', 'Text'),
  },
  {
    key: 'index',
    label: '#',
    class: 'text-end text-nowrap',
  },
])

const progressPercent = computed(() => {
  if (!progressData.value.totalPages) return 0
  return Math.round((progressData.value.indexedPages / progressData.value.totalPages) * 100)
})

watch(() => [route.query.page, route.query.limit, route.query.pageCursor], () => {
  if (fetchedOnce) resourceList.value?.refresh()
})

function items () {
  if (!fetchedOnce) {
    fetchedOnce = true
    return $ComposeAPI.ragPagesList()
      .then(res => {
        if (Array.isArray(res)) {
          chunks.value = res
        } else if (res?.set && Array.isArray(res.set)) {
          chunks.value = res.set
        } else {
          chunks.value = []
        }
        error.value = ''
        return sliceChunks()
      })
      .catch(e => {
        error.value = e.message || 'Failed to load RAG data'
        return []
      })
  }
  return Promise.resolve(sliceChunks())
}

function sliceChunks () {
  const q = (filter.value.query || '').toLowerCase()
  let list = chunks.value
  if (q) {
    list = list.filter(c => {
      return [c.title, c.blockKind, c.blockTitle, c.text]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(q)
    })
  }

  if (sorting.value.sortBy) {
    const valueOf = (c) => {
      if (sorting.value.sortBy === 'page') return c.title || c.pageID || ''
      if (sorting.value.sortBy === 'block') return `${c.blockKind || ''} ${c.blockTitle || ''}`
      return c[sorting.value.sortBy] || ''
    }
    list = [...list].sort((a, b) => {
      const r = String(valueOf(a)).localeCompare(String(valueOf(b)))
      return sorting.value.sortDesc ? -r : r
    })
  }

  const total = list.length
  const page = parseInt(String(route.query.page || '1'), 10) || 1
  const limit = parseInt(String(route.query.limit || '25'), 10) || 25
  const maxPage = Math.max(1, Math.ceil(total / limit))

  pagination.value.total = total
  pagination.value.limit = limit
  pagination.value.page = page
  pagination.value.prevPage = page > 1 ? String(page - 1) : ''
  pagination.value.nextPage = page < maxPage ? String(page + 1) : ''

  return list.slice((page - 1) * limit, page * limit)
}

function onSearch () {
  resourceList.value?.refresh()
}

function startProgressPolling () {
  stopProgressPolling()
  progressTimer = setInterval(() => {
    $ComposeAPI.ragPagesReindexProgress()
      .then(p => {
        progressData.value = p || {}
      })
      .catch(() => {})
  }, 500)
}

function stopProgressPolling () {
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
  }
}

function reindex () {
  reindexing.value = true
  reindexMessage.value = ''
  error.value = ''
  progressData.value = { totalPages: 0, indexedPages: 0, currentPage: '', totalBlocks: 0, indexedBlocks: 0, complete: false }
  startProgressPolling()

  $ComposeAPI.ragPagesReindex()
    .then(() => {
      reindexMessage.value = t('pages.rag.reindexDone', 'Reindex completed successfully')
      stopProgressPolling()
      fetchedOnce = false
      resourceList.value?.refresh()
    })
    .catch(e => {
      error.value = e.message || 'Reindex failed'
    })
    .finally(() => {
      reindexing.value = false
      stopProgressPolling()
    })
}
</script>
