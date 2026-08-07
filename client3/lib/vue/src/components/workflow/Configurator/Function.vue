<template>
  <div v-if="!processing">
    <div v-if="showFunctionList" class="card flex-grow-1 border-bottom border-light rounded-0">
      <div class="card-header p-0 mb-3">
        <h5 class="mb-0">{{ t('configurator.configuration') }}</h5>
      </div>
      <div v-if="functionTypes.length" class="card-body p-0">
        <div class="mb-0">
          <label class="text-primary form-label">{{ t('steps.function.configurator.type*') }}</label>
          <c-input-select v-model="functionRef" :options="functionTypes" :get-option-key="getOptionTypeKey"
            label="text" :selectable="f => !f.disabled" :reduce="f => f.value"
            :filter="functionFilter" :placeholder="t('steps.function.configurator.select-function')"
            @input="functionChanged" />
        </div>
        <p v-if="functionDescription" class="mt-3 mb-0">{{ functionDescription }}</p>
      </div>
    </div>

    <div v-if="args.length" class="card flex-grow-1 border-bottom border-light rounded-0">
      <div class="card-header d-flex align-items-center">
        <h5 class="mb-0">{{ t('steps.function.configurator.arguments') }}</h5>
      </div>
      <div class="card-body p-0">
        <table class="table table-borderless table-hover mb-4">
          <thead class="table-secondary">
            <tr>
              <th class="ps-3 py-2">{{ t('steps.function.configurator.name') }}</th>
              <th class="pe-3 py-2">{{ t('steps.function.configurator.value') }}</th>
              <th class="text-center" style="width: 3rem;"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(a, index) in args" :key="index"
              :class="a._showDetails ? 'border-thick' : 'border-thick-transparent'"
              @click="a._showDetails = !a._showDetails">
              <td class="text-truncate pointer">
                <var>{{ a.target }}{{ a.required ? '*' : '' }}</var>
                <samp v-if="!isWhileIterator"> ({{ a.type }})</samp>
              </td>
              <td class="text-truncate pointer"><samp>{{ a[a.valueType] }}</samp></td>
              <td class="text-center pointer">
                <span v-if="a.valueType === 'expr'" class="circle-badge" :title="t('steps.function.configurator.expression')">e</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-for="(a, index) in args" :key="'detail-' + index">
      <div v-if="a._showDetails" class="mb-3 px-3">
        <div class="arrow-up"></div>
        <div class="card bg-light">
          <div class="card-body px-4 pb-3">
            <div v-if="(paramTypes[functionRef] && paramTypes[functionRef][a.target] || []).length > 1" class="mb-3">
              <label class="text-primary form-label">{{ t('steps.function.configurator.type') }}</label>
              <c-input-select v-model="a.type" :options="(paramTypes[functionRef] && paramTypes[functionRef][a.target] || [])"
                :get-option-key="getOptionParamKey" :filter="argTypeFilter" :clearable="false"
                @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
            </div>

            <div class="mb-0">
              <label class="d-flex align-items-center text-primary form-label">{{ t('steps.function.configurator.expression') }}
                <a :href="documentationURL" target="_blank" class="d-flex align-items-center h6 mb-0 ms-1">
                  <font-awesome-icon :icon="['far', 'question-circle']" />
                </a>
              </label>
              <div v-if="a.valueType === 'value'">
                <c-input-select v-if="a.target === 'workflow'" v-model="a.value" :options="workflowOptions"
                  :get-option-label="getWorkflowLabel" :get-option-key="getWorkflowKey"
                  :reduce="wf => a.type === 'ID' ? wf.workflowID : wf.handle"
                  :placeholder="t('steps.function.configurator.search-workflow')" :filterable="false"
                  @input="window.dispatchEvent(new CustomEvent('change-detected'))" @search="searchWorkflows" />
                <c-input-select v-else-if="a.input && a.input.type === 'select'" v-model="a.value"
                  :options="a.input.properties.options" :get-option-key="getOptionTypeKey" label="text"
                  :filter="varFilter" :reduce="a => a.value"
                  :placeholder="t('steps.function.configurator.option-select')" :clearable="false"
                  @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
                <div v-else-if="a.type === 'Boolean'" class="form-check">
                  <input class="form-check-input" type="checkbox" v-model="a.value" true-value="true" false-value="false"
                    :id="'arg-' + index" @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
                  <label class="form-check-label" :for="'arg-' + index">{{ a.target }}</label>
                </div>
                <expression-editor v-else v-model="a.value" :auto-complete="false" @open="openInEditor(index)"
                  @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
              </div>
              <expression-editor v-else-if="a.valueType === 'expr'" v-model="a.expr" show-line-numbers
                @open="openInEditor(index)" @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
            </div>

            <div v-if="!isWhileIterator" class="form-check form-switch float-end me-2 mt-2">
              <input class="form-check-input" type="checkbox" v-model="a.valueType" true-value="expr" false-value="value"
                :id="'switch-' + index" @change="valueTypeChanged($event.target.checked ? 'expr' : 'value', index)" />
              <label class="form-check-label" :for="'switch-' + index">{{ t('steps.function.configurator.expression') }}</label>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="expressionResults || results.length" class="card flex-grow-1 border-bottom border-light rounded-0">
      <div class="card-header d-flex align-items-center">
        <h5 class="mb-0">{{ t('steps.function.configurator.results') }}</h5>
      </div>
      <div v-if="results.length" class="card-body p-0">
        <expression-table v-if="expressionResults" value-field="expr" :items="results" :fields="resultFields"
          :types="fieldTypes" @remove="removeResult" @open-editor="openInEditor" />
        <table v-else class="table table-borderless table-hover mb-4">
          <thead class="table-secondary">
            <tr>
              <th class="ps-3">{{ t('steps.function.configurator.target') }}</th>
              <th>{{ t('steps.function.configurator.type') }}</th>
              <th class="pe-3">{{ t('steps.function.configurator.result') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(r, index) in results" :key="index"
              :class="r._showDetails ? 'border-thick' : 'border-thick-transparent'"
              @click="r._showDetails = !r._showDetails">
              <td class="text-truncate pointer">{{ r.target }}</td>
              <td class="text-truncate pointer"><var>{{ r.type }}</var></td>
              <td class="position-relative pointer"><samp>{{ r.expr }}</samp></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-for="(r, index) in results" :key="'result-detail-' + index">
      <div v-if="!expressionResults && r._showDetails" class="mb-3 px-3">
        <div class="arrow-up"></div>
        <div class="card bg-light">
          <div class="card-body px-4 pb-3">
            <div class="mb-0">
              <label class="text-primary form-label">{{ t('configurator.target') }}</label>
              <input class="form-control" v-model="r.target"
                @input="window.dispatchEvent(new CustomEvent('change-detected'))" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <Teleport to="#sidebar-footer">
      <button v-if="expressionResults" class="btn btn-primary align-top border-0 ms-auto" @click="addResult()">
        {{ t('steps.function.configurator.add-result') }}
      </button>
    </Teleport>

    <div class="modal fade" id="expression-editor-function" tabindex="-1">
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('editor.editor') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body p-0">
            <expression-editor v-model="currentExpressionValue" :lang="expressionEditor.lang" min-height="80vh"
              font-size="18px" show-line-numbers :border="false" :show-popout="false" />
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button class="btn btn-primary" @click="saveExpression">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, inject, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import ExpressionTable from '../ExpressionTable.vue'
import ExpressionEditor from '../ExpressionEditor.vue'
import { objectSearchMaker, stringSearchMaker } from '../lib/filter'
import { getDocumentationURL } from '../lib/version'

const { t } = useI18n()
const toast = useToast()
const $AutomationAPI = inject('automationAPI', {})

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value'])

const processing = ref(true)
const showFunctionList = ref(true)
const expressionResults = ref(false)
const functionRef = ref(undefined)

const functions = ref([])
const args = ref([])
const results = ref([])
const fieldTypes = ref([])
const paramTypes = ref({})
const resultTypes = ref({})
const expressionEditor = reactive({ currentIndex: undefined, currentExpression: undefined, lang: 'javascript' })

const currentExpressionValue = computed({
  get: () => expressionEditor.currentExpression ? expressionEditor.currentExpression[expressionEditor.currentExpression.valueType] : '',
  set: (value) => { if (expressionEditor.currentExpression) expressionEditor.currentExpression[expressionEditor.currentExpression.valueType] = value },
})

const functionTypes = computed(() => functions.value.map(({ ref, meta, disabled = false }) => ({ value: ref, text: meta.short, disabled })))

const functionDescription = computed(() => (functions.value.find(({ ref }) => ref === functionRef.value) || { meta: {} }).meta.description)

const isWhileIterator = computed(() => props.item.config && props.item.config.kind === 'iterator' && functionRef.value === 'loopDo')

const documentationURL = computed(() => getDocumentationURL('integrator-guide/expr/index.html'))

const resultFields = [
  { key: 'target', label: t('steps.function.configurator.target'), thClass: 'ps-3', tdClass: 'text-truncate pointer' },
  { key: 'type', label: t('steps.function.configurator.type'), tdClass: 'text-truncate pointer' },
  { key: 'expr', label: t('steps.function.configurator.result'), thClass: 'me-3', tdClass: 'position-relative pointer' },
]

const valueTypes = [
  { text: t('steps.function.configurator.expression'), value: 'expr' },
  { text: t('steps.function.configurator.constant'), value: 'value' },
]

const functionFilter = objectSearchMaker('text')
const argTypeFilter = stringSearchMaker()
const varFilter = objectSearchMaker('text')

const workflowOptions = ref([])

watch(() => props.item.config.stepID, async () => {
  processing.value = true
  props.item.config.arguments = props.item.config.arguments || []
  props.item.config.results = props.item.config.results || []
  await getFunctionTypes()
  await getTypes()
  functionRef.value = props.item.config.ref || functionRef.value
  setParams(functionRef.value, true)
  processing.value = false
}, { immediate: true })

watch(args, (newArgs) => {
  props.item.config.arguments = newArgs.filter(({ value, source, expr }) => value || source || expr).map(arg => {
    const argMapped = { target: arg.target, type: arg.type }
    argMapped[arg.valueType] = arg[arg.valueType]
    return argMapped
  })
}, { deep: true })

watch(results, (newResults) => {
  props.item.config.results = newResults.filter(({ target }) => target).map(({ target, expr, type }) => ({ target, type, expr }))
}, { deep: true })

async function getFunctionTypes() {
  return $AutomationAPI.functionList()
    .then(({ set }) => { functions.value = set.filter(({ kind = '' }) => kind !== 'iterator').sort((a, b) => a.meta.short.localeCompare(b.meta.short)) })
}

async function getTypes() {
  return $AutomationAPI.typeList()
    .then(({ set }) => { fieldTypes.value = set })
}

function setParams(fName, immediate = false) {
  args.value = []
  results.value = []

  if (!immediate) window.dispatchEvent(new CustomEvent('change-detected'))

  if (fName) {
    const func = functions.value.find(({ ref }) => ref === fName)
    if (!paramTypes.value[func.ref] && func.parameters) {
      paramTypes.value[func.ref] = {}
      func.parameters.forEach(({ name, types }) => { paramTypes.value[func.ref][name] = types || [] })
    }

    args.value = (func.parameters || []).map(param => {
      const arg = (props.item.config.arguments || []).find(({ target }) => target === param.name) || {}
      const { input = {} } = (param.meta || {}).visual || {}
      return {
        name: param.name,
        target: param.name,
        type: arg.type || (paramTypes.value[func.ref] ? paramTypes.value[func.ref][param.name]?.[0] : undefined),
        valueType: arg.expr !== undefined ? 'expr' : 'value',
        value: arg.value || input.default || null,
        expr: arg.expr || arg.source || null,
        required: param.required || false,
        input,
        _showDetails: false,
      }
    })

    if (!expressionResults.value) {
      if (!resultTypes.value[func.ref] && func.results) {
        resultTypes.value[func.ref] = {}
        func.results.forEach(({ name, types }) => { resultTypes.value[func.ref][name] = types || [] })
      }
      results.value = (func.results || []).map(result => {
        const res = (props.item.config.results || []).find(({ expr }) => expr === result.name) || {}
        return {
          name: result.name,
          valueType: 'expr',
          target: res.target || undefined,
          type: resultTypes.value[func.ref] ? resultTypes.value[func.ref][result.name]?.[0] : undefined,
          expr: res.expr || result.name,
          _showDetails: false,
        }
      })
    }
  }
}

function openInEditor(index = -1) {
  expressionEditor.currentIndex = index >= -1 ? index : undefined
  expressionEditor.currentExpression = index >= 0 ? { ...args.value[index] } : undefined
  expressionEditor.lang = expressionEditor.currentExpression?.valueType === 'expr' ? 'javascript' : 'text'

  const modal = new bootstrap.Modal(document.getElementById('expression-editor-function'))
  modal.show()
}

function saveExpression() {
  const { currentIndex = -1, currentExpression } = expressionEditor
  if (currentIndex >= 0) {
    args.value[currentIndex] = currentExpression
    window.dispatchEvent(new CustomEvent('change-detected'))
  }
  expressionEditor.currentIndex = undefined
  expressionEditor.currentExpression = undefined
  expressionEditor.lang = 'javascript'
  const modal = bootstrap.Modal.getInstance(document.getElementById('expression-editor-function'))
  if (modal) modal.hide()
}

function functionChanged(newRef) {
  props.item.config.ref = newRef
  setParams(newRef)
  emit('update-default-value', {
    value: (functionTypes.value.find(({ value }) => value === newRef) || {}).text,
    force: !props.item.node.value,
  })
}

function valueTypeChanged(valueType, index) {
  const oldType = valueType === 'value' ? 'expr' : 'value'
  args.value[index][valueType] = args.value[index][oldType]
  if (!args.value[index].value && args.value[index].type === 'Boolean' && valueType === 'value') {
    args.value[index].value = 'false'
  }
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function addResult() {
  results.value.push({ target: '', expr: '', type: 'Any', _showDetails: true })
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function removeResult(index) {
  results.value.splice(index, 1)
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function getOptionTypeKey({ value }) { return value }
function getOptionEWorkflowLabelKey({ workflowID }) { return workflowID }
function getOptionParamKey(type) { return type }
function getWorkflowLabel({ workflowID, handle, meta = {} }) { return meta.name || handle || workflowID }
function getWorkflowKey({ workflowID, handle }) { return handle || workflowID }

function searchWorkflows(query = '', loading) {
  if (loading) loading(true)
  $AutomationAPI.workflowList({ query, subWorkflow: 2 }).then(({ set }) => { workflowOptions.value = set.map(m => Object.freeze(m)) })
    .finally(() => { if (loading) loading(false) })
}
</script>
