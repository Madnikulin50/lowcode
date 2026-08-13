<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('automation.sessions.editor.title')" />
    <c-session-editor-info :session="session" :user="user" :processing="info.processing" @cancel="cancelSession()" />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'automation.sessions', keyPrefix: 'editor' } })
import { ref, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { system } from 'corteza-lib/js/dist'
import CSessionEditorInfo from '../../../components/Session/CSessionEditorInfo.vue'
const props = defineProps({ sessionID: { type: String, required: false, default: undefined } })
const { t } = useI18n()
const session = ref({})
const user = ref({})
const info = reactive({ processing: false })
function incLoader() {} function decLoader() {}
watch(() => props.sessionID, () => { if (props.sessionID) fetchSession() }, { immediate: true })
function fetchSession() { incLoader(); info.processing = true; window.__AutomationAPI.sessionRead({ sessionID: props.sessionID }).then(s => { session.value = s; fetchUser() }).finally(() => { decLoader(); info.processing = false }) }
function cancelSession(sessionID = props.sessionID) { incLoader(); info.processing = true; window.__AutomationAPI.sessionCancel({ sessionID }).then(() => { fetchSession() }).finally(() => { decLoader(); info.processing = false }) }
function fetchUser() { incLoader(); window.__SystemAPI.userRead({ userID: session.value.createdBy }).then(u => { user.value = new system.User(u) }).finally(() => decLoader()) }
</script>
