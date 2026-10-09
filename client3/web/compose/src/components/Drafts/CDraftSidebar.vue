<template>
  <div v-if="isVisible" class="offcanvas offcanvas-end show draft-sidebar" tabindex="-1">
    <div class="offcanvas-header d-flex align-items-center justify-content-between bg-white px-3 py-3 border-bottom">
      <h5 class="text-primary mb-0">
        <b>{{ $t('title') }}</b>
      </h5>
      <button
        class="btn btn-outline-light d-flex align-items-center justify-content-center p-2 border-0 text-secondary"
        title="Close"
        @click="isVisible = false"
      >
        <font-awesome-icon :icon="['fas', 'times']" class="h6 mb-0" />
      </button>
    </div>
    <div class="offcanvas-body d-flex flex-column overflow-hidden bg-white p-0">
      <Drafts />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'drafts' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDraftsStore } from 'corteza-lib/vue/dist'
import { throttle } from 'lodash'
import Drafts from 'corteza-webapp-compose/src/components/Drafts/Drafts.vue'

const { t: $t } = useI18n({ useScope: 'global' })
const draftsStore = useDraftsStore()

const isMobile = ref(false)

const isVisible = computed({
  get: () => draftsStore.visible,
  set: (val) => draftsStore.setVisible(val),
})

watch(isVisible, (visible) => {
  if (visible) {
    window.dispatchEvent(new CustomEvent('right-sidebar:opened', { detail: 'drafts' }))
  }
})

function handleSidebarOpened(e) {
  if (e.detail !== 'drafts') {
    isVisible.value = false
  }
}

const checkIfMobile = throttle(function () {
  isMobile.value = window.innerWidth < 1024
}, 500)

onMounted(() => {
  window.addEventListener('right-sidebar:opened', handleSidebarOpened)
  window.addEventListener('resize', checkIfMobile)
  checkIfMobile()
})

onBeforeUnmount(() => {
  window.removeEventListener('right-sidebar:opened', handleSidebarOpened)
  window.removeEventListener('resize', checkIfMobile)
})
</script>

<style lang="scss">
.draft-sidebar {
  top: calc(var(--topbar-height) + 0.5rem) !important;
  right: 0.5rem !important;
  height: calc(100% - var(--topbar-height) - 1rem) !important;
  border-radius: 1rem !important;
  border: none !important;
  z-index: 1048 !important;
  width: 400px;

  .offcanvas-header {
    border-top-left-radius: 1rem !important;
    border-top-right-radius: 1rem !important;
  }

  .offcanvas-body {
    border-bottom-left-radius: 1rem !important;
    border-bottom-right-radius: 1rem !important;
  }
}
</style>
