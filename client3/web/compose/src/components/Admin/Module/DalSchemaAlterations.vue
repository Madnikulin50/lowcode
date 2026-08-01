<template>
  <div
    id="schema-alterations-modal"
    ref="modal"
    class="modal fade"
    :class="{ show: showModal }"
    :style="{ display: showModal ? 'block' : 'none', backgroundColor: showModal ? 'rgba(0,0,0,0.5)' : 'transparent' }"
    tabindex="-1"
  >
    <div class="modal-dialog modal-xl">
      <div class="modal-content">
        <div class="modal-header p-3 pb-0 border-bottom-0">
          <h5 class="modal-title">{{ t('title') }}</h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
          />
        </div>
        <div class="modal-body p-0 border-top-0 position-relative">
          <table
            class="table table-borderless mb-0"
            style="min-height: 300px; max-height: 75vh;"
          >
            <thead>
              <tr class="table-light">
                <th>{{ t('columns.alteration') }}</th>
                <th style="max-width: 300px;">{{ t('columns.change') }}</th>
                <th class="text-center">{{ t('columns.status') }}</th>
                <th style="min-width: 200px;" />
              </tr>
            </thead>

            <tbody v-if="sortedAlterations.length && !loading">
              <tr
                v-for="a in sortedAlterations"
                :key="a.alterationID"
                class="border-top"
                :class="{ 'bg-extra-light': a.alterationID === dependOnHover }"
                @mouseover="dependOnHover = a.dependsOn"
                @mouseleave="dependOnHover = undefined"
              >
                <td>{{ a.alterationID }}</td>
                <td>{{ stringifyParams(a.params) }}</td>
                <td class="text-center align-top">
                  <span
                    v-if="a.error"
                    class="badge bg-danger"
                  >{{ a.error || '' }}</span>
                  <span
                    v-else-if="a.completedAt"
                    class="badge bg-success"
                  >{{ t('resolved') }}</span>
                  <span
                    v-else-if="a.dependsOn"
                    class="badge bg-extra-light"
                  >{{ t('waitingFor', { id: a.dependsOn }) }}</span>
                </td>
                <td class="text-end">
                  <span
                    v-if="a.processing"
                    class="spinner-border spinner-border-sm text-primary"
                  />
                  <template v-else>
                    <c-input-confirm
                      v-if="!a.completedAt"
                      :disabled="!canResolve(a) || a.processing || processing"
                      :text="t('resolve')"
                      variant="primary"
                      size="sm"
                      class="mx-1"
                      @click.stop
                      @confirmed="onResolve(a)"
                    />
                    <c-input-confirm
                      v-if="!a.completedAt"
                      :disabled="!canDismiss(a) || a.processing || processing"
                      :text="t('dismiss')"
                      variant="outline-secondary"
                      size="sm"
                      class="mx-1"
                      @click.stop
                      @confirmed="onDismiss(a)"
                    />
                  </template>
                </td>
              </tr>
            </tbody>
          </table>

          <div
            v-if="!sortedAlterations.length || loading"
            class="position-absolute text-center mt-5 d-print-none"
            style="left: 0; right: 0;"
          >
            <span
              v-if="loading"
              class="spinner-border spinner-border-sm"
            />
            <p
              v-else-if="!sortedAlterations.length"
              class="mb-0 mx-2"
            >
              {{ t('noAlterations') }}
            </p>
          </div>
        </div>
        <div class="modal-footer">
          <button
            :disabled="processing"
            class="btn btn-outline-secondary text-primary border-0"
            @click="showModal = false"
          >
            {{ canResolveAlterations ? $t('label.cancel') : $t('label.close') }}
          </button>

          <c-input-confirm
            v-if="canResolveAlterations"
            :text="t('resolveAuto')"
            :processing="processing"
            variant="primary"
            :disabled="processing"
            size="md"
            @confirmed="onResolve()"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'

const prefixed$ = 'edit.schemaAlterations.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

const props = defineProps({
  modal: {
    type: Boolean,
    required: false,
  },
  module: {
    type: compose.Module,
    required: true,
  },
  batch: {
    type: Array,
    required: false,
    default: undefined,
  },
})

const $SystemAPI = window.__systemAPI

const showModal = ref(false)
const loading = ref(false)
const processing = ref(false)
const dependOnHover = ref(undefined)
const alterations = ref([])

const sortedAlterations = computed(() => {
  return alterations.value.toSorted((a, b) => (a.batchID || '').localeCompare(b.batchID || '') || (a.dependsOn || '').localeCompare(b.dependsOn || ''))
})

const canResolveAlterations = computed(() => {
  return sortedAlterations.value.some(a => canResolve(a))
})

watch(() => props.batch, (batch) => {
  if (batch && batch.length) {
    load(...batch)
  }
})

async function onDismiss (alteration) {
  processing.value = true
  alteration = alteration ? [alteration] : alterations.value

  const alterationID = []
  for (const a of alteration) {
    alterationID.push(a.alterationID)
    a.processing = true
  }

  $SystemAPI.dalSchemaAlterationDismiss({ alterationID })
    .then(() => {
      toastSuccess(t('notification.module.schemaAlterations.dismiss.success'))
    })
    .catch(toastErrorHandler(t('notification.module.schemaAlterations.dismiss.error')))
    .finally(() => {
      for (const a of alteration) {
        a.processing = false
      }
      load(...props.batch)
      processing.value = false
    })
}

async function onResolve (alteration) {
  processing.value = true
  alteration = alteration ? [alteration] : alterations.value

  const alterationID = []
  for (const a of alteration) {
    alterationID.push(a.alterationID)
    a.processing = true
  }

  $SystemAPI.dalSchemaAlterationApply({ alterationID })
    .then(() => {
      toastSuccess(t('notification.module.schemaAlterations.resolve.success'))
    })
    .catch(toastErrorHandler(t('notification.module.schemaAlterations.resolve.error')))
    .finally(() => {
      for (const a of alteration) {
        a.processing = false
      }
      load(...props.batch)
      processing.value = false
    })
}

async function load (...batchID) {
  if (!batchID || (batchID && !batchID.length)) {
    alterations.value = []
    return
  }

  loading.value = true

  return $SystemAPI.dalSchemaAlterationList({ batchID })
    .then(({ set }) => {
      alterations.value = set
      if (alterations.value.length) {
        showModal.value = true
      }
    })
    .catch(toastErrorHandler(t('notification.module.schemaAlterations.load.error')))
    .finally(() => {
      loading.value = false
    })
}

function stringifyParams (params) {
  switch (true) {
    case !!params.attributeAdd:
      return stringifyAttributeAddParams(params.attributeAdd)
    case !!params.attributeDelete:
      return stringifyAttributeDeleteParams(params.attributeDelete)
    case !!params.attributeReType:
      return stringifyAttributeReTypeParams(params.attributeReType)
    case !!params.attributeReEncode:
      return stringifyAttributeReEncodeParams(params.attributeReEncode)
    case !!params.modelAdd:
      return stringifyModelAddParams(params.modelAdd)
    case !!params.modelDelete:
      return stringifyModelDeleteParams(params.modelDelete)
  }
  throw new Error('Unknown alteration type')
}

function stringifyAttributeAddParams ({ attr = {} }) {
  return t('params.attribute.add', { ident: attr.ident, storeType: attr.store.type, attrType: attr.type.type })
}

function stringifyAttributeDeleteParams ({ attr = {} }) {
  return t('params.attribute.delete', { ident: attr.ident, storeType: attr.store.type })
}

function stringifyAttributeReTypeParams ({ attr = {}, to = {} }) {
  return t('params.attribute.reType', { ident: attr.ident, toType: to.type })
}

function stringifyAttributeReEncodeParams ({ attr = {}, to = {} }) {
  return t('params.attribute.reEncode', { ident: attr.ident, toType: to.type })
}

function stringifyModelAddParams ({ model = {} }) {
  return t('params.model.add', { ident: model.ident })
}

function stringifyModelDeleteParams ({ model = {} }) {
  return t('params.model.delete', { ident: model.ident })
}

function canDismiss (alteration) {
  if (alteration.completedAt) return false
  if (alteration.dependsOn) {
    return alterations.value.some(a => a.alterationID === alteration.dependsOn && !a.completedAt)
  }
  return true
}

function canResolve (alteration) {
  return canDismiss(alteration)
}

function toastSuccess (...args) {
  console.log('toastSuccess', ...args)
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
