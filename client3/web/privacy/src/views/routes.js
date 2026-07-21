export default [
  {
    path: '',
    name: 'root',
    redirect: 'dashboard',
    component: () => import('./Layout.vue'),
    children: [
      { name: 'dashboard', path: '/dashboard', component: () => import('./Privacy/Dashboard.vue') },
      { name: 'sensitive-data', path: '/sensitive-data', component: () => import('./Privacy/SensitiveData.vue') },
      { name: 'data-overview', path: '/data-overview', component: () => import('./Privacy/DataOverview/index.vue') },
      { name: 'data-overview.application', path: '/data-overview/application', component: () => import('./Privacy/DataOverview/Application.vue') },
      { name: 'request.list', path: '/request/list', component: () => import('./Privacy/Request/List.vue') },
      { name: 'request.view', path: '/request/:requestID?', component: () => import('./Privacy/Request/View.vue'), props: true },
      { name: 'request.create', path: '/request/:kind/new', component: () => import('./Privacy/Request/Create.vue'), props: true },
    ],
  },

  { path: '/:pathMatch(.*)*', redirect: { name: 'root' } },
]
