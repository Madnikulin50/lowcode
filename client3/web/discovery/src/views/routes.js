export default [
  {
    name: 'root',
    path: '',
    component: () => import('./Layout.vue'),
  },

  // When everything else fails, go to root
  { path: '/auth/callback', name: 'auth.callback', component: () => import('./Layout.vue') },
  { path: '/discovery/auth/callback', name: 'auth.callback.discovery', component: () => import('./Layout.vue') },
  { path: '/:pathMatch(.*)*', redirect: () => ({ name: 'root' }) },
]
