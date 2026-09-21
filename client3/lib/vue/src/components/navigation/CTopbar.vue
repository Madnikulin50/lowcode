<template>
  <div class="header-navigation d-flex flex-wrap align-items-center p-2 gap-2">
    <button
      v-if="showMenuToggle"
      class="btn btn-outline-light d-flex align-items-center justify-content-center border-0 text-dark rounded-circle nav-icon menu-toggle-btn flex-shrink-0"
      data-test-id="button-topbar-menu-toggle"
      aria-label="Toggle menu"
      @click="$emit('update:expanded', !expanded)"
    >
      <font-awesome-icon
        :icon="['fas', 'bars']"
        class="h5 mb-0"
      />
    </button>

    <h2 class="title mb-0 d-flex align-items-center gap-2 flex-wrap">
      <slot name="title" />
      <div id="topbar-title" />
      <div id="topbar-title-target" />
    </h2>

    <!--
      Dedicated slot for a search/filter control. Kept out of tools-wrapper
      so its own width doesn't force the (much narrower) button group in
      #topbar-tools onto a lonely, mostly-empty line below md — see
      .topbar-search-slot below for how it claims a full row on mobile.
    -->
    <div
      id="topbar-search"
      class="topbar-search-slot"
    />

    <div class="tools-wrapper ms-auto">
      <div class="d-flex align-items-center flex-nowrap gap-1">
        <slot name="tools" />
        <div
          id="topbar-tools"
          class="d-flex align-items-center flex-nowrap gap-1"
        />
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
          class="btn btn-outline-light text-decoration-none text-dark rounded-circle border-0 w-100 h-100 dropdown-toggle no-caret"
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
            class="avatar d-flex h-100 w-100 overflow-hidden"
          >
            <img
              class="h-100 w-100"
              :src="profileAvatarUrl"
              alt=""
              @error="avatarBroken = true"
            >
          </div>

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
import { faSun, faMoon, faBars, faChevronLeft, faChevronRight } from '@fortawesome/free-solid-svg-icons'
import CNotificationButton from '../notifications/CNotificationButton.vue'
import { checkValidURL } from '../../filters/url'
import { applyColorMode } from '../../libs/theme'

library.add(faSun, faMoon, faBars, faChevronLeft, faChevronRight)

declare const VERSION: string

const vm = getCurrentInstance()!
const $auth = (vm.appContext.config.globalProperties as any).$auth
const $SystemAPI = (vm.appContext.config.globalProperties as any).$SystemAPI

defineEmits(['update:expanded'])

const props = defineProps({
  expanded: {
    type: Boolean,
    default: false,
  },
  // Renders an inline hamburger button, at every width, that toggles
  // `expanded` via v-model, for apps that have a CSidebar and pass
  // `hide-floating-toggle` to it (see CSidebar.vue) to give up its floating
  // collapsed-rail trigger in favour of this one — so the collapsed sidebar
  // never needs to reserve rail width just to avoid overlapping the page.
  // Off by default to avoid a dead button in apps without a sidebar (e.g.
  // One's launcher) or whose CSidebar still relies on its own rail.
  showMenuToggle: {
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
const avatarBroken = ref(false)

const userProfileURL = computed(() => {
  return $auth.cortezaAuthURL
})

const changePasswordURL = computed(() => {
  return `${$auth.cortezaAuthURL}/change-password`
})

const helpLinks = computed(() => {
  const { helpLinks = [] } = props.settings || {}
  return (helpLinks || []).filter(({ handle, url }: { handle: string; url: string }) => handle && url)
})

const profileLinks = computed(() => {
  const { profileLinks = [] } = props.settings || {}
  return (profileLinks || []).filter(({ handle, url }: { handle: string; url: string }) => handle && url)
})

const onlyVersion = computed(() => !helpLinks.value.length)

const frontendVersion = computed(() => VERSION)

const profileAvatarUrl = computed(() => {
  const avatarID = $auth.user?.meta?.avatarID
  if (!avatarID || avatarID === '0') {
    return ''
  }
  return `${$SystemAPI.baseURL}/attachment/avatar/${avatarID}/original/profile-photo-avatar`
})

const avatarExists = computed(() => {
  const avatarID = $auth.user?.meta?.avatarID
  return !avatarBroken.value && !!avatarID && avatarID !== '0'
})

watch(() => $auth.user?.meta?.avatarID, () => {
  avatarBroken.value = false
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

watch(() => $auth.user?.meta?.theme, (theme: string) => {
  if (!theme) return
  currentTheme.value = theme
  applyColorMode(theme)
}, { immediate: true })

async function saveThemeMode (theme: string) {
  currentTheme.value = theme
  if ($auth.user?.meta) {
    $auth.user.meta.theme = theme
  }
  applyColorMode(theme)

  $SystemAPI.userUpdate($auth.user).catch(console.error)
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

.menu-toggle-btn {
  min-width: 44px;
  min-height: 44px;
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

  img {
    object-fit: cover;
    border-radius: 50%;
  }

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
  // Refuse to shrink itself — on wide screens (flex-nowrap, see the
  // template and the media query below) this must stay wide enough for its
  // own content, or its content will be individually squeezed by the
  // browser instead. If it and its content genuinely don't fit, the root
  // header-navigation row (d-flex flex-wrap) already wraps this whole block
  // onto its own line rather than distorting anything inside it.
  flex-shrink: 0;

  > * {
    display: flex;
    justify-content: end;
    align-items: center;
    flex-wrap: wrap;
  }
}

// #topbar-tools can carry arbitrary teleported content (the Page Builder
// puts a scenario picker + layout <select>, each up to 300px, plus a
// Показать страницу/Справка/Изменить/Переводы btn-group) — none of it
// should shrink, for the same reason a select box shrunk to a sliver reads
// as broken, not "responsive".
//
// This used to be `flex-shrink: 0` (content sizes itself, refuses to
// compress) but that hit a genuine browser flex quirk: Bootstrap's own
// `.btn-group > .btn` rule makes every grouped button flex-shrink:1, and —
// even with that overridden back to 0 — this element's own auto/
// fit-content intrinsic width still came out ~200px short of what its
// btn-group actually renders at, so the group's last button silently
// overflowed past the visible edge instead of the whole thing wrapping to
// a new line. Sidestepping content-based sizing entirely fixes it: grow to
// fill whatever space `.tools-wrapper > div` actually has (flex-basis 0 +
// grow 1, ignoring the buggy intrinsic-size math), then right-align the
// content within that guaranteed-correct box.
#topbar-tools {
  flex: 1 0 0;
  min-width: 0;
  justify-content: flex-end;

  > * {
    flex-shrink: 0;
  }
}

.topbar-search-slot {
  min-width: 0;

  > * {
    width: 100%;
  }

  &:empty {
    display: none;
  }
}

// Below md, the search control and the #topbar-tools button group each get
// their own full-width row instead of sharing one with the Меню/theme/avatar
// cluster: that used to mean two `ms-auto` items competing for the same
// line's leftover space, which left the tool buttons bunched at the left
// edge with an oversized, accidental-looking gap before the chrome icons
// rather than a deliberate layout.
@media (max-width: 767.98px) {
  .topbar-search-slot {
    flex-basis: 100%;
  }

  .tools-wrapper {
    flex-basis: 100%;
    margin-left: 0 !important;

    // Wide screens keep this on one line (flex-nowrap, see the template) —
    // e.g. the Page Builder teleports a scenario picker + layout select
    // (300px each) + several buttons in here, which is exactly the content
    // that needs to wrap once it no longer fits below md, rather than
    // overflowing its now-full-width row.
    > div,
    #topbar-tools {
      flex-wrap: wrap !important;
    }
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
