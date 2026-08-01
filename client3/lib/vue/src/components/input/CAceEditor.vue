<template>
  <div
    ref="container"
    class="position-relative ace-editor-wrapper"
    :class="{ 'resizable': resizable }"
    :style="containerStyle"
  >
    <v-ace-editor
      ref="aceeditor"
      v-bind="$attrs"
      :value="modelValue"
      :lang="lang"
      :theme="theme"
      :options="editorOptions"
      :placeholder="placeholder"
      :readonly="readOnly"
      :wrap="true"
      :print-print-margin="showPrintMargin"
      width="100%"
      :height="effectiveHeight"
      :class="{ 'border-0 rounded-0': !border }"
      @init="editorInit"
      @update:value="onValueUpdate"
    />

    <button
      v-if="showPopout"
      class="popout position-absolute px-2 py-1 me-3"
      variant="link"
      @click="$emit('open')"
    >
      <font-awesome-icon
        :icon="['fas', 'expand-alt']"
      />
    </button>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { VAceEditor } from 'vue3-ace-editor'
import ace from 'ace-builds'
import 'ace-builds/src-noconflict/ext-language_tools'
import 'ace-builds/src-noconflict/mode-text'
import 'ace-builds/src-noconflict/mode-css'
import 'ace-builds/src-noconflict/mode-html'
import 'ace-builds/src-noconflict/mode-json'
import 'ace-builds/src-noconflict/mode-javascript'
import 'ace-builds/src-noconflict/mode-scss'
import 'ace-builds/src-noconflict/theme-chrome'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faExpandAlt } from '@fortawesome/free-solid-svg-icons'

ace.config.set('basePath', '')

library.add(faExpandAlt)

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  lang: {
    type: String,
    default: 'text',
  },
  theme: {
    type: String,
    default: 'chrome',
  },
  minHeight: {
    type: String,
    default: '2.35rem',
  },
  showLineNumbers: {
    type: Boolean,
    default: false,
  },
  fontSize: {
    type: String,
    default: '14px',
  },
  border: {
    type: Boolean,
    default: true,
  },
  showPopout: {
    type: Boolean,
    default: false,
  },
  readOnly: {
    type: Boolean,
    default: false,
  },
  autoComplete: {
    type: Boolean,
    default: false,
  },
  highlightActiveLine: {
    type: Boolean,
    default: false,
  },
  showPrintMargin: {
    type: Boolean,
    default: false,
  },
  autoCompleteSuggestions: {
    type: [Array, Object],
    default: () => ([]),
  },
  initExpressions: {
    type: Boolean,
    required: false,
  },
  fontFamily: {
    type: String,
    default: '',
  },
  placeholder: {
    type: String,
    default: '',
  },
  resizable: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue', 'open'])

const container = ref(null)
const aceeditor = ref(null)

const manualHeight = ref(null)
const contentHeight = ref(null)
const programmaticUpdate = ref(false)
const resizeTimeout = ref(null)
let resizeObserver = null

const editorOptions = computed(() => {
  const opts = {
    tabSize: 2,
    fontSize: props.fontSize,
    wrap: true,
    indentedSoftWrap: false,
    showPrintMargin: props.showPrintMargin,
    showLineNumbers: props.showLineNumbers,
    showGutter: props.showLineNumbers,
    displayIndentGuides: props.lang !== 'text',
    useWorker: false,
    readOnly: props.readOnly,
    highlightActiveLine: props.highlightActiveLine,
    cursorStyle: 'smooth',
  }

  if (props.fontFamily) {
    opts.fontFamily = props.fontFamily
  }

  if (props.fontSize) {
    opts.fontSize = props.fontSize
  }

  const minHeightPx = parseHeight(props.minHeight)
  const computedMinLines = Math.max(1, Math.floor(minHeightPx / 16))
  opts.minLines = computedMinLines

  if (!props.resizable) {
    opts.maxLines = computedMinLines
  }

  return opts
})

const containerStyle = computed(() => {
  const style = {}
  if (props.minHeight) {
    style.minHeight = props.minHeight
  }
  if (props.resizable) {
    style.height = effectiveHeight.value
  }
  return style
})

const effectiveHeight = computed(() => {
  if (props.resizable) {
    const manualHeightPx = manualHeight.value ? parseHeight(manualHeight.value) : 0
    const contentHeightPx = contentHeight.value ? parseHeight(contentHeight.value) : 0
    const minHeightPx = parseHeight(props.minHeight)
    const maxHeight = Math.max(manualHeightPx, contentHeightPx, minHeightPx)
    return `${maxHeight}px`
  }
  return props.minHeight
})

function parseHeight(height) {
  if (typeof height === 'number') return height
  if (typeof height === 'string') {
    return parseFloat(height.replace('px', '').replace('rem', '')) * (height.includes('rem') ? 16 : 1)
  }
  return 0
}

function onValueUpdate(value = '') {
  emit('update:modelValue', value)
}

watch(() => props.modelValue, () => {
  if (props.resizable && aceeditor.value && aceeditor.value._editor) {
    nextTick(() => {
      calculateContentHeight(aceeditor.value._editor)
    })
  }
})

onMounted(() => {
  if (props.resizable) {
    setupResizeObserver()
  }
})

onBeforeUnmount(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  if (resizeTimeout.value) {
    clearTimeout(resizeTimeout.value)
  }
})

function calculateContentHeight(editor) {
  if (!editor) return

  const session = editor.session
  const lineCount = session.getLength()
  const lineHeight = editor.renderer.lineHeight || 16
  const padding = 14

  const calcHeight = (lineCount * lineHeight) + padding
  const minHeightPx = parseHeight(props.minHeight)
  const finalHeight = Math.max(calcHeight, minHeightPx)

  programmaticUpdate.value = true
  contentHeight.value = `${finalHeight}px`

  nextTick(() => {
    setTimeout(() => {
      programmaticUpdate.value = false
    }, 50)
  })
}

function setupResizeObserver() {
  let lastHeight = null
  let isInitialized = false

  resizeObserver = new ResizeObserver((entries) => {
    if (programmaticUpdate.value) return

    for (const entry of entries) {
      const newHeight = entry.contentRect.height

      if (!isInitialized) {
        lastHeight = newHeight
        isInitialized = true
        return
      }

      if (lastHeight !== null && Math.abs(lastHeight - newHeight) > 2) {
        if (resizeTimeout.value) {
          clearTimeout(resizeTimeout.value)
        }

        resizeTimeout.value = setTimeout(() => {
          manualHeight.value = `${newHeight}px`
          updateEditorSize()
          lastHeight = newHeight
        }, 10)
      }
    }
  })

  if (container.value) {
    resizeObserver.observe(container.value)
  }
}

function updateEditorSize() {
  nextTick(() => {
    if (aceeditor.value && aceeditor.value._editor) {
      aceeditor.value._editor.resize()
    }
  })
}

function editorInit(editor) {
  editor.setOptions({
    ...editorOptions.value,
    ...(props.autoComplete && {
      enableBasicAutocompletion: true,
      enableLiveAutocompletion: true,
      enableSnippets: true,
    }),
  })

  if (props.resizable) {
    nextTick(() => {
      calculateContentHeight(editor)
    })

    editor.session.on('change', () => {
      calculateContentHeight(editor)
    })
  }

  if (props.initExpressions) {
    processExpressionAutoComplete(editor)
  } else if (props.autoComplete) {
    const staticWordCompleter = {
      getCompletions: (editor, session, pos, prefix, callback) => {
        callback(
          null,
          props.autoCompleteSuggestions.map(({ caption, value, meta }) => ({
            caption,
            value,
            meta,
          })),
        )
      },
    }

    editor.completers.push(staticWordCompleter)
  }
}

function processExpressionAutoComplete(editor) {
  const getTextContext = (pos) => {
    const session = editor.session
    const line = session.getLine(pos.row)
    const lastSpaceIndex = Math.max(0, line.lastIndexOf(' '))
    const textAfterSpace = line.slice(lastSpaceIndex, pos.column).trim()
    const lastDotIndex = textAfterSpace.lastIndexOf('.')
    const searchTextForCaption = lastDotIndex >= 0 ? textAfterSpace.slice(lastDotIndex + 1) : textAfterSpace
    return { line, textAfterSpace, searchTextForCaption }
  }

  const matchesSuggestion = (suggestion, textAfterSpace, searchTextForCaption) => {
    const suggestionValue = typeof suggestion === 'string' ? suggestion : suggestion.value
    const suggestionCaption = typeof suggestion === 'string' ? suggestion : suggestion.caption
    return suggestionValue.toLowerCase().startsWith(textAfterSpace.toLowerCase()) ||
           suggestionCaption.toLowerCase().startsWith(searchTextForCaption.toLowerCase())
  }

  const checkAndTriggerAutocomplete = () => {
    setTimeout(() => {
      const pos = editor.getCursorPosition()
      const { line, textAfterSpace, searchTextForCaption } = getTextContext(pos)
      const charBeforeCursor = pos.column > 0 ? line[pos.column - 1] : ''
      if (charBeforeCursor === '}' || charBeforeCursor === ')' || textAfterSpace.length === 0) return
      const context = getContext(editor, editor.session, pos)
      const suggestions = getSuggestionsForContext(context)
      const hasMatches = suggestions.some(s => matchesSuggestion(s, textAfterSpace, searchTextForCaption))
      if (hasMatches) {
        editor.execCommand('startAutocomplete')
      }
    }, 10)
  }

  const staticWordCompleter = {
    identifierRegexps: [/[${\w]+/],
    getCompletions: (editor, session, pos, prefix, callback) => {
      const context = getContext(editor, session, pos)
      const suggestions = getSuggestionsForContext(context)
      const { textAfterSpace, searchTextForCaption } = getTextContext(pos)
      const filteredSuggestions = suggestions
        .filter(s => matchesSuggestion(s, textAfterSpace, searchTextForCaption))
        .map(suggestion => {
          const caption = typeof suggestion === 'string' ? suggestion : suggestion.caption
          const value = typeof suggestion === 'string' ? suggestion : suggestion.value
          const captionMatch = caption.toLowerCase().indexOf(searchTextForCaption.toLowerCase())
          return {
            caption,
            value,
            score: captionMatch === 0 ? 10000 : 1000,
            meta: 'variable',
            completer: {
              insertMatch: (insertEditor, data) => {
                insertEditor.jumpToMatching()
                const line = session.getLine(pos.row)
                const spaceIndex = line.lastIndexOf(' ')
                const startCol = spaceIndex > 0 ? spaceIndex + 1 : 0
                insertEditor.session.replace({
                  start: { row: pos.row, column: startCol },
                  end: { row: pos.row, column: pos.column },
                }, data.value)
                checkAndTriggerAutocomplete()
              },
            },
          }
        })
      callback(null, filteredSuggestions)
    },
  }

  editor.completers = [staticWordCompleter]

  editor.commands.on('afterExec', (e) => {
    if (['insertstring', 'Return'].includes(e.command.name) || /^[\w.($]$/.test(e.args)) {
      checkAndTriggerAutocomplete()
    }
  })

  editor.renderer.setScrollMargin(7, 7)
  editor.renderer.setPadding(10)
}

function getContext(editor, session, pos) {
  const line = session.getLine(pos.row)
  const lastSpaceIndex = Math.max(0, line.lastIndexOf(' '))
  const textBeforeCursor = line.slice(lastSpaceIndex, pos.column).trim()
  return textBeforeCursor.split('.').slice(0, -1).join('.')
}

function getSuggestionsForContext(context) {
  return props.autoCompleteSuggestions[context] || []
}
</script>

<style scoped>
.popout {
  z-index: 7;
  bottom: 0;
  right: 0;
}

.ace-editor-wrapper {
  &.resizable {
    resize: vertical;
    overflow: auto;
  }
}
</style>
