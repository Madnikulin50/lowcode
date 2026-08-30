export default [
  {
    name: 'root',
    path: '',
    component: () => import('./Layout.vue'),
    redirect: 'namespaces',
    children: [
      {
        name: 'namespaces',
        path: '/',
        component: () => import('./Namespace/Index.vue'),
        redirect: 'namespaces.list',
        children: [
          { name: 'namespace.list', path: '/namespaces', component: () => import('./Namespace/List.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.namespaces.list' } },
          { name: 'namespace.manage', path: '/namespaces/manage', component: () => import('./Namespace/Manage.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.namespaces.list' } },
          { name: 'namespace.create', path: '/admin/namespace/create', component: () => import('./Namespace/Edit.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.namespace.edit' } },
          { name: 'namespace.edit', path: '/admin/namespace/edit/:namespaceID', component: () => import('./Namespace/Edit.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.namespace.edit' } },
          {
            name: 'namespace',
            path: '/ns/:slug',
            component: () => import('./Namespace/View.vue'),
            props: r => ({ ...r.params }),
            redirect: { name: 'pages' },

            children: [
              {
                name: 'pages',
                path: 'pages',
                component: () => import('./Public/Index.vue'),
                props: r => ({ ...r.params }),
                children: [
                  {
                    name: 'page',
                    path: ':pageID?',
                    component: () => import('./Public/Pages/View.vue'),
                    props: r => ({ ...r.params }),

                    children: [
                      { name: 'page.record', path: 'record/:recordID', component: () => import('./Public/Pages/Records/View.vue'), props: r => ({ edit: false, ...r.params }) },
                      { name: 'page.record.edit', path: 'record/:recordID/edit', component: () => import('./Public/Pages/Records/View.vue'), props: r => ({ edit: true, ...r.params }) },
                      { name: 'page.record.create', path: 'record', component: () => import('./Public/Pages/Records/View.vue'), props: r => ({ edit: true, ...r.params }) },
                    ],
                  },
                ],
              },
              {
                name: 'admin',
                path: 'admin',
                component: () => import('./Admin/Index.vue'),
                props: r => ({ ...r.params }),
                redirect: { name: 'admin.modules' },

                children: [
                  { name: 'admin.modules', path: 'modules', component: () => import('./Admin/Modules/List.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.modules.create', path: 'modules/new', component: () => import('./Admin/Modules/Edit.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.modules.edit', path: 'modules/:moduleID/edit', component: () => import('./Admin/Modules/Edit.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.modules.record.list', path: 'modules/:moduleID/record/list', component: () => import('./Admin/Modules/Records/List.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.modules.record.view', path: 'modules/:moduleID/record/:recordID', component: () => import('./Admin/Modules/Records/View.vue'), props: r => ({ edit: false, ...r.params }) },
                  { name: 'admin.modules.record.create', path: 'modules/:moduleID/record', component: () => import('./Admin/Modules/Records/View.vue'), props: r => ({ edit: true, ...r.params }) },
                  { name: 'admin.modules.record.edit', path: 'modules/:moduleID/record/:recordID/edit', component: () => import('./Admin/Modules/Records/View.vue'), props: r => ({ edit: true, ...r.params }) },

                  { name: 'admin.pages', path: 'pages', component: () => import('./Admin/Pages/List.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.pages.list' } },
                  { name: 'admin.pages.edit', path: 'pages/:pageID/edit', component: () => import('./Admin/Pages/Edit.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.page.edit' } },
                  { name: 'admin.pages.builder', path: 'pages/:pageID/builder', component: () => import('./Admin/Pages/Builder.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.page.builder' } },
                  { name: 'admin.pages.rag', path: 'pages/rag', component: () => import('./Admin/Pages/RAG.vue'), props: r => ({ ...r.params }) },

                  { name: 'admin.charts', path: 'charts', component: () => import('./Admin/Charts/List.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.charts.list' } },
                  { name: 'admin.charts.create', path: 'charts/new/:category?', component: () => import('./Admin/Charts/Edit.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.chart.edit' } },
                  { name: 'admin.charts.edit', path: 'charts/:chartID/edit', component: () => import('./Admin/Charts/Edit.vue'), props: r => ({ ...r.params }), meta: { helpTopic: 'compose.chart.edit' } },

                  { name: 'admin.etl', path: 'etl', component: () => import('./Admin/ETL/List.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.etl.create', path: 'etl/new', component: () => import('./Admin/ETL/Edit.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.etl.edit', path: 'etl/:etlID/edit', component: () => import('./Admin/ETL/Edit.vue'), props: r => ({ ...r.params }) },

                  { name: 'admin.rulechains', path: 'rulechains', component: () => import('./Admin/RuleChains/List.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.rulechains.create', path: 'rulechains/new', component: () => import('./Admin/RuleChains/Edit.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.rulechains.edit', path: 'rulechains/:chainID/edit', component: () => import('./Admin/RuleChains/Edit.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.workflows', path: 'workflows', component: () => import('./Admin/Workflows/List.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.workflows.create', path: 'workflows/new', component: () => import('./Admin/Workflows/Edit.vue'), props: r => ({ ...r.params }) },
                  { name: 'admin.workflows.edit', path: 'workflows/:workflowID/edit', component: () => import('./Admin/Workflows/Edit.vue'), props: r => ({ ...r.params }) },

                  { name: 'admin.configuration', path: 'configuration', component: () => import('./Admin/Configuration/Index.vue'), props: r => ({ ...r.params }) },
                ],
              },

              { path: ':pathMatch(.*)*', redirect: { name: 'pages' } },
            ],
          },
        ],
      },
    ],
  },

  { path: '/:pathMatch(.*)*', redirect: { name: 'root' } },
]
