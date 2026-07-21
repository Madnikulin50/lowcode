<template>
  <div class="container pt-2 pb-3">
    <c-content-header
      :title="$t('system.apigw.title')"
      class="mb-2"
    >
      <button
        v-if="routeID && canCreate"
        data-test-id="button-add"
        class="btn btn-primary"
        @click="$router.push({ name: 'system.apigw.new' })"
      >
        {{ $t('new') }}
      </button>

      <c-permissions-button
        v-if="routeID && canGrant"
        :title="route.endpoint || routeID"
        :target="route.endpoint || routeID"
        :resource="`corteza::system:apigw-route/${routeID}`"
      >
        <font-awesome-icon :icon="['fas', 'lock']" />
        {{ $t('permissions') }}
      </c-permissions-button>
    </c-content-header>

    <c-route-editor-info
      v-if="Object.keys(route).length"
      :route="route"
      :processing="info.processing"
      :success="info.success"
      :can-create="canCreate"
      @submit="onInfoSubmit"
      @delete="onInfoDelete"
    />

    <c-filters-stepper
      v-if="routeID"
      ref="stepperRef"
      :fetching="stepper.fetching"
      :processing="stepper.processing"
      :success="stepper.success"
      v-model:filters="filters"
      :available-filters="availableFilters"
      :steps="steps"
      @submit="onFiltersSubmit"
    />

    <c-profiler-route-hits
      v-if="routeID && showProfiler"
      :route="routeEndpoint"
      class="mt-3"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, inject } from 'vue'
import { useRouter, useRoute, onBeforeRouteUpdate, onBeforeRouteLeave } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual, cloneDeep } from 'lodash'
import { useStore } from 'corteza-webapp-admin/src/store'
import { useEditorHelpers } from 'corteza-webapp-admin/src/mixins/editorHelpers'
import { NoID } from 'corteza-lib/js/dist'
import CRouteEditorInfo from 'corteza-webapp-admin/src/components/Apigw/CRouteEditorInfo'
import CFiltersStepper from 'corteza-webapp-admin/src/components/Apigw/CFiltersStepper'
import CProfilerRouteHits from 'corteza-webapp-admin/src/components/Apigw/Profiler/CProfilerRouteHits'

const { t } = useI18n()
const router = useRouter()
const $route = useRoute()
const store = useStore()
const $Settings = inject('$Settings', {})
const { incLoader, decLoader, animateSuccess } = useEditorHelpers()

const props = defineProps({
  routeID: { type: String, required: false, default: undefined },
})

const route = ref({})
const initialRouteState = ref({})
const routeEndpoint = ref(undefined)
const info = reactive({ processing: false, success: false })
const stepper = reactive({ fetching: false, processing: false, success: false })
const filters = ref([])
const initialFiltersState = ref([])
const availableFilters = ref([])
const steps = ref([])
const stepperRef = ref(null)

const canCreate = computed(() => store.rbac.can('system/', 'apigw-route.create'))
const canGrant = computed(() => store.rbac.can('system/', 'grant'))
const showProfiler = computed(() => $Settings.get('apigw.profiler.enabled', false) && ($Settings.get('apigw.profiler.global', false) || filters.value.some(({ ref, enabled = false }) => ref === 'profiler' && enabled)))

watch(() => props.routeID, {
  immediate: true,
  handler() {
    routeEndpoint.value = undefined

    if (props.routeID) {
      fetchSteps()
      fetchRoute()
      fetchFilters()
    } else {
      route.value = {
        endpoint: '',
        method: 'GET',
      }
      initialRouteState.value = cloneDeep(route.value)
    }
  },
})

onBeforeRouteUpdate((to, from, next) => {
  checkUnsavedChanges(next, to)
})

onBeforeRouteLeave((to, from, next) => {
  checkUnsavedChanges(next, to)
})

function fetchRoute() {
  incLoader()

  window.__systemAPI.apigwRouteRead({ routeID: props.routeID, incFlags: 1 })
    .then((api) => {
      route.value = api
      initialRouteState.value = cloneDeep(api)
      routeEndpoint.value = btoa(api.endpoint)
    })
    .catch(window.__toastError(t('notification.gateway.fetch.error')))
    .finally(() => {
      decLoader()
    })
}

function onInfoSubmit(r) {
  info.processing = true

  if (props.routeID) {
    window.__systemAPI
      .apigwRouteUpdate(r)
      .then(() => {
        fetchRoute()
        animateSuccess(info)
        window.__toastSuccess(t('notification.gateway.update.success'))
      })
      .catch(window.__toastError(t('notification.gateway.update.error')))
      .finally(() => {
        info.processing = false
      })
  } else {
    window.__systemAPI
      .apigwRouteCreate(r)
      .then(({ routeID }) => {
        animateSuccess(info)
        window.__toastSuccess(t('notification.gateway.create.success'))

        router.push({
          name: 'system.apigw.edit',
          params: { routeID },
        })
      })
      .catch(window.__toastError(t('notification.gateway.create.error')))
      .finally(() => {
        info.processing = false
      })
  }
}

function onInfoDelete() {
  incLoader()

  if (route.value.deletedAt) {
    window.__systemAPI
      .apigwRouteUndelete({ routeID: props.routeID })
      .then(() => {
        fetchRoute()
        window.__toastSuccess(t('notification.gateway.undelete.success'))
      })
      .catch(window.__toastError(t('notification.gateway.undelete.error')))
      .finally(() => {
        decLoader()
      })
  } else {
    window.__systemAPI
      .apigwRouteDelete({ routeID: props.routeID })
      .then(() => {
        fetchRoute()
        route.value.deletedAt = new Date()
        window.__toastSuccess(t('notification.gateway.delete.success'))
        router.push({ name: 'system.apigw' })
      })
      .catch(window.__toastError(t('notification.gateway.delete.error')))
      .finally(() => {
        decLoader()
      })
  }
}

function onFiltersSubmit() {
  if (props.routeID) {
    stepper.processing = true

    Promise.all(filters.value.map(filter => {
      if (filter.created || filter.updated || filter.deleted) {
        filter.params = encodeParams(filter.params)
        filter.weight = filter.weight.toString()

        if (filter.filterID && filter.filterID !== NoID) {
          return filter.deleted ? deleteFilter(filter) : updateFilter(filter)
        } else {
          return filter.deleted ? Promise.resolve() : createFilter(filter)
        }
      }

      return Promise.resolve()
    })).then(async () => {
      await fetchFilters()

      animateSuccess(stepper)
      window.__toastSuccess(t('notification.gateway.filter.update.success'))
    })
      .catch(window.__toastError(t('notification.gateway.filter.update.error')))
      .finally(() => {
        stepper.processing = false
      })
  }
}

function createFilter(filter) {
  return window.__systemAPI.apigwFilterCreate({ ...filter, routeID: props.routeID })
}

function updateFilter(filter) {
  return window.__systemAPI.apigwFilterUpdate({ ...filter, routeID: props.routeID })
}

function deleteFilter({ filterID = '' }) {
  if (filterID) {
    return window.__systemAPI.apigwFilterDelete({ filterID })
  }
}

function fetchFilters() {
  incLoader()
  stepper.fetching = true

  window.__systemAPI.apigwFilterList({ routeID: props.routeID })
    .then(({ set = [] }) => {
      return setRouteFilters(set)
    })
    .catch(window.__toastError(t('notification.gateway.filter.fetch.error')))
    .finally(() => {
      decLoader()
      stepper.fetching = false
    })
}

function setRouteFilters(routeFilters = []) {
  return fetchAllAvailableFilters().then(() => {
    filters.value = (routeFilters || []).map(filter => {
      const f = { ...availableFilters.value.find((af) => af.ref === filter.ref) }
      f.params = decodeParams(f, { ...filter.params })
      f.weight = parseInt(filter.weight)
      f.filterID = filter.filterID
      f.enabled = !!filter.enabled
      return { ...f }
    })
    initialFiltersState.value = cloneDeep(filters.value)
  })
}

function decodeParams(filter = {}, values = {}) {
  const { params = [] } = filter
  return params.map(({ label, type }) => {
    return {
      label,
      type,
      value: values[label],
    }
  })
}

function encodeParams(params = []) {
  return params.reduce((result, p) => {
    result[p.label] = p.value
    return result
  }, {})
}

function fetchAllAvailableFilters() {
  incLoader()

  return window.__systemAPI.apigwFilterDefFilter()
    .then((api) => {
      availableFilters.value = api.map((f) => {
        return { ...f, ref: f.name, enabled: true, options: { checked: false } }
      })
    })
    .catch(window.__toastError(t('notification.gateway.filter.fetch.error')))
    .finally(() => {
      decLoader()
    })
}

function fetchSteps() {
  steps.value = ['prefilter', 'processer', 'postfilter']
}

function checkUnsavedChanges(next, to) {
  const isNewPage = $route.path.includes('/new') && to.name.includes('edit')
  const { deletedAt } = route.value || {}

  if (isNewPage || deletedAt) {
    next(true)
  } else if (!to.name.includes('edit')) {
    const routeState = !isEqual(route.value, initialRouteState.value)
    const filtersState = !isEqual(filters.value, initialFiltersState.value)

    next((routeState || filtersState) ? window.confirm(t('editor.unsavedChanges')) : true)
  }
}
</script>
