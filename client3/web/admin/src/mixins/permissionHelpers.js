import { useUiStore } from '../store/ui'

const systemRoles = ['1', '2']
const lsKey = 'permissionList.roles'

function getIncludedRoles() {
  return JSON.parse(localStorage.getItem(lsKey) || '[]')
}

function setIncludedRoles(roles = []) {
  roles = roles.filter(r => !systemRoles.includes(r))
  localStorage.setItem(lsKey, JSON.stringify(roles))
}

export function usePermissionHelpers(api) {
  const ui = useUiStore()

  const roles = []
  const allRoles = []
  const rolePermissions = []
  const permissions = {}
  const effective = {}
  const resources = []
  const loaded = { roles: false, permissions: false }
  const permission = { processing: false, success: false }

  function incLoader() { ui.incLoader() }
  function decLoader() { ui.decLoader() }

  function prepareRoles() {
    incLoader()
    rolePermissions.length = 0

    return Promise.all(getIncludedRoles().map(({ mode, name, roleID, userID }) => {
      if (mode === 'edit') {
        return readPermissions({ name, roleID })
      } else {
        return evaluatePermissions({ name, roleID, userID })
      }
    })).finally(() => {
      loaded.roles = true
      decLoader()
    })
  }

  function fetchPermissions(apiInstance = api) {
    incLoader()

    if (!apiInstance || !apiInstance.permissionsList) {
      decLoader()
      return Promise.resolve()
    }

    apiInstance.permissionsList()
      .then(permissionsList => {
        const resSet = new Set()
        Object.assign(permissions, (permissionsList || []).reduce((map, { type, any, op }) => {
          resSet.add(any)
          if (!map[type]) {
            map[type] = { any, ops: [] }
          }
          map[type].ops.push(op)
          return map
        }, {}))
        resources.length = 0
        resources.push(...resSet)
      })
      .then(() => prepareRoles())
      .finally(() => {
        loaded.permissions = true
        decLoader()
      })
  }

  function onSubmit(roleRules, apiInstance = api) {
    permission.processing = true
    Promise.all(roleRules.filter(({ ID }) => ID.includes('edit')).map(({ ID, rules }) => {
      const roleID = ID.split('-')[1]
      const externalRules = []
      Object.entries(rules).forEach(([key, value]) => {
        const [operation, resource] = key.split('@', 2)
        externalRules.push({ roleID, resource, operation, access: value })
      })
      return apiInstance.permissionsUpdate({ roleID, rules: externalRules })
    })).then(() => {
      animateSuccess(permission)
    }).finally(() => {
      Promise.all(roles.filter(({ mode }) => mode === 'eval').map(({ roleID, userID }) => {
        return apiInstance.permissionsTrace({ roleID, userID }).then(rr => {
          const ID = userID ? `eval-${userID}` : `eval-${roleID.join('-')}`
          const idx = rolePermissions.findIndex(rp => rp.ID !== ID)
          if (idx >= 0) rolePermissions.splice(idx, 1)
          rolePermissions.push({ resource: '', ID, rules: roleRules(rr, 'eval') })
        })
      })).finally(() => {
        permission.processing = false
      })
    })
  }

  function addRole(add) {
    loaded.roles = false
    const { mode } = add || {}

    if (mode === 'edit') {
      const { roleID, name } = add.roleID || {}
      const ID = `edit-${roleID}`
      if (roles.some(r => r.ID === ID)) {
        loaded.roles = true
        return
      }
      readPermissions({ roleID, name: [name] }).finally(() => {
        setIncludedRoles(roles)
        loaded.roles = true
      })
    } else if (mode === 'eval') {
      let { userID, roleID } = add
      let name = ''
      if (userID) {
        name = [userID.name]
        userID = userID.userID
      } else {
        name = roleID.map(({ name }) => name)
        roleID = roleID.map(({ roleID }) => roleID)
      }
      const ID = userID ? `eval-${userID}` : `eval-${roleID.join('-')}`
      if (roles.some(r => r.ID === ID)) {
        loaded.roles = true
        return
      }
      evaluatePermissions({ name, roleID, userID }).finally(() => {
        setIncludedRoles(roles)
        loaded.roles = true
      })
    }
  }

  function hideRole(role) {
    loaded.roles = false
    const idx = roles.findIndex(r => r.ID === role.ID)
    if (idx >= 0) roles.splice(idx, 1)
    setIncludedRoles(roles)
    loaded.roles = true
  }

  async function readPermissions({ name, roleID }, apiInstance = api) {
    const resource = [...resources]
    return apiInstance.permissionsRead({ resource, roleID })
      .then(rr => {
        rolePermissions.push({ resource: '', ID: `edit-${roleID}`, rules: roleRules(rr, 'edit') })
        roles.push({ mode: 'edit', ID: `edit-${roleID}`, roleID, name })
      })
  }

  async function evaluatePermissions({ name, roleID, userID }, apiInstance = api) {
    const resource = [...resources]
    return apiInstance.permissionsTrace({ resource, roleID, userID })
      .then(rr => {
        const ID = userID ? `eval-${userID}` : `eval-${roleID.join('-')}`
        rolePermissions.push({ resource: '', ID, rules: roleRules(rr, 'eval') })
        roles.push({ mode: 'eval', ID, roleID, userID, name })
      })
  }

  function roleRules(rules, mode = 'edit') {
    return (rules || []).reduce((map, { resource, operation, access, resolution }) => {
      const [type] = resource.split('/', 2)
      if ((permissions[type] || { ops: [] }).ops.indexOf(operation) > -1) {
        if (mode === 'eval') {
          if (resolution === 'unknown-context') {
            access = 'unknown-context'
          } else if (access === 'inherit') {
            access = 'deny'
          }
        }
        map[`${operation}@${resource}`] = access
      }
      return map
    }, {})
  }

  function animateSuccess(key) {
    key.success = true
    setTimeout(() => { key.success = false }, 2000)
  }

  return {
    roles,
    rolePermissions,
    permissions,
    effective,
    resources,
    loaded,
    permission,
    incLoader,
    decLoader,
    prepareRoles,
    fetchPermissions,
    onSubmit,
    addRole,
    hideRole,
    readPermissions,
    evaluatePermissions,
    roleRules,
    animateSuccess,
  }
}
