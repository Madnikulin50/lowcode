<template>
  <div class="header-navigation d-flex flex-wrap align-items-center p-2 gap-2">
    <h2 class="title mb-0 d-flex align-items-center gap-2 flex-wrap">
      <slot name="title" />
      <div id="topbar-title" />
      <div id="topbar-title-target" />
    </h2>

    <div class="tools-wrapper ms-auto">
      <div class="d-flex align-items-center flex-wrap gap-1">
        <slot name="tools" />
        <div id="topbar-tools" />
      </div>

    </div>

    <div class="d-flex align-items-center ms-auto gap-1">
      <a
        v-if="!hideAppSelector && !settings.hideAppSelector"
        data-test-id="app-selector"
        class="btn btn-outline-light text-dark border-0 px-1"
        :href="appSelectorURL"
      >
        {{ labels.appMenu }}
      </a>

      <slot name="right-tools" />

      <button
        class="btn btn-outline-light text-decoration-none text-dark rounded-circle border-0 nav-icon d-flex align-items-center justify-content-center"
        data-test-id="theme-toggle"
        :title="currentTheme === 'dark' ? labels.lightTheme : labels.darkTheme"
        @click="saveThemeMode(currentTheme === 'dark' ? 'light' : 'dark')"
      >
        <font-awesome-icon
          class="m-0 h5"
          :icon="['fas', currentTheme === 'dark' ? 'sun' : 'moon']"
        />
        <span class="visually-hidden">
          {{ currentTheme === 'dark' ? labels.lightTheme : labels.darkTheme }}
        </span>
      </button>

      <c-notification-button
        v-if="!settings.hideNotifications"
      />

      <div
        v-if="!settings.hideHelp"
        class="dropdown nav-icon text-sm-nowrap"
      >
        <button
          class="btn btn-outline-light text-decoration-none text-dark rounded-circle border-0 w-100 dropdown-toggle no-caret"
          data-bs-toggle="dropdown"
          data-test-id="dropdown-helper"
          aria-expanded="false"
        >
          <div
            class="d-flex align-items-center justify-content-center"
          >
            <font-awesome-icon
              class="m-0 h5"
              :icon="['far', 'question-circle']"
            />
            <span class="visually-hidden">
              {{ labels.helpForum }}
            </span>
          </div>
        </button>

        <ul class="dropdown-menu topbar-dropdown-menu border-0 shadow-sm text-dark mt-2">
          <li>
            <div>
              <slot name="help-dropdown" />
            </div>
          </li>

          <li
            v-for="(helpLink, index) in helpLinks"
            :key="index"
          >
            <a
              class="dropdown-item"
              :href="checkValidURL(helpLink.url)"
              :target="helpLink.newTab ? '_blank' : ''"
            >
              {{ helpLink.handle }}
            </a>
          </li>

          <li v-if="!settings.hideForumLink">
            <a
              class="dropdown-item"
              data-test-id="dropdown-helper-forum"
              href="https://forum.cortezaproject.org/"
              target="_blank"
            >
              {{ labels.helpForum }}
            </a>
          </li>

          <li v-if="!settings.hideDocumentationLink">
            <a
              class="dropdown-item"
              data-test-id="dropdown-helper-docs"
              :href="documentationURL"
              target="_blank"
            >
              {{ labels.helpDocumentation }}
            </a>
          </li>

          <li v-if="!settings.hideFeedbackLink">
            <a
              class="dropdown-item"
              data-test-id="dropdown-helper-feedback"
              href="mailto:info@cortezaproject.org"
              target="_blank"
            >
              {{ labels.helpFeedback }}
            </a>
          </li>

          <li v-if="!onlyVersion">
            <hr class="dropdown-divider">
          </li>

          <li>
            <button
              class="dropdown-item small"
              disabled
            >
              {{ labels.helpVersion }}
              <br>
              {{ frontendVersion }}
            </button>
          </li>
        </ul>
      </div>

      <div
        v-if="!settings.hideProfile"
        class="dropdown nav-user-icon"
      >
        <button
          class="btn dropdown-toggle d-flex align-items-center no-caret nav-icon"
          :class="avatarExists ? 'btn-link p-0 rounded-circle border' : 'btn-outline-light text-dark'"
          data-bs-toggle="dropdown"
          data-test-id="dropdown-profile"
          aria-expanded="false"
        >
          <div
            v-if="avatarExists"
            class="avatar d-flex h-100 w-100"
            :style="{
              'background-image': `url(${profileAvatarUrl})`,
            }"
          />

          <div
            v-else
            class="d-flex align-items-center justify-content-center"
          >
            <font-awesome-icon
              class="m-0 h5"
              :icon="['far', 'user']"
            />
            <span class="visually-hidden">
              {{ labels.helpForum }}
            </span>
          </div>
        </button>

        <ul class="dropdown-menu topbar-dropdown-menu border-0 shadow-sm text-dark mt-2">
          <li>
            <span
              class="dropdown-item-text text-muted mb-2"
              data-test-id="dropdown-item-username"
            >
              {{ labels.userSettingsLoggedInAs }}
            </span>
          </li>

          <li>
            <div>
              <slot name="avatar-dropdown" />
            </div>
          </li>

          <li
            v-for="(profileLink, index) in profileLinks"
            :key="index"
          >
            <a
              class="dropdown-item"
              :href="checkValidURL(profileLink.url)"
              :target="profileLink.newTab ? '_blank' : ''"
            >
              {{ profileLink.handle }}
            </a>
          </li>

          <li v-if="!settings.hideProfileLink">
            <a
              class="dropdown-item"
              data-test-id="dropdown-profile-user"
              :href="userProfileURL"
              target="_blank"
            >
              {{ labels.userSettingsProfile }}
            </a>
          </li>

          <li v-if="!settings.hideChangePasswordLink">
            <a
              class="dropdown-item"
              data-test-id="dropdown-profile-change-password"
              :href="changePasswordURL"
              target="_blank"
            >
              {{ labels.userSettingsChangePassword }}
            </a>
          </li>

          <li v-if="!settings.hideThemeSelector">
            <div
              class="dropdown-item d-flex align-items-center justify-content-between"
              style="cursor: pointer;"
              @click.stop="isThemeDropdownVisible = !isThemeDropdownVisible"
            >
              <span>{{ labels.userSettingsTheme }}</span>
              <font-awesome-icon
                v-if="!isThemeDropdownVisible"
                class="text-dark"
                :icon="['fas', 'chevron-right']"
              />
              <font-awesome-icon
                v-else
                class="text-primary"
                :icon="['fas', 'chevron-left']"
              />
            </div>

            <div v-show="isThemeDropdownVisible" class="ps-2">
              <button
                v-for="theme in themes"
                :key="theme.id"
                class="dropdown-item"
                :disabled="currentTheme === theme.id"
                @click.stop="saveThemeMode(theme.id)"
              >
                {{ theme.label }}
              </button>
            </div>
          </li>

          <li><hr class="dropdown-divider"></li>

          <li>
            <button
              class="dropdown-item mt-2"
              data-test-id="dropdown-profile-logout"
              @click="$auth.logout()"
            >
              {{ labels.userSettingsLogout }}
            </button>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, getCurrentInstance } from 'vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faSun, faMoon } from '@fortawesome/free-solid-svg-icons'
import CNotificationButton from '../notifications/CNotificationButton.vue'
import { checkValidURL } from '../../filters/url'

library.add(faSun, faMoon)

declare const VERSION: string

const vm = getCurrentInstance()!
const $auth = (vm.appContext.config.globalProperties as any).$auth
const $SystemAPI = (vm.appContext.config.globalProperties as any).$SystemAPI

const props = defineProps({
  expanded: {
    type: Boolean,
    default: false,
  },
  hideAppSelector: {
    type: Boolean,
    default: false,
  },
  appSelectorURL: {
    type: String,
    default: '../',
  },
  settings: {
    type: Object,
    required: true,
  },
  labels: {
    type: Object,
    required: true,
  },
})

const currentTheme = ref('light')
const isThemeDropdownVisible = ref(false)

const userProfileURL = computed(() => {
  return $auth.cortezaAuthURL
})

const changePasswordURL = computed(() => {
  return `${$auth.cortezaAuthURL}/change-password`
})

const documentationURL = computed(() => {
  const [year, month] = VERSION.split('.')
  return `https://docs.cortezaproject.org/corteza-docs/${year}.${month}/index.html`
})

const helpLinks = computed(() => {
  const { helpLinks = [] } = props.settings || {}
  return (helpLinks || []).filter(({ handle, url }: { handle: string; url: string }) => handle && url)
})

const profileLinks = computed(() => {
  const { profileLinks = [] } = props.settings || {}
  return (profileLinks || []).filter(({ handle, url }: { handle: string; url: string }) => handle && url)
})

const onlyVersion = computed(() => {
  const {
    hideForumLink,
    hideDocumentationLink,
    hideFeedbackLink,
  } = props.settings || {}

  return !helpLinks.value.length && hideForumLink && hideDocumentationLink && hideFeedbackLink
})

const frontendVersion = computed(() => VERSION)

const profileAvatarUrl = computed(() => {
  return `${$SystemAPI.baseURL}/attachment/avatar/${$auth.user.meta.avatarID}/original/profile-photo-avatar`
})

const avatarExists = computed(() => {
  return $auth.user.meta.avatarID !== '0' && $auth.user.meta.avatarID
})

const themes = computed(() => [
  {
    id: 'light',
    label: props.labels.lightTheme,
  },
  {
    id: 'dark',
    label: props.labels.darkTheme,
  },
])

watch(() => $auth.user.meta.theme, (theme: string) => {
  currentTheme.value = theme
}, { immediate: true })

async function saveThemeMode (theme: string) {
  currentTheme.value = theme
  $auth.user.meta.theme = theme

  $SystemAPI.userUpdate($auth.user).then(() => {
    const html = document.getElementsByTagName('html')[0]
    html.setAttribute('data-color-mode', theme)
    html.setAttribute('data-bs-theme', theme)
  }).catch(console.error)
}
</script>

<style lang="scss" scoped>
$nav-icon-size: calc(var(--topbar-height) - 24px);
$nav-user-icon-size: calc(var(--topbar-height) - 16px);

.nav-icon {
  width: $nav-icon-size;
  height: $nav-icon-size;
}

.nav-user-icon {
  min-width: $nav-user-icon-size;
  min-height: $nav-user-icon-size;
}

.header-navigation {
  width: 100%;
  min-height: var(--topbar-height);
  background-color: var(--topbar-bg, transparent);
}

.avatar {
  border-radius: 50%;
  background-size: cover;
  background-repeat: no-repeat;
  background-position: center;

  &:hover {
    opacity: 0.8;
    transition: opacity .25s ease-in-out;
    -moz-transition: opacity .25s ease-in-out;
    -webkit-transition: opacity .25s ease-in-out;
  }
}

.title {
  display: flex;
  align-items: center;
  min-height: $nav-user-icon-size;
  padding-left: 20px;

  > * {
    padding: 0.25rem 0;
    display: -webkit-box;
    display: -ms-flexbox;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.tools-wrapper {
  flex-grow: 1;

  > * {
    display: flex;
    justify-content: end;
    align-items: center;
    flex-wrap: wrap;
  }
}

.dropdown-toggle.no-caret::after {
  display: none !important;
}
</style>

<style lang="scss">
.topbar-dropdown-menu {
  z-index: 1051;
}
</style>
