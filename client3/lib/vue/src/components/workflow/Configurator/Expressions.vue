<template>
  <div>
    <div class="card flex-grow-1 rounded-0">
      <div class="card-header d-flex align-items-center py-4">
        <h5 class="d-flex align-items-center mb-0">
          {{ t('steps.expressions.label') }}
          <a
            v-if="documentationURL"
            :href="documentationURL"
            target="_blank"
            class="d-flex align-items-center h6 mb-0 ms-1"
          >
            <font-awesome-icon
              :icon="['far', 'question-circle']"
            />
          </a>
        </h5>

        <Teleport to="#sidebar-footer">
          <button
            class="btn btn-primary align-top border-0 ms-auto"
            @click="addArgument()"
          >
            {{ t('steps.expressions.configurator.add-expression') }}
          </button>
        </Teleport>
      </div>

      <div
        v-if="hasArguments"
        class="card-body p-0"
      >
        <expression-table
          value-field="expr"
          :items="item.config.arguments"
          :fields="argumentFields"
          :types="fieldTypes"
          @remove="removeArgument"
          @open-editor="openInEditor"
        />
      </div>
    </div>

    <Teleport to="body">
      <div
        v-if="!!expressionEditor.currentExpression"
        class="modal fade show d-block"
        id="expression-editor"
        tabindex="-1"
        style="background: rgba(0,0,0,0.5);"
      >
        <div class="modal-dialog modal-xl modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ t('editor.editor') }}</h5>
              <button type="button" class="btn-close" @click="resetExpression"></button>
            </div>
            <div class="modal-body p-0">
              <expression-editor
                v-model="currentExpressionValue"
                min-height="80vh"
                font-size="18px"
                show-line-numbers
                :border="false"
                :show-popout="false"
              />
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline-secondary" @click="resetExpression">{{ t('cancel') }}</button>
              <button type="button" class="btn btn-primary" @click="saveExpression">{{ t('save') }}</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import ExpressionTable from '../ExpressionTable.vue'
import ExpressionEditor from '../ExpressionEditor.vue'
import { getDocumentationURL } from '../lib/version'

const { t } = useI18n()

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value'])
const $AutomationAPI = inject('automationAPI', {})

const fieldTypes = ref([])
const expressionEditor = ref({
  currentIndex: undefined,
  currentExpression: undefined,
})

const currentExpressionValue = computed({
  get () {
    return expressionEditor.value.currentExpression ? expressionEditor.value.currentExpression.expr : ''
  },
  set (value) {
    if (expressionEditor.value.currentExpression) {
      expressionEditor.value.currentExpression.expr = value
    }
  },
})

const argumentFields = computed(() => [
  {
    key: 'target',
    label: t('steps.expressions.configurator.target'),
    thClass: 'ps-4 ms-1',
    formatter: (item) => {
      return `${item.target}(${item.type})`
    },
  },
  {
    key: 'expr',
    label: t('steps.expressions.configurator.expression'),
    thClass: 'ps-1 me-3',
  },
])

const hasArguments = computed(() => {
  const { config } = props.item || {}
  return (config && (config.arguments || []).length) || []
})

const documentationURL = computed(() => {
  return getDocumentationURL('integrator-guide/expr/index.html')
})

watch(() => props.item.config.stepID, () => {
  if (!props.item.config.arguments) {
    props.item.config.arguments = []
  }
}, { immediate: true })

onMounted(() => {
  getTypes()
})

function addArgument () {
  props.item.config.arguments.push({
    target: '',
    expr: '',
    type: 'Any',
    _showDetails: true,
  })
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function removeArgument (index) {
  props.item.config.arguments.splice(index, 1)
  window.dispatchEvent(new CustomEvent('change-detected'))
}

function openInEditor (index = -1) {
  expressionEditor.value = {
    currentIndex: index >= -1 ? index : undefined,
    currentExpression: index >= 0 ? { ...props.item.config.arguments[index] } : undefined,
  }
}

function saveExpression () {
  if (expressionEditor.value.currentIndex >= 0) {
    const args = [...props.item.config.arguments]
    args[expressionEditor.value.currentIndex] = expressionEditor.value.currentExpression
    props.item.config.arguments = args
    window.dispatchEvent(new CustomEvent('change-detected'))
  }

  resetExpression()
}

function resetExpression () {
  expressionEditor.value = {
    currentIndex: undefined,
    currentExpression: undefined,
  }
}

async function getTypes () {
  return $AutomationAPI.typeList()
    .then(({ set }) => {
      fieldTypes.value = set
    })
    .catch(toastErrorHandler(t('notification.fetch-types-failed')))
}

function getTypeDescription (type) {
  const typeDescriptions = {
    ID: "Make sure to provide the ID in double quotes if you're using a literal value. Example \"123\"",
  }
  return typeDescriptions[type]
}

function toastErrorHandler (msg) {
  return (e) => {
    console.error(msg, e)
  }
}
</script>
