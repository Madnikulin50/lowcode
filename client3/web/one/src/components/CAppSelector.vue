<template>
  <div class="app-selector d-flex flex-column h-100 py-2">
    <div class="d-flex flex-column justify-content-center align-items-center mx-4 my-2">
      <img :src="logo" class="logo px-2" alt="logo">

      <div class="search w-100 mx-auto my-4">
        <c-input-search
          v-model.trim="query"
          :aria-label="t('search')"
          :placeholder="t('search')"
          :debounce="200"
        />
      </div>
    </div>

    <div class="flex-fill overflow-auto">
      <div class="container-fluid h-100">
        <draggable
            item-key="id"
          v-if="filteredApps.length"
          v-model="appList"
          :disabled="!canCreateApplication || query || isMobileResolution"
          group="apps"
          class="h-100 w-100"
          @end="onDrop"
        >
          <transition-group
            name="apps"
            tag="div"
            class="d-flex flex-wrap align-items-stretch justify-content-center mx-2 row"
          >
            <div
              v-for="app in filteredApps"
              :key="app.applicationID"
              class="col-12 col-md-6 col-lg-4 col-xl-3 p-2"
            >
              <div
                class="card app h-100 position-relative"
                @mouseover="hovered = app.applicationID"
                @mouseleave="hovered = undefined"
              >
                <div class="card-body d-flex flex-column align-items-center justify-content-center text-center">
                  <img
                    class="rounded-bottom thumbnail mb-3"
                    :src="logoUrl(app)"
                    :alt="app.unify.name || app.name"
                  >
                  <h6 class="my-4">{{ app.unify.name || app.name }}</h6>
                </div>

                <a
                  :data-test-id="app.name"
                  :href="app.enabled ? app.unify.url : undefined"
                  :target="openAppInNewTab(app.unify.url)"
                  :class="{ disabled: !app.enabled }"
                  :style="{ cursor: app.enabled ? 'pointer' : canCreateApplication ? 'grab' : 'default' }"
                  class="stretched-link"
                />
              </div>
            </div>
          </transition-group>
        </draggable>

        <div
          v-else
          class="d-flex justify-content-center align-items-center mt-5 w-100"
        >
          <h4 data-test-id="heading-no-apps">
            {{ query ? t('no-applications-found') : t('no-applications') }}
          </h4>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'layout' } })
import { ref, computed, watch, onMounted } from 'vue'

import { useAuth, useNsI18n } from 'corteza-lib/vue/dist'
import { url, components } from 'corteza-lib/vue/dist'
import draggable from 'vuedraggable'
import { useApplicationsStore } from '../store'

const { CInputSearch } = components

const t = useNsI18n()
const { auth } = useAuth()
const applicationsStore = useApplicationsStore()

const props = defineProps({
  logo: {
    type: String,
    default: '',
  },
})

const query = ref('')
const appList = ref([])
const canCreateApplication = ref(false)
const canPin = ref(false)
const hovered = ref(undefined)
const isMobileResolution = ref(false)

const filteredApps = computed(() => {
  const q = (query.value || '').toUpperCase()
  return query.value
    ? appList.value.filter(({ name }) => (name.toUpperCase()).includes(q))
    : appList.value
})

watch(() => applicationsStore.unifyOnly, (apps) => {
  appList.value = apps
}, { immediate: true })

onMounted(() => {
  fetchEffective()
  if (window.innerWidth < 576) {
    isMobileResolution.value = true
  }
})

function fetchEffective() {
  window.__systemAPI.permissionsEffective({ resource: 'application' })
    .then(p => {
      canCreateApplication.value = p.find(per => per.operation === 'application.create').allow || false
    })
}

function handlePin(pin = true, applicationID) {
  if (pin) {
    applicationsStore.unpin({ applicationID, ownedBy: auth.user.userID })
  } else {
    applicationsStore.pin({ applicationID, ownedBy: auth.user.userID })
  }
}

async function onDrop() {
  const ids = appList.value.map(({ applicationID }) => applicationID)
  await applicationsStore.reorder(ids)
}

function logoUrl(app) {
  if (!app.unify.logo) {
    return 'applications/default-app.png'
  }

  const apiSystem = '/api/system'
  const apiBaseUrl = (new URL(url.Make({ url: window.__systemAPI.baseURL }))).toString()

  if (app.unify.logo.startsWith(apiSystem)) {
    return apiBaseUrl.substring(0, apiBaseUrl.length - apiSystem.length) + app.unify.logo
  }

  return app.unify.logo
}

function openAppInNewTab(route) {
  return !route.includes('jitsi') ? '' : '_blank'
}
</script>

<style lang="scss" scoped>
.app-selector {
  .logo {
    max-height: 20vh;
    max-width: 500px;
    width: auto;
  }

  @media only screen and (max-width: 576px) {
    .logo {
      max-width: 100%;
    }
  }

  .search {
    max-width: 600px;
  }

  .app {
    min-height: 13rem;
    transition: all 0.2s ease;
    box-shadow: 0;
    top: 0;

    .thumbnail {
      max-width: 100%;
      max-height: 150px;
      object-fit: contain;
    }

    &:hover {
      transition: all 0.2s ease;
      box-shadow: 0px 4px 8px rgba(38, 38, 38, 0.2);
      top: -2px;
    }
  }

  .star {
    position: absolute;
    top: .2rem;
    left: .2rem;
    padding: 0;
    margin: 0;
    background-color: transparent;
    border: none;
    .star-icon {
      fill: var(--warning);
      width: 1.2rem;
      height: 1.2rem;
    }
  }

  .apps-leave-active {
    position: absolute;
    transition: opacity 0.25s ease;
  }
  .apps-enter, .apps-leave-to {
    opacity: 0;
  }

  .apps-move {
    transition: transform 0.25s ease;
  }
}
</style>
