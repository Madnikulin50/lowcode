<template>
  <div class="d-flex w-100">
    <namespace-sidebar
      v-if="namespaces.length"
      :namespaces="namespaces"
    />

    <portal to="topbar-avatar-dropdown">
      <button
        class="dropdown-item"
        data-test-id="dropdown-item-reminders"
        @click="remindersVisible = true"
      >
        {{ $t('reminder.listLabel') }}
      </button>
    </portal>

    <c-reminder-sidebar
      :title="$t('reminder.listLabel')"
      :visible="remindersVisible"
      @update:visible="remindersVisible = $event"
    >
      <reminders />
    </c-reminder-sidebar>

    <router-view
      v-if="loaded"
    />
  </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useNamespaceStore } from '../../store/namespace'
import { useUserStore } from '../../store/user'
import { composables } from 'corteza-lib/vue/dist'
import NamespaceSidebar from 'corteza-webapp-compose/src/components/Namespaces/NamespaceSidebar'
import Reminders from 'corteza-webapp-compose/src/components/Namespaces/Reminders'
import { components } from 'corteza-lib/vue/dist'

const { CReminderSidebar } = components

const { t } = useI18n()
const { toastErrorHandler } = composables.useToast()
const nsStore = useNamespaceStore()
const userStore = useUserStore()
const $Settings = window.__settings

const loaded = ref(false)
const query = ref('')
const remindersVisible = ref(false)

const namespaces = computed(() => nsStore.set)

const showDrafts = computed(() => $Settings.get('ui.topbar.showDrafts', false))

userStore.load({ limit: 500 })

nsStore.load({ force: true }).finally(() => {
  loaded.value = true
}).catch(toastErrorHandler(t('notification.general.composeAccessNotAllowed')))

window.addEventListener('reminders.show', () => {
  remindersVisible.value = true
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function setDefaultValues() {
  loaded.value = false
  query.value = ''
  remindersVisible.value = false
}
</script>

<style scoped>
</style>
