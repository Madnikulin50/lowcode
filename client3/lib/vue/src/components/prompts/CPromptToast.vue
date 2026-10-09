<template>
  <div
    v-if="!hideToasts"
    class="toast-container position-fixed top-0 end-0 p-3"
    style="z-index: 1060"
  >
    <div
      v-for="({ prompt, component, passive }) in toasts"
      :id="'wfPromptToast-'+prompt.stateID"
      :key="'wfPromptToast-'+prompt.stateID"
      class="toast"
      :class="'bg-'+ (pVal(prompt.payload, 'variant', 'primary'))"
      role="alert"
    >
      <div class="toast-header">
        <strong class="me-auto">{{ pVal(prompt.payload, 'title', 'Workflow prompt') }}</strong>
        <button
          type="button"
          class="btn-close"
          @click="onToastHide({ prompt, passive })"
        />
      </div>
      <div class="toast-body">
        <component
          :is="component"
          v-if="component"
          :payload="prompt.payload"
          :loading="isLoading"
          @submit="resumeToast({ input: $event, prompt })"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useWfPromptsStore } from '../../store/wf-prompts'
import definitions from './kinds/index.ts'
import { pVal } from './utils.ts'

const props = defineProps<{
  hideToasts?: boolean
}>()

const wfPrompts = useWfPromptsStore()

const passivePrompts = ref<any[]>([])
const hasFocus = ref<boolean | null>(null)
const hasFocusObserver = ref(0)

const prompts = computed(() => wfPrompts.prompts)
const isActive = computed(() => wfPrompts.active !== false)
const isLoading = computed(() => wfPrompts.loading)

const withHandlers = computed(() => (hasFocus.value ? prompts.value : [])
  .filter(({ ref }: any) => !!definitions[ref] && !!definitions[ref].handler)
  .map((prompt: any) => ({ ...definitions[prompt.ref], prompt }))
)

const withComponents = computed(() => (hasFocus.value ? prompts.value : [])
  .filter(({ ref }: any) => !!definitions[ref] && !!definitions[ref].component)
  .map((prompt: any) => ({ ...definitions[prompt.ref], prompt }))
)

const activePrompts = computed(() => withComponents.value.filter(({ passive }: any) => !passive))

const toasts = computed(() => props.hideToasts
  ? []
  : [
    ...passivePrompts.value,
    ...activePrompts.value,
  ]
)

const defaultTimeout = 7

watch(withHandlers, (hh) => {
  if (hh.length > 0) {
    const { handler, prompt } = hh.shift()
    resume({ input: {}, prompt }).then(() => {
      handler.call(undefined, prompt.payload)
    })
  }
})

watch(withComponents, (wc) => {
  wc.forEach((p: any) => {
    if (p.passive && !passivePrompts.value.some(({ prompt }: any) => prompt.stateID === p.prompt.stateID)) {
      passivePrompts.value.push(p)
    }
  })
}, { immediate: true })

onMounted(() => {
  setDocumentFocusObserver()
})

onBeforeUnmount(() => {
  clearDocumentFocusObserver()
  setDefaultValues()
})

function resume(values: any) {
  return wfPrompts.resume(values)
}

function cancel(prompt: any) {
  wfPrompts.cancel(prompt)
}

function activate(prompt?: any) {
  wfPrompts.activate(prompt)
}

function resumeToast(values: any) {
  if (values.input && values.input.keep) {
    values.input = {}
  }
  resume(values)
}

function onToastHide({ prompt, passive }: any) {
  setTimeout(() => {
    if (passive) {
      passivePrompts.value = passivePrompts.value.filter(({ prompt: p }: any) => p.stateID !== prompt.stateID)
    } else {
      cancel(prompt)
    }
  }, 300)
}

function clearDocumentFocusObserver() {
  if (hasFocusObserver.value) {
    window.clearInterval(hasFocusObserver.value)
  }
}

function setDocumentFocusObserver() {
  clearDocumentFocusObserver()
  hasFocusObserver.value = window.setInterval(() => {
    const f = document.hasFocus()
    if (hasFocus.value !== f) {
      hasFocus.value = f
    }
  }, 1000)
}

function setDefaultValues() {
  passivePrompts.value = []
  hasFocus.value = null
  hasFocusObserver.value = 0
}
</script>

<style lang="scss">
.toast-header {
  align-items: start;
  padding: 0.375rem 0.75rem;

  strong {
    word-break: break-word;
  }

  .btn-close {
    margin-bottom: 0 !important;
  }
}

.toast-container {
  .toast {
    &.toaster-enter-active,
    &.toaster-leave-active,
    &.toaster-move {
      transition: transform 0.3s ease-in-out;
      opacity: 1;
    }

    &.toaster-enter {
      transform: translate(0, -100%);
      opacity: 0;
    }

    &.toaster-enter-to,
    &.toaster-enter-active {
      transform: translate(0, 0);
      opacity: 1;
    }

    &.toaster-leave-active {
      position: absolute;
      transform: translate(0, -100%);
      opacity: 0;
    }

    &.toaster-leave-to {
      opacity: 0;
    }
  }
}
</style>
