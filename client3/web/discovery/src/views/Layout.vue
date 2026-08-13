<template>
  <div
    class="h-viewport overflow-hidden"
    style="display: grid; grid-template-columns: auto 1fr; grid-template-rows: 1fr; width: 100%"
  >
    <aside class="sidebar-container" :style="{ width: expanded ? '320px' : '35px', transition: 'width 0.2s' }">
      <c-sidebar
        :expanded="expanded"
        :icon="icon"
        :logo="logo"
        expand-on-click
        @update:expanded="expanded = $event"
      >
        <template #body-expanded>
          <Filters />
        </template>
      </c-sidebar>
    </aside>

    <div class="d-flex flex-column overflow-hidden" style="min-width: 0">
      <header>
        <c-topbar
          :expanded="expanded"
          :settings="settings.get('ui.topbar', {})"
          :labels="{
            appMenu: t('appMenu'),
            helpForum: t('help.forum'),
            helpDocumentation: t('help.documentation'),
            helpFeedback: t('help.feedback'),
            helpVersion: t('help.version'),
            userSettingsLoggedInAs: t('userSettings.loggedInAs', { user }),
            userSettingsProfile: t('userSettings.profile'),
            userSettingsChangePassword: t('userSettings.changePassword'),
            userSettingsLogout: t('userSettings.logout'),
            userSettingsTheme: t('userSettings.theme'),
            lightTheme: t('themes.labels.light'),
            darkTheme: t('themes.labels.dark'),
          }"
        >
          <template #title>
            {{ t("discovery") }}
          </template>
        </c-topbar>
      </header>

      <main class="d-flex flex-column flex-grow-1 overflow-hidden" style="min-width: 0">
        <Search />
      </main>
    </div>

    <c-extend-session
      v-if="isAutoLogoutEnabled"
      :timeout="settings.get('auth.autoLogout.timeout')"
      :labels="{
        extFend: t('extendSession.labels.extend'),
        warning: (countdownTime) =>
          t('extendSession.labels.warning', { countdownTime }),
      }"
    />
    <c-notification-sidebar
      v-if="!settings.get('ui.topbar', {}).hideNotifications"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'navigation' } })
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuth, useSettings } from 'corteza-lib/vue/dist'
import { components } from 'corteza-lib/vue/dist'
import Search from '../components/Search.vue'
import Filters from '../components/Filters.vue'

const { CTopbar, CSidebar, CExtendSession, CNotificationSidebar } = components

const { t } = useI18n({
  useScope: 'local',
  messages: {},
})
const { auth } = useAuth()
const { $Settings: settings } = useSettings()

const expanded = ref(false)

const user = computed(() => {
  const { user } = auth
  return user.name || user.handle || user.email || ''
})

const icon = computed(() => settings.attachment('ui.iconLogo'))

const logo = computed(() => settings.attachment('ui.mainLogo'))

const isAutoLogoutEnabled = computed(() => settings.get('auth.autoLogout.enabled'))
</script>

<style scoped>
.h-viewport {
  height: 100vh;
  height: 100dvh;
}

.sidebar-container {
  bottom: 0;
  z-index: 1037;
}

.sidebar-container > div {
  height: 100%;
}

.sidebar-container :deep(.sidebar) {
  position: relative !important;
  left: 0 !important;
  right: auto !important;
}

.sidebar-container :deep(.b-sidebar-backdrop) {
  display: none !important;
}
</style>

<style lang="scss">
.sidebar-container .sidebar {
  position: relative !important;
  left: 0 !important;
  right: auto !important;
  width: 320px;
  height: 100%;
  transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar-container .sidebar:not(.expanded) {
  width: 66px;
  height: 0;
  overflow: hidden;
}

.sidebar-container .b-sidebar-backdrop {
  display: none !important;
}
</style>
