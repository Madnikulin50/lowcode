<template>
    <div
      id="resource-list-wrapper"
      class="card shadow-sm d-flex flex-column"
      style="min-height: 45rem;"
    >
    <div
      class="card-header border-0"
      :class="cardHeaderClass"
    >
      <div class="container-fluid d-flex flex-column p-0 gap-2 d-print-none">
        <div class="row g-0 d-flex align-items-center justify-content-between gap-1">
          <div :class="`d-flex align-items-center flex-grow-1 flex-wrap flex-fill-child gap-1 ${headerClass}`">
            <slot
              name="header"
              :selected="selected"
            />
          </div>
          <div
            v-if="!hideSearch"
            class="flex-fill"
          >
            <c-input-search
              :model-value="(filter as Record<string, any>)[queryField]"
              :placeholder="translations.searchPlaceholder"
              :submittable="true"
              @search="handleSearch"
            />
          </div>
        </div>
        <div
          v-if="$slots.toolbar"
          class="row gap-1"
        >
          <slot name="toolbar" />
        </div>
      </div>
    </div>

    <div class="card-body p-0 d-flex flex-column">
      <BTable
        id="resource-list"
        :items="tableItems"
        :fields="_fields"
        :primary-key="primaryKey"
        :busy="loading"
        :hover="true"
        :sticky-header="stickyHeader ? stickyHeaderHeight : false"
        :no-local-sorting="true"
        :no-provider-sorting="true"
        :no-provider-paging="true"
        :no-provider-filtering="true"
        :show-empty="true"
        :tbody-tr-class="getTbodyTrClass"
        :empty-text="translations.noItems"
        table-class="mb-0 resource-list-table"
        thead-class="resource-list-thead"
        v-model:sort-by="sortByModel"
        @row-clicked="onRowClicked"
      >
        <template #head()="{ label, field }">
          <div class="d-flex align-items-center">
            <div class="flex-fill text-nowrap">
              {{ label }}
              <span
                v-if="field.sortable"
                class="btn d-inline-flex align-items-center text-secondary d-print-none border-0 px-1 ms-1 btn-outline-extra-light"
                style="margin-right: -0.25rem;"
              >
                <div class="d-print-none fa-layers">
                  <font-awesome-icon
                    :icon="['fas', 'angle-up']"
                    class="mb-1"
                    :class="{ 'text-primary': sorting.sortBy === field.key && !sorting.sortDesc }"
                  />
                  <font-awesome-icon
                    :icon="['fas', 'angle-down']"
                    class="mt-1"
                    :class="{ 'text-primary': sorting.sortBy === field.key && sorting.sortDesc }"
                  />
                </div>
              </span>
            </div>
          </div>
        </template>

        <template
          v-if="selectable"
          #head(select)="{}"
        >
          <div class="form-check">
            <input
              :disabled="disableSelectAll"
              :checked="allRowsSelected && !disableSelectAll"
              type="checkbox"
              class="form-check-input-v3"
              @change="selectAllRows"
            >
          </div>
        </template>

        <template
          v-if="selectable"
          #cell(select)="{ item }"
        >
          <div class="form-check">
            <input
              v-if="isItemSelectable(item)"
              type="checkbox"
              class="form-check-input-v3"
              :checked="selected.includes(item[primaryKey])"
              @change="onSelectRow(($event.target as HTMLInputElement).checked, item[primaryKey])"
            >
          </div>
        </template>

        <template #table-busy>
          <div class="text-center m-5">
            <span class="spinner-border spinner-border-sm align-middle m-2" />
            {{ translations.loading }}
          </div>
        </template>

        <template #empty>
          <p
            data-test-id="no-matches"
            class="text-center text-dark"
            style="margin-top: 1vh;"
          >
            {{ translations.noItems }}
          </p>
        </template>

        <template
          v-for="key in customFieldSlots"
          :key="key"
          #[cellSlotName(key)]="{ item }"
        >
          <slot
            :name="key"
            :item="item"
          />
        </template>
      </BTable>
    </div>

    <div
      v-if="showFooter"
      class="card-footer bg-light p-0"
    >
      <div class="resource-list-footer d-flex align-items-center flex-wrap px-3 py-2 gap-2">
        <div class="d-flex align-items-center flex-wrap gap-2">
          <div
            v-if="!hideTotal"
            class="text-nowrap text-truncate py-1"
          >
            {{ getPagination }}
          </div>
          <div
            v-if="!hidePerPageOption"
            class="d-flex align-items-center gap-1 text-nowrap"
          >
            <span>{{ translations.recordsPerPage || $t('resourceList.pagination.recordsPerPage') }}</span>
            <select
              :value="pagination.limit"
              class="form-select form-control form-select-sm"
              @change="handlePerPageChange(Number(($event.target as HTMLSelectElement).value))"
            >
              <option
                v-for="opt in perPageOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.text }}
              </option>
            </select>
          </div>
        </div>
        <div
          v-if="!hidePagination"
          class="d-flex align-items-center ms-auto"
        >
          <div class="btn-group gap-1">
            <button
              :disabled="!hasPrevPage"
              type="button"
              class="btn btn-outline-extra-light d-flex align-items-center text-dark border-0 p-1"
              @click="goToPage('firstPage')"
            >
              <font-awesome-icon :icon="['fas', 'angle-double-left']" />
            </button>
            <button
              :disabled="!hasPrevPage"
              type="button"
              class="btn btn-outline-extra-light d-flex align-items-center text-dark border-0 p-1"
              @click="goToPage('prevPage')"
            >
              <font-awesome-icon
                :icon="['fas', 'angle-left']"
                class="me-1"
              />
              {{ translations.prevPagination }}
            </button>
            <button
              :disabled="!hasNextPage"
              type="button"
              class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-dark border-0 p-1"
              @click="goToPage('nextPage')"
            >
              {{ translations.nextPagination }}
              <font-awesome-icon
                :icon="['fas', 'angle-right']"
                class="ms-1"
              />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, getCurrentInstance, useSlots, inject } from 'vue'
import { routeLocationKey, routerKey } from 'vue-router'
import { BTable } from 'bootstrap-vue-next'
import CInputSearch from '../input/CInputSearch.vue'

interface Field {
  key: string
  label: string
  sortable?: boolean
  sort?: boolean
  thClass?: string
  thStyle?: string
  column?: string
}

interface Pagination {
  total?: number
  limit?: number
  page?: number
  prevPage?: string
  nextPage?: string
}

interface Sorting {
  sortBy: string
  sortDesc: boolean
}

const props = withDefaults(defineProps<{
  primaryKey: string
  filter: Record<string, any>
  sorting: Sorting
  pagination: Pagination
  fields: Field[]
  items: () => Promise<Array<Record<string, any>>>
  hideSearch?: boolean
  hideTotal?: boolean
  hidePagination?: boolean
  stickyHeader?: boolean
  clickable?: boolean
  selectable?: boolean
  isItemSelectable?: (item: Record<string, any>) => boolean
  cardHeaderClass?: string
  headerClass?: string
  rowClass?: (item: Record<string, any>) => Record<string, boolean>
  translations: Record<string, string>
  queryField?: string
  hidePerPageOption?: boolean
}>(), {
  hideSearch: false,
  hideTotal: false,
  hidePagination: false,
  stickyHeader: false,
  clickable: false,
  selectable: false,
  isItemSelectable: () => true,
  cardHeaderClass: '',
  headerClass: '',
  rowClass: () => ({}),
  queryField: 'query',
  hidePerPageOption: false,
})

let requestId = 0

const emit = defineEmits<{
  (e: 'row-clicked', item: Record<string, any>): void
  (e: 'search'): void
}>()

const instance = getCurrentInstance()
const $t = instance!.appContext.config.globalProperties.$t
const route = inject(routeLocationKey, {} as any) || {}
const router = inject(routerKey, {} as any) || {}
const slots = useSlots()
const ariaSortObserver = ref<MutationObserver | null>(null)

const selected = ref<string[]>([])
const selectableItemIDs = ref<string[]>([])
const tableItems = ref<Array<Record<string, any>>>([])
const loading = ref(false)

const stickyHeaderHeight = computed(() => '60vh')

const _fields = computed(() => {
  const selectField = props.selectable
    ? [{
      key: 'select',
      label: '',
      sortable: false,
      thClass: 'border-0 table-b-table-default b-table-sticky-column',
      thStyle: 'width: 0; white-space: nowrap;',
    }]
    : []

  return [
    ...selectField,
    ...props.fields.map(f => ({
      ...f,
      sortable: f.sortable || false,
      thClass: (f.thClass || 'border-0') + ' table-b-table-default b-table-sticky-column',
      thStyle: f.thStyle || '',
    })),
  ]
})

const sortByModel = computed({
  get(): Array<{ key: string; order: 'asc' | 'desc' }> | undefined {
    if (!props.sorting.sortBy) return undefined
    return [{ key: props.sorting.sortBy, order: props.sorting.sortDesc ? 'desc' : 'asc' }]
  },
  set(val: Array<{ key: string; order: string }> | undefined) {
    const item = val?.[0]
    if (item) {
      props.sorting.sortBy = item.key
      props.sorting.sortDesc = item.order === 'desc'
    } else {
      props.sorting.sortBy = ''
      props.sorting.sortDesc = false
    }
    props.pagination.page = 1
    refresh()
  },
})

const customFieldSlots = computed(() => {
  return Object.keys(slots).filter(s => s !== 'header' && s !== 'toolbar')
})

function cellSlotName(key: string): string {
  return `cell(${key})`
}

function getTbodyTrClass(item: Record<string, any> | null, _type: string) {
  if (!item) return undefined
  return {
    pointer: props.clickable,
    ...props.rowClass(item),
  }
}

const perPageOptions = computed(() => {
  const defaultText = props.pagination.limit === 0 ? $t('label.all') : props.pagination.limit!.toString()
  return [
    { text: defaultText, value: props.pagination.limit ?? 0 },
    { text: '25', value: 25 },
    { text: '50', value: 50 },
    { text: '100', value: 100 },
  ]
    .filter((v, i) => i === 0 || v.value !== props.pagination.limit)
    .sort((a, b) => {
      if (a.value === 0) return 1
      if (b.value === 0) return -1
      return a.value - b.value
    })
})

const disableSelectAll = computed(() => !selectableItemIDs.value.length)

const allRowsSelected = computed(() => selected.value.length === selectableItemIDs.value.length)

const getPagination = computed(() => {
  let { total = 0, limit = 10, page = 1 } = props.pagination
  total = isNaN(total) ? 0 : total
  const pagination = {
    from: ((page - 1) * limit) + 1,
    to: limit > 0 ? Math.min((page * limit), total) : total,
    count: total,
    data: total === 1 ? props.translations.resourceSingle : props.translations.resourcePlural,
  }
  return $t(props.translations[total > limit ? 'showingPagination' : 'singlePluralPagination'], pagination)
})

const hasPrevPage = computed(() => !!props.pagination.prevPage)

const hasNextPage = computed(() => !!props.pagination.nextPage)

const showFooter = computed(() => !(props.hideTotal && props.hidePagination && props.hidePerPageOption))

async function loadItems() {
  const id = ++requestId
  selected.value = []
  selectableItemIDs.value = []
  loading.value = true
  try {
    const items = await props.items()
    if (id !== requestId) return
    selectableItemIDs.value = items.filter(props.isItemSelectable).map(i => i[props.primaryKey])
    tableItems.value = items
  } finally {
    if (id === requestId) loading.value = false
  }
}

const onRefreshTable = () => { loadItems() }

window.addEventListener('bv::refresh::table', onRefreshTable)

function refresh() {
  loadItems()
}

function onRowClicked(payload: { item: Record<string, any> }) {
  emit('row-clicked', payload.item)
}

function onSelectRow(selectedFlag: boolean, itemID: string) {
  if (selectedFlag) {
    if (!selected.value.includes(itemID)) {
      selected.value.push(itemID)
    }
  } else {
    const i = selected.value.indexOf(itemID)
    if (i >= 0) {
      selected.value.splice(i, 1)
    }
  }
}

function selectAllRows(e: Event) {
  const allSelected = (e.target as HTMLInputElement).checked
  if (allSelected) {
    selected.value = [...selectableItemIDs.value]
  } else {
    selected.value = []
  }
}

function goToPage(destination: string) {
  const pageCursor = (props.pagination as Record<string, any>)[destination] || ''
  let { page = 1 } = props.pagination

  if (destination === 'nextPage') {
    page = (page as number) + 1
  } else if (destination === 'prevPage') {
    page = (page as number) - 1
  } else {
    page = 1
  }

  router.replace({ query: { ...route.query, page, pageCursor } })
}

function handlePerPageChange(limit: number) {
  router.replace({ query: { ...route.query, page: 1, limit } })
  refresh()
}

function handleSearch(searchQuery: string) {
  (props.filter as Record<string, any>)[props.queryField] = searchQuery ? searchQuery.trim() : ''
  emit('search')
}

defineExpose({ refresh })

onMounted(() => {
  loadItems()
  nextTick(() => {
    const table = document.getElementById('resource-list')
    if (!table) return
    const removeAriaSort = () => {
      table.querySelectorAll('th[aria-sort]').forEach(th => th.removeAttribute('aria-sort'))
    }
    removeAriaSort()
    ariaSortObserver.value = new MutationObserver(removeAriaSort)
    ariaSortObserver.value.observe(table, { subtree: true, attributeFilter: ['aria-sort'] })
  })
})
onBeforeUnmount(() => {
  window.removeEventListener('bv::refresh::table', onRefreshTable)
  ariaSortObserver.value?.disconnect()
  selected.value = []
  selectableItemIDs.value = []
})
</script>

<style lang="scss">
.visually-hidden {
  position: absolute !important;
  width: 1px !important;
  height: 1px !important;
  padding: 0 !important;
  margin: -1px !important;
  overflow: hidden !important;
  clip: rect(0, 0, 0, 0) !important;
  white-space: nowrap !important;
  border: 0 !important;
}

#resource-list-wrapper {
  .b-table-sticky-header {
    margin-bottom: 0 !important;
    flex-grow: 1 !important;
    max-height: none !important;
  }
}

.pointer {
  cursor: pointer;
}

#resource-list.resource-list-table > :not(caption) > * > * {
  padding: 0.75rem 0.5rem;
}

.resource-list-thead {
}

#resource-list td.actions {
  padding-top: 8px;
  right: 0;
  opacity: 0;
  position: sticky;
  transition: opacity 0.25s;
  width: 1%;
  font-family: var(--font-regular) !important;
}

#resource-list tr:hover td.actions {
  opacity: 1;
  z-index: 1;
  background-color: var(--light);
}

.resource-list-footer {
  font-family: var(--font-medium);
}
</style>
