<template>
  <div
    :style="previewStyle"
    :class="[...previewClass, 'pdf-preview', inline ? 'inline' : '', attrs.onClick ? 'clickable' : '']"
    @click.stop="onPreviewClick"
  >
    <div
      v-show="show"
      ref="pages"
      class="pages shadow-sm"
    />

    <div
      v-if="loadError"
      class="doc-msg doc-err"
    >
      <p class="err-message">
        {{ loadError.message }}
      </p>
    </div>

    <div
      v-else-if="!show && labels.loading"
      class="doc-msg"
    >
      <p class="d-flex align-items-center gap-1">
        <span class="spinner-border spinner-border-sm" />
        {{ labels.loading }}
      </p>
    </div>

    <div
      v-else-if="!pageCount && labels.noPages"
      class="doc-msg doc-err"
    >
      <p>{{ labels.noPages }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, useAttrs } from 'vue'
import * as pdfjsLib from 'pdfjs-dist'
import { makePlaceholder, makeFailedPage, Page, Document } from './helpers'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()

const props = defineProps<{
  inline?: boolean
  retryBackoff?: number
  maxRetries?: number
  maxPages?: number
  initialScale?: number
}>()

const emit = defineEmits<{
  (e: 'openPreview', payload: { document: any }): void
  (e: 'error', err: Error): void
}>()

pdfjsLib.GlobalWorkerOptions.workerSrc = new URL('pdfjs-dist/build/pdf.worker.min.mjs', import.meta.url).toString()

function sleep(t: number) {
  return new Promise(resolve => setTimeout(resolve, t))
}

const pages = ref<HTMLElement | null>(null)

const document_ = ref<any>(null)
const pagesArr = ref<any[]>([])
const show = ref(false)
const loadError = ref<Error | undefined>(undefined)

const src = computed(() => (attrs as any).src)
const labels = computed(() => (attrs as any).labels || {})
const previewStyle = computed(() => (attrs as any).previewStyle || {})
const previewClass = computed(() => (attrs as any).previewClass || [])
const inline = computed(() => (attrs as any).inline ?? props.inline)

const pageCount = computed(() => {
  if (!document_.value || !document_.value.pdf) {
    return 0
  }
  return document_.value.pdf.numPages
})

const hasMore = computed(() => (props.maxPages ?? 25) < pageCount.value)

function onPreviewClick() {
  if (loadError.value) {
    init()
  } else {
    emit('openPreview', { document: document_.value })
  }
}

async function init() {
  document_.value = null
  pagesArr.value = []
  show.value = false
  loadError.value = undefined

  return loadDocument(src.value)
    .then(renderDocument)
    .catch(stdErr)
}

async function pdfjsLoad(src: string) {
  return pdfjsLib.getDocument({
    url: src,
    useWorkerFetch: true,
    isEvalSupported: true,
    useSystemFonts: true,
  }).promise
}

async function loadDocument(src: any) {
  if (src instanceof Document) {
    document_.value = new Document({ ...src, scale: props.initialScale ?? 1 })
  } else if (typeof src === 'string') {
    let retries = 0
    let err: any
    const maxRetries = props.maxRetries ?? 10
    const retryBackoff = props.retryBackoff ?? 300
    const pdfl = async () => {
      return sleep(retries * retryBackoff)
        .then(() => pdfjsLoad(src))
        .then(pdf => {
          document_.value = new Document({ pdf, src, scale: props.initialScale ?? 1 })
        })
    }

    while (!document_.value && retries < maxRetries) {
      await pdfl().catch((e: any) => {
        retries++
        err = e
      })
    }

    if (!document_.value) {
      throw err
    }
  } else {
    throw new Error('src.notValid')
  }
  return document_.value
}

async function renderDocument(doc: any) {
  const rf = pages.value!
  const maxPages = props.maxPages ?? 25
  const pgCount = Math.min(pageCount.value, maxPages)
  pagesArr.value = [...new Array(pgCount)].map((_, i) => new Page({ index: i }))

  if (pgCount <= 0) {
    show.value = true
    return
  }

  for (let i = 0; i < pgCount; i++) {
    const node = makePlaceholder(labels.value)
    rf.appendChild(node)
    pagesArr.value.splice(i, 1, new Page({ ...pagesArr.value[i], node, loading: true }))

    renderPage(pagesArr.value[i], doc, rf)
      .then(page => {
        pagesArr.value.splice(page.index, 1, page)

        if (page.index === 0) {
          show.value = true
        }
      })
      .catch(stdErr)
  }
}

async function renderPage(page: any, doc: any, rf: Node) {
  return doc.pdf.getPage(page.index + 1).then((p: any) => {
    const np = new Page(page)
    np.loading = false
    np.loaded = true
    np.page = p

    const canvas = document.createElement('canvas')
    const scale = doc.scale
    const viewport = np.page.getViewport({ scale })
    const canvasContext = canvas.getContext('2d')!
    const renderContext = { canvasContext, viewport }

    canvas.height = viewport.height
    canvas.width = viewport.width

    return np.page.render(renderContext).promise.then(() => {
      np.node = canvas
      np.rendered = true
      if (inline.value) {
        canvas.classList.add('inline')
      }
      return np
    })

      .catch(() => {
        const node = makeFailedPage(labels.value)
        np.node = node
        np.rendered = false
        np.failed = true
        return np
      })

      .then((np: any) => {
        rf.replaceChild(np.node, page.node)
        return np
      })
  })
}

function stdErr(err: Error) {
  console.error(err)
  loadError.value = err
  emit('error', err)
}

function setDefaultValues() {
  document_.value = null
  pagesArr.value = []
  show.value = false
  loadError.value = undefined
}

onMounted(() => {
  if (!src.value) {
    stdErr(new Error('src.missing'))
    return
  }

  nextTick(() => init())
})

onBeforeUnmount(() => {
  setDefaultValues()
})
</script>

<style lang="scss" scoped>
.doc-msg {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  width: 100%;
  background-color: var(--white);
}
.doc-err {
  cursor: pointer;

  .err-message {
    color: var(--danger);
  }
}
</style>

<style lang="scss">
.pdf-preview {
  text-align: center;
  width: 100%;
  height: auto;

  &.inline {
    cursor: zoom-in;

    canvas {
      width: 100%;
      height: auto;
    }
  }

  canvas {
    margin-bottom: 10px;
    width: 80%;
    height: auto;

    &:last-of-type {
      margin-bottom: unset;
    }
  }

  .loader {
    margin-bottom: 10px;
  }
}
</style>
