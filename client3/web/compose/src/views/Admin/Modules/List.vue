<template>
  <div class="container-fluid d-flex flex-column py-3">
    <Teleport to="#topbar-title">
      {{ $t('navigation.module') }}
    </Teleport>

    <c-resource-list
      data-test-id="table-modules-list"
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('searchPlaceholder'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('resourceList.noItems', 'No items'),
        loading: $t('label.loading', 'Loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: $t('resourceList.pagination.single', 'resource'),
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('label.module.single', 'Module'),
        resourcePlural: $t('label.module.plural', 'Modules'),
      }"
      clickable
      sticky-header
      class="h-100 flex-fill"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <router-link
          v-if="namespace?.canCreateModule"
          data-test-id="button-create"
          :to="{ name: 'admin.modules.create' }"
          class="btn btn-primary"
        >
          {{ $t('createLabel') }}
        </router-link>

        <import
          v-if="namespace?.canCreateModule"
          :namespace="namespace"
          type="module"
          @importSuccessful="onImportSuccessful"
        />

        <export
          v-if="namespace?.canExportModules"
          :list="modules"
          type="module"
        />

        <div
          v-if="namespace?.canGrant"
          class="dropdown"
        >
          <button
            class="btn btn-outline-secondary dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            <font-awesome-icon :icon="['fas', 'lock']" />
            <span>
              {{ $t('label.permissions') }}
            </span>
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-permissions-button
                :resource="`corteza::compose:module/${namespace.namespaceID}/*`"
                :button-label="$t('label.module.single', 'Module')"
                :show-button-icon="false"
                class="dropdown-item"
              />
            </li>
            <li>
              <c-permissions-button
                :resource="`corteza::compose:module-field/${namespace.namespaceID}/*/*`"
                :button-label="$t('label.field')"
                :show-button-icon="false"
                class="dropdown-item"
              />
            </li>
            <li>
              <c-permissions-button
                :resource="`corteza::compose:record/${namespace.namespaceID}/*/*`"
                :button-label="$t('label.record')"
                :show-button-icon="false"
                class="dropdown-item"
              />
            </li>
          </ul>
        </div>
      </template>

      <template #actions="{ item: m }">
        <related-pages
          :namespace="namespace"
          :module="m"
          size="sm"
          boundary="scrollParent"
          class="d-inline-block"
        />

        <div
          v-if="m.canGrant"
          class="dropdown d-inline-block ms-2"
        >
          <button
            data-test-id="dropdown-permissions"
            class="btn btn-extra-light btn-sm dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2"
            type="button"
            data-bs-toggle="dropdown"
            :title="$t('permissions.resources.compose.module.tooltip')"
            aria-expanded="false"
          >
            <font-awesome-icon :icon="['fas', 'lock']" />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-permissions-button
                :title="m.name || m.handle || m.moduleID"
                :target="m.name || m.handle || m.moduleID"
                :resource="`corteza::compose:module/${namespace.namespaceID}/${m.moduleID}`"
                :button-label="$t('label.module.single', 'Module')"
                :show-button-icon="false"
                class="dropdown-item"
              />
            </li>
            <li>
              <c-permissions-button
                :title="m.name || m.handle || m.moduleID"
                :target="m.name || m.handle || m.moduleID"
                :resource="`corteza::compose:module-field/${namespace.namespaceID}/${m.moduleID}/*`"
                :button-label="$t('label.field')"
                :show-button-icon="false"
                all-specific
                class="dropdown-item"
              />
            </li>
            <li>
              <c-permissions-button
                :title="m.name || m.handle || m.moduleID"
                :target="m.name || m.handle || m.moduleID"
                :resource="`corteza::compose:record/${namespace.namespaceID}/${m.moduleID}/*`"
                :button-label="$t('label.record')"
                :show-button-icon="false"
                all-specific
                class="dropdown-item"
              />
            </li>
          </ul>
        </div>

        <div
          class="dropdown d-inline-block"
        >
          <button
            class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2 ms-2"
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
              <router-link
                data-test-id="button-all-records"
                :to="{name: 'admin.modules.record.list', params: { moduleID: m.moduleID }}"
                class="dropdown-item"
              >
                <font-awesome-icon
                  :icon="['fas', 'columns']"
                  class="text-primary"
                />
                {{ $t('allRecords.label') }}
              </router-link>
            </li>
            <li>
              <c-input-confirm
                v-if="m.canDeleteModule"
                :text="$t('list.delete')"
                show-icon
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(m)"
              />
            </li>
          </ul>
        </div>
      </template>

      <template #name="{ item: m }">
        <div
          class="d-flex align-items-center"
        >
          <font-awesome-icon
            :icon="moduleIcon(m)"
            class="me-2"
          />
          {{ m.name }}
          <h5
            class="ms-2 mb-0"
          >
            <span
              v-if="Object.keys(m.labels || {}).includes('federation')"
              class="badge rounded-pill bg-primary"
            >
              {{ $t('federated') }}
            </span>
          </h5>
        </div>
      </template>

      <template #changedAt="{ item }">
        {{ $locFullDateTime(item.deletedAt || item.updatedAt || item.createdAt) }}
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useStore } from '../../../store'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import listHelpers from 'corteza-webapp-compose/src/mixins/listHelpers'
import RelatedPages from 'corteza-webapp-compose/src/components/Admin/Module/RelatedPages'
import Import from 'corteza-webapp-compose/src/components/Admin/Import'
import Export from 'corteza-webapp-compose/src/components/Admin/Export'

const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const props = defineProps({
  namespace: {
    type: compose.Namespace,
    required: true,
  },
})

const primaryKey = 'moduleID'

const sorting = ref({
  sortBy: 'name',
  sortDesc: false,
})

const creatingRecordPage = ref(false)

const modules = computed(() => store.module.set)
const pages = computed(() => store.page.set)

const fields = computed(() => [
  {
    key: 'name',
    label: t('list.columns.name'),
    sortable: true,
    tdClass: 'text-nowrap',
  },
  {
    key: 'handle',
    label: t('list.columns.handle'),
    sortable: true,
  },
  {
    key: 'changedAt',
    label: t('list.columns.changedAt'),
    sortable: true,
    class: 'text-end text-nowrap',
  },
  {
    key: 'actions',
    label: '',
    tdClass: 'text-end text-nowrap actions',
  },
])

const recordPage = computed(() => {
  return (moduleID) => pages.value.find(p => p.moduleID === moduleID)
})

const { procListResults, encodeListParams, pagination, filterList, filter, toastSuccess, toastErrorHandler } = listHelpers.setup({
  store,
  router,
  route,
})

watch(() => props.namespace?.namespaceID, (nsID) => { if (nsID) filter.value.namespaceID = nsID }, { immediate: true })

onMounted(() => {
  document.title = t('label.app-name.module.list', { label: props.namespace?.name, interpolation: { escapeValue: false } })
})

function handleRowClicked ({ moduleID, canUpdateModule, canDeleteModule }) {
  if (!(canUpdateModule || canDeleteModule)) return
  router.push({
    name: 'admin.modules.edit',
    params: { moduleID },
    query: null,
  })
}

function encodeRouteParams () {
  const { query } = filter.value
  const { limit, pageCursor, page } = pagination.value
  return {
    query: {
      limit,
      ...sorting.value,
      query,
      page,
      pageCursor,
    },
  }
}

function items () {
  if (!filter.value.namespaceID) return Promise.resolve([])
  return procListResults(window.__composeAPI.moduleListCancellable(encodeListParams()))
}

function onImportSuccessful () {
  filterList()
  toastSuccess(t('notification.general.import.successful'))
}

function handleDelete (module) {
  store.module.delete(module).then(() => {
    const moduleRecordPage = pages.value.find(p => p.moduleID === module.moduleID)
    if (moduleRecordPage) {
      return store.page.delete({ ...moduleRecordPage, strategy: 'rebase' })
    }
  }).catch(toastErrorHandler(t('notification.module.deleteFailed')))
    .finally(() => {
      toastSuccess(t('notification.module.deleted'))
      filterList()
    })
}

function moduleIcon (m) {
  const type = m.config?.type || 'basic'
  if (type === 'datasource') return ['fas', 'cube']
  if (type === 'dbref') return ['fas', 'code-branch']
  return ['fas', 'database']
}
</script>
