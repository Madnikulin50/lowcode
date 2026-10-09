<template>
  <div
    ref="modalRef"
    class="modal fade"
    tabindex="-1"
    aria-hidden="true"
  >
    <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            {{ current ? current.title : 'Workflow prompts' }}
          </h5>
          <button
            type="button"
            class="btn-close"
            @click="deactivate()"
          />
        </div>
        <div class="modal-body">
          <div
            v-if="isLoading"
            class="d-flex justify-content-center py-3"
          >
            <span class="spinner-border" />
          </div>

          <component
            :is="current.component"
            v-else-if="current"
            :payload="current.prompt.payload"
            :loading="isLoading"
            @submit="resume({ input: $event, prompt: current.prompt })"
          />

          <div v-else>
            <div
              v-for="({ key, title, age, prompt }) in list"
              :key="key"
              class="d-flex flex-grow-1 align-items-baseline mb-2"
            >
              <a
                class="p-0 ms-auto"
                @click="activate(prompt)"
              >
                {{ title }} -
                <time
                  class="muted small"
                  :datetime="prompt.createdAt"
                >
                  {{ age }}
                </time>
              </a>
            </div>
          </div>
        </div>
        <div
          v-if="current"
          class="modal-footer d-flex"
        >
          <button
            class="btn btn-link ms-auto"
            @click="activate(true)"
          >
            &laquo; Back to list
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useWfPromptsStore } from '../../store/wf-prompts'
import definitions from './kinds/index.ts'
import { pVal } from './utils.ts'
import moment from 'moment'
import * as bootstrap from 'bootstrap'

const wfPrompts = useWfPromptsStore()

const modalRef = ref<HTMLDivElement>()
let bsModal: bootstrap.Modal | null = null

const isLoading = computed(() => wfPrompts.loading)
const isActive = computed(() => wfPrompts.active !== false)
const prompts = computed(() => wfPrompts.prompts)

const list = computed(() => prompts.value
  .filter(({ ref }: any) => !!definitions[ref] && !!definitions[ref].component)
  .map((prompt: any) => ({ ...definitions[prompt.ref], prompt }))
  .filter(({ passive }: any) => !passive)
  .map((p: any) => ({
    key: p.prompt.stateID,
    title: pVal(p.prompt.payload, 'title', 'Workflow prompt'),
    age: moment(p.prompt.createdAt).fromNow(),
    ...p,
  }))
)

const current = computed(() => {
  if (typeof wfPrompts.active === 'boolean') return undefined
  const c = wfPrompts.active
  if (!c) return undefined
  return list.value.find(({ prompt }: any) => prompt.stateID === c.stateID)
})

function remove(prompt: any) {
  wfPrompts.remove(prompt)
}

function resume(values: any) {
  wfPrompts.resume(values)
}

function activate(prompt?: any) {
  wfPrompts.activate(prompt)
}

function deactivate() {
  wfPrompts.deactivate()
}

function clear() {
  deactivate()
}

watch(isActive, (active) => {
  if (active) {
    bsModal?.show()
  } else {
    bsModal?.hide()
  }
})

onMounted(() => {
  if (modalRef.value) {
    bsModal = new bootstrap.Modal(modalRef.value, { backdrop: 'static' })
    modalRef.value.addEventListener('hidden.bs.modal', () => deactivate())
  }
})

onBeforeUnmount(() => {
  bsModal?.dispose()
})
</script>
