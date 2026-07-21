<template>
  <div>
    <div
      :class="['offcanvas offcanvas-end', { show: isVisible }]"
      class="notification-sidebar bg-white shadow"
      tabindex="-1"
      style="width: 400px;"
      :data-bs-backdrop="isMobile ? 'true' : 'false'"
    >
      <div class="offcanvas-header d-flex align-items-center justify-content-between bg-white ps-2 pe-3 py-3">
        <h5 class="text-primary mb-0">
          <b>{{ $t('title') }}</b>
        </h5>
        <button
          type="button"
          class="btn btn-outline-light d-flex align-items-center justify-content-center p-2 border-0 text-secondary"
          @click="notificationsStore.visible = false"
        >
          <font-awesome-icon
            :icon="['fas', 'times']"
            class="h6 mb-0"
          />
        </button>
      </div>
      <div class="offcanvas-body d-flex flex-column overflow-hidden bg-white p-0">
        <notifications />
      </div>
    </div>
    <div
      v-if="isMobile && isVisible"
      class="offcanvas-backdrop fade show"
      @click="notificationsStore.visible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useNotificationsStore } from '@/store'
import { throttle } from 'lodash'
import Notifications from './Notifications.vue'

const notificationsStore = useNotificationsStore()
notificationsStore.visible = false

const isMobile = ref(false)

const isVisible = computed({
  get() {
    return notificationsStore.visible
  },
  set(val: boolean) {
    notificationsStore.visible = val
  },
})

watch(isVisible, (val) => {
  if (val) {
    window.dispatchEvent(new CustomEvent('right-sidebar:opened', { detail: 'notifications' }))
  }
})

function handleSidebarOpened(e: Event) {
  const detail = (e as CustomEvent).detail
  if (detail !== 'notifications') {
    isVisible.value = false
  }
}

const checkIfMobile = throttle(function () {
  isMobile.value = window.innerWidth < 1024
}, 500)

onMounted(() => {
  checkIfMobile()
  window.addEventListener('resize', checkIfMobile)
  window.addEventListener('right-sidebar:opened', handleSidebarOpened)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', checkIfMobile)
  window.removeEventListener('right-sidebar:opened', handleSidebarOpened)
})
</script>

<style lang="scss">
.offcanvas-backdrop {
  opacity: 0.75 !important;
}

.offcanvas.notification-sidebar {
  display: none !important;
}

.offcanvas.notification-sidebar.show {
  display: flex !important;
  visibility: visible !important;
  transform: none !important;
}

@media (min-width: 1024px) {
  .offcanvas.notification-sidebar {
    top: calc(var(--topbar-height) + 0.5rem) !important;
    right: 0.5rem !important;
    height: calc(100% - var(--topbar-height) - 1rem) !important;
    border-radius: 1rem !important;
    border: none !important;
    z-index: 1048 !important;
  }

  .offcanvas.notification-sidebar .offcanvas-header {
    border-top-left-radius: 1rem !important;
    border-top-right-radius: 1rem !important;
  }

  .offcanvas.notification-sidebar .offcanvas-body {
    border-bottom-left-radius: 1rem !important;
    border-bottom-right-radius: 1rem !important;
  }
}
</style>
