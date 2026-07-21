<template>
  <div class="d-flex flex-column w-100 h-viewport overflow-hidden">
    <header v-show="loaded">
      <c-topbar
        hide-app-selector
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
      />
    </header>

    <main v-show="loaded" class="flex-fill overflow-hidden">
      <c-app-selector :logo="logo" />
    </main>

    <c-loader-logo v-if="!loaded" :logo="logo" />

    <c-prompts />

    <c-extend-session
      v-if="isAutoLogoutEnabled"
      :timeout="settings.get('auth.autoLogout.timeout')"
      :labels="{
        extend: t('extendSession.labels.extend'),
        warning: (countdownTime) => t('extendSession.labels.warning', { countdownTime }),
      }"
    />

    <c-notification-sidebar v-if="!settings.get('ui.topbar', {}).hideNotifications" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuth, useSettings } from 'corteza-lib/vue/dist'
import { components } from 'corteza-lib/vue/dist'
import CAppSelector from '../components/CAppSelector.vue'
import { useApplicationsStore } from '../store'

const { CTopbar, CLoaderLogo, CPrompts, CExtendSession, CNotificationSidebar } = components

const { t } = useI18n()
const { auth } = useAuth()
const { $Settings: settings } = useSettings()
const applicationsStore = useApplicationsStore()

const loaded = ref(false)

const icon = computed(() => settings.attachment('ui.iconLogo'))
const logo = computed(() => settings.attachment('ui.mainLogo'))
const user = computed(() => {
  const u = auth.user
  return u?.name || u?.handle || u?.email || ''
})
const isAutoLogoutEnabled = computed(() => settings.get('auth.autoLogout.enabled'))

onMounted(() => {
  applicationsStore.load().then(() => {
    setTimeout(() => {
      loaded.value = true
    }, 2000)
  })
})
</script>

<style scoped>
.h-viewport {
  height: 100vh;
  height: 100dvh;
}
</style>
