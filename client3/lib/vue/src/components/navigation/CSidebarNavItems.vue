<template>
  <div
    class="nav-sidebar"
    :class="{ 'mt-1': root }"
  >
    <div
      v-for="({page = {}, params = {}, children = []}) of items"
      :key="pageKey(page)"
      :class="{ 'mb-1': root }"
    >
      <div class="d-flex align-items-start pointer pb-1">
        <router-link
          active-class="nav-active"
          exact-active-class="nav-exact-active"
          :title="page.title"
          :to="{ name: page.name || defaultRouteName, params }"
          class="nav-item d-flex align-items-center text-decoration-none rounded flex-grow-1 text-start ps-1 py-1 gap-1"
          @click="onItemClick()"
        >
          <template v-if="page.icon">
            <font-awesome-icon
              v-if="Array.isArray(page.icon)"
              :icon="page.icon"
              class="icon"
              style="height: 1rem; width: 1rem;"
            />
            <img
              v-else
              :src="page.icon"
              class="me-1"
              style="height: 1rem; width: 1rem;"
            >
          </template>

          <label
            class="title pointer mb-0"
            :class="{ 'root': root }"
          >
            {{ page.title }}
          </label>
        </router-link>

        <button
          v-if="children.length"
          class="btn btn-outline-light p-0 border-0 ms-auto"
          style="min-width: 2rem; min-height: 2rem;"
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
          :root="false"
          class="py-1 ms-2"
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
})

const route = (() => { try { return useRoute() } catch(e) { return {} } })() || {}
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
  .nav-item {
    transition: background-color 0.2s ease-out;

    .icon {
      color: var(--black);
      transition: color 0.3s ease-in-out;
    }

    .title {
      color: var(--black);
      font-family: var(--font-regular) !important;
      transition: color 0.3s ease-in-out;
      text-align: left;
    }

    &:hover {
      background-color: var(--light);
    }
  }

  .nav-active {
    .icon {
      color: var(--primary);
    }

    .title {
      font-family: var(--font-medium) !important;
      color: var(--primary);
    }
  }

}
</style>
