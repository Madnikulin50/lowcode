<template>
  <div
    id="federation-modal"
    ref="modal"
    class="modal fade"
    tabindex="-1"
  >
    <div class="modal-dialog modal-lg">
      <div class="modal-content">
        <div class="modal-header p-3 pb-0 border-bottom-0">
          <h5 class="modal-title">{{ federationModalTitle }}</h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
          />
        </div>
        <div class="modal-body p-0 border-top-0">
          <ul class="nav nav-tabs" role="tablist">
            <li class="nav-item" role="presentation">
              <button
                class="nav-link active"
                id="upstream-tab"
                data-bs-toggle="tab"
                data-bs-target="#upstream"
                type="button"
                role="tab"
              >
                {{ $t('edit.federationSettings.upstream.title') }}
              </button>
            </li>
            <li class="nav-item" role="presentation">
              <button
                class="nav-link"
                id="downstream-tab"
                data-bs-toggle="tab"
                data-bs-target="#downstream"
                type="button"
                role="tab"
              >
                {{ $t('edit.federationSettings.downstream.title') }}
              </button>
            </li>
          </ul>
          <div class="tab-content">
            <div
              id="upstream"
              class="tab-pane active"
              role="tabpanel"
            >
              <ul class="list-group overflow-auto server-list">
                <li
                  v-for="f in servers"
                  :key="f.nodeID"
                  class="list-group-item border d-flex flex-column"
                  :class="{ 'border border-primary': f.nodeID === upstream.active }"
                  @click="upstream.active = f.nodeID"
                >
                  <p class="mb-0 text-truncate">{{ f.name }}</p>
                  <small class="text-truncate">{{ f.baseURL }}</small>
                </li>
              </ul>

              <div
                v-if="upstream.processing"
                class="d-flex flex-grow-1 justify-content-center align-items-center"
              >
                <span class="spinner-border spinner-border-sm text-primary" />
              </div>

              <div
                v-else-if="upstream[upstream.active]"
                class="list-group flex-grow-1 ms-4"
              >
                <div v-if="upstream[upstream.active].canManageModule">
                  <p>{{ $t('edit.federationSettings.upstream.description') }}</p>
                  <div class="mb-3">
                    <label class="form-label col-sm-4 col-lg-5">{{ $t('edit.federationSettings.upstream.copyFrom') }}</label>
                    <select
                      :key="upstream.active"
                      v-model="upstream[upstream.active].copy"
                      class="form-select form-control"
                      @change="copyUpstreamFrom"
                    >
                      <option
                        v-for="opt in upstream[upstream.active].options"
                        :key="opt.nodeID"
                        :value="opt.nodeID"
                      >{{ opt.name }}</option>
                    </select>
                  </div>

                  <div class="form-check mb-2">
                    <input
                      :id="`upstream-allFields-${upstream.active}`"
                      type="checkbox"
                      class="form-check-input"
                      :checked="upstream[upstream.active].allFields"
                      @change="selectAllFields($event, 'upstream')"
                    >
                    <label
                      class="form-check-label"
                      :for="`upstream-allFields-${upstream.active}`"
                    ><strong>{{ $t('edit.federationSettings.upstream.allFields') }}</strong></label>
                  </div>

                  <div class="overflow-auto">
                    <div
                      v-for="f in upstream[upstream.active].fields"
                      :key="`${upstream.active}${f.name}`"
                      class="form-check mb-2"
                    >
                      <input
                        :id="`upstream-field-${upstream.active}-${f.name}`"
                        v-model="f.value"
                        type="checkbox"
                        class="form-check-input"
                        @change="checkChange($event, 'upstream')"
                      >
                      <label
                        class="form-check-label"
                        :for="`upstream-field-${upstream.active}-${f.name}`"
                      >{{ f.label }}</label>
                    </div>
                  </div>
                </div>

                <div
                  v-else
                  class="d-flex flex-grow-1 align-items-center justify-content-center"
                >
                  {{ $t('edit.federationSettings.noPermission') }}
                </div>
              </div>

              <div
                v-else
                class="d-flex flex-grow-1 align-items-center justify-content-center"
              >
                {{ $t('edit.federationSettings.noNodes') }}
              </div>
            </div>

            <div
              id="downstream"
              class="tab-pane"
              role="tabpanel"
            >
              <ul class="list-group overflow-auto server-list">
                <li
                  v-for="f in servers"
                  :key="f.nodeID"
                  class="list-group-item border d-flex flex-column"
                  :class="{ 'border border-primary': f.nodeID === downstream.active }"
                  @click="downstream.active = f.nodeID"
                >
                  <p class="mb-0 text-truncate">{{ f.name }}</p>
                  <small class="text-truncate">{{ f.baseURL }}</small>
                </li>
              </ul>

              <div
                v-if="downstream.processing"
                class="d-flex flex-grow-1 justify-content-center align-items-center"
              >
                <span class="spinner-border spinner-border-sm text-primary" />
              </div>

              <div
                v-else-if="downstream[downstream.active]"
                class="list-group flex-grow-1 ms-4"
              >
                <div class="mb-3">
                  <select
                    :key="downstream.active"
                    v-model="downstream[downstream.active].module"
                    class="form-select form-control w-50"
                  >
                    <option
                      v-for="opt in downstream[downstream.active].options"
                      :key="opt.moduleID"
                      :value="opt.moduleID"
                    >{{ opt.name }}</option>
                  </select>
                </div>

                <div
                  v-if="downstream[downstream.active].module"
                  class="mb-2"
                >
                  <p>{{ $t('edit.federationSettings.downstream.description') }}</p>
                  <div class="form-check">
                    <input
                      :id="`downstream-allFields-${downstream.active}`"
                      type="checkbox"
                      class="form-check-input"
                      :checked="downstream[downstream.active].allFields[downstream[downstream.active].module]"
                      @change="selectAllFields($event, 'downstream')"
                    >
                    <label
                      class="form-check-label"
                      :for="`downstream-allFields-${downstream.active}`"
                    ><strong>{{ $t('edit.federationSettings.downstream.allFields') }}</strong></label>
                  </div>
                </div>
                <div
                  v-if="downstream[downstream.active].module"
                  class="overflow-auto"
                >
                  <div
                    v-for="sharedModuleFields in activeSharedModules"
                    :key="`${downstream.active}_${sharedModuleFields.name}`"
                    class="d-flex align-items-center justify-content-between"
                  >
                    <div class="form-check my-2">
                      <input
                        :id="`downstream-field-${downstream.active}-${sharedModuleFields.name}`"
                        v-model="sharedModuleFields.map"
                        type="checkbox"
                        class="form-check-input"
                        @change="checkChange($event, 'downstream')"
                      >
                      <label
                        class="form-check-label"
                        :for="`downstream-field-${downstream.active}-${sharedModuleFields.name}`"
                      >{{ sharedModuleFields.label }}</label>
                    </div>

                    <select
                      v-show="sharedModuleFields.map"
                      :key="`${downstream.active}_${sharedModuleFields.name}`"
                      v-model="sharedModuleFields.mapped"
                      class="form-select form-control w-50"
                      @change="setUpdated('downstream')"
                    >
                      <option
                        v-for="opt in transformedModuleFields"
                        :key="opt.name"
                        :value="opt.name"
                      >{{ opt.label }}</option>
                    </select>
                  </div>
                </div>
              </div>

              <div
                v-else
                class="d-flex flex-grow-1 align-items-center justify-content-center"
              >
                {{ $t('edit.federationSettings.noNodes') }}
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button
            class="btn btn-primary"
            @click="handleFederationSettingsSave()"
          >
            {{ $t('label.saveAndClose') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'module',
  },
})

const props = defineProps({
  modal: {
    type: Boolean,
    required: false,
  },
  module: {
    type: compose.Module,
    required: true,
  },
})

const emit = defineEmits(['change'])

const $FederationAPI = window.__federationAPI

const showModal = ref(false)
const servers = ref([])
const moduleFields = ref([])
const sharedModule = ref(null)

const sharedModules = reactive({})
const sharedModulesMapped = reactive({})
const exposedModules = reactive({})
const moduleMappings = reactive({})

const downstream = reactive({
  active: undefined,
  processing: false,
  enabled: false,
})

const upstream = reactive({
  active: undefined,
  processing: false,
  enabled: false,
})

const activeSharedModules = computed(() => {
  if (!downstream[downstream.active]?.module) return []
  return (sharedModulesMapped[downstream.active] || {})[downstream[downstream.active].module] || []
})

const transformedModuleMappings = computed(() => {
  const tf = transformFields(moduleFields.value)
  const mm = ((sharedModules[downstream.active] || {})[sharedModule.value] || {}).fields || []
  return tf.map((el) => {
    el.origin.value = false
    if (mm.find((e) => e.origin.name === el.origin.name)) {
      el.destination.name = el.origin.name
      el.origin.value = true
    }
    return el
  })
})

const transformedModuleFields = computed(() => {
  return [
    { name: null, label: t('edit.federationSettings.pickModuleField') },
    ...transformedModuleMappings.value.map((el) => ({
      name: el.origin.name,
      label: el.origin.label,
    })),
  ]
})

const federationModalTitle = computed(() => {
  const { handle } = props.module
  return handle ? t('edit.federationSettings.specificTitle', { handle }) : t('edit.federationSettings.title')
})

watch(() => props.modal, {
  immediate: true,
  handler (show = false) {
    showModal.value = show
  },
})

watch(() => props.module.fields, {
  immediate: true,
  handler (fields) {
    moduleFields.value = fields
      .map(f => ({ kind: f.kind, name: f.name, label: f.label, isMulti: f.isMulti, value: false, map: null }))
      .sort((a, b) => a.label.localeCompare(b.label))
  },
})

watch(() => upstream.active, {
  handler (nodeID) {
    getNodeUpstream(nodeID)
  },
})

watch(() => downstream.active, {
  handler (nodeID) {
    getNodeDownstream(nodeID)
  },
})

onMounted(() => {
  preload()
})

onBeforeUnmount(() => {
  setDefaultValues()
})

async function preload () {
  await $FederationAPI.nodeSearch({ status: 'paired' })
    .then(({ set = [] }) => {
      servers.value = set.filter(({ canManageNode }) => canManageNode)
    })
    .catch(toastErrorHandler(t('edit.federationSettings.error.fetch.node')))

  for (const node of servers.value) {
    await loadExposedModules(node.nodeID).catch(toastErrorHandler(t('edit.federationSettings.error.fetch.exposed')))
    await loadSharedModules(node.nodeID).catch(toastErrorHandler(t('edit.federationSettings.error.fetch.shared')))
    await loadModuleMappings(node.nodeID).catch(toastErrorHandler(t('edit.federationSettings.error.fetch.mmap')))
  }

  Object.assign(sharedModulesMapped, getSharedModulesMapped())

  if (servers.value.length) {
    downstream.active = servers.value[0].nodeID
    upstream.active = servers.value[0].nodeID
  }
}

function getSharedModulesMapped () {
  const list = {}
  for (const nodeID in sharedModules) {
    list[nodeID] = {}
    for (const sm of sharedModules[nodeID]) {
      let f = [...sm.fields].sort((a, b) => a.label.localeCompare(b.label))
      const mappedFields = ((moduleMappings[nodeID] || {})[sm.moduleID] || {}).fields || []

      f = f.map((el) => {
        let found = false
        let mapped = (moduleFields.value.find(({ name }) => name === el.name) || {}).name || null
        if (mappedFields) {
          const m = mappedFields.find((mf) => el.name === mf.origin.name)
          mapped = ((m || {}).destination || {}).name || null
          found = !!mapped
        }
        return { ...el, map: found, mapped }
      })

      list[nodeID][sm.moduleID] = f
    }
  }
  return list
}

async function handleFederationSettingsSave () {
  for (const nodeID in sharedModulesMapped) {
    for (const moduleID in sharedModulesMapped[nodeID]) {
      const crtModule = sharedModules[nodeID].find(m => m.moduleID === moduleID)
      if (!crtModule || !crtModule.updated) continue
      const fields = toModuleMappingFormat(sharedModulesMapped[nodeID][moduleID])
      const payload = {
        nodeID,
        moduleID,
        composeModuleID: props.module.moduleID,
        composeNamespaceID: props.module.namespaceID,
        fields,
      }
      await persistModuleMappings(payload)
        .then(() => { crtModule.updated = false })
        .catch(toastErrorHandler(t('edit.federationSettings.error.persist.mmap')))
    }
  }

  const nodes = servers.value.map(s => s.nodeID)
  for (const nodeID of nodes) {
    if (!upstream[nodeID] || !upstream[nodeID].updated) continue
    const fields = (upstream[nodeID].fields || []).filter((el) => el.value)
    const payload = {
      nodeID,
      moduleID: (exposedModules[nodeID] || {}).moduleID,
      composeModuleID: props.module.moduleID,
      composeNamespaceID: props.module.namespaceID,
      name: props.module.name,
      handle: props.module.handle,
      fields,
    }
    const response = await persistExposedModule(payload)
      .then(() => { upstream[nodeID].updated = false })
      .catch(toastErrorHandler(t('edit.federationSettings.error.persist.exposed')))
    if (!response || !response.moduleID) return
    exposedModules[nodeID] = response
  }
}

function toModuleMappingFormat (fields) {
  return fields
    .filter((el) => el.map)
    .filter((el) => !!el.mapped)
    .map((el) => ({
      origin: { kind: el.kind, name: el.name, label: el.label, isMulti: el.isMulti },
      destination: { kind: el.kind, name: el.mapped, label: el.label, isMulti: el.isMulti },
    }))
}

function transformFields (fields) {
  return fields.map((el) => ({
    origin: { kind: el.kind, name: el.name, label: el.label || 'N/A', isMulti: false },
    destination: { kind: el.kind, name: '', label: '', isMulti: false },
  }))
}

function getNodeUpstream (nodeID) {
  if (upstream[nodeID]) return
  upstream.processing = true

  const exposedModule = exposedModules[nodeID] || {}
  const fields = (moduleFields.value || []).map(f => ({ ...f, value: false }))
  const exposedFields = exposedModule.fields || []
  exposedFields.forEach(({ name }) => {
    const found = fields.find(f => f.name === name)
    if (found) found.value = true
  })

  const u = {
    options: [
      { nodeID: null, name: t('edit.federationSettings.pickServer') },
      ...servers.value.filter(s => s.nodeID !== nodeID),
    ],
    copy: null,
    allFields: false,
    fields,
    updated: false,
    canManageModule: !!exposedModule.canManageModule || !!(servers.value.find(s => s.nodeID === nodeID) || {}).canCreateModule,
  }
  u.allFields = u.fields.filter(f => f.value).length === u.fields.length
  upstream[nodeID] = u
  upstream.processing = false
}

function getNodeDownstream (nodeID) {
  if (downstream[nodeID]) return
  downstream.processing = true

  const fields = (moduleFields.value || []).map(f => ({ ...f, value: false }))
  const d = {
    options: [
      { moduleID: null, name: t('edit.federationSettings.pickModule') },
      ...Object.values(sharedModules[nodeID] || {})
        .filter(({ canMapModule }) => canMapModule)
        .map(m => ({ moduleID: m.moduleID, name: m.name })),
    ],
    module: ((sharedModules[nodeID] || []).find(({ handle }) => handle === props.module.handle) || {}).moduleID || null,
    allFields: {},
    fields,
  }

  Object.entries(sharedModulesMapped[nodeID] || {}).forEach(([key, value]) => {
    d.allFields[key] = value.length ? value.filter(f => f.map).length === value.length : false
  })

  downstream[nodeID] = d
  downstream.processing = false
}

function selectAllFields (event, target) {
  const active = target === 'upstream' ? upstream.active : downstream.active
  const value = event.target ? event.target.checked : event

  if (target === 'upstream') {
    upstream[active].fields = upstream[active].fields.map(f => ({ ...f, value }))
    upstream[active].allFields = value
    upstream[active].updated = true
  } else if (target === 'downstream') {
    sharedModulesMapped[active][downstream[active].module] = sharedModulesMapped[active][downstream[active].module].map(f => ({ ...f, map: value }))
    downstream[active].allFields[downstream[active].module] = value
    sharedModules[active].find(m => m.moduleID === downstream[active].module).updated = true
  }
}

function setUpdated (target) {
  const active = target === 'upstream' ? upstream.active : downstream.active
  if (target === 'upstream') {
    upstream[active].updated = true
  } else if (target === 'downstream') {
    const mod = sharedModules[active]?.find(m => m.moduleID === downstream[active].module)
    if (mod) mod.updated = true
  }
}

function copyUpstreamFrom (nodeID) {
  upstream[upstream.active].fields = upstream[upstream.active].fields.map(f => {
    let value = false
    if (upstream[nodeID]) {
      value = !!upstream[nodeID].fields.find(({ name, value }) => name === f.name && value)
    } else {
      value = !!exposedModules[nodeID]?.fields.find(({ name }) => name === f.name)
    }
    return { ...f, value }
  })
}

function checkChange (event, target) {
  const active = target === 'upstream' ? upstream.active : downstream.active
  const value = event.target ? event.target.checked : event

  if (target === 'upstream') {
    upstream[active].allFields = value ? !upstream[active].fields.find(f => f.value === !value) : false
  } else if (target === 'downstream') {
    const allSameValue = !sharedModulesMapped[active][downstream[active].module]?.find(({ map }) => map === !value)
    downstream[active].allFields[downstream[active].module] = value ? allSameValue : false
  }
  setUpdated(target)
}

async function persistExposedModule (payload) {
  if (payload.moduleID) {
    return $FederationAPI.manageStructureUpdateExposed(payload)
  }
  return $FederationAPI.manageStructureCreateExposed(payload)
}

async function persistModuleMappings (payload) {
  return $FederationAPI.manageStructureCreateMappings(payload)
}

async function loadSharedModules (nodeID) {
  if (sharedModules[nodeID]) return
  await $FederationAPI.manageStructureListAll({ nodeID, shared: 1 })
    .then((data = []) => {
      sharedModules[nodeID] = data.map(d => ({ ...d, updated: false }))
    })
    .catch(toastErrorHandler(t('edit.federationSettings.error.fetch.shared')))
}

async function loadExposedModules (nodeID) {
  if (exposedModules[nodeID]) return
  await $FederationAPI.manageStructureListAll({ nodeID, exposed: 1 })
    .then((data = []) => {
      const exposedModule = data.find(({ composeModuleID }) => composeModuleID === props.module.moduleID)
      if (exposedModule) {
        exposedModules[nodeID] = exposedModule
      }
    })
    .catch(toastErrorHandler(t('edit.federationSettings.error.fetch.exposed')))
}

async function loadModuleMappings (nodeID) {
  if (moduleMappings[nodeID] || !sharedModules[nodeID]) return
  const mm = {}
  for (const { moduleID } of sharedModules[nodeID]) {
    mm[moduleID] = []
    await $FederationAPI.manageStructureReadMappings({ nodeID, moduleID, composeModuleID: props.module.moduleID })
      .then((data) => { mm[moduleID] = data })
      .catch(() => {})
  }
  moduleMappings[nodeID] = mm
}

function setDefaultValues () {
  showModal.value = false
  servers.value = []
  moduleFields.value = []
  sharedModule.value = null
  Object.keys(sharedModules).forEach(k => delete sharedModules[k])
  Object.keys(sharedModulesMapped).forEach(k => delete sharedModulesMapped[k])
  Object.keys(exposedModules).forEach(k => delete exposedModules[k])
  Object.keys(moduleMappings).forEach(k => delete moduleMappings[k])
  Object.keys(downstream).forEach(k => { if (typeof downstream[k] !== 'boolean') delete downstream[k] })
  Object.keys(upstream).forEach(k => { if (typeof upstream[k] !== 'boolean') delete upstream[k] })
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>

<style lang="scss" scoped>
.tab-content {
  min-height: 0;
  max-height: 70vh;
  display: flex;
}

.server-list {
  max-width: 35%;
}
</style>
