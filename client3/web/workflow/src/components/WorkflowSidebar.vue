<template>
  <portal to="sidebar-body-expanded">
    <div class="py-3">
      <h6 class="mt-2 text-uppercase">
        <router-link
          :to="{ name: 'workflow.list', query: { limit: '100', sortBy: 'changedAt', sortDesc: 'true', query: '', deleted: '0', subWorkflow: '1', disabled: '0', page: '1' } }"
          class="text-decoration-none text-dark"
        >
          <font-awesome-icon
            :icon="['fas', 'project-diagram']"
            class="text-primary me-1"
          />
          {{ $t('list.title', 'Workflows') }}
        </router-link>
      </h6>

      <div class="px-2 pb-2">
        <c-input-search
          v-model.trim="query"
          :disabled="loading"
          :placeholder="$t('searchPlaceholder', 'Search workflows')"
          autocomplete="off"
        />
      </div>

      <div v-if="!loading">
        <c-sidebar-nav-items
          :items="filteredNavItems"
          :start-expanded="!!query"
          default-route-name="workflow.edit"
        />
        <div
          v-if="!filteredNavItems.length"
          class="d-flex justify-content-center mt-3"
        >
          {{ $t('sidebar.noResults', 'No results') }}
        </div>
      </div>
      <div
        v-else
        class="d-flex align-items-center justify-content-center mt-5"
      >
        <span class="spinner-border spinner-border-sm" />
      </div>
    </div>
  </portal>
</template>

<script setup>
import { ref, computed, inject, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { components, filter } from 'corteza-lib/vue/dist'

const { CSidebarNavItems, CInputSearch } = components
const { t } = useI18n()
const router = useRouter()
const $AutomationAPI = inject('$AutomationAPI', {})

const workflows = ref([])
const loading = ref(false)
const query = ref('')
const aliases = ref([])

const filteredNavItems = computed(() => {
  const items = workflows.value.map(w => ({
    page: {
      pageID: `workflow-${w.workflowID}`,
      name: `wf-${w.workflowID}`,
      title: w.meta?.name || w.handle || w.workflowID,
      icon: ['fas', 'project-diagram'],
      visible: true,
    },
    children: [],
    params: { workflowID: w.workflowID },
  }))

  if (!query.value) return items
  return items.filter(({ page }) => filter.Assert(page, query.value, 'title'))
})

onMounted(() => {
  loading.value = true
  $AutomationAPI.workflowList({
    disabled: 1,
    deleted: 0,
    subWorkflow: 1,
    limit: 0,
  })
    .then(({ set }) => {
      if (set) {
        workflows.value = set
        set.forEach(w => {
          const name = `wf-${w.workflowID}`
          router.addRoute('root', {
            name,
            path: ':workflowID/edit',
            component: () => import('../views/Workflow/Editor.vue'),
          })
          aliases.value.push(name)
        })
      }
    })
    .finally(() => {
      loading.value = false
    })
})

onBeforeUnmount(() => {
  aliases.value.forEach(name => router.removeRoute(name))
})
</script>

<style>
.sidebar-body .nav-sidebar .nav-item .icon {
  color: var(--black) !important;
}
.sidebar-body .nav-sidebar .nav-item .title {
  color: var(--black) !important;
}
.sidebar-body .nav-sidebar .nav-exact-active .icon,
.sidebar-body .nav-sidebar .nav-exact-active .title {
  color: var(--primary) !important;
}
</style>
