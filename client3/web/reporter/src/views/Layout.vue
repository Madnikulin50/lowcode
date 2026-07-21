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
        :disabled-routes="disabledRoutes"
        expand-on-click
        @update:expanded="expanded = $event"
      >
        <template #body-expanded>
          <report-sidebar />
        </template>
      </c-sidebar>
    </aside>

    <div class="d-flex flex-column overflow-hidden" style="min-width: 0">
      <header>
        <c-topbar
          :expanded="expanded"
          :settings="settings?.get('ui.topbar', {})"
          :labels="{
            appMenu: t('navigation.appMenu', 'App menu'),
            helpForum: t('navigation.help.forum', 'Forum'),
            helpDocumentation: t('navigation.help.documentation', 'Documentation'),
            helpFeedback: t('navigation.help.feedback', 'Feedback'),
            helpVersion: t('navigation.help.version', 'Version'),
            userSettingsLoggedInAs: t('navigation.userSettings.loggedInAs', { user }),
            userSettingsProfile: t('navigation.userSettings.profile', 'Profile'),
            userSettingsChangePassword: t('navigation.userSettings.changePassword', 'Change password'),
            userSettingsLogout: t('navigation.userSettings.logout', 'Logout'),
            userSettingsTheme: t('navigation.userSettings.theme', 'Theme'),
            lightTheme: t('themes.labels.light', 'Light'),
            darkTheme: t('themes.labels.dark', 'Dark'),
          }"
        />
      </header>

      <main class="d-flex flex-column flex-grow-1 overflow-auto" style="min-width: 0">
        <router-view class="flex-grow-1 overflow-auto" />
      </main>
    </div>

    <c-permissions-modal
      :labels="{
        save: t('permissions.ui.save'),
        cancel: t('permissions.ui.cancel'),
        loading: t('permissions.ui.loading'),
        edit: {
          label: t('permissions.ui.edit.label'),
          description: t('permissions.ui.edit.description'),
        },
        evaluate: {
          title: t('permissions.ui.evaluate.title'),
          description: t('permissions.ui.evaluate.description'),
        },
        add: {
          label: t('permissions.ui.add.label'),
          title: t('permissions.ui.add.title'),
          save: t('permissions.ui.add.save'),
          role: {
            label: t('permissions.ui.add.role.label'),
            placeholder: t('permissions.ui.add.role.placeholder'),
          },
          user: {
            label: t('permissions.ui.add.user.label'),
            placeholder: t('permissions.ui.add.user.placeholder'),
          },
        },
      }"
    />

    <c-extend-session
      v-if="isAutoLogoutEnabled"
      :timeout="settings?.get('auth.autoLogout.timeout')"
      :labels="{
        extend: t('extendSession.labels.extend'),
        warning: (countdownTime) => t('extendSession.labels.warning', { countdownTime }),
      }"
    />

    <c-notification-sidebar v-if="!settings?.get('ui.topbar', {}).hideNotifications" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuth, useSettings, components } from 'corteza-lib/vue/dist'
import ReportSidebar from '../components/ReportSidebar.vue'

const { CPermissionsModal, CTopbar, CSidebar, CExtendSession, CNotificationSidebar } = components
const { t } = useI18n()
const { auth } = useAuth()
const { $Settings: settings } = useSettings()

const expanded = ref(false)
const disabledRoutes = ['report.list', 'report.create', 'report.edit']

const user = computed(() => {
  const u = auth.user
  return u?.name || u?.handle || u?.email || ''
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
