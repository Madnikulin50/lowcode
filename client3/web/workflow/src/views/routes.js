export default [
  {
    path: '',
    name: 'root',
    redirect: { name: 'workflow.list' },
    component: () => import('./Layout.vue'),
    children: [
      { name: 'workflow.list', path: '/list', component: () => import('./Workflow/List.vue') },
      { name: 'workflow.create', path: 'new', component: () => import('./Workflow/Editor.vue') },
      { name: 'workflow.edit', path: ':workflowID/edit', component: () => import('./Workflow/Editor.vue') },
    ],
  },

  { path: '/auth/callback', name: 'auth.callback', component: () => import('./Layout.vue') },
  { path: '/workflow/auth/callback', name: 'auth.callback.workflow', component: () => import('./Layout.vue') },
  { path: '/:pathMatch(.*)*', redirect: () => ({ name: 'root' }) },
]
