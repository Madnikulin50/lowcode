import Layout from './Layout.vue'
import Dashboard from './Dashboard.vue'
import SystemSettingsIndex from './System/Settings/Index.vue'
import SystemEmailIndex from './System/Email/Index.vue'
import SystemApigwIndex from './System/Apigw/Index.vue'
import SystemApigwEditor from './System/Apigw/Editor.vue'
import SystemApigwProfilerIndex from './System/Apigw/Profiler/Index.vue'
import SystemApigwProfilerRoute from './System/Apigw/Profiler/Route.vue'
import SystemApigwProfilerHit from './System/Apigw/Profiler/Hit.vue'
import SystemPermissionsIndex from './System/Permissions/Index.vue'
import SystemActionlogIndex from './System/Actionlog/Index.vue'
import SystemConnectionIndex from './System/Connection/Index.vue'
import SystemConnectionEditor from './System/Connection/Editor.vue'
import SystemCodeSnippetsIndex from './System/CodeSnippets/Index.vue'
import SystemAIIndex from './System/AI/Index.vue'
import ComposeSettingsIndex from './Compose/Settings/Index.vue'
import ComposePermissionsIndex from './Compose/Permissions/Index.vue'
import AutomationScriptsIndex from './Automation/Scripts/Index.vue'
import AutomationPermissionsIndex from './Automation/Permissions/Index.vue'
import FederationPermissionsIndex from './Federation/Permissions/Index.vue'
import UIThemingIndex from './UI/Theming/Index.vue'
import UINavigationIndex from './UI/Navigation/Index.vue'
import UILocationIndex from './UI/Location/Index.vue'

// combo components
import SystemUserList from './System/User/List.vue'
import SystemUserEditor from './System/User/Editor.vue'
import SystemRoleList from './System/Role/List.vue'
import SystemRoleEditor from './System/Role/Editor.vue'
import SystemApplicationList from './System/Application/List.vue'
import SystemApplicationEditor from './System/Application/Editor.vue'
import SystemTemplateList from './System/Template/List.vue'
import SystemTemplateEditor from './System/Template/Editor.vue'
import SystemAuthClientList from './System/AuthClient/List.vue'
import SystemAuthClientEditor from './System/AuthClient/Editor.vue'
import SystemUserGroupList from './System/UserGroup/List.vue'
import SystemUserGroupEditor from './System/UserGroup/Editor.vue'
import SystemSensitivityLevelList from './System/SensitivityLevel/List.vue'
import SystemSensitivityLevelEditor from './System/SensitivityLevel/Editor.vue'
import SystemQueueList from './System/Queue/List.vue'
import SystemQueueEditor from './System/Queue/Editor.vue'
import AutomationWorkflowList from './Automation/Workflow/List.vue'
import AutomationWorkflowEditor from './Automation/Workflow/Editor.vue'
import AutomationSessionList from './Automation/Session/List.vue'
import AutomationSessionEditor from './Automation/Session/Editor.vue'
import RuleChainIndex from './Automation/RuleChain/Index.vue'
import FederationNodesList from './Federation/Nodes/List.vue'
import FederationNodesEditor from './Federation/Nodes/Editor.vue'

const lookup = {
  'Dashboard': Dashboard,
  'System/Settings/Index': SystemSettingsIndex,
  'System/Email/Index': SystemEmailIndex,
  'System/Apigw/Index': SystemApigwIndex,
  'System/Apigw/Editor': SystemApigwEditor,
  'System/Apigw/Profiler/Index': SystemApigwProfilerIndex,
  'System/Apigw/Profiler/Route': SystemApigwProfilerRoute,
  'System/Apigw/Profiler/Hit': SystemApigwProfilerHit,
  'System/Permissions/Index': SystemPermissionsIndex,
  'System/Actionlog/Index': SystemActionlogIndex,
  'System/Connection/Index': SystemConnectionIndex,
  'System/Connection/Editor': SystemConnectionEditor,
  'System/CodeSnippets/Index': SystemCodeSnippetsIndex,
  'System/AI/Index': SystemAIIndex,
  'Compose/Settings/Index': ComposeSettingsIndex,
  'Compose/Permissions/Index': ComposePermissionsIndex,
  'Automation/Scripts/Index': AutomationScriptsIndex,
  'Automation/Permissions/Index': AutomationPermissionsIndex,
  'Federation/Permissions/Index': FederationPermissionsIndex,
  'UI/Theming/Index': UIThemingIndex,
  'UI/Navigation/Index': UINavigationIndex,
  'UI/Location/Index': UILocationIndex,
  'System/User/List': SystemUserList,
  'System/User/Editor': SystemUserEditor,
  'System/Role/List': SystemRoleList,
  'System/Role/Editor': SystemRoleEditor,
  'System/Application/List': SystemApplicationList,
  'System/Application/Editor': SystemApplicationEditor,
  'System/Template/List': SystemTemplateList,
  'System/Template/Editor': SystemTemplateEditor,
  'System/AuthClient/List': SystemAuthClientList,
  'System/AuthClient/Editor': SystemAuthClientEditor,
  'System/UserGroup/List': SystemUserGroupList,
  'System/UserGroup/Editor': SystemUserGroupEditor,
  'System/SensitivityLevel/List': SystemSensitivityLevelList,
  'System/SensitivityLevel/Editor': SystemSensitivityLevelEditor,
  'System/Queue/List': SystemQueueList,
  'System/Queue/Editor': SystemQueueEditor,
  'Automation/Workflow/List': AutomationWorkflowList,
  'Automation/Workflow/Editor': AutomationWorkflowEditor,
  'Automation/Session/List': AutomationSessionList,
  'Automation/Session/Editor': AutomationSessionEditor,
  'Automation/RuleChain/Index': RuleChainIndex,
  'Federation/Nodes/List': FederationNodesList,
  'Federation/Nodes/Editor': FederationNodesEditor,
}

function r (name, path, component) {
  return {
    path,
    name,
    component: lookup[component] || Dashboard,
    props: true,
  }
}

function wrap (name, path) {
  return {
    path,
    name,
    component: { name: name + 'Wrap', template: '<router-view />' },
    props: true,
  }
}

function combo (ns, name, opt = {}) {
  const cptlz = (s) => s.slice(0, 1).toUpperCase() + s.slice(1)

  opt = {
    pkey: `${name}ID`,
    plural: `${name}s`,
    cmpDir: cptlz(ns) + '/' + cptlz(name),
    ...opt,
  }

  return {
    ...wrap(`${ns}.${name}`, `/${ns}/${name}`),
    redirect: `/${ns}/${name}/list`,
    children: [
      r(`${ns}.${name}.list`, 'list', `${opt.cmpDir}/List`),
      r(`${ns}.${name}.new`, 'new', `${opt.cmpDir}/Editor`),
      r(`${ns}.${name}.edit`, `edit/:${opt.pkey}`, `${opt.cmpDir}/Editor`),
    ],
  }
}

export default [
  {
    name: 'root',
    path: '/',
    component: Layout,
    redirect: 'dashboard',
    children: [
      r('dashboard', 'dashboard', 'Dashboard'),
      {
        ...wrap('system', '/system'),

        children: [
          combo('system', 'user'),
          combo('system', 'role'),
          combo('system', 'application'),
          combo('system', 'template'),

          r('system.settings', 'settings', 'System/Settings/Index'),
          r('system.email', 'email', 'System/Email/Index'),

          combo('system', 'authClient', { pkey: 'authClientID' }),
          combo('system', 'userGroup', { pkey: 'userGroupID' }),

          r('system.apigw', 'apigw', 'System/Apigw/Index'),
          r('system.apigw.new', 'apigw/new', 'System/Apigw/Editor'),
          r('system.apigw.edit', 'apigw/edit/:routeID', 'System/Apigw/Editor'),
          r('system.apigw.profiler', 'apigw/profiler', 'System/Apigw/Profiler/Index'),
          r('system.apigw.profiler.route.list', 'apigw/profiler/route/:routeID', 'System/Apigw/Profiler/Route'),
          r('system.apigw.profiler.hit', 'apigw/profiler/hit/:hitID', 'System/Apigw/Profiler/Hit'),

          r('system.permissions', 'permissions', 'System/Permissions/Index'),
          r('system.actionlog', 'actionlog', 'System/Actionlog/Index'),

          r('system.connection', 'connection', 'System/Connection/Index'),
          r('system.connection.new', 'connection/new', 'System/Connection/Editor'),
          r('system.connection.edit', 'connection/edit/:connectionID', 'System/Connection/Editor'),

          r('system.codesnippets', 'codesnippets', 'System/CodeSnippets/Index'),
          r('system.ai', 'ai', 'System/AI/Index'),

          combo('system', 'sensitivityLevel'),

          combo('system', 'queue', { pkey: 'queueID' }),
        ],
      },

      {
        ...wrap('compose', '/compose'),
        children: [
          r('compose.settings', 'settings', 'Compose/Settings/Index'),
          r('compose.permissions', 'permissions', 'Compose/Permissions/Index'),
        ],
      },

      {
        ...wrap('automation', '/automation'),
        children: [
          combo('automation', 'workflow'),
          r('automation.ruleChain', 'rulechain', 'Automation/RuleChain/Index'),
          r('automation.scripts', 'scripts', 'Automation/Scripts/Index'),
          combo('automation', 'session'),
          r('automation.permissions', 'permissions', 'Automation/Permissions/Index'),
        ],
      },

      {
        ...wrap('federation', '/federation'),
        children: [
          combo('federation', 'nodes', { pkey: 'nodeID' }),
          r('federation.permissions', 'permissions', 'Federation/Permissions/Index'),
        ],
      },

      {
        ...wrap('ui', '/ui'),
        children: [
          r('theming.settings', 'theming', 'UI/Theming/Index'),
          r('navigation.settings', 'navigation', 'UI/Navigation/Index'),
          r('location.settings', 'location', 'UI/Location/Index'),
        ],
      },
    ],
  },

  { path: '/:pathMatch(.*)*', redirect: { name: 'dashboard' } },
]
