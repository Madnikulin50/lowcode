<template>
  <div>
    <div
      class="sidebar b-sidebar"
      :class="{ expanded: isExpanded, 'sidebar-right': right }"
      :style="{ zIndex: 1037 }"
    >
      <template v-if="isExpanded">
        <div class="d-block sidebar-header2 expanded border-bottom p-2">
          <div
            class="d-flex align-items-center justify-content-between ps-2"
            style="height: 47px;"
          >
            <img
              v-if="logo"
              data-test-id="img-main-logo"
              class="logo w-auto border-0"
              :src="logo"
            >
            <button
              class="btn btn-outline-light d-flex align-items-center justify-content-center p-2 border-0 text-secondary"
              @click="closeSidebar()"
            >
              <font-awesome-icon
                :icon="['fas', 'times']"
                class="h6 mb-0"
              />
            </button>
          </div>
          <div class="px-2">
            <slot name="header-expanded" />
          </div>
        </div>
      </template>
      <div :class="isExpanded ? 'px-3' : ''" class="sidebar-body">
        <slot v-if="isExpanded" name="body-expanded" />
      </div>

      <div :class="isExpanded ? 'px-2' : ''" class="rounded-right">
        <slot v-if="isExpanded" name="footer-expanded" />
      </div>
    </div>

    <div
      v-if="isExpanded && isMobile"
      class="b-sidebar-backdrop"
      @click="closeSidebar()"
    />

    <div class="d-flex align-items-center justify-content-center tab position-absolute p-2">
      <div
        v-if="!isExpanded && disabledRoutes.includes(route.name)"
        class="d-flex align-items-center border-0 p-2"
      >
        <img
          v-if="logo"
          class="icon w-auto border-0 me-2"
          :src="logo"
        >
        <img
          class="icon w-auto border-0"
          :src="icon"
        >
      </div>

      <button
        v-else-if="!isExpanded && expandOnClick"
        data-test-id="button-sidebar-open"
        class="btn btn-outline-light d-flex align-items-center border-0 text-primary"
        @click="openSidebar()"
      >
        <font-awesome-icon
          :icon="['fas', 'bars']"
          class="h4 mb-0"
        />
      </button>

      <router-link
        v-else-if="!isExpanded"
        data-test-id="button-home"
        class="btn btn-outline-light btn-lg d-flex align-items-center p-2 border-0 text-primary"
        :to="{ name: 'root' }"
      >
        <font-awesome-icon
          :icon="['fas', 'home']"
          class="h4 mb-0"
        />
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { routeLocationKey } from 'vue-router'
import { throttle } from 'lodash'

const props = defineProps({
  expanded: {
    type: Boolean,
    default: false,
  },
  expandOnClick: {
    type: Boolean,
    default: false,
  },
  disabledRoutes: {
    type: Array,
    default: () => [],
  },
  icon: {
    type: String,
    default: '',
  },
  logo: {
    type: String,
    default: '',
  },
  right: {
    type: Boolean,
    default: false,
  },
  storageKey: {
    type: String,
    default: 'sidebar-expanded',
  },
})

const emit = defineEmits(['update:expanded'])

const route = inject(routeLocationKey, {} as any) || {}
const isMobile = ref(false)

const isExpanded = computed({
  get: () => props.expanded,
  set: (val) => emit('update:expanded', val),
})

const checkIfMobile = throttle(() => {
  isMobile.value = window.innerWidth < 1024
}, 500)

function checkSidebar (initial = false) {
  if ((props.disabledRoutes as string[]).includes(route?.name as string)) {
    isExpanded.value = false
  } else if (!isMobile.value && initial) {
    const stored = localStorage.getItem(props.storageKey)
    isExpanded.value = stored ? stored === 'true' : true
  }
}

function openSidebar () {
  isExpanded.value = true
  localStorage.setItem(props.storageKey, 'true')
}

function closeSidebar () {
  isExpanded.value = false
  localStorage.setItem(props.storageKey, 'false')
}

watch(() => route.name, () => {
  checkSidebar()
})

watch(() => props.disabledRoutes, () => {
  checkSidebar()
})

onMounted(() => {
  checkSidebar(true)
  checkIfMobile()

  window.addEventListener('close-sidebar', closeSidebar)
  window.addEventListener('resize', checkIfMobile)
})

onBeforeUnmount(() => {
  window.removeEventListener('close-sidebar', closeSidebar)
  window.removeEventListener('resize', checkIfMobile)
})
</script>

<style lang="scss" scoped>
$header-height: 64px;

.tab {
  z-index: 1021;
  top: 0;
  height: $header-height;
  width: 66px;
}

.icon {
  max-height: 40px;
  max-width: 40px;
}

.logo {
  max-height: 40px;
}

.sidebar-header {
  height: $header-height;
}

.sidebar-body {
  flex: 1;
  overflow-y: auto;
}
</style>

<style lang="scss">
$nav-width: 320px;
$nav-width-mobile: 400px;

.b-sidebar {
  background-color: var(--white, #fff) !important;
}

.b-sidebar-backdrop {
  opacity: 0.75 !important;
}

.sidebar {
  display: flex !important;
  flex-direction: column;
  overflow: hidden;
  left: calc(-#{$nav-width}) !important;
  transition: left 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  position: fixed;
  top: 0;
  bottom: 0;
  width: $nav-width;
  background-color: var(--white, #fff);

  &.expanded {
    left: 0 !important;
    transition: left 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }
}

// Mobile sidebar should be 400px wide to match notifications and drafts
@media (max-width: 1023px) {
  .sidebar {
    width: $nav-width-mobile !important;
    left: calc(-#{$nav-width-mobile}) !important;
  }
}

.b-sidebar-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1036;
  background-color: rgba(0, 0, 0, 0.5);
}

.sidebar-spacer {
  display: none;
  min-width: calc((var(--sidebar-width, 320px)) - 60px);

  @media (min-width: 1024px) {
    &.expanded {
      display: block;
    }
  }
}
</style>
