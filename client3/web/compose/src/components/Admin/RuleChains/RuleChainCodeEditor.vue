<template>
  <div
    v-show="!failed"
    ref="host"
    class="rulechain-ace border rounded"
  />
  <textarea
    v-if="failed"
    class="form-control form-control-sm font-monospace"
    rows="14"
    spellcheck="false"
    :value="modelValue"
    @input="$emit('update:modelValue', $event.target.value)"
  />
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import aceModule from 'ace-builds'
import 'ace-builds/src-noconflict/mode-javascript'
import 'ace-builds/src-noconflict/mode-golang'
import 'ace-builds/src-noconflict/theme-chrome'
import 'ace-builds/src-noconflict/ext-language_tools'

const ace = aceModule?.edit ? aceModule : (aceModule?.default || aceModule)

ace.config.set('basePath', '')
ace.config.set('loadWorkerFromBlob', false)

const props = defineProps({
  modelValue: { type: String, default: '' },
  lang: { type: String, default: 'javascript' },
})

const emit = defineEmits(['update:modelValue'])

const host = ref(null)
const failed = ref(false)
let editor = null
let silent = false
let ro = null

function aceMode (lang) {
  return lang === 'golang' || lang === 'go' ? 'ace/mode/golang' : 'ace/mode/javascript'
}

onMounted(async () => {
  await nextTick()
  if (!host.value) {
    failed.value = true
    return
  }
  try {
    editor = ace.edit(host.value)
    editor.setTheme('ace/theme/chrome')
    editor.session.setMode(aceMode(props.lang))
    editor.session.setUseWorker(false)
    editor.setOptions({
      fontSize: '13px',
      tabSize: 2,
      useSoftTabs: true,
      wrap: true,
      showPrintMargin: false,
      highlightActiveLine: true,
      enableBasicAutocompletion: true,
      enableLiveAutocompletion: true,
    })
    silent = true
    editor.setValue(props.modelValue || '', -1)
    silent = false
    editor.session.on('change', () => {
      if (silent) return
      emit('update:modelValue', editor.getValue())
    })
    editor.resize()
    ro = new ResizeObserver(() => editor?.resize())
    ro.observe(host.value)
  } catch (e) {
    console.error('RuleChain code editor failed to start', e)
    failed.value = true
  }
})

watch(() => props.modelValue, (v) => {
  if (!editor) return
  const next = v || ''
  if (editor.getValue() === next) return
  silent = true
  editor.setValue(next, -1)
  silent = false
})

watch(() => props.lang, (lang) => {
  editor?.session.setMode(aceMode(lang))
})

onBeforeUnmount(() => {
  ro?.disconnect()
  ro = null
  editor?.destroy()
  editor = null
})
</script>

<style scoped>
.rulechain-ace {
  position: relative;
  height: 280px;
  min-height: 280px;
  width: 100%;
}
</style>
