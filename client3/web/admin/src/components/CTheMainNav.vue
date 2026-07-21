<template>
  <div class="py-3">
    <div v-for="(grp, g) in navigation" :key="g">
      <h6 v-if="grp.header" class="mt-3 text-uppercase">{{ $t(grp.header.label) }}</h6>
      <c-sidebar-nav-items
        :items="grp.items"
        default-route-name="dashboard"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CSidebarNavItems } = components
const { t } = useI18n()
const can = inject('can', () => true)

const nav = [
  {
    items: [
      { label: 'dashboard', route: 'dashboard', icon: 'tachometer-alt' },
    ],
  },
  {
    header: { label: 'system.group' },
    items: [
      { label: 'system.items.users', route: 'system.user', icon: 'users', can: [['system/', 'users.search'], ['system/', 'user.create']] },
      { label: 'system.items.roles', route: 'system.role', icon: 'hat-cowboy', can: [['system/', 'roles.search'], ['system/', 'role.create']] },
      { label: 'system.items.usergroups', route: 'system.userGroup', icon: 'user-group', can: ['system/', 'user-groups.search'] },
      { label: 'system.items.applications', route: 'system.application', icon: 'th-large', can: [['system/', 'applications.search'], ['system/', 'application.create']] },
      { label: 'system.items.templates', route: 'system.template', icon: 'file-code', can: ['system/', 'templates.search'] },
      { label: 'system.items.settings', route: 'system.settings', icon: 'sliders-h', can: ['system/', 'settings.read'] },
      { label: 'system.items.email', route: 'system.email', icon: 'envelope-open', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
      { label: 'system.items.authclients', route: 'system.authClient', icon: 'key', can: [['system/', 'auth-clients.search'], ['system/', 'auth-client.create']] },
      { label: 'system.items.actionlog', route: 'system.actionlog', icon: 'glasses', can: ['system/', 'action-log.read'] },
      { label: 'system.items.queues', route: 'system.queue', icon: 'stream', can: [['system/', 'queues.search'], ['system/', 'queue.create']] },
      { label: 'system.items.apigw', route: 'system.apigw', icon: 'archway', can: [['system/', 'apigw-routes.search'], ['system/', 'apigw-route.create']] },
      { label: 'system.items.connections', route: 'system.connection', icon: 'cloud', can: [['system/', 'dal-connections.search'], ['system/', 'dal-connection.create']] },
      { label: 'system.items.code-snippets', route: 'system.codesnippets', icon: 'file-code', can: [['system/', 'settings.read'], ['system/', 'settings.manage']] },
      { label: 'system.items.sensitivityLevel', route: 'system.sensitivityLevel', icon: 'stamp', can: ['system/', 'dal-sensitivity-level.manage'] },
      { label: 'system.items.permissions', route: 'system.permissions', icon: 'lock', can: ['system/', 'grant'] },
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

const navigation = computed(() => {
  return nav.map(grp => {
    grp = JSON.parse(JSON.stringify(grp))
    grp.items = grp.items
      .map(itm => {
        const page = {
          name: itm.route,
          title: t(itm.label),
          icon: ['fas', itm.icon],
        }

        let allowed = true
        if (Array.isArray(itm.can)) {
          if (Array.isArray(itm.can[0])) {
            allowed = itm.can.every(c => can(c[0], c[1]))
          } else {
            allowed = can(itm.can[0], itm.can[1])
          }
        }

        return { page, can: allowed }
      })
      .filter(itm => itm && itm.can)

    if (grp.items.length === 0) return null
    return grp
  }).filter(i => i)
})
</script>