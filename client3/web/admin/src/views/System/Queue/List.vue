<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.queues.list.title')" />
    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :row-class="genericRowClass"
      :translations="{ searchPlaceholder: $t('system.queues.list.filterForm.handle.placeholder'), notFound: $t('admin.general.notFound'), noItems: $t('general.resource-list.no-items'), loading: $t('system.queues.list.loading'), showingPagination: 'general.pagination.showing', singlePluralPagination: 'general.pagination.single', prevPagination: $t('admin.general.pagination.prev'), nextPagination: $t('admin.general.pagination.next'), resourceSingle: $t('label.queue.single'), resourcePlural: $t('label.queue.plural') }"
      clickable
      sticky-header
      class="custom-resource-list-height flex-fill"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <button v-if="canCreate" class="btn btn-primary btn-lg" @click="$router.push({ name: 'system.queue.new' })">{{ $t('system.queues.list.new') }}</button>
      </template>
      <template #toolbar>
        <c-resource-list-status-filter
          v-model="filter.deleted"
          data-test-id="filter-deleted-queues"
          :label="$t('system.queues.list.filterForm.deleted.label')"
          :excluded-label="$t('system.queues.list.filterForm.excluded.label')"
          :inclusive-label="$t('system.queues.list.filterForm.inclusive.label')"
          :exclusive-label="$t('system.queues.list.filterForm.exclusive.label')"
          @change="filterList"
        />
      </template>
      <template #actions="{ item: q }">
        <div v-if="q.canDeleteQueue" class="dropdown">
          <button class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2" data-bs-toggle="dropdown"><font-awesome-icon :icon="['fas', 'ellipsis-v']" /></button>
          <ul class="dropdown-menu m-0"><li><c-input-confirm :text="getActionText(q)" show-icon :icon="getActionIcon(q)" borderless variant="link" size="md" button-class="dropdown-item" icon-class="text-danger" class="w-100" @confirmed="handleDelete(q)" /></li></ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.queues', keyPrefix: 'list' } })
import { computed, reactive, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components } from 'corteza-lib/vue/dist'
import { useListHelpers } from 'corteza-webapp-admin/src/mixins/listHelpers'
import { useStore } from '../../../store'

const { CResourceList } = components
const { t } = useI18n()
const store = useStore()
const $SystemAPI = inject('$SystemAPI', window.__systemAPI)

const primaryKey = 'queueID'
const editRoute = 'system.queue.edit'

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
} = useListHelpers({ editRoute, primaryKey })

const filter = reactive({ query: '', archived: 0, deleted: 0 })
const sorting = reactive({ sortBy: 'createdAt', sortDesc: true })

const fields = computed(() => [
  { key: 'queue', sortable: true, label: t('columns.queue') },
  { key: 'consumer', sortable: false, label: t('columns.consumer') },
  { key: 'createdAt', sortable: true, label: t('columns.createdAt'), formatter: (v) => moment(v).fromNow() },
  { key: 'actions', label: '', class: 'actions' },
])

const canCreate = computed(() => store.rbac.can('system/', 'queue.create'))

function items () {
  return procListResults(
    $SystemAPI.queuesListCancellable(encodeListParams(filter, sorting, pagination, { value: undefined })),
    true,
    pagination,
    sorting,
    filter,
    { value: undefined },
  )
}

function handleDelete (queue) {
  handleItemDelete({
    resource: queue,
    resourceName: 'queues',
    locale: 'queue',
  }, filter, pagination)
}
</script>
