<template>
  <div class="card border border-light rounded w-100">
    <div
      v-if="editor && !hideToolbar"
      class="p-0 border-bottom"
    >
      <r-toolbar
        :editor="editor"
        :formats="toolbar"
        :labels="labels"
        :current-value="currentValue"
      />
    </div>

    <editor-content
      :editor="editor"
      :class="bodyClass"
      class="card-body p-2 rt-editor-content rt-content w-100"
      :style="{ minHeight: minBodyHeight, maxHeight: maxBodyHeight }"
      @drop.native="onDrop"
      @paste.native="onPaste"
      @dragover.native.prevent
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { Editor, EditorContent } from '@tiptap/vue-3'
import RToolbar from './RToolbar/index.vue'
import { getFormats, getToolbar } from './lib'

const props = withDefaults(defineProps<{
  modelValue?: string | null
  labels?: Record<string, any>
  minBodyHeight?: string
  maxHeight?: string
  bodyClass?: string
  placeholder?: string
  hideToolbar?: boolean
  outputFormat?: 'html' | 'markdown'
  toMarkdown?: (html: string) => string
  toHtml?: (md: string) => string
}>(), {
  modelValue: null,
  labels: () => ({}),
  minBodyHeight: '10rem',
  maxHeight: '',
  bodyClass: '',
  placeholder: '',
  hideToolbar: false,
  outputFormat: 'html',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'upload': [files: FileList]
}>()

const knownHtmlTags = new Set([
  'a','abbr','address','area','article','aside','b','base','bdi','bdo','blockquote','br',
  'button','canvas','caption','cite','code','col','colgroup','data','dd','del','details',
  'dfn','dialog','div','dl','dt','em','embed','fieldset','figcaption','figure','footer',
  'form','h1','h2','h3','h4','h5','h6','header','hgroup','hr','i','iframe','img','input',
  'ins','kbd','label','legend','li','link','main','map','mark','menu','meta','meter','nav',
  'noscript','object','ol','optgroup','option','output','p','picture','pre','progress','q',
  'rp','rt','ruby','s','samp','script','section','select','slot','small','source','span',
  'strong','style','sub','summary','sup','table','tbody','td','template','textarea','tfoot',
  'th','thead','time','title','tr','track','u','ul','var','video','wbr',
])

function escapeXmlTags(html: string): string {
  return html.replace(/<\/?([a-zA-Z][a-zA-Z0-9]*)[^>]*>/g, (match) => {
    const tagName = /<(\/?)([a-zA-Z][a-zA-Z0-9]*)/.exec(match)
    if (!tagName) return match
    if (knownHtmlTags.has(tagName[2].toLowerCase())) return match
    return match.replace(/</g, '&lt;').replace(/>/g, '&gt;')
  })
}

function unescapeXmlTags(html: string): string {
  const blocks: string[] = []
  html = html.replace(/<code[^>]*>[\s\S]*?<\/code>/g, (m) => {
    blocks.push(m)
    return `\x00CB${blocks.length - 1}\x00`
  })
  html = html.replace(/&lt;(\/?[a-zA-Z][a-zA-Z0-9]*(?:\s[^&]*?)?)&gt;/g, '<$1>')
  html = html.replace(/\x00CB(\d+)\x00/g, (_, i) => blocks[+i])
  return html
}

function prepareContent(val: string | null | undefined): string {
  return escapeXmlTags(val || '')
}

const formats = getFormats({ placeholder: props.placeholder })
const toolbar = getToolbar()
const editor = ref<Editor>()
const currentValue = ref('')
const emittedContent = ref(false)

const vm = getCurrentInstance()!
const $SystemAPI = (vm.appContext.config.globalProperties as any).$SystemAPI

function inputToHtml(val: string | null | undefined): string {
  const raw = val || ''
  if (props.outputFormat === 'markdown' && props.toHtml) {
    return escapeXmlTags(props.toHtml(raw))
  }
  return prepareContent(raw)
}

watch(() => props.modelValue, (val) => {
  if (!emittedContent.value && editor.value) {
    editor.value.commands.setContent(inputToHtml(val), false)
  }
  emittedContent.value = false
}, { deep: true })

onMounted(() => {
  editor.value = new Editor({
    extensions: formats,
    content: inputToHtml(props.modelValue),
    parseOptions: {
      preserveWhitespace: 'full',
    },
    onUpdate: onUpdate,
  })
})

onBeforeUnmount(() => {
  if (editor.value) editor.value.destroy()
})

function onUpdate() {
  let editorValue = editor.value!.getHTML().replace(/<p><\/p>/g, '<p><br></p>')
  editorValue = editorValue === '<p><br></p>' ? '' : unescapeXmlTags(editorValue)
  if (props.outputFormat === 'markdown' && props.toMarkdown) {
    editorValue = props.toMarkdown(editorValue)
  }
  currentValue.value = editorValue
  emittedContent.value = true
  emit('update:modelValue', currentValue.value)
}

function focus() {
  if (editor.value) {
    editor.value.commands.focus()
  }
}

function onDrop(event: DragEvent) {
  if (event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files.length > 0) {
    event.preventDefault()
    emit('upload', event.dataTransfer.files)
  }
}

function onPaste(event: ClipboardEvent) {
  if (event.clipboardData && event.clipboardData.files && event.clipboardData.files.length > 0) {
    event.preventDefault()
    emit('upload', event.clipboardData.files)
  }
}

defineExpose({ focus })
</script>

<style lang="scss">
.rt-editor-content {
  height: 100%;

  .ProseMirror {
    height: 100%;
  }

  input[type="checkbox"] {
    pointer-events: auto !important;
    cursor: pointer !important;
  }
}
</style>
