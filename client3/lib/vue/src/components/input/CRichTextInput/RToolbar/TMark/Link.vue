<template>
  <div id="link-popover-container" class="position-relative">
    <button
      id="link-popover"
      class="btn btn-link text-dark fw-bold text-decoration-none"
      @click="showPopover"
    >
      <span :class="activeClasses(format.attrs)">
        <font-awesome-icon icon="link" />
      </span>
    </button>

    <div
      v-if="visible"
      ref="popover"
      class="link-popover-dropdown"
    >
      <div class="input-group" style="min-width: 250px;">
        <input
          v-model="attrs.href"
          type="url"
          class="form-control"
          autofocus
          :placeholder="labels.urlPlaceholder"
          @keydown.enter.prevent.stop="link"
          @keydown.esc.prevent.stop="close"
        />
        <div class="input-group-append">
          <button
            class="btn btn-outline-success"
            @click="link"
          >
            {{ labels.ok }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  editor: any
  format: any
  isActive?: any
  getMarkAttrs?: (...args: any[]) => any
  currentValue?: string
  labels: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'click', payload: { type: string; attrs: Record<string, any> }): void
}>()

const visible = ref(false)
const attrs = ref<{ href: string | null; target: string }>({ href: null, target: '_self' })

const urlValid = computed(() => {
  if (!attrs.value.href) {
    return false
  }
  return !!attrs.value.href
})

function showPopover() {
  if (props.currentValue) {
    visible.value = true
    attrs.value = { ...(props.getMarkAttrs ? props.getMarkAttrs(props.format.type) : {}), target: '_self' }
  }
}

function link() {
  onClick(props.format.type, attrs.value)
  close()
}

function close() {
  attrs.value.href = null
  visible.value = false
}

function onClick(type: string, attrs: Record<string, any>) {
  emit('click', { type, attrs })
}

function activeClasses(attrs?: Record<string, any>) {
  const isActive = props.editor.isActive(props.format.type, attrs)
  if (isActive) {
    return ['text-primary']
  }
  return undefined
}
</script>

<style lang="scss">
.link-popover-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  z-index: 1060;
  background: var(--white, #fff);
  border: 1px solid var(--light, #f8f9fa);
  border-radius: 0.25rem;
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
  padding: 0.5rem;
}
</style>
