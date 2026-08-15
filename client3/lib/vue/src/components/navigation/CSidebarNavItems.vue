<template>
  <div
    class="nav-sidebar"
    :class="[
      { 'mt-1': root },
      `nav-density-${density}`,
    ]"
  >
    <div
      v-for="({page = {}, params = {}, children = []}) of items"
      :key="pageKey(page)"
      :class="{
        'mb-1': root && !page.section,
        'nav-section-block': root && page.section,
      }"
    >
      <div
        v-if="page.section"
        class="nav-section-head d-flex align-items-center"
      >
        <router-link
          active-class="nav-active"
          exact-active-class="nav-exact-active"
          :title="page.title"
          :to="{ name: page.name || defaultRouteName, params }"
          class="nav-section-title text-decoration-none flex-grow-1 text-start"
          @click="onItemClick()"
        >
          {{ page.title }}
        </router-link>
        <button
          v-if="children.length"
          class="btn btn-outline-light p-0 border-0 ms-auto nav-chevron"
          type="button"
          @click="toggle(page)"
        >
          <font-awesome-icon
            v-if="!collapses[pageKey(page)]"
            :icon="['fas', 'chevron-down']"
          />
          <font-awesome-icon
            v-else
            :icon="['fas', 'chevron-up']"
          />
        </button>
      </div>

      <div
        v-else
        class="d-flex align-items-start pointer"
        :class="density === 'compact' ? 'pb-0' : 'pb-1'"
      >
        <router-link
          active-class="nav-active"
          exact-active-class="nav-exact-active"
          :title="page.title"
          :to="{ name: page.name || defaultRouteName, params }"
          class="nav-item d-flex align-items-center text-decoration-none rounded flex-grow-1 text-start gap-1"
          :class="root ? 'nav-item-root' : 'nav-item-child'"
          @click="onItemClick()"
        >
          <template v-if="page.icon">
            <font-awesome-icon
              v-if="Array.isArray(page.icon)"
              :icon="page.icon"
              class="icon"
            />
            <img
              v-else
              :src="page.icon"
              class="icon-img"
            >
          </template>

          <label
            class="title pointer mb-0"
            :class="{ root: root }"
          >
            {{ page.title }}
          </label>
        </router-link>

        <button
          v-if="children.length"
          class="btn btn-outline-light p-0 border-0 ms-auto nav-chevron"
          type="button"
          @click="toggle(page)"
        >
          <font-awesome-icon
            v-if="!collapses[pageKey(page)]"
            class="text-dark"
            :icon="['fas', 'chevron-down']"
          />
          <font-awesome-icon
            v-else
            class="text-primary"
            :icon="['fas', 'chevron-up']"
          />
        </button>
      </div>

      <div
        v-if="children.length"
        v-show="collapses[pageKey(page)]"
      >
        <c-sidebar-nav-items
          :items="children"
          :start-expanded="startExpanded"
          :default-route-name="defaultRouteName"
          :density="density"
          :root="false"
          class="nav-children"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'CSidebarNavItems' })

import { reactive, watch } from 'vue'
import { useRoute } from 'vue-router'

const props = defineProps({
  items: {
    type: Array,
    required: true,
    default: () => [],
  },
  root: {
    type: Boolean,
    default: true,
  },
  defaultRouteName: {
    type: String,
    required: true,
  },
  startExpanded: {
    type: Boolean,
    default: false,
  },
  density: {
    type: String,
    default: 'comfortable',
  },
})

const route = (() => { try { return useRoute() } catch (e) { return {} } })() || {}
const collapses = reactive({})

watch(() => props.items, (items: any[] = []) => {
  items.forEach(({ page, params, children }: any) => {
    const px = pageKey(page)
    collapses[px] = props.startExpanded || page.expanded || showChildren({ params, children })
  })
}, { immediate: true })

function onItemClick () {
  if (window.innerWidth < 1024) {
    window.dispatchEvent(new CustomEvent('close-sidebar'))
  }
}

function pageKey (p: any): string {
  return p.pageID || p.name || p.title
}

function toggle (p: any): void {
  const px = pageKey(p)
  collapses[px] = !collapses[px]
}

function showChildren ({ params = {}, children = [] }: any): boolean {
  const partialParamsMatch = Object.entries(params).some(([key, value]) => {
    return route?.params?.[key] === value
  })

  if (partialParamsMatch) {
    return partialParamsMatch
  }

  return children.map((c: any) => showChildren(c)).some((isOpen: boolean) => isOpen)
}
</script>

<style scoped lang="scss">
.nav-sidebar {
  .nav-section-block {
    margin-top: 0.75rem;
    margin-bottom: 0.25rem;

    &:first-child {
      margin-top: 0.25rem;
    }
  }

  .nav-section-head {
    padding: 0.35rem 0.25rem 0.35rem 0.5rem;
    border-bottom: 1px solid var(--bs-border-color-translucent, rgba(0, 0, 0, 0.08));
    margin-bottom: 0.25rem;
  }

  .nav-section-title {
    font-size: 0.7rem;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--bs-secondary-color, #6c757d);
    line-height: 1.4;

    &.nav-active,
    &.nav-exact-active {
      color: var(--primary);
    }
  }

  .nav-chevron {
    min-width: 1.75rem;
    min-height: 1.75rem;
    color: var(--bs-secondary-color, #6c757d);
  }

  .nav-children {
    padding: 0.15rem 0 0.25rem 0.5rem;
    margin-left: 0.25rem;
    border-left: 1px solid var(--bs-border-color-translucent, rgba(0, 0, 0, 0.06));
  }

  .nav-item {
    position: relative;
    border-radius: 0.375rem;
    margin-bottom: 1px;
    padding: 0.35rem 0.5rem 0.35rem 0.35rem;
    transition: background-color 0.15s ease;

    &::after {
      content: '';
      position: absolute;
      right: 0;
      top: 5px;
      bottom: 5px;
      width: 3px;
      border-radius: 2px 0 0 2px;
      background: transparent;
      transition: background 0.15s ease;
    }

    .icon {
      min-width: 1.1rem;
      height: 1rem;
      width: 1rem;
      text-align: center;
      color: var(--bs-secondary-color, #6c757d);
      opacity: 0.85;
      margin-left: 0.15rem;
      transition: opacity 0.15s ease, color 0.15s ease;
    }

    .icon-img {
      height: 1rem;
      width: 1rem;
      margin-left: 0.15rem;
      object-fit: contain;
    }

    .title {
      color: var(--bs-body-color, #212529);
      font-family: var(--font-regular) !important;
      font-size: 0.875rem;
      transition: color 0.15s ease;
      text-align: left;
      margin-left: 0.35rem;
      min-width: 0;
    }

    &:hover {
      background-color: var(--bs-tertiary-bg, var(--light, #f8f9fa));

      .icon {
        opacity: 1;
      }

      &::after {
        background: rgba(var(--bs-primary-rgb, 13 110 253), 0.35);
      }
    }
  }

  .nav-item-root .title {
    font-weight: 500;
  }

  .nav-active {
    background-color: rgba(var(--bs-primary-rgb, 13 110 253), 0.08) !important;

    &::after {
      background: var(--primary) !important;
    }

    .icon {
      opacity: 1;
      color: var(--primary);
    }

    .title {
      font-family: var(--font-medium) !important;
      color: var(--primary);
    }
  }

  .nav-exact-active {
    background-color: rgba(var(--bs-primary-rgb, 13 110 253), 0.1) !important;

    .title {
      font-family: var(--font-medium) !important;
    }
  }

  &.nav-density-compact {
    .nav-section-block {
      margin-top: 0.5rem;
    }

    .nav-section-head {
      padding-top: 0.2rem;
      padding-bottom: 0.2rem;
      margin-bottom: 0.15rem;
    }

    .nav-item {
      padding-top: 0.2rem;
      padding-bottom: 0.2rem;

      .title {
        font-size: 0.8125rem;
      }
    }

    .nav-chevron {
      min-width: 1.5rem;
      min-height: 1.5rem;
    }

    .nav-children {
      padding-top: 0.1rem;
      padding-bottom: 0.15rem;
    }
  }
}
</style>
