<template>
  <div class="card shadow-sm" header-class="border-bottom" body-class="p-0">
    <div class="card-header border-bottom">
      <h4 class="mb-0">
        {{ $t('list.title') }}
      </h4>
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
         searchPlaceholder: $t('list.filterForm.query.placeholder'),
        notFound: $t('admin.general.notFound'),
        noItems: $t('general.resource-list.no-items'),
         loading: $t('list.loading'),
        showingPagination: 'general.pagination.showing',
        singlePluralPagination: 'general.pagination.single',
        prevPagination: $t('admin.general.pagination.prev'),
        nextPagination: $t('admin.general.pagination.next'),
        resourceSingle: $t('label.route.single'),
        resourcePlural: $t('label.route.plural')
      }"
      clickable
      card-header-class="rounded-0"
      class="h-100 bg-transparent"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <button
          v-if="canCreate"
          data-test-id="button-add"
          class="btn btn-primary btn-lg"
          @click="goToNew"
        >
          {{ $t('list.new') }}
        </button>

        <button
          v-if="$Settings.get('apigw.profiler.enabled', false)"
          data-test-id="button-profiler"
          class="btn btn-info btn-lg"
          @click="goToProfiler"
        >
          {{ $t('list.profiler') }}
        </button>

        <c-permissions-button
          v-if="canGrant"
          data-test-id="button-permissions"
          resource="corteza::system:apigw-route/*"
          :button-label="$t('permissions')"
          size="lg"
        />
      </template>

      <template #toolbar>
        <c-resource-list-status-filter
          v-model="filter.deleted"
          data-test-id="filter-deleted-routes"
          :label="$t('list.filterForm.deleted.label')"
          :excluded-label="$t('list.filterForm.excluded.label')"
          :inclusive-label="$t('list.filterForm.inclusive.label')"
          :exclusive-label="$t('list.filterForm.exclusive.label')"
          @change="filterList"
        />
      </template>

      <template #actions="{ item: r }">
        <div
          v-if="(areActionsVisible({ resource: r, conditions: ['canDeleteApigwRoute', 'canGrant'] }))"
          class="dropdown"
        >
          <button
            class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            <font-awesome-icon
              :icon="['fas', 'ellipsis-v']"
            />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-permissions-button
                v-if="r.routeID && canGrant"
                :title="r.endpoint || r.routeID"
                :target="r.endpoint || r.routeID"
                :resource="`corteza::system:apigw-route/${r.routeID}`"
          :button-label="$t('list.permissions')"
                class="dropdown-item"
              />
            </li>
            <li>
              <c-input-confirm
                v-if="r.canDeleteApigwRoute"
                :text="getActionText(r)"
                show-icon
                :icon="getActionIcon(r)"
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(r)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.apigw', keyPrefix: 'list' } })
import { computed, reactive, getCurrentInstance, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
import moment from 'moment'
import { useListHelpers } from 'corteza-webapp-admin/src/mixins/listHelpers'
import { useStore } from '../../store'

const { CResourceList } = components
const { t } = useI18n()
const router = useRouter()
const $Settings = inject('$Settings', {})
const store = useStore()
const { proxy } = getCurrentInstance()

const {
  pagination,
  genericRowClass,
  handleRowClicked,
  areActionsVisible,
  getActionText,
  getActionIcon,
  procListResults,
  encodeListParams,
  filterList,
  handleItemDelete,
} = useListHelpers()

const primaryKey = 'routeID'
const editRoute = 'system.apigw.edit'

const filter = reactive({
  query: '',
  deleted: 0,
})

const sorting = reactive({
  sortBy: 'createdAt',
  sortDesc: true,
})

const fields = [
  {
    key: 'endpoint',
    sortable: true,
  },
  {
    key: 'method',
    sortable: false,
  },
  {
    key: 'enabled',
    formatter: (v) => v ? 'Yes' : 'No',
  },
  {
    key: 'createdAt',
    sortable: true,
    formatter: (v) => moment(v).fromNow(),
  },
  {
    key: 'actions',
    class: 'actions',
  },
].map(c => ({
  ...c,
  label: t(`columns.${c.key}`),
}))

const canCreate = computed(() => store.rbac.can('system/', 'apigw-route.create'))

const canGrant = computed(() => store.rbac.can('system/', 'grant'))

function goToNew () {
  router.push({ name: 'system.apigw.new' })
}

function goToProfiler () {
  router.push({ name: 'system.apigw.profiler' })
}

function items () {
  return procListResults(proxy.$SystemAPI.apigwRouteListCancellable(encodeListParams(filter, sorting, pagination, { value: undefined })), true, pagination, sorting, filter, { value: undefined })
}

function handleDelete (route) {
  handleItemDelete({
    resource: route,
    resourceName: 'apigwRoute',
    locale: 'gateway',
  }, filter, pagination)
}
</script>

<style lang="scss">
.route-list {
  .card-header {
    border-radius: 0;
  }
}
</style>
