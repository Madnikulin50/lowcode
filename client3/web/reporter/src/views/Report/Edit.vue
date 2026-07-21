<template>
  <div class="py-3">
    <Teleport to="#topbar-title">
      <div class="d-flex w-100">{{ pageTitle }}</div>
    </Teleport>

    <Teleport to="#topbar-tools">
      <div v-if="report && !isNew" class="btn-group btn-group-sm">
        <button
          class="btn btn-primary d-flex align-items-center justify-content-center"
          data-test-id="button-report-builder"
          :disabled="!report.canUpdateReport"
          @click="$router.push(reportBuilder)"
        >
          {{ t('report.builder') }}
          <font-awesome-icon class="ms-2" :icon="['fas', 'tools']" />
        </button>

        <button
          class="btn btn-primary d-flex align-items-center justify-content-center"
          style="margin-left:2px"
          :title="t('tooltip.view-report')"
          :disabled="!canRead"
          @click="$router.push(reportViewer)"
        >
          <font-awesome-icon :icon="['far', 'eye']" />
        </button>
      </div>
    </Teleport>

    <div v-if="report" class="container">
      <div class="row g-0">
        <div class="col">
          <div class="card shadow-sm">
            <div v-if="!isNew" class="card-header bg-white border-bottom">
              <div class="row g-0 align-items-center">
                <div>
                  <button
                    v-if="canCreate"
                    class="btn btn-primary btn-lg me-1"
                    data-test-id="button-create-report"
                    @click="$router.push({ name: 'report.create' })"
                  >
                    {{ t('new-report') }}
                  </button>

                  <c-permissions-button
                    v-if="canGrant"
                    :title="report.meta.name || report.handle || report.reportID"
                    :target="report.meta.name || report.handle || report.reportID"
                    :resource="`corteza::system:report/${report.reportID}`"
                    :button-label="t('permissions')"
                    class="btn-lg ms-1"
                  />
                </div>
              </div>
            </div>

            <div class="container-fluid py-3">
              <div class="row">
                <div class="col-12 col-lg-6 col-xl-4">
                  <div class="mb-3">
                    <label class="text-primary form-label">{{ t('name-with-star') }}</label>
                    <input
                      v-model="report.meta.name"
                      class="form-control"
                      data-test-id="input-name"
                      :placeholder="t('name')"
                      required
                      :class="{ 'is-invalid': nameState === false }"
                      @input="handleDetectStateChange"
                    />
                  </div>
                </div>
                <div class="col-12 col-lg-6 col-xl-4">
                  <div class="mb-3">
                    <label class="text-primary form-label">{{ t('handle-with-star') }}</label>
                    <input
                      v-model="report.handle"
                      class="form-control"
                      data-test-id="input-handle"
                      :placeholder="t('placeholder-handle')"
                      required
                      :class="{ 'is-invalid': handleState === false }"
                      @input="handleDetectStateChange"
                    />
                    <div v-if="handleState === false" class="invalid-feedback" data-test-id="input-handle-invalid-state">
                      {{ t('invalid-handle-characters') }}
                    </div>
                  </div>
                </div>
              </div>

              <div class="mb-3">
                <label class="text-primary form-label">{{ t('description') }}</label>
                <textarea
                  v-model="report.meta.description"
                  class="form-control"
                  data-test-id="input-description"
                  :placeholder="t('report.description')"
                  rows="5"
                  @input="handleDetectStateChange"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="d-flex align-items-center justify-content-center w-100 h-100">
      <div class="spinner-border" />
    </div>

    <Teleport to="#report-toolbar">
      <editor-toolbar
        :back-link="{ name: 'report.list' }"
        :hide-delete="isNew"
        :delete-disabled="!canDelete"
        :save-disabled="!canSave"
        :clone-disabled="!canSave"
        :processing="processing"
        :processing-save="processingSave"
        :processing-delete="processingDelete"
        :processing-clone="processingClone"
        @clone="handleReportCloning"
        @delete="handleDeleteFn"
        @save="handleSave"
      />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, inject } from 'vue'
import { useRouter, useRoute, onBeforeRouteUpdate, onBeforeRouteLeave } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import { system } from 'corteza-lib/js/dist'
import { handle } from 'corteza-lib/vue/dist'
import { isEqual } from 'lodash'
import { useReportHelpers } from '../../mixins/report'
import EditorToolbar from '../../components/EditorToolbar.vue'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const toast = useToast()
const toastSuccess = toast.success
const toastErrorHandler = toast.toastErrorHandler
const { fetchReport, handleDelete: doDelete, handleClone: doClone } = useReportHelpers()

const processing = ref(false)
const processingSave = ref(false)
const processingDelete = ref(false)
const processingClone = ref(false)
const report = ref(undefined)
const initialReportState = ref(undefined)
const detectStateChange = ref(false)

const reportID = computed(() => route.params.reportID)
const isNew = computed(() => !reportID.value)

const can = inject('can', () => false)

const canGrant = computed(() => can('system/', 'grant'))
const canCreate = computed(() => can('system/', 'report.create'))
const canRead = computed(() => report.value ? report.value.canReadReport : false)
const canDelete = computed(() => report.value ? report.value.canDeleteReport : false)
const canUpdate = computed(() => isNew.value ? canCreate.value : (report.value && report.value.canUpdateReport) || false)

const pageTitle = computed(() => isNew.value ? t('report.create') : t('report.edit'))

const reportBuilder = computed(() => report.value ? { name: 'report.builder', params: { reportID: report.value.reportID } } : undefined)
const reportViewer = computed(() => report.value ? { name: 'report.view', params: { reportID: report.value.reportID } } : undefined)

const nameState = computed(() => {
  const { name = '' } = report.value?.meta || {}
  return name.length ? null : false
})

const handleState = computed(() => handle.handleState(report.value?.handle))

const canSave = computed(() => canUpdate.value && ![nameState.value, handleState.value].includes(false))

watch(reportID, (id) => {
  if (id) {
    fetchReport(id, toastErrorHandler, t).then(r => {
      report.value = r
      initialReportState.value = r.clone()
    })
  } else {
    report.value = new system.Report()
    initialReportState.value = new system.Report()
  }
}, { immediate: true })

function handleDetectStateChange() {
  detectStateChange.value = true
}

function checkUnsavedChart(next) {
  if (report.value.deletedAt) return next(true)

  const rState = {
    handle: report.value.handle,
    meta: {
      name: report.value.meta.name,
      description: report.value.meta.description,
    },
  }
  const initState = {
    handle: initialReportState.value.handle,
    meta: {
      name: initialReportState.value.meta.name,
      description: initialReportState.value.meta.description,
    },
  }
  next(!isEqual(rState, initState) ? window.confirm(t('unsavedChanges')) : true)
}

onBeforeRouteUpdate((to, from, next) => checkUnsavedChart(next))
onBeforeRouteLeave((to, from, next) => checkUnsavedChart(next))

async function handleSave() {
  processing.value = true
  processingSave.value = true
  try {
    const r = await doSave(report.value, isNew.value, toastSuccess, toastErrorHandler, t, router)
    if (r) {
      report.value = r
      initialReportState.value = r.clone()
      detectStateChange.value = false
    }
  } finally {
    processing.value = false
    processingSave.value = false
  }
}

async function doSave(reportData, isNewReport, success, error, trans, rtr) {
  const { blocks } = reportData
  const cleaned = {
    ...reportData,
    blocks: blocks.map(b => ({
      ...b,
      elements: b.elements.map(el => {
        const { dataframes, ...rest } = el
        return rest
      }),
    })),
  }

  if (isNewReport) {
    const r = await window.__systemAPI.reportCreate(cleaned)
    success(trans('notification:report.created'))
    rtr.push({ name: 'report.builder', params: { reportID: r.reportID } })
    return new system.Report(r)
  } else {
    const r = await window.__systemAPI.reportUpdate(cleaned)
    success(trans('notification:report.updated'))
    return new system.Report(r)
  }
}

async function handleDeleteFn() {
  processing.value = true
  processingDelete.value = true
  try {
    await doDelete(report.value, toastSuccess, toastErrorHandler, t, router)
  } finally {
    processing.value = false
    processingDelete.value = false
  }
}

async function handleReportCloning() {
  processingClone.value = true
  try {
    const r = await doClone(report.value, toastSuccess, toastErrorHandler, t)
    if (r) {
      router.push({ name: 'report.builder', params: { reportID: r.reportID } })
    }
  } finally {
    processingClone.value = false
  }
}
</script>
