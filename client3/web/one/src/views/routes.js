import Layout from './Layout.vue'
import BridgeIndex from './Bridge/index.vue'
import BridgeJitsi from './Bridge/Jitsi.vue'

export default [
  { name: 'layout', path: '/', component: Layout, props: true },

  {
    ...{ name: 'bridge', path: '/bridge', component: BridgeIndex, props: true },
    children: [
      { name: 'bridge-jitsi', path: 'jitsi', component: BridgeJitsi, props: true },
    ],
  },

  { path: '/:pathMatch(.*)*', redirect: { name: 'layout' } },
]
