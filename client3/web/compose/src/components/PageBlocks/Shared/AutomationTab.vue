<template>
  <div class="p-0">
    <div class="d-flex align-items-center gap-2">
      {{ $t('label') }}
      <span v-if="buttons.length > 0" class="badge rounded-pill bg-secondary ms-1">{{ buttons.length }}</span>
    </div>

    <div class="container-fluid py-3">
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="card border">
            <div class="card-body">
              <h5 class="card-title">{{ $t('configuredButtons') }}</h5>
              <draggable item-key="id" :list="buttons" group="buttons" filter=".disabled">
                <template #item="{ element, index }">
                  <button :key="index" :class="`btn btn-${element.variant || 'primary'} cursor-move m-1`" @click="currentButton = element">
                    {{ buttonLabel(element.label) || '-' }}
                  </button>
                </template>
              </draggable>
            </div>
            <div class="card-footer text-end">
              <button class="btn btn-link" @click="appendButton({ label: $t('dummyButtonLabel'), variant: 'danger' })">
                {{ $t('addPlaceholderLabel') }}
              </button>
              <c-input-confirm v-if="buttons.length" :text="$t('removeAll')" variant="link" size="md" @confirmed="removeAllButtons" />
            </div>
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <button-editor v-if="currentButton" :page="page" :block="block" :button="currentButton" :script="currentScript" :trigger="currentTrigger" :record="record" :module="module" @delete="deleteButton(currentButton)" />
        </div>
      </div>

      <div class="row mt-4">
        <div class="col">
          <div v-if="loading" class="spinner-border d-block mx-auto my-5" />
          <div v-else-if="available.length > 0" class="card border">
            <div class="card-body">
              <h5 class="card-title">{{ $t('availableScriptsAndWorkflow', { count: available.length }) }}</h5>
              <c-input-search v-model="queryAvailable" class="mb-1" :placeholder="$t('searchPlaceholder')" />
              <div v-for="(b, index) in filtered" :key="index" class="mb-2 cursor-pointer list-group list-group-flush" @click.prevent="appendButton(b)">
                <div class="list-group-item list-group-item-action">
                  <div class="d-flex align-items-center w-100 justify-content-between">
                    <h6>
                      {{ b.label || b.script }}
                      <span v-if="b.workflowID" class="badge bg-info ms-1">{{ $t('badge.workflow') }}</span>
                      <span v-else-if="b.script" class="badge border border-secondary text-secondary ms-1">{{ $t('badge.script') }}</span>
                    </h6>
                    <code v-if="b.label && b.script">{{ b.script }}</code>
                  </div>
                  <p class="my-2">
                    <span v-if="b.description">{{ b.description }}</span>
                    <i v-else>{{ $t('noDescription') }}</i>
                  </p>
                  <var v-if="b.stepID">{{ $t('stepID', { stepID: b.stepID }) }}</var>
                </div>
              </div>
            </div>
          </div>
          <p v-else-if="buttons.length === 0">{{ $t('noScripts') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { inject } from 'vue'
import { compose, NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import draggable from 'vuedraggable'
import { words } from 'lodash'
import ButtonEditor from './AutomationTabButtonEditor'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'

const { CInputSearch } = components

const { t: $t } = useI18n({ useScope: 'global', messages: {}, keyPrefix: 'block:automation' })

const props = defineProps({
  buttons: { type: Array, required: true },
  page: { type: compose.Page, required: true },
  block: { type: compose.PageBlock, required: true },
  module: { type: compose.Module, required: false, default: undefined },
  record: { type: compose.Record, required: false, default: undefined },
})

const $auth = inject('$auth')
const $UIHooks = inject('$UIHooks')
const $AutomationAPI = inject('$AutomationAPI')

const loading = ref(false)
const currentButton = ref(null)
const queryAvailable = ref('')
const triggerButtons = ref([])

const currentScript = computed(() => {
  const c = currentButton.value
  if (!c?.script) return undefined
  return scriptButtons.value.filter(({ script }) => script).find(({ script }) => script === c.script)
})

const currentTrigger = computed(() => {
  const c = currentButton.value
  if (!c?.workflowID) return undefined
  return triggerButtons.value.filter(({ workflowID, stepID }) => workflowID && stepID).find(t => t.workflowID === c.workflowID && t.stepID === c.stepID)
})

const resourceTypes = computed(() => {
  const types = ['compose', 'compose:namespace', 'compose:page']
  if (props.module) types.push('compose:module')
  if (props.record) types.push('compose:record')
  return types
})

const scriptButtons = computed(() => $UIHooks?.Find(resourceTypes.value) || [])

const available = computed(() => {
  const existingScripts = props.buttons.map(b => b.script || `${b.workflowID}-${b.stepID}`)
  return [...scriptButtons.value, ...triggerButtons.value].filter(b => !existingScripts.includes(b.script || `${b.workflowID}-${b.stepID}`))
})

const filtered = computed(() => {
  if (!queryAvailable.value) return available.value
  const q = words(queryAvailable.value.toLowerCase())
  return available.value.filter(({ script = '', label, description }) => q.every(query => `${script} ${label} ${description}`.toLowerCase().includes(query)))
})

onMounted(() => { fetchTriggers() })

function appendButton(newButton) {
  currentButton.value = { ...newButton, variant: newButton.variant || 'primary' }
  props.buttons.push(currentButton.value)
}

function deleteButton(button = {}) {
  const i = props.buttons.indexOf(button)
  if (i > -1) props.buttons.splice(i, 1)
  currentButton.value = null
}

async function fetchTriggers() {
  loading.value = true
  let aux = []
  return $AutomationAPI.triggerList({ eventType: 'onManual' })
    .then(({ set } = {}) => {
      aux = set.map(({ triggerID, workflowID, resourceType, stepID }) => ({ triggerID, workflowID, resourceType, stepID }))
      return set.map(({ workflowID }) => workflowID)
    })
    .then((workflowID) => $AutomationAPI.workflowList({ workflowID }))
    .then(({ set = [] } = {}) => {
      triggerButtons.value = aux.map(trigger => {
        const { triggerID, workflowID, stepID, resourceType } = trigger
        const workflow = set.find(wf => wf.workflowID === workflowID)
        if (!workflow) {
          console.warn('trigger referencing a non existing workflow', { triggerID, workflowID: trigger.workflowID })
          return null
        }
        const { handle, meta: { name, description } = {} } = workflow
        let label = name || handle
        const step = workflow.steps.find(s => s.stepID === stepID)
        if (!step) {
          console.warn('trigger referencing a non existing step', { triggerID, workflowID, stepID })
          return null
        } else if (step.meta?.label) {
          label = `${label} (${step.meta.label})`
        }
        return { label, workflowID, stepID, resourceType, description, workflow }
      }).filter(t => !!t)
    })
    .catch(err => console.error(err))
    .finally(() => { loading.value = false })
}

function removeAllButtons() {
  props.buttons.splice(0)
  currentButton.value = null
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
  } catch (e) { return e }
}
</script>
<style lang="scss" scoped>
.cursor-move { cursor: move !important; }
.cursor-pointer { cursor: pointer !important; }
</style>
