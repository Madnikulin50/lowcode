<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      Библиотека факторов риска
    </Teleport>

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('riskFactor.searchPlaceholder', 'Поиск факторов'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('riskFactor.noItems', 'Факторы риска не найдены'),
        loading: $t('label.loading', 'Loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('riskFactor.label.single', 'фактор риска'),
        resourcePlural: $t('riskFactor.label.plural', 'факторы риска'),
      }"
      clickable
      sticky-header
      class="h-100 flex-fill"
      ref="resourceList"
      @search="handleSearch"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <router-link
          :to="{ name: 'admin.risk.factors.create' }"
          class="btn btn-primary"
        >
          <font-awesome-icon :icon="['fas', 'plus']" />
          <span class="ms-1">
            {{ $t('riskFactor.add', 'Новый фактор') }}
          </span>
        </router-link>
      </template>

      <template #handle="{ item }">
        <font-awesome-icon
          :icon="['fas', 'list']"
          class="me-2"
          style="height: 1rem; width: 1rem; opacity: 0.65;"
        />
        <code>{{ item.handle }}</code>
      </template>

      <template #scale="{ item }">
        <span class="badge bg-info-subtle text-info-emphasis">
          {{ scaleLabel(item.scale) }}
        </span>
        <span
          v-if="item.scale === 'states'"
          class="text-secondary ms-1 small"
        >
          ({{ item.statesCount }})
        </span>
      </template>

      <template #source="{ item }">
        {{ item.source }}<span v-if="item.fieldName" class="text-secondary"> · {{ item.fieldName }}</span>
      </template>

      <template #actions="{ item }">
        <div class="dropdown">
          <button
            class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            <font-awesome-icon :icon="['fas', 'ellipsis-v']" />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-input-confirm
                :text="$t('riskFactor.edit.delete', 'Удалить фактор')"
                show-icon
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(item)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'

const { useToast } = composables
const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const primaryKey = 'factorID'

const filter = ref({ query: '' })
const sorting = ref({
  sortBy: 'handle',
  sortDesc: false,
})
const pagination = ref({
  total: 0,
  limit: 25,
  page: 1,
  prevPage: '',
  nextPage: '',
})

const factors = ref([])
const resourceList = ref(null)

const scaleLabels = { numeric: 'Числовая', bool: 'Да/нет', states: 'Состояния' }
function scaleLabel (s) {
  return scaleLabels[s] || s
}

const tableFields = computed(() => [
  {
    key: 'handle',
    label: t('riskFactor.columns.handle', 'Код'),
    sortable: true,
    tdClass: 'text-nowrap',
  },
  {
    key: 'label',
    label: t('riskFactor.columns.label', 'Название'),
    sortable: true,
  },
  {
    key: 'scale',
    label: t('riskFactor.columns.scale', 'Шкала'),
    tdClass: 'text-nowrap',
  },
  {
    key: 'source',
    label: t('riskFactor.columns.source', 'Источник'),
  },
  {
    key: 'actions',
    label: '',
    tdClass: 'text-end text-nowrap actions',
  },
])

const { toastSuccess, toastErrorHandler } = useToast()

let fetchedOnce = false

onMounted(() => {
  document.title = t('label.app-name.riskFactors', { label: props.namespace?.name, interpolation: { escapeValue: false } })
})

watch(() => route.query.page, (page) => {
  pagination.value.page = Number(page) || 1
}, { immediate: true })

watch(() => route.query.limit, (limit) => {
  pagination.value.limit = Number(limit) || 25
}, { immediate: true })

function items () {
  if (!fetchedOnce) {
    fetchedOnce = true
    return $ComposeAPI.riskFactorList({ namespaceID: props.namespace?.namespaceID })
      .then(({ factors: set }) => {
        factors.value = (set || []).map((f) => ({ ...f, statesCount: (f.states || []).length }))
        return sliceFactors()
      })
      .catch((e) => {
        toastErrorHandler(t('riskFactor.notification.listFailed', 'Не удалось загрузить факторы'))(e)
        return []
      })
  }
  return Promise.resolve(sliceFactors())
}

function handleSearch () {
  pagination.value.page = 1
  router.replace({ query: { ...route.query, page: '1' } })
  resourceList.value?.refresh()
}

function sliceFactors () {
  const q = (filter.value.query || '').toLowerCase()
  let list = factors.value
  if (q) {
    list = list.filter((f) => (f.handle || '').toLowerCase().includes(q) || (f.label || '').toLowerCase().includes(q))
  }
  const { sortBy, sortDesc } = sorting.value
  if (sortBy === 'handle' || sortBy === 'label') {
    list = [...list].sort((a, b) => {
      const r = String(a[sortBy] || '').localeCompare(String(b[sortBy] || ''))
      return sortDesc ? -r : r
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

function handleRowClicked (factor) {
  router.push({ name: 'admin.risk.factors.edit', params: { factorID: factor.factorID }, query: null })
}

function handleDelete (factor) {
  $ComposeAPI.riskFactorDelete({ factorID: factor.factorID })
    .then(() => {
      toastSuccess(t('riskFactor.notification.deleted', 'Фактор удалён'))
      fetchedOnce = false
      items()
      resourceList.value?.refresh()
    })
    .catch(toastErrorHandler(t('riskFactor.notification.deleteFailed', 'Не удалось удалить фактор')))
}
</script>
