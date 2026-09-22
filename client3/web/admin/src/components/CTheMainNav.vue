<template>
  <div class="admin-nav px-1 pb-3">
    <div class="admin-nav-search sticky-top py-2">
      <c-input-search
        v-model.trim="query"
        :placeholder="searchPlaceholder"
        :autocomplete="'off'"
        size="sm"
      />
    </div>

    <c-sidebar-nav-items
      :items="filteredNav"
      :start-expanded="true"
      default-route-name="dashboard"
    />

    <div
      v-if="!filteredNav.length"
      class="admin-nav-empty text-muted text-center mt-4 px-2"
    >
      {{ noResultsLabel }}
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'navigation' } })
import { computed, inject, ref } from 'vue'
import { components, useNsI18n } from 'corteza-lib/vue/dist'

const { CSidebarNavItems, CInputSearch } = components
const t = useNsI18n()
const can = inject('can', () => true)
const query = ref('')

function tr (key, fallback) {
  const v = t(key)
  if (!v || v === key || v.endsWith(`.${key}`)) return fallback || key
  return v
}

const searchPlaceholder = computed(() => tr('search', 'Search'))
const noResultsLabel = computed(() => tr('noResults', 'No matching pages'))

const nav = [
  {
    items: [
      { label: 'dashboard', route: 'dashboard', icon: 'tachometer-alt' },
    ],
  },
  {
    header: { label: 'groups.userManagement', fallback: 'Управление пользователями' },
    items: [
      { label: 'system.items.users', route: 'system.user', icon: 'users', can: [['system/', 'users.search'], ['system/', 'user.create']] },
      { label: 'system.items.roles', route: 'system.role', icon: 'hat-cowboy', can: [['system/', 'roles.search'], ['system/', 'role.create']] },
      { label: 'system.items.usergroups', route: 'system.userGroup', icon: 'user-group', can: ['system/', 'user-groups.search'] },
    ],
  },
  {
    header: { label: 'groups.access', fallback: 'Access' },
    items: [
      { label: 'system.items.authclients', route: 'system.authClient', icon: 'key', can: [['system/', 'auth-clients.search'], ['system/', 'auth-client.create']] },
      { label: 'system.items.permissions', route: 'system.permissions', icon: 'lock', can: ['system/', 'grant'] },
    ],
  },
  {
    header: { label: 'groups.platform', fallback: 'Platform' },
    items: [
      { label: 'system.items.settings', route: 'system.settings', icon: 'sliders-h', can: ['system/', 'settings.read'] },
      { label: 'system.items.email', route: 'system.email', icon: 'envelope-open', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
      { label: 'system.items.templates', route: 'system.template', icon: 'file-code', can: ['system/', 'templates.search'] },
      { label: 'system.items.code-snippets', route: 'system.codesnippets', icon: 'code', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
      { label: 'system.items.ai', route: 'system.ai', icon: 'microchip', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
      { label: 'system.items.sensitivityLevel', route: 'system.sensitivityLevel', icon: 'stamp', can: ['system/', 'dal-sensitivity-level.manage'] },
      { label: 'system.items.actionlog', route: 'system.actionlog', icon: 'glasses', can: ['system/', 'action-log.read'] },
    ],
  },
  {
    header: { label: 'groups.integrations', fallback: 'Integrations' },
    items: [
      { label: 'system.items.applications', route: 'system.application', icon: 'th-large', can: [['system/', 'applications.search'], ['system/', 'application.create']] },
      { label: 'system.items.queues', route: 'system.queue', icon: 'stream', can: [['system/', 'queues.search'], ['system/', 'queue.create']] },
      { label: 'system.items.apigw', route: 'system.apigw', icon: 'archway', can: [['system/', 'apigw-routes.search'], ['system/', 'apigw-route.create']] },
      { label: 'system.items.connections', route: 'system.connection', icon: 'cloud', can: [['system/', 'dal-connections.search'], ['system/', 'dal-connection.create']] },
    ],
  },
  {
    header: { label: 'compose.group' },
    items: [
      { label: 'compose.items.settings', route: 'compose.settings', icon: 'sliders-h', can: [['compose/', 'settings.read'], ['compose/', 'settings.manage']] },
      { label: 'compose.items.permissions', route: 'compose.permissions', icon: 'lock', can: ['compose/', 'grant'] },
    ],
  },
  {
    header: { label: 'automation.group' },
    items: [
      { label: 'automation.items.workflows', route: 'automation.workflow', icon: 'project-diagram', can: [['automation/', 'workflows.search'], ['automation/', 'workflow.create']] },
      { label: 'automation.items.ruleChains', route: 'automation.ruleChain', icon: 'link', can: [['automation/', 'workflows.search']] },
      { label: 'automation.items.sessions', route: 'automation.session', icon: 'business-time', can: ['automation/', 'sessions.search'] },
      { label: 'automation.items.scripts', route: 'automation.scripts', icon: 'scroll', can: ['automation/', 'workflows.search'] },
      { label: 'automation.items.permissions', route: 'automation.permissions', icon: 'lock', can: ['automation/', 'grant'] },
    ],
  },
  {
    header: { label: 'federation.group' },
    items: [
      { label: 'federation.items.nodes', route: 'federation.nodes', icon: 'share-alt', can: ['federation/', 'pair'] },
      { label: 'federation.items.permissions', route: 'federation.permissions', icon: 'lock', can: ['federation/', 'grant'] },
    ],
  },
  {
    header: { label: 'ui.group' },
    items: [
      { label: 'ui.items.theming', route: 'theming.settings', icon: 'palette', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
      { label: 'ui.items.navigation', route: 'navigation.settings', icon: 'bars', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
      { label: 'ui.items.location', route: 'location.settings', icon: 'map-marker-alt', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
    ],
  },
]

function allowed (itm) {
  if (!Array.isArray(itm.can)) return true
  if (Array.isArray(itm.can[0])) {
    return itm.can.every(c => can(c[0], c[1]))
  }
  return can(itm.can[0], itm.can[1])
}

function toPage (itm) {
  return {
    page: {
      name: itm.route,
      title: tr(itm.label, itm.label),
      icon: ['fas', itm.icon],
    },
  }
}

const navigation = computed(() => {
  const items = []

  nav.forEach(grp => {
    const children = grp.items.filter(allowed).map(toPage)
    if (!children.length) return

    if (grp.header) {
      items.push({
        page: { title: tr(grp.header.label, grp.header.fallback || grp.header.label), section: true },
        children,
      })
    } else {
      items.push(...children)
    }
  })

  return items
})

function matches (item, q) {
  return (item.page?.title || '').toLowerCase().includes(q)
}

const filteredNav = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return navigation.value

  const out = []
  navigation.value.forEach(item => {
    if (item.page?.section) {
      const kids = (item.children || []).filter(child => matches(child, q))
      if (matches(item, q)) {
        out.push(item)
      } else if (kids.length) {
        out.push({ ...item, children: kids })
      }
    } else if (matches(item, q)) {
      out.push(item)
    }
  })
  return out
})
</script>

<style scoped>
.admin-nav-search {
  z-index: 2;
  background: var(--sidebar-bg, var(--white, #fff));
}

.admin-nav-empty {
  font-size: 0.875rem;
}
</style>
