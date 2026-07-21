<template>
  <div v-if="template" class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="templateID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'system.template.new' })">{{ $t('new') }}</button>
      <c-permissions-button v-if="templateID && canGrant" :title="template.meta.short || template.handle || template.templateID" :target="template.meta.short || template.handle || template.templateID" :resource="`corteza::system:template/${templateID}`"><font-awesome-icon :icon="['fas', 'lock']" /> {{ $t('permissions') }}</c-permissions-button>
      <c-corredor-manual-buttons ui-page="template/editor" ui-slot="toolbar" resource-type="system:template" default-variant="link" class="me-1" @click="dispatchCortezaSystemTemplateEvent($event, { template })" />
    </c-content-header>
    <c-template-editor-info :template="template" :processing="info.processing" :success="info.success" :can-create="canCreate" @delete="onDelete" @submit="onInfoSubmit" />
    <c-template-editor-content v-if="template && template.templateID != '0'" class="mt-3" :template="template" :partials="partials" :processing="info.processing" :success="info.success" :can-create="canCreate" @submit="onInfoSubmit" />
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual } from 'lodash'
import { system } from 'corteza-lib/js/dist'
import CTemplateEditorInfo from '../../../components/Template/CTemplateEditorInfo.vue'
import CTemplateEditorContent from '../../../components/Template/CTemplateEditorContent/Index.vue'
const props = defineProps({ templateID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const template = ref(undefined)
const initialTemplateState = ref(undefined)
const info = reactive({ processing: false, success: false })
const partials = ref([])
const canCreate = computed(() => can('system/', 'template.create'))
const canGrant = computed(() => can('system/', 'grant'))
const title = computed(() => props.templateID ? t('title.edit') : t('title.create'))
function can(resource, operation) { return true }
function incLoader() {}
function decLoader() {}
watch(() => props.templateID, () => {
  fetchPartials()
  if (props.templateID) { fetchTemplate() } else { template.value = new system.Template(); initialTemplateState.value = template.value.clone() }
}, { immediate: true })
function fetchTemplate() { incLoader(); window.__systemAPI.templateRead({ templateID: props.templateID }).then(t => { template.value = new system.Template(t); initialTemplateState.value = template.value.clone() }).finally(() => decLoader()) }
function fetchPartials() { incLoader(); window.__systemAPI.templateList({ partial: true }).then(({ set: tt }) => { partials.value = tt.map(t => new system.Template(t)) }).finally(() => decLoader()) }
function onDelete() {
  incLoader()
  if (template.value.deletedAt) { window.__systemAPI.templateUndelete({ templateID: props.templateID }).then(() => fetchTemplate()).finally(() => decLoader()) }
  else { window.__systemAPI.templateDelete({ templateID: props.templateID }).then(() => { fetchTemplate(); template.value.deletedAt = new Date(); router.push({ name: 'system.template' }) }).finally(() => decLoader()) }
}
function onInfoSubmit(t) {
  info.processing = true; incLoader()
  if (props.templateID) { window.__systemAPI.templateUpdate(t).then(t => { template.value = new system.Template(t); initialTemplateState.value = template.value.clone(); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => { decLoader(); info.processing = false }) }
  else { window.__systemAPI.templateCreate(t).then(({ templateID }) => { info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'system.template.edit', params: { templateID } }) }).finally(() => { decLoader(); info.processing = false }) }
}
function dispatchCortezaSystemTemplateEvent($event, { template }) {}
</script>
