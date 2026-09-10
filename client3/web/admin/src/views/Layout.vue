<template>
  <div class="h-viewport overflow-hidden" style="display: grid; grid-template-columns: auto 1fr; grid-template-rows: 1fr; width: 100%">
    <aside
      v-if="allowed"
      class="sidebar-container"
      :style="{ width: expanded ? '320px' : '0px', transition: 'width 0.2s' }"
    >
      <c-sidebar
        :expanded="expanded"
        :icon="icon"
        :logo="logo"
        expand-on-click
        hide-floating-toggle
        @update:expanded="expanded = $event"
      >
        <template #body-expanded>
          <c-the-main-nav />
        </template>
      </c-sidebar>
    </aside>

    <div class="d-flex flex-column overflow-hidden" style="min-width: 0">
      <header>
        <c-topbar
          :expanded="expanded"
          show-menu-toggle
          :settings="settings.get('ui.topbar', {})"
          :labels="{
            appMenu: $t('navigation.appMenu'),
            helpForum: $t('navigation.help.forum'),
            helpDocumentation: $t('navigation.help.documentation'),
            helpFeedback: $t('navigation.help.feedback'),
            helpVersion: $t('navigation.help.version'),
            userSettingsLoggedInAs: $t('navigation.userSettings.loggedInAs', { user }),
            userSettingsProfile: $t('navigation.userSettings.profile'),
            userSettingsChangePassword: $t('navigation.userSettings.changePassword'),
            userSettingsLogout: $t('navigation.userSettings.logout'),
            userSettingsTheme: $t('navigation.userSettings.theme'),
            lightTheme: $t('themes.labels.light'),
            darkTheme: $t('themes.labels.dark'),
          }"
          @update:expanded="expanded = $event"
        />
      </header>

      <main
        v-if="allowed"
        class="d-flex flex-column flex-grow-1 overflow-auto"
        style="min-width: 0"
      >
        <router-view />
      </main>
    </div>

    <c-prompts />
    <c-permissions-modal
      :labels="{
        save: $t('permissions.ui.save'),
        cancel: $t('permissions.ui.cancel'),
        loading: $t('permissions.ui.loading'),
        edit: {
          label: $t('permissions.ui.edit.label'),
          description: $t('permissions.ui.edit.description'),
        },
        evaluate: {
          title: $t('permissions.ui.evaluate.title'),
          description: $t('permissions.ui.evaluate.description'),
        },
        add: {
          label: $t('permissions.ui.add.label'),
          title: $t('permissions.ui.add.title'),
          save: $t('permissions.ui.add.save'),
          role: {
            label: $t('permissions.ui.add.role.label'),
            placeholder: $t('permissions.ui.add.role.placeholder'),
          },
          user: {
            label: $t('permissions.ui.add.user.label'),
            placeholder: $t('permissions.ui.add.user.placeholder'),
          },
        },
      }"
    />
    <c-extend-session
      v-if="isAutoLogoutEnabled"
      :timeout="settings.get('auth.autoLogout.timeout')"
      :labels="{
        extend: $t('extendSession.labels.extend'),
        warning: (countdownTime) => $t('extendSession.labels.warning', { countdownTime }),
      }"
    />
    <c-notification-sidebar v-if="!settings.get('ui.topbar', {}).hideNotifications" />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'admin' } })
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CExtendSession, CPermissionsModal, CPrompts, CTopbar, CSidebar, CNotificationSidebar } = components
const { t } = useI18n()
const $auth = window.__auth
const settings = {
  get: (key, def) => window.__settings?.get?.(key, def) ?? def,
  attachment: (key) => window.__settings?.attachment?.(key) ?? '',
}

const expanded = ref(false)
const allowed = ref(true)

const user = computed(() => {
  const u = $auth.user
  return u.name || u.handle || u.email || ''
})

const icon = computed(() => settings.attachment('ui.iconLogo'))
const logo = computed(() => settings.attachment('ui.mainLogo'))
const isAutoLogoutEnabled = computed(() => settings.get('auth.autoLogout.enabled'))

function can(resource, operation) {
  return true
}
</script>

<style scoped>
.h-viewport {
  height: 100vh;
  height: 100dvh;
}

/*!rtl:ignore*/
/* Push-layout (sidebar reserves grid-column space, content shifts over) only
   applies at md and up. Below that, CSidebar's own fixed-position overlay
   drawer + backdrop take over instead — see the max-width block further down. */
@media (min-width: 768px) {
  .sidebar-container :deep(.sidebar) {
    position: relative !important;
    left: 0 !important;
    right: auto !important;
  }

  .sidebar-container :deep(.b-sidebar-backdrop) {
    display: none !important;
  }
}
/*!rtl:end:ignore*/
</style>

<style>
.sidebar-container {
  height: 100%;
}

.sidebar-container > div {
  height: 100%;
}

@media (min-width: 768px) {
  .sidebar-container .sidebar {
    position: relative !important;
    left: 0 !important;
    right: auto !important;
    width: 320px;
    height: 100%;
  }

  .sidebar-container .sidebar-body {
    position: absolute;
    top: 64px;
    bottom: 0;
    left: 0;
    right: 0;
    overflow-y: auto !important;
  }

  .sidebar-container .sidebar:not(.expanded) {
    /* The collapsed rail used to reserve 66px here for CSidebar's floating
       toggle (see hide-floating-toggle above) so it wouldn't overlap the
       page. That toggle now lives inline in CTopbar instead, so collapsed
       reserves nothing at any width, matching the <aside> inline style. */
    width: 0;
    overflow: hidden;
    height: 0;
  }

  .sidebar-container .sidebar:not(.expanded) .sidebar-body {
    display: none;
  }

  .sidebar-container .b-sidebar-backdrop {
    display: none !important;
  }
}

/* Below md the *expanded* sidebar drawer is an overlay, not a grid column
   (collapsed is already 0 above at every width) — so the page doesn't need
   to make room for it here either (overrides the inline `width` style set
   on <aside> above — a stylesheet !important beats an inline style without
   one). */
@media (max-width: 767.98px) {
  .sidebar-container {
    width: 0 !important;
  }
}

#resource-list td.actions {
  padding: 0.5rem !important;
  vertical-align: middle;
}
</style>
