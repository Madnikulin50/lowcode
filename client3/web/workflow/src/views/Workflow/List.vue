<template>
  <div class="container-fluid d-flex flex-column py-3">
    <Teleport to="#topbar-title">
      {{ $t('workflow-list') }}
    </Teleport>

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="workflowList"
      :row-class="genericRowClass"
      :translations="{
        searchPlaceholder: $t('searchPlaceholder'),
        notFound: $t('resourceList.notFound'),
        noItems: $t('resourceList.noItems'),
        loading: $t('loading'),
        showingPagination: $t('resourceList.pagination.showing'),
        singlePluralPagination: $t('resourceList.pagination.single'),
        prevPagination: $t('resourceList.pagination.prev'),
        nextPagination: $t('resourceList.pagination.next'),
        resourceSingle: $t('workflow.single'),
        resourcePlural: $t('workflow.plural')
      }"
      clickable
      sticky-header
      class="h-100 flex-fill"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <button
          v-if="canCreate"
          data-test-id="button-create-workflow"
          class="btn btn-primary btn-lg"
          @click="$router.push({ name: 'workflow.create' })"
        >
          {{ $t('new-workflow') }}
        </button>

        <import
          v-if="canCreate"
          :disabled="importProcessing"
          class="d-flex"
          @import="importJSON"
        />

        <export size="lg" />

        <button
          class="btn btn-outline-secondary btn-lg"
          data-bs-toggle="modal"
          data-bs-target="#workflow-filter"
        >
          <font-awesome-icon
            :icon="['fas', 'filter']"
            :class="['me-1', { 'text-primary': labelFilterCount > 0 }]"
          />
          {{ $t('filter.label') }}
        </button>

        <c-permissions-button
          v-if="canGrant"
          resource="corteza::automation:workflow/*"
          :button-label="$t('permissions')"
          size="lg"
        />
      </template>

      <template #toolbar>
        <div class="col">
          <div class="btn-group btn-group-sm" role="group">
            <input type="radio" class="btn-check" name="subWorkflowRadio" id="subWorkflow0" :value="0" v-model="filter.subWorkflow" @change="filterList">
            <label class="btn btn-outline-primary" for="subWorkflow0">{{ $t('without') }}</label>
            <input type="radio" class="btn-check" name="subWorkflowRadio" id="subWorkflow1" :value="1" v-model="filter.subWorkflow" @change="filterList">
            <label class="btn btn-outline-primary" for="subWorkflow1">{{ $t('including') }}</label>
            <input type="radio" class="btn-check" name="subWorkflowRadio" id="subWorkflow2" :value="2" v-model="filter.subWorkflow" @change="filterList">
            <label class="btn btn-outline-primary" for="subWorkflow2">{{ $t('only') }}</label>
          </div>
          {{ $t('subworkflows') }}
        </div>
        <div class="col">
          <div class="btn-group btn-group-sm" role="group">
            <input type="radio" class="btn-check" name="disabledRadio" id="disabled0" :value="0" v-model="filter.disabled" @change="filterList">
            <label class="btn btn-outline-primary" for="disabled0">{{ $t('without') }}</label>
            <input type="radio" class="btn-check" name="disabledRadio" id="disabled1" :value="1" v-model="filter.disabled" @change="filterList">
            <label class="btn btn-outline-primary" for="disabled1">{{ $t('including') }}</label>
            <input type="radio" class="btn-check" name="disabledRadio" id="disabled2" :value="2" v-model="filter.disabled" @change="filterList">
            <label class="btn btn-outline-primary" for="disabled2">{{ $t('only') }}</label>
          </div>
          {{ $t('disabled') }}
        </div>
        <div class="col">
          <div class="btn-group btn-group-sm" role="group">
            <input type="radio" class="btn-check" name="deletedRadio" id="deleted0" :value="0" v-model="filter.deleted" @change="filterList">
            <label class="btn btn-outline-primary" for="deleted0">{{ $t('without') }}</label>
            <input type="radio" class="btn-check" name="deletedRadio" id="deleted1" :value="1" v-model="filter.deleted" @change="filterList">
            <label class="btn btn-outline-primary" for="deleted1">{{ $t('including') }}</label>
            <input type="radio" class="btn-check" name="deletedRadio" id="deleted2" :value="2" v-model="filter.deleted" @change="filterList">
            <label class="btn btn-outline-primary" for="deleted2">{{ $t('only') }}</label>
          </div>
          {{ $t('deleted') }}
        </div>
      </template>

      <template #name="{ item: w }">
        <div>
          {{ w.meta.name || w.handle }}
          <span
            v-if="w.meta.subWorkflow"
            class="badge bg-info ms-2"
          >
            {{ $t('subworkflow') }}
          </span>
        </div>
        <div
          v-for="group in getWorkflowLabels(w)"
          :key="'group-' + group.namespaceID"
          class="d-flex align-items-center flex-wrap gap-1 mt-2"
        >
          <span
            :title="$t('filter.namespace.label')"
            class="badge bg-primary"
            style="font-size: 90%;"
          >
            {{ group.namespaceName }}
          </span>

          <span
            v-for="mod in group.modules"
            :key="'mod-' + group.namespaceID + '-' + mod.id"
            :title="$t('filter.module.label')"
            class="badge bg-extra-light text-dark"
            style="font-size: 90%;"
          >
            {{ mod.name }}
          </span>
        </div>
      </template>

      <template #enabled="{ item: w }">
        <font-awesome-icon
          :icon="['fas', w.enabled ? 'check' : 'times']"
          :class="w.enabled ? 'text-primary' : 'text-extra-light'"
        />
      </template>

      <template #changedAt="{ item }">
        {{ $locFullDateTime(item.deletedAt || item.updatedAt || item.createdAt) }}
      </template>

      <template #actions="{ item: w }">
        <div class="dropdown">
          <button class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-primary border-0 py-2 ms-1 dropdown-toggle" data-bs-toggle="dropdown" aria-expanded="false">
            <font-awesome-icon :icon="['fas', 'ellipsis-v']" />
          </button>
          <ul class="dropdown-menu m-0">
            <li><button class="dropdown-item" @click="handleStatusChange(w)">
              <font-awesome-icon :icon="['fas', w.enabled ? 'toggle-off' : 'toggle-on']" />
              {{ statusText(w) }}
            </button></li>
            <li>
              <export
                data-test-id="button-export-workflow"
                :workflows="([w.workflowID])"
                :file-name="w.meta.name || w.handle"
                size="md"
              >
                <font-awesome-icon :icon="['fas', 'file-export']" />
              </export>
            </li>
            <li>
              <c-permissions-button
                v-if="w.canGrant"
                :tooltip="$t('permissions.resources.automation.workflow.tooltip')"
                :title="w.meta.name || w.handle || w.workflowID"
                :target="w.meta.name || w.handle || w.workflowID"
                :resource="`corteza::automation:workflow/${w.workflowID}`"
                :button-label="$t('permissions.ui.label')"
                class="dropdown-item"
              />
            </li>
            <li>
              <c-input-confirm
                v-if="w.canDeleteWorkflow && !w.deletedAt"
                borderless
                variant="link"
                size="md"
                show-icon
                :text="$t('delete')"
                text-class="p-1"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(w)"
              />
            </li>
            <li>
              <c-input-confirm
                v-if="w.canUndeleteWorkflow && w.deletedAt"
                borderless
                variant="link"
                size="md"
                show-icon
                :text="$t('undelete')"
                text-class="p-1"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(w)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>

    <workflow-filter-modal
      :namespace-labels="selectedNamespaceLabels"
      :module-labels="selectedModuleLabels"
      @apply="handleFilterApply"
    />
  </div>
</template>

<script setup>
import { ref, computed, inject, reactive, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useLabelsStore } from '../../store'
import Import from '../../components/Import.vue'
import Export from '../../components/Export.vue'
import WorkflowFilterModal from '../../components/WorkflowFilterModal.vue'
import { components } from 'corteza-lib/vue/dist'
import { useToast, useAuth, useRBACStore } from 'corteza-lib/vue/dist'

const { CResourceList } = components

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const toast = useToast()
const { auth } = useAuth()
const labelsStore = useLabelsStore()

const $SystemAPI = inject('$SystemAPI', {})
const $ComposeAPI = inject('$ComposeAPI', {})
const $AutomationAPI = inject('$AutomationAPI', {})
const $Settings = inject('$Settings', {})

const primaryKey = ref('workflowID')

const filter = reactive({
  query: '',
  deleted: 0,
  subWorkflow: 1,
  disabled: 0,
  labels: [],
})

const sorting = reactive({
  sortBy: 'changedAt',
  sortDesc: true,
})

const pagination = reactive({
  limit: 100,
  pageCursor: undefined,
  prevPage: '',
  nextPage: '',
  total: 0,
  page: 1,
})

const abortableRequests = ref([])
const cancelled = ref(false)

const newWorkflow = ref({})
const importProcessing = ref(false)
const selectedNamespaceLabels = ref([])
const selectedModuleLabels = ref([])
const tempQuery = ref(undefined)

const rbac = useRBACStore()
const can = computed(() => (resource, action) => rbac.can(resource, action))

const canGrant = computed(() => can.value('automation/', 'grant'))
const canCreate = computed(() => can.value('automation/', 'workflow.create'))

const labelFilterCount = computed(() => {
  return selectedNamespaceLabels.value.length + selectedModuleLabels.value.length
})

const tableFields = computed(() => [
  {
    key: 'name',
    label: t('columns.name'),
    sortable: false,
    tdClass: 'text-nowrap',
  },
  {
    key: 'enabled',
    label: t('columns.enabled'),
    sortable: true,
    class: 'text-center',
  },
  {
    key: 'steps',
    label: t('columns.steps'),
    class: 'text-center',
    formatter: steps => {
      return (steps || []).length
    },
  },
  {
    key: 'changedAt',
    label: t('columns.changedAt'),
    sortable: true,
    class: 'text-right text-nowrap',
  },
  {
    key: 'actions',
    label: '',
    tdClass: 'text-right text-nowrap actions',
  },
])

const workflowIDs = computed(() => workflows.map(({ workflowID }) => workflowID))

const userID = computed(() => {
  if (auth.user) {
    return auth.user.userID
  }
  return undefined
})

// Keep track of current workflows for label resolution
let workflows = []

watch(() => route.fullPath, () => {
  handleQueryParams()
})

onMounted(() => {
  handleQueryParams(true)
})

onBeforeUnmount(() => {
  abortRequests()
})

function handleQueryParams(initial = false) {
  let {
    limit = pagination.limit,
    pageCursor = pagination.pageCursor,
    prevPage = pagination.prevPage,
    nextPage = pagination.nextPage,
    total = pagination.total,
    page = pagination.page,
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

  const refresh = route.query.pageCursor !== pagination.pageCursor
  Object.assign(pagination, { limit, pageCursor, prevPage, nextPage, total, page })

  let { sortBy = sorting.sortBy, sortDesc = sorting.sortDesc, ...r2 } = r1
  sortDesc = sortDesc === true || sortDesc === 'true'

  if (!initial && (sortBy !== sorting.sortBy || sortDesc !== sorting.sortDesc)) {
    pagination.pageCursor = ''
    pagination.page = 1
  }
  Object.assign(sorting, { sortBy, sortDesc })

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

function filterList() {
  pagination.pageCursor = ''
  pagination.page = 1
  abortRequests()
  window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
}

function encodeListParams() {
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
    incTotal: !pageCursor || tempQuery.value,
  }
}

function encodeRouteParams() {
  const { limit, pageCursor, page } = pagination
  return {
    query: {
      limit,
      ...sorting,
      ...filter,
      page,
      pageCursor,
    },
  }
}

function procListResults(p, updateQuery = true) {
  abortRequests()

  const { response, cancel } = p
  abortableRequests.value.push(cancel)

  if (updateQuery && !tempQuery.value) {
    router.replace(encodeRouteParams())
  }

  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(async ([{ set, filter: responseFilter }]) => {
      if (responseFilter.incTotal) {
        pagination.total = responseFilter.total
      }

      if (tempQuery.value) {
        const query = tempQuery.value
        tempQuery.value = undefined
        router.replace({ query })
        return []
      }

      pagination.pageCursor = undefined
      pagination.nextPage = responseFilter.nextPage
      pagination.prevPage = responseFilter.prevPage

      return set
    }).catch(error => {
      if (error && error.message && error.message.includes('cancel')) {
        cancelled.value = true
      } else {
        toast.error(t('notification.list.load.error'))
      }
    }).finally(() => {
      cancelled.value = false
    })
}

function genericRowClass(item) {
  return { 'text-secondary': item && !!item.deletedAt }
}

function handleFilterApply({ namespaceLabels, moduleLabels }) {
  selectedNamespaceLabels.value = namespaceLabels || []
  selectedModuleLabels.value = moduleLabels || []
  updateLabelsFilter()
}

function updateLabelsFilter() {
  const labels = []

  if (selectedNamespaceLabels.value.length > 0) {
    labels.push(`ref_namespace=${JSON.stringify(selectedNamespaceLabels.value)}`)
  }

  if (selectedModuleLabels.value.length > 0) {
    labels.push(`ref_module=${JSON.stringify(selectedModuleLabels.value)}`)
  }

  filter.labels = labels
  filterList()
}

async function importJSON(workflows = []) {
  importProcessing.value = true

  const skippedWorkflows = []

  await Promise.all(workflows.map(({ triggers = [], ...wf }) => {
    return $AutomationAPI.workflowCreate({ ownedBy: userID.value, runAs: '0', ...wf })
      .then(({ workflowID }) => {
        return Promise.all(triggers.map(trigger => {
          return $AutomationAPI.triggerCreate({
            ...trigger,
            workflowID,
            workflowStepID: trigger.stepID,
            ownedBy: userID.value,
          })
        }))
      })
      .catch(({ message }) => {
        if (wf.handle) {
          skippedWorkflows.push(`${wf.handle}${message ? ' - ' + message : ''};`)
        }
      })
  }))
    .then(() => {
      if (skippedWorkflows.length) {
        toast.info(`${skippedWorkflows.join(' ')}`, t('notification.import.skipped-workflows'))
      } else {
        toast.success(t('notification.import.imported-workflows'))
      }
    })
    .catch((e) => toast.error(t('notification.import.failed-import'), e))

  window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))

  importProcessing.value = false
}

function workflowList() {
  return procListResults(
    $AutomationAPI.workflowListCancellable(encodeListParams()),
  ).then(result => {
    if (result && result.length > 0) {
      workflows = result
      return resolveLabelsForWorkflows(result).then(() => result)
    }
    return result
  })
}

async function resolveLabelsForWorkflows(workflows) {
  const namespaceIDs = new Set()
  const modules = []

  workflows.forEach(workflow => {
    if (workflow.labels?.ref_namespace) {
      const nsValues = Array.isArray(workflow.labels.ref_namespace)
        ? workflow.labels.ref_namespace
        : [workflow.labels.ref_namespace]
      nsValues.forEach(label => {
        const nsID = label.split('/')[1]
        if (nsID) namespaceIDs.add(nsID)
      })
    }

    if (workflow.labels?.ref_module) {
      const modValues = Array.isArray(workflow.labels.ref_module)
        ? workflow.labels.ref_module
        : [workflow.labels.ref_module]
      modValues.forEach(label => {
        const parts = label.split('/')
        const nsID = parts[1]
        const modID = parts[2]
        if (nsID) namespaceIDs.add(nsID)
        if (modID) modules.push({ moduleID: modID, namespaceID: nsID })
      })
    }
  })

  await Promise.all([
    labelsStore.resolveMultipleNamespaces(Array.from(namespaceIDs), $ComposeAPI),
    labelsStore.resolveMultipleModules(modules, $ComposeAPI),
  ])
}

function handleRowClicked(workflow) {
  router.push({ name: 'workflow.edit', params: { workflowID: workflow.workflowID } })
}

function handleDelete(workflow) {
  const { deletedAt = '' } = workflow
  const method = deletedAt ? 'workflowUndelete' : 'workflowDelete'
  const event = deletedAt ? 'undelete' : 'delete'
  const { workflowID } = workflow
  $AutomationAPI[method]({ workflowID })
    .then(() => {
      toast.success(t(`notification.${event}.success`))
      filterList()
    })
    .catch((e) => toast.error(t(`notification.${event}.failed`), e))
}

function statusText(w) {
  return w.enabled ? t('disable') : t('enable')
}

function getWorkflowLabels(workflow) {
  if (!workflow.labels) {
    return []
  }

  const namespaceIDs = []
  const modulesByNamespace = {}

  if (workflow.labels.ref_namespace) {
    const nsValues = Array.isArray(workflow.labels.ref_namespace)
      ? workflow.labels.ref_namespace
      : [workflow.labels.ref_namespace]

    nsValues.forEach(label => {
      const nsID = label.split('/')[1]
      if (nsID && !namespaceIDs.includes(nsID)) {
        namespaceIDs.push(nsID)
      }
    })
  }

  if (workflow.labels.ref_module) {
    const modValues = Array.isArray(workflow.labels.ref_module)
      ? workflow.labels.ref_module
      : [workflow.labels.ref_module]

    modValues.forEach(label => {
      const parts = label.split('/')
      const nsID = parts[1]
      const modID = parts[2]

      if (!nsID || !modID) return

      if (!namespaceIDs.includes(nsID)) {
        namespaceIDs.push(nsID)
      }

      if (!modulesByNamespace[nsID]) {
        modulesByNamespace[nsID] = []
      }

      const name = labelsStore.getModule(modID)
      modulesByNamespace[nsID].push({
        id: modID,
        name: name || modID,
      })
    })
  }

  if (namespaceIDs.length === 0) {
    return []
  }

  return namespaceIDs.map(nsID => {
    const name = labelsStore.getNamespace(nsID)
    return {
      namespaceID: nsID,
      namespaceName: name || nsID,
      modules: modulesByNamespace[nsID] || [],
    }
  })
}

function handleStatusChange({ workflowID, enabled }) {
  enabled = !enabled
  const notificationKey = enabled ? 'enable' : 'disable'

  $AutomationAPI.workflowRead({ workflowID }).then((w) => {
    return $AutomationAPI.workflowUpdate({ ...w, enabled }).then((w) => {
      toast.success(t(`notification.list.${notificationKey}.success`))
      filterList()
    })
  }).catch((e) => toast.error(t(`notification.list.${notificationKey}.failed`), e))
}

function abortRequests() {
  abortableRequests.value.forEach((cancel) => {
    cancel()
  })
  abortableRequests.value = []
}
</script>
