<template>
  <div
    :id="namespaceID"
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
        <template #header-expanded>
          <portal-target name="sidebar-header-expanded" />
        </template>

        <template #body-expanded>
          <portal-target name="sidebar-body-expanded" />
        </template>

        <template #footer-expanded>
          <portal-target name="sidebar-footer-expanded" />
        </template>
      </c-sidebar>
    </aside>

    <div class="d-flex flex-column overflow-hidden" style="min-width: 0">
    <header>
      <c-topbar
        :expanded="expanded"
        :settings="settings.get('ui.topbar', {})"
        :labels="{
          appMenu: $t('appMenu'),
          helpForum: $t('help.forum'),
          helpDocumentation: $t('help.documentation'),
          helpFeedback: $t('help.feedback'),
          helpVersion: $t('help.version'),
          userSettingsLoggedInAs: $t('userSettings.loggedInAs', { user }),
          userSettingsProfile: $t('userSettings.profile'),
          userSettingsChangePassword: $t('userSettings.changePassword'),
          userSettingsLogout: $t('userSettings.logout'),
          userSettingsTheme: $t('userSettings.theme'),
          lightTheme: $t('themes.labels.light'),
          darkTheme: $t('themes.labels.dark'),
        }"
      >
        <template #title>
        </template>

        <template #tools>
        </template>

        <template #avatar-dropdown>
          <portal-target name="topbar-avatar-dropdown" />
        </template>

        <template #right-tools>
          <c-search-button
            v-if="settings.get('discovery.enabled', false) && settings.get('ui.topbar.showSearch', false)"
            :labels="{
              search: $t('label.search'),
            }"
          />
          <c-draft-button
        v-if="settings.get('ui.topbar.showDrafts', false)"
          />
        </template>
      </c-topbar>
    </header>

    <main class="d-flex flex-column" style="flex: 1; min-height: 0; min-width: 0">
      <router-view class="flex-grow-1 overflow-auto" style="min-height: 0" />
      <div id="admin-toolbar"></div>
    </main>
  </div>
  </div>

  <c-prompts />

  <c-toaster
    :toasts="toasts"
  />

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

  <c-translation-modal />

  <c-extend-session
    v-if="isAutoLogoutEnabled"
    :timeout="settings.get('auth.autoLogout.timeout')"
    :labels="{
        extend: $t('extendSession.labels.extend'),
        warning: (countdownTime) => $t('extendSession.labels.warning', { countdownTime }),
      }"
  />

  <c-notification-sidebar v-if="!settings.get('ui.topbar.hideNotifications', false)" />

  <c-draft-sidebar v-if="settings.get('ui.topbar.showDrafts', false)" />

  <c-topbar-search
    :labels="{
        placeholder: $t('search.placeholder'),
        noResults: () => $t('search.noResults'),
        notFoundNamespace: $t('search.notFoundNamespace'),
        notFoundPage: $t('search.notFoundPage'),
        recordRedirectError: $t('search.recordRedirectError'),
        recentSearches: $t('search.recentSearches'),
        clearHistory: $t('search.clearHistory'),
        openInNewTab: $t('search.openInNewTab'),
        numberOfResults: (count) => $t('search.numberOfResults', { count }),
      }"
  />
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { debounce } from 'lodash'
import moment from 'moment'
import { useUiStore } from '../store/ui'
import { useNamespaceStore } from '../store/namespace'
import { useResourceTranslations } from '../mixins/resource-translations'
import { components } from 'corteza-lib/vue/dist'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faFile } from '@fortawesome/free-regular-svg-icons'
import CTranslationModal from '../components/Translator/CTranslatorModal.vue'
import CDraftSidebar from '../components/Drafts/CDraftSidebar.vue'
import CDraftButton from '../components/Drafts/CDraftButton.vue'

library.add(faFile)

const { CToaster, CPrompts, CPermissionsModal, CTopbar, CSidebar, CExtendSession, CNotificationSidebar, CTopbarSearch, CSearchButton } = components

const { t } = useI18n()
const route = useRoute()
const uiStore = useUiStore()
const nsStore = useNamespaceStore()

const { textDirectionality, currentLanguage } = useResourceTranslations()

const app = getCurrentInstance()?.appContext?.app
const $auth = app?.config?.globalProperties?.$auth || window.__auth
const $Settings = app?.config?.globalProperties?.$Settings || window.__settings
const $SystemAPI = app?.config?.globalProperties?.$SystemAPI || window.__systemAPI
const settings = $Settings

const expanded = ref(false)
const toasts = ref([])
const disabledRoutes = ref([
  'namespaces',
  'namespace.list',
  'namespace.edit',
  'namespace.create',
  'namespace.clone',
  'namespace.manage',
])

const user = computed(() => {
  const { user } = $auth
  return user.name || user.handle || user.email || ''
})

const icon = computed(() => settings.attachment('ui.iconLogo'))

const currentNamespace = computed(() => {
  const { slug } = route.params
  return slug ? nsStore.getByUrlPart(slug) : undefined
})

const logo = computed(() => {
  const ns = currentNamespace.value
  if (!ns && route.params.slug) {
    return ''
  }
  if (ns && ns.meta && ns.meta.logoEnabled && ns.meta.logo) {
    return ns.meta.logo
  }
  return settings.attachment('ui.mainLogo')
})

const namespaceID = computed(() => {
  const { params = {} } = route
  return params.slug
})

const bodyClass = computed(() => {
  const classes = []
  if (uiStore.namespaceSlug) {
    classes.push(`namespace-${uiStore.namespaceSlug}-body`)
  }
  if (uiStore.pageHandle) {
    classes.push(`page-${uiStore.pageHandle}-body`)
  }
  if (uiStore.layoutHandle) {
    classes.push(`page-layout-${uiStore.layoutHandle}-body`)
  }
  return classes.join(' ')
})

const isAutoLogoutEnabled = computed(() => settings.get('auth.autoLogout.enabled'))

watch(() => route.params.slug, (slug, oldSlug) => {
  if (slug !== oldSlug) {
    uiStore.setNamespaceSlug(slug)
  }
}, { immediate: true })

watch(bodyClass, debounce((cls) => {
  document.body.className = cls
}, 300), { immediate: true })

onMounted(() => {
  window.addEventListener('alert', showAlert)
  window.addEventListener('reminder.show', showReminder)
  window.addEventListener('check-namespace-sidebar', checkNamespaceSidebar)
})

onBeforeUnmount(() => {
  destroyEvents()
  setDefaultValues()
})

function checkNamespaceSidebar(showSidebar) {
  const defaultDisabledRoutes = [
    'namespaces',
    'namespace.list',
    'namespace.edit',
    'namespace.create',
    'namespace.clone',
    'namespace.manage',
  ]
  const namespaceRoutes = ['page', 'pages', 'page.record', 'page.record.create', 'page.record.edit']
  disabledRoutes.value = [...defaultDisabledRoutes, ...(showSidebar ? [] : namespaceRoutes)]
}

function addToast(message, params = {}) {
  toasts.value.push({ message, ...params })
}

function removeToast(reminderID) {
  const i = toasts.value.findIndex(r => r.reminderID === reminderID)
  if (i > -1) {
    toasts.value.splice(i, 1)
  }
}

function onReminderHide({ reminderID }) {
  $SystemAPI.reminderDismiss({ reminderID })
    .then(() => {
      removeToast(reminderID)
      window.dispatchEvent(new CustomEvent('reminder.updated', { detail: reminderID }))
    })
}

function onReminderSnooze({ reminderID }, { duration }) {
  const remindAt = moment().add(duration, 'ms').toISOString()
  $SystemAPI.reminderSnooze({ reminderID, remindAt })
    .then(() => {
      removeToast(reminderID)
      window.dispatchEvent(new CustomEvent('reminder.updated', { detail: reminderID }))
    })
}

function showAlert({ message, ...params }) {
  addToast(message, params)
}

function showReminder(r) {
  const i = toasts.value.findIndex(({ reminderID }) => reminderID === r.reminderID)
  if (i > -1 && (!r.editedAt || r.editedAt === toasts.value[i].editedAt)) {
    return
  }

  r.options = {
    variant: 'secondary',
    'no-auto-hide': true,
    solid: true,
    ...r.options,
  }

  r.actions.dismiss = {
    cb: onReminderHide,
    kind: 'Button',
    label: `<b>${t('reminder.dismiss')}</b>`,
    options: {
      variant: 'warning',
      class: ['float-end'],
    },
  }

  r.actions.snooze = {
    cb: onReminderSnooze,
    label: `<b>${t('reminder.snooze.label')}</b>`,
    kind: 'Select',
    options: {
      variant: 'outline-warning',
      class: ['float-start'],
      size: 'sm',
      items: [
        { kind: 'item-button', label: t('label.timeMinute', { t: 5 }), value: { duration: 1000 * 60 * 5 } },
        { kind: 'item-button', label: t('label.timeMinute', { t: 15 }), value: { duration: 1000 * 60 * 15 } },
        { kind: 'item-button', label: t('label.timeMinute', { t: 30 }), value: { duration: 1000 * 60 * 30 } },
        { kind: 'item-button', label: t('label.timeHour', { t: 1 }), value: { duration: 1000 * 60 * 60 * 1 } },
        { kind: 'item-button', label: t('label.timeHour', { t: 2 }), value: { duration: 1000 * 60 * 60 * 2 } },
        { kind: 'item-button', label: t('label.timeHour', { t: 24 }), value: { duration: 1000 * 60 * 60 * 24 } },
      ],
    },
  }

  if (i > -1) {
    toasts.value.splice(i, 1, r)
  } else {
    toasts.value.push(r)
  }
}

function setDefaultValues() {
  expanded.value = false
  toasts.value = []
  disabledRoutes.value = []
}

function destroyEvents() {
  window.removeEventListener('alert', showAlert)
  window.removeEventListener('reminder.show', showReminder)
  window.removeEventListener('check-namespace-sidebar', checkNamespaceSidebar)
}
</script>

<style lang="scss" scoped>
.h-viewport {
  height: 100vh;
  height: 100dvh;
}

.sidebar-container {
  bottom: 0;
  z-index: 1020;
  > div {
    height: 100%;
  }
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

.sidebar-body {
  overflow: hidden !important;
  min-height: 0 !important;
}
.sidebar-scroll::-webkit-scrollbar {
  width: 6px;
}
.sidebar-scroll::-webkit-scrollbar-thumb {
  background: transparent;
  border-radius: 3px;
}
.sidebar-scroll:hover::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.25);
}
.sidebar-scroll {
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}
.sidebar-scroll:hover {
  scrollbar-color: rgba(0, 0, 0, 0.25) transparent;
}

</style>
