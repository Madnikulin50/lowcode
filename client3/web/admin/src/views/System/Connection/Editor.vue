<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="connectionID ? $t('system.connections.editor.title.edit') : $t('system.connections.editor.title.create')" />
    <form v-if="connection && sensitivityLevels" @submit.prevent="onSubmit">
      <c-connection-editor-info :connection="connection" :sensitivity-levels="sensitivityLevels" :processing="info.processing" :success="info.success" :can-create="canCreate" :disabled="disabled" @submit="updateInfo" @delete="toggleDelete" />
      <c-connection-editor-properties v-if="connectionID && connection.meta.properties" :properties="connection.meta.properties" :processing="properties.processing" :success="properties.success" class="mt-4" @submit="updateProperties" />
      <c-connection-editor-dal v-if="connectionID && connection.config.dal && canManage" :dal="connection.config.dal" :issues="connection.issues || []" :can-manage="connection.canManageDalConfig" class="mt-4" @submit="updateDal" />
    </form>
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual } from 'lodash'
import { system } from 'corteza-lib/js/dist'
import CConnectionEditorInfo from '../../../components/Connection/CConnectionEditorInfo.vue'
import CConnectionEditorProperties from '../../../components/Connection/CConnectionEditorProperties.vue'
import CConnectionEditorDal from '../../../components/Connection/CConnectionEditorDAL.vue'
const props = defineProps({ connectionID: { type: String, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const info = reactive({ processing: false, success: false })
const properties = reactive({ processing: false, success: false })
const dal = reactive({ processing: false, success: false })
const connection = ref(undefined)
const initialConnectionState = ref(undefined)
const sensitivityLevels = ref(undefined)
const canCreate = computed(() => can('system/', 'dal-connection.create'))
const canManage = computed(() => connection.value?.canManageDalConfig)
const disabled = computed(() => info.processing || properties.processing || dal.processing)
function can(resource, operation) { return true }
function incLoader() {} function decLoader() {}
watch(() => props.connectionID, (connectionID) => { if (connectionID) fetchConnection(connectionID); else { connection.value = new system.DalConnection(); initialConnectionState.value = connection.value.clone() } }, { immediate: true })
onMounted(() => { fetchSensitivityLevels() })
function fetchConnection(connectionID) { incLoader(); return window.__systemAPI.dalConnectionRead({ connectionID }).then(c => { connection.value = new system.DalConnection(c); initialConnectionState.value = connection.value.clone() }).finally(() => decLoader()) }
async function fetchSensitivityLevels() { info.processing = true; return window.__systemAPI.dalSensitivityLevelList().then(({ set = [] }) => { sensitivityLevels.value = set }).finally(() => { info.processing = false }) }
function updateInfo() { const updating = !!props.connectionID; const fn = updating ? 'dalConnectionUpdate' : 'dalConnectionCreate'; info.processing = true; incLoader(); const c = new system.DalConnection(initialConnectionState.value); c.meta.name = connection.value.meta.name; c.handle = connection.value.handle; c.meta.location.properties.name = connection.value.meta.location.properties.name; c.meta.location.geometry.coordinates = connection.value.meta.location.geometry.coordinates; c.meta.ownership = connection.value.meta.ownership; c.config.privacy.sensitivityLevelID = connection.value.config.privacy.sensitivityLevelID; return window.__systemAPI[fn](c).then(c => { info.success = true; setTimeout(() => { info.success = false }, 2000); if (!updating) { const { connectionID } = c; router.push({ name: 'system.connection.edit', params: { connectionID } }) } else { c.config.dal = connection.value.config.dal; c.meta.properties = connection.value.meta.properties; connection.value = new system.DalConnection(c); initialConnectionState.value = connection.value.clone() } }).finally(() => { info.processing = false }) }
function updateProperties() { properties.processing = true; incLoader(); const c = new system.DalConnection(initialConnectionState.value); c.meta.properties = connection.value.meta.properties; return window.__systemAPI.dalConnectionUpdate(c).then(c => { properties.success = true; setTimeout(() => { properties.success = false }, 2000); c.meta.name = connection.value.meta.name; c.handle = connection.value.handle; c.meta.location.properties.name = connection.value.meta.location.properties.name; c.meta.location.geometry.coordinates = connection.value.meta.location.geometry.coordinates; c.meta.ownership = connection.value.meta.ownership; c.config.privacy.sensitivityLevelID = connection.value.config.privacy.sensitivityLevelID; c.config.dal = connection.value.config.dal; connection.value = new system.DalConnection(c); initialConnectionState.value = connection.value.clone() }).finally(() => { properties.processing = false }) }
function updateDal() { dal.processing = true; incLoader(); const c = new system.DalConnection(initialConnectionState.value); c.config.dal = connection.value.config.dal; return window.__systemAPI.dalConnectionUpdate(c).then(c => { dal.success = true; setTimeout(() => { dal.success = false }, 2000); c.meta.name = connection.value.meta.name; c.handle = connection.value.handle; c.meta.location.properties.name = connection.value.meta.location.properties.name; c.meta.location.geometry.coordinates = connection.value.meta.location.geometry.coordinates; c.meta.ownership = connection.value.meta.ownership; c.config.privacy.sensitivityLevelID = connection.value.config.privacy.sensitivityLevelID; c.meta.properties = connection.value.meta.properties; connection.value = new system.DalConnection(c); initialConnectionState.value = connection.value.clone() }).finally(() => { dal.processing = false }) }
function toggleDelete() { const { deletedAt } = connection.value; const deleting = !deletedAt; const fn = deleting ? 'dalConnectionDelete' : 'dalConnectionUndelete'; info.processing = true; incLoader(); return window.__systemAPI[fn](connection.value).then(() => { if (deleting) { connection.value.deletedAt = new Date(); initialConnectionState.value.deletedAt = connection.value.deletedAt; router.push({ name: 'system.connection' }) } else { connection.value.deletedAt = null; initialConnectionState.value.deletedAt = null } }).finally(() => { info.processing = false }) }
</script>
