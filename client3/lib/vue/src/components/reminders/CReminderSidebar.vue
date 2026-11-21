<template>
  <div
    class="offcanvas offcanvas-end reminder-sidebar"
    :class="{ show: isVisible }"
    tabindex="-1"
    data-bs-backdrop="static"
    :data-bs-scroll="!isMobile"
    style="width: 400px"
  >
    <div class="offcanvas-header d-flex align-items-center justify-content-between bg-white pe-2 ps-3 py-3 border-bottom">
      <h5 class="text-primary mb-0">
        <b>{{ title }}</b>
      </h5>
      <button
        class="btn btn-outline-light d-flex align-items-center justify-content-center p-2 border-0 text-secondary"
        @click="isVisible = false"
      >
        <font-awesome-icon
          :icon="['fas', 'times']"
          class="h6 mb-0"
        />
      </button>
    </div>
    <div class="offcanvas-body d-flex flex-column overflow-hidden bg-white">
      <slot />
    </div>
  </div>
  <div
    v-if="isMobile && isVisible"
    class="offcanvas-backdrop fade show"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { throttle } from 'lodash'

const props = withDefaults(defineProps<{
  title?: string
  visible?: boolean
}>(), {
  title: '',
  visible: false,
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const isMobile = ref(false)

const isVisible = computed({
  get: () => props.visible,
  set: (visible: boolean) => emit('update:visible', visible),
})

watch(isVisible, (visible) => {
  if (visible) {
    window.dispatchEvent(new CustomEvent('right-sidebar:opened', { detail: 'reminders' }))
  }
})

onMounted(() => {
  checkIfMobile()
  window.addEventListener('resize', checkIfMobile)
  window.addEventListener('right-sidebar:opened', handleSidebarOpened as EventListener)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', checkIfMobile)
  window.removeEventListener('right-sidebar:opened', handleSidebarOpened as EventListener)
})

const checkIfMobile = throttle(() => {
  isMobile.value = window.innerWidth < 1024
}, 500)

function handleSidebarOpened(e: CustomEvent) {
  if (e.detail !== 'reminders') {
    isVisible.value = false
  }
}
</script>

<style lang="scss">
.offcanvas-backdrop {
  opacity: 0.75 !important;
}

.offcanvas.reminder-sidebar {
  display: none !important;
}

.offcanvas.reminder-sidebar.show {
  display: flex !important;
  visibility: visible !important;
  transform: none !important;
}

@media (min-width: 1024px) {
  .offcanvas.reminder-sidebar {
    top: calc(var(--topbar-height) + 0.5rem) !important;
    right: 0.5rem !important;
    height: calc(100% - var(--topbar-height) - 1rem) !important;
    border-radius: 1rem !important;
    border: none !important;
    z-index: 1048 !important;
  }

  .offcanvas.reminder-sidebar .offcanvas-header {
    border-top-left-radius: 1rem !important;
    border-top-right-radius: 1rem !important;
  }

  .offcanvas.reminder-sidebar .offcanvas-body {
    border-bottom-left-radius: 1rem !important;
    border-bottom-right-radius: 1rem !important;
  }
}
</style>
