<template>
  <div v-if="sensitivityLevel" class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="sensitivityLevelID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'system.sensitivityLevel.new' })">{{ $t('system.sensitivityLevel.editor.new') }}</button>
    </c-content-header>
    <c-sensitivity-level-editor-info :sensitivity-level="sensitivityLevel" :processing="info.processing" :success="info.success" :can-delete="canDelete" :can-create="canCreate" @submit="onSubmit($event)" @delete="onDelete($event)" />
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual, cloneDeep } from 'lodash'
import CSensitivityLevelEditorInfo from '../../../components/SensitivityLevel/CSensitivityLevelEditorInfo.vue'
const props = defineProps({ sensitivityLevelID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const sensitivityLevel = ref(undefined)
const initialSensitivityLevelState = ref(undefined)
const info = reactive({ processing: false, success: false })
const canCreate = computed(() => can('system/', 'dal-sensitivity-level.manage'))
const canDelete = computed(() => sensitivityLevel.value && sensitivityLevel.value.sensitivityLevelID && canCreate.value)
const title = computed(() => props.sensitivityLevelID ? t('system.sensitivityLevel.editor.title.edit') : t('system.sensitivityLevel.editor.title.create'))
function can(resource, operation) { return true }
function incLoader() {}
function decLoader() {}
watch(() => props.sensitivityLevelID, () => {
  if (props.sensitivityLevelID) { fetchSensitivityLevel() } else { sensitivityLevel.value = { handle: '', level: 1, meta: { name: '', description: '' } }; initialSensitivityLevelState.value = cloneDeep(sensitivityLevel.value) }
}, { immediate: true })
function fetchSensitivityLevel(sensitivityLevelID = props.sensitivityLevelID) { incLoader(); window.__systemAPI.dalSensitivityLevelRead({ sensitivityLevelID }).then(s => { sensitivityLevel.value = s; initialSensitivityLevelState.value = cloneDeep(s) }).finally(() => decLoader()) }
function onSubmit(s) {
  info.processing = true
  if (props.sensitivityLevelID) { window.__systemAPI.dalSensitivityLevelUpdate(s).then(s => { sensitivityLevel.value = s; initialSensitivityLevelState.value = cloneDeep(s); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => { info.processing = false }) }
  else { window.__systemAPI.dalSensitivityLevelCreate(s).then(s => { sensitivityLevel.value = s; initialSensitivityLevelState.value = cloneDeep(s); const { sensitivityLevelID } = s; info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'system.sensitivityLevel.edit', params: { sensitivityLevelID } }) }).finally(() => { info.processing = false }) }
}
function onDelete(sensitivityLevelID = props.sensitivityLevelID) {
  incLoader()
  if (sensitivityLevel.value.deletedAt) { window.__systemAPI.dalSensitivityLevelUndelete({ sensitivityLevelID }).then(() => fetchSensitivityLevel()).finally(() => decLoader()) }
  else { window.__systemAPI.dalSensitivityLevelDelete({ sensitivityLevelID }).then(() => { fetchSensitivityLevel(); sensitivityLevel.value.deletedAt = new Date(); router.push({ name: 'system.sensitivityLevel' }) }).finally(() => decLoader()) }
}
</script>
