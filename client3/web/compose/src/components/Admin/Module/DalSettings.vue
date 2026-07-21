<template>
  <div v-if="module">
    <div class="mb-3">
      <label class="form-label text-primary">{{ t('connection.label') }}</label>
      <div class="form-text mb-2">{{ t('connection.description') }}</div>
      <c-input-select
        v-model="module.config.dal.connectionID"
        :options="connections"
        :get-option-key="getOptionKey"
        :disabled="processing"
        :clearable="false"
        :reduce="s => s.connectionID"
        :placeholder="t('connection.placeholder')"
        :get-option-label="getConnectionLabel"
      />
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('ident.label') }}</label>
      <div class="form-text mb-2">{{ identDescription }}</div>
      <input
        v-model="module.config.dal.ident"
        class="form-control form-control-sm"
        :placeholder="t('ident.placeholder')"
      >
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('module-fields.label') }}</label>
      <div class="form-text mb-2">{{ t('module-fields.description') }}</div>
      <dal-field-store-encoding
        v-for="({ field, storeIdent, label, isMulti }, i) in moduleFields"
        :key="i"
        :config="moduleFieldEncoding[field] || {}"
        :field="field"
        :label="label"
        :is-multi="isMulti"
        :default-strategy="moduleFieldDefaultEncodingStrategy"
        :store-ident="storeIdent"
        @change="applyModuleFieldStrategyConfig(field, $event)"
      />
    </div>

    <div
      v-if="isBasic"
      class="mb-3"
    >
      <label class="form-label text-primary">{{ t('system-fields.label') }}</label>
      <div class="form-text mb-2">{{ t('system-fields.description') }}</div>
      <div class="d-flex justify-content-end align-items-center flex-wrap">
        <div
          class="btn-group btn-group-sm mb-3"
          data-bs-toggle="buttons"
        >
          <label
            v-for="opt in optionsGroups"
            :key="opt.value"
            class="btn btn-outline-secondary"
            :class="{ active: selectedGroup === opt.value }"
          >
            <input
              v-model="selectedGroup"
              type="radio"
              class="btn-check"
              :value="opt.value"
              @change="applySelectedSystemFields"
            >
            {{ opt.text }}
          </label>
        </div>
      </div>

      <dal-field-store-encoding
        v-for="({ field, storeIdent, label, disabled }, i) in systemFields"
        :key="i"
        :config="systemFieldEncoding[field] || {}"
        :field="field"
        :label="label"
        :store-ident="storeIdent"
        :allow-omit-strategy="true"
        :disabled="disabled"
        @change="applySystemFieldStrategyConfig(field, $event)"
      />
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed, reactive, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import { moduleFieldStrategyConfig, systemFieldStrategyConfig, types } from './encoding-strategy'
import DalFieldStoreEncoding from 'corteza-webapp-compose/src/components/Admin/Module/DalFieldStoreEncoding'

const prefixed$ = 'edit.config.dal.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)
const tG = (key) => $t(key)

const props = defineProps({
  module: {
    type: compose.Module,
    required: true,
  },
})

const $SystemAPI = window.__systemAPI
const PrimaryConnType = 'corteza::system:primary-dal-connection'

const processing = ref(false)
const connections = ref([])
const moduleFields = ref([])
const moduleFieldEncoding = ref([])
const selectedGroup = ref('')

const systemFields = ref([])
const systemFieldEncoding = ref({})

const optionsGroups = ref([
  { text: t('system-fields.grouptypes.all'), value: 'all' },
  { text: t('system-fields.grouptypes.partition'), value: 'partition' },
  { text: t('system-fields.grouptypes.userReference'), value: 'user_reference' },
  { text: t('system-fields.grouptypes.timestamps'), value: 'timestamps' },
  { text: t('system-fields.grouptypes.extras'), value: 'extras' },
])

const moduleFieldDefaultEncodingStrategy = computed(() => types.JSON)

const moduleType = computed(() => {
  const ds = props.module.config.type ?? 'basic'
  return ds ?? false
})

const isDbRef = computed(() => moduleType.value === 'dbref')

const isBasic = computed(() => moduleType.value === 'basic')

const identDescription = computed(() => t('ident.description', { interpolation: { prefix: '{{{', suffix: '}}}' } }))

function initSystemFields () {
  const systemFieldEncodingConfig = props.module.config.dal.systemFieldEncoding || {}
  const sf = [
    { field: 'id', storeIdent: 'id', disabled: true },
    { field: 'namespaceID', storeIdent: 'rel_namespace', group: 'partition' },
    { field: 'moduleID', storeIdent: 'rel_module', group: 'partition' },
    { field: 'revision', storeIdent: 'revision', group: 'extras' },
    { field: 'meta', storeIdent: 'meta', group: 'extras' },
    { field: 'ownedBy', storeIdent: 'owned_by', group: 'user_reference' },
    { field: 'createdAt', storeIdent: 'created_at', group: 'timestamps' },
    { field: 'createdBy', storeIdent: 'created_by', group: 'user_reference' },
    { field: 'updatedAt', storeIdent: 'updated_at', group: 'timestamps' },
    { field: 'updatedBy', storeIdent: 'updated_by', group: 'user_reference' },
    { field: 'deletedAt', storeIdent: 'deleted_at', group: 'timestamps' },
    { field: 'deletedBy', storeIdent: 'deleted_by', group: 'user_reference' },
    ].map(sf => ({ ...sf, label: tG(`system.${sf.field}`) }))

  systemFields.value = sf
  systemFieldEncoding.value = sf.reduce((enc, { field }) => {
    enc[field] = systemFieldEncodingConfig[field] || {}
    return enc
  }, {})
}

initSystemFields()

watch(() => props.module.fields, (fields) => {
  moduleFields.value = []
  for (const f of props.module.fields) {
    const a = {
      field: f.name,
      label: f.label || f.name,
      storeIdent: f.name,
      isMulti: f.isMulti,
    }
    const strat = f.config.dal.encodingStrategy
    if (!strat || strat[types.JSON]) {
      a.storeIdent = 'values'
    }
    moduleFields.value.push(a)
  }

  moduleFieldEncoding.value = moduleFields.value.reduce((enc, { field }) => {
    const f = props.module.findField(field)
    if (f) {
      enc[field] = f.config.dal.encodingStrategy || {}
    }
    return enc
  }, {})
}, { deep: true, immediate: true })

onMounted(() => {
  fetchConnections()
})

onBeforeUnmount(() => {
  setDefaultValues()
})

async function fetchConnections () {
  processing.value = true
  return $SystemAPI.dalConnectionList()
    .then(({ set = [] }) => {
      connections.value = set
      const { connectionID } = props.module.config.dal || {}
      if (!connectionID || connectionID === NoID) {
        const primaryConnectionID = (connections.value.find(c => c.type === PrimaryConnType) || { connectionID: NoID }).connectionID
        props.module.config.dal.connectionID = primaryConnectionID
      }
    })
    .catch(toastErrorHandler(t('connections.fetch-failed')))
    .finally(() => {
      processing.value = false
    })
}

function getConnectionLabel ({ connectionID, handle, meta = {} }) {
  return meta.name || handle || connectionID
}

function applyModuleFieldStrategyConfig (field, { strategy, config }) {
  const value = moduleFieldStrategyConfig(strategy, config)
  moduleFieldEncoding.value = { ...moduleFieldEncoding.value, [field]: value }
  const moduleField = props.module.findField(field)
  if (moduleField) {
    moduleField.config.dal.encodingStrategy = value
  }
}

function applySystemFieldStrategyConfig (field, { strategy, config }) {
  const value = systemFieldStrategyConfig(strategy, config)
  systemFieldEncoding.value = { ...systemFieldEncoding.value, [field]: value }
  props.module.config.dal.systemFieldEncoding = Object.entries(systemFieldEncoding.value)
    .reduce((enc, [f, c]) => {
      if (c === null || Object.keys(c).length) {
        enc[f] = c
      }
      return enc
    }, {})
}

function applySelectedSystemFields () {
  if (isDbRef.value) {
    systemFieldEncoding.value = {}
    return
  }
  systemFieldEncoding.value = systemFields.value.reduce((enc, { field, group }) => {
    if (field !== 'id') {
      if (selectedGroup.value === 'all') {
        enc[field] = {}
      } else {
        enc[field] = group === selectedGroup.value ? {} : { omit: true }
      }
    }
    return enc
  }, {})
}

function getOptionKey ({ connectionID }) {
  return connectionID
}

function setDefaultValues () {
  processing.value = false
  connections.value = []
  moduleFields.value = []
  moduleFieldEncoding.value = []
  selectedGroup.value = ''
  systemFields.value = []
  systemFieldEncoding.value = {}
  optionsGroups.value = []
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
