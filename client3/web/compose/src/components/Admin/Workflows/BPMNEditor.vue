<template>
  <div class="d-flex flex-column flex-grow-1" style="min-height: 0">
    <div class="d-flex align-items-center gap-2 border-bottom p-2">
      <button
        class="btn btn-sm btn-outline-primary"
        :disabled="compiling"
        @click="compile"
      >
        <span v-if="compiling" class="spinner-border spinner-border-sm me-1" />
        {{ $t('workflow.bpmn.compile', 'Compile') }}
      </button>
      <span class="small text-muted">
        {{ $t('workflow.bpmn.hint', 'Draws a BPMN 2.0 diagram; Compile previews the automation steps it translates to (start/end events, tasks, gateways, timers - see the extension "properties" panel on a task/gateway for chainID/functionRef/correlationKey).') }}
      </span>
    </div>

    <div class="d-flex flex-grow-1" style="min-height: 0">
      <div
        ref="canvasEl"
        class="flex-grow-1"
        style="min-height: 0"
      />
      <div
        v-if="result || error"
        class="border-start p-2 overflow-auto"
        style="width: 360px; flex-shrink: 0"
      >
        <div v-if="error" class="alert alert-danger py-2 small mb-0">
          {{ error }}
        </div>
        <template v-else-if="result">
          <h6 class="small text-muted text-uppercase mb-1">
            {{ $t('workflow.bpmn.steps', 'Steps') }} ({{ result.steps.length }})
          </h6>
          <pre class="bg-light border rounded p-2 small" style="white-space: pre-wrap;">{{ JSON.stringify(result.steps, null, 2) }}</pre>
          <h6 class="small text-muted text-uppercase mb-1 mt-2">
            {{ $t('workflow.bpmn.paths', 'Paths') }} ({{ result.paths.length }})
          </h6>
          <pre class="bg-light border rounded p-2 small" style="white-space: pre-wrap;">{{ JSON.stringify(result.paths, null, 2) }}</pre>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import 'bpmn-js/dist/assets/diagram-js.css'
import 'bpmn-js/dist/assets/bpmn-js.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css'

const $AutomationAPI = window.__automationAPI

const props = defineProps({
  // initial BPMN 2.0 XML; falls back to a blank one-startEvent diagram
  xml: {
    type: String,
    required: false,
    default: '',
  },
})

const emit = defineEmits(['change'])

const BLANK_DIAGRAM = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                   xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI"
                   xmlns:dc="http://www.omg.org/spec/DD/20100524/DC"
                   id="Definitions_1" targetNamespace="http://lowcode/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1" name="Start" />
  </bpmn:process>
  <bpmndi:BPMNDiagram id="BPMNDiagram_1">
    <bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="Process_1">
      <bpmndi:BPMNShape id="StartEvent_1_di" bpmnElement="StartEvent_1">
        <dc:Bounds x="152" y="102" width="36" height="36" />
      </bpmndi:BPMNShape>
    </bpmndi:BPMNPlane>
  </bpmndi:BPMNDiagram>
</bpmn:definitions>`

const canvasEl = ref(null)
const compiling = ref(false)
const result = ref(null)
const error = ref('')

let modeler = null

onMounted(async () => {
  const { default: Modeler } = await import('bpmn-js/lib/Modeler')
  modeler = new Modeler({ container: canvasEl.value })

  try {
    await modeler.importXML(props.xml || BLANK_DIAGRAM)
    modeler.get('canvas').zoom('fit-viewport')
  } catch (e) {
    error.value = `${e.message || e}`
  }

  modeler.on('commandStack.changed', () => {
    emit('change')
  })
})

onBeforeUnmount(() => {
  modeler?.destroy()
})

async function currentXML () {
  const { xml } = await modeler.saveXML({ format: true })
  return xml
}

async function compile () {
  compiling.value = true
  error.value = ''
  result.value = null
  try {
    const xml = await currentXML()
    const { steps, paths } = await $AutomationAPI.bpmnCompile({ xml })
    result.value = { steps: steps || [], paths: paths || [] }
  } catch (e) {
    error.value = e?.response?.data?.error?.message || String(e.message || e)
  } finally {
    compiling.value = false
  }
}

defineExpose({ currentXML, compile })
</script>
