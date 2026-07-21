<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :row-class="genericRowClass"
      :translations="{
        notFound: $t('admin.general.notFound'),
        noItems: $t('general.resource-list.no-items'),
        loading: $t('loading'),
        showingPagination: 'general.pagination.showing',
        singlePluralPagination: 'general.pagination.single',
        prevPagination: $t('admin.general.pagination.prev'),
        nextPagination: $t('admin.general.pagination.next'),
        resourceSingle: $t('label.connection.single'),
        resourcePlural: $t('label.connection.plural')
      }"
      clickable
      hide-search
      class="h-100 bg-transparent"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <router-link
          class="btn btn-primary btn-lg"
          :to="{ name: 'system.connection.new' }"
        >
          {{ $t('add-button') }}
        </router-link>
      </template>

      <template #toolbar>
        <c-resource-list-status-filter
          v-model="filter.deleted"
          data-test-id="filter-deleted-connections"
          :label="$t('deleted.label')"
          :excluded-label="$t('excluded.label')"
          :inclusive-label="$t('inclusive.label')"
          :exclusive-label="$t('exclusive.label')"
          @change="filterList"
        />
      </template>

      <template #actions="{ item: c }">
        <div v-if="c.canDeleteConnection" class="dropdown">
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
                :text="getActionText(c)"
                show-icon
                :icon="getActionIcon(c)"
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(c)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useListHelpers } from 'corteza-webapp-admin/src/mixins/listHelpers'
import { useI18n } from 'vue-i18n'
import { components } from '../../../../../lib/vue/dist'
import moment from 'moment'

const { CResourceList } = components
const { t } = useI18n()

const {
  pagination,
  genericRowClass,
  handleRowClicked,
  getActionText,
  getActionIcon,
  procListResults,
  encodeListParams,
  filterList,
  handleItemDelete,
} = useListHelpers()

const primaryKey = 'connectionID'
const editRoute = 'system.connection.edit'

const filter = reactive({
  type: 'corteza::system:dal-connection',
  query: '',
  deleted: 0,
})

const sorting = reactive({
  sortBy: 'createdAt',
  sortDesc: true,
})

const fields = [
  {
    key: 'name',
    sortable: false,
    formatter: (value, col, conn) => conn.meta.name || conn.handle,
  },
  {
    key: 'location',
    sortable: false,
    formatter: (value, col, conn) => conn.meta.location.properties.name,
  },
  {
    key: 'ownership',
    sortable: false,
    formatter: (value, col, conn) => conn.meta.ownership,
  },
  {
    key: 'createdAt',
    sortable: false,
    formatter: (v) => moment(v).fromNow(),
  },
  {
    key: 'actions',
    class: 'actions',
  },
].map(c => ({
  label: c.label || t(`columns.${c.key}`),
  ...c,
}))

function items() {
  return procListResults(window.__SystemAPI.dalConnectionListCancellable(encodeListParams(filter, sorting, pagination, { value: undefined })), true, pagination, sorting, filter, { value: undefined })
}

function handleDelete(connection) {
  handleItemDelete({
    resource: connection,
    resourceName: 'dalConnection',
    locale: 'connection',
  }, filter, pagination)
}
</script>
