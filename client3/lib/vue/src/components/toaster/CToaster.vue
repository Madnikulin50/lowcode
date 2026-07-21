<template>
  <div
    class="toast-container position-fixed top-0 end-0 p-3"
    style="z-index: 1080"
  >
    <div
      v-for="t in toasts"
      :key="t.id"
      :ref="(el: any) => setToastRef(el, t)"
      class="toast overflow-unset"
      :class="t.options?.toastClass"
      role="alert"
      aria-live="assertive"
      aria-atomic="true"
    >
      <div class="toast-header">
        <strong class="me-auto">{{ t.payload.title }}</strong>
        <button
          v-if="t.actions?.hide"
          type="button"
          class="btn-close"
          data-bs-dismiss="toast"
          aria-label="Close"
        />
      </div>
      <div class="toast-body">
        <div class="card border-0 bg-transparent">
          <div class="card-body p-0">
            <p
              v-if="t.payload.notes"
              class="card-text"
            >
              {{ t.payload.notes }}
            </p>
            <p
              v-if="t.payload.link"
              class="card-text"
            >
              <c-toaster-link :link="t.payload.link" />
            </p>
            <component
              :is="actComponent(act)"
              v-for="([name, act]) in extraActions(t)"
              :key="name"
              class="mr-1"
              v-bind="act"
              @action="act.cb ? act.cb(t, $event) : undefined"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch, onBeforeUnmount } from 'vue'
import { Toast } from 'bootstrap'
import * as actions from './actions'
import CToasterLink from './CToasterLink.vue'

const props = defineProps<{
  toasts: Array<Record<string, any>>
}>()

const toastInstances = new Map<string, { instance: Toast; el: HTMLElement }>()

function extraActions({ actions: acts = {} }: Record<string, any>) {
  const { hide, ...act } = acts
  return Object.entries(act)
}

function actComponent({ kind }: Record<string, any>) {
  const act = (actions as Record<string, any>)[kind]
  if (!act) {
    throw new Error('toast.actionKind.unknown')
  }
  return act
}

function onHidden(t: Record<string, any>) {
  if (t.actions?.hide?.cb) {
    t.actions.hide.cb(t)
  }
  toastInstances.delete(t.id)
}

function setToastRef(el: any, t: Record<string, any>) {
  if (!el) return
  if (toastInstances.has(t.id)) return

  const instance = new Toast(el)
  instance.show()
  toastInstances.set(t.id, { instance, el })

  el.addEventListener('hidden.bs.toast', () => onHidden(t), { once: true })
}

onBeforeUnmount(() => {
  toastInstances.forEach(({ instance }) => {
    instance.hide()
  })
  toastInstances.clear()
})
</script>

<style lang="scss">
.toast.overflow-unset {
  overflow: unset !important;
}
</style>
