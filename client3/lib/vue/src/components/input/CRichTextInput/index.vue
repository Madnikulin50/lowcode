<template>
  <div class="card border border-light rounded">
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
      class="card-body p-2 rt-editor-content rt-content"
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
  maxBodyHeight?: string
  bodyClass?: string
  placeholder?: string
  hideToolbar?: boolean
}>(), {
  modelValue: null,
  labels: () => ({}),
  minBodyHeight: '10rem',
  maxBodyHeight: '',
  bodyClass: '',
  placeholder: '',
  hideToolbar: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'upload': [files: FileList]
}>()

const formats = getFormats({ placeholder: props.placeholder })
const toolbar = getToolbar()
const editor = ref<Editor>()
const currentValue = ref('')
const emittedContent = ref(false)

const vm = getCurrentInstance()!
const $SystemAPI = (vm.appContext.config.globalProperties as any).$SystemAPI

watch(() => props.modelValue, (val) => {
  if (!emittedContent.value && editor.value) {
    editor.value.commands.setContent(val || '', false)
  }
  emittedContent.value = false
}, { deep: true })

onMounted(() => {
  editor.value = new Editor({
    extensions: formats,
    content: props.modelValue || '',
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
  const editorValue = editor.value!.getHTML().replace(/<p><\/p>/g, '<p><br></p>')
  currentValue.value = editorValue === '<p><br></p>' ? '' : editorValue
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
