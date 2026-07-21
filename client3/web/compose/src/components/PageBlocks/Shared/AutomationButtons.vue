<template>
  <div class="d-flex gap-2">
    <button
      v-for="(b, i) in buttons"
      :key="i"
      class="btn"
      :class="[`btn-${variant(b)}`, buttonClass]"
      :disabled="!isValid(b) || processingIDs.includes(i)"
      @click.prevent="handle(b, i)"
    >
      {{ buttonLabel(b.label) || '-' }}
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, automation, NoID } from 'corteza-lib/js/dist'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'

const { t: $t } = useI18n({ useScope: 'global' })

const $auth = window.__auth
const $AutomationAPI = window.__automationAPI
const $EventBus = window.__eventBus
const $UIHooks = window.__uiHooks

const props = defineProps({
  buttons: { type: Array, required: true },
  automationScripts: { type: Array, required: false, default: () => [] },
  buttonClass: { type: String, default: '' },
  extraEventArgs: { type: Object, default: () => ({}) },
  namespace: { type: Object, required: false },
  module: { type: Object, required: false },
  record: { type: Object, required: false },
  page: { type: Object, required: false },
})

const processingIDs = ref([])

function variant(b) {
  if (!b.script) return b.variant
  if (!isValid(b)) return 'outline-danger'
  return b.variant || 'primary'
}

function isValid(b) {
  const { resourceType } = b
  let paramsExist = true
  if (resourceType === 'compose:record') paramsExist = props.record && props.module
  else if (resourceType === 'compose:module') paramsExist = !!props.module
  else if (resourceType === 'compose:namespace') paramsExist = !!props.namespace
  else if (resourceType === 'compose:page') paramsExist = !!props.page
  if (!paramsExist) return false
  if (b.workflowID) return true
  if (b.script) {
    if ($UIHooks?.FindByScript?.(b.script)) return true
    if (!props.automationScripts) return false
    return props.automationScripts.find(({ name }) => name === b.script)
  }
  return false
}

async function handle(b, i) {
  try {
    processingIDs.value.push(i)
    let ev = { args: props.extraEventArgs || {} }
    switch (b.resourceType) {
      case 'compose:record':
        ev.args.namespace = props.namespace
        if (props.record && props.module) {
          ev.args.module = props.module
          ev = compose.RecordEvent(props.record, ev)
        }
        break
      case 'compose:module':
        ev.args.namespace = props.namespace
        ev = compose.ModuleEvent(props.module, ev)
        break
      case 'compose:namespace':
        ev = compose.NamespaceEvent(props.namespace, ev)
        break
      case 'compose:page':
        ev.args.namespace = props.namespace
        ev = compose.PageEvent(props.page.pageID, ev)
        break
      case 'compose':
        ev = compose.ComposeEvent(ev)
    }
    if (b.workflowID) {
      const { workflowID, stepID } = b
      const input = automation.Encode(ev.args)
      await $AutomationAPI.workflowExec({ workflowID, stepID, input })
    } else if (b.script) {
      await $EventBus.Dispatch(ev, b.script)
    }
  } catch (e) {
    console.error('Automation button error:', e)
  } finally {
    processingIDs.value = processingIDs.value.filter(id => id !== i)
  }
}

function buttonLabel(label = '') {
  try {
    return evaluatePrefilter(label, {
      record: props.record,
      user: $auth?.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth?.user || {}).userID || NoID,
    })
  } catch (e) {
    return e
  }
}
</script>
