<template>
  <div class="h-100">
    <div class="card shadow h-100">
      <template v-if="loaded && canGrant">
        <div class="border-bottom">
          <div class="row g-0 align-items-stretch">
            <div class="col-4 text-start p-3">
              <small>{{ $t('ui.click-on-cell-to-allow') }}</small>
            </div>
            <div
              v-for="role in roles"
              :key="role.ID"
              data-test-id="button-hide-role"
              class="col d-flex flex-column align-items-center justify-content-center pointer hide-role border-start p-3 overflow-hidden"
              @click="onHideRole(role)"
            >
              <label
                v-for="(n, index) in role.name"
                :key="index"
                :title="n"
                class="pointer text-center text-primary text-break mb-1"
              >{{ n }}</label>
              <font-awesome-icon :icon="['fas', 'plus']" class="text-light rotate" />
            </div>
            <div
              v-if="roles.length < 8"
              data-test-id="button-add-role"
              class="col d-flex flex-column align-items-center justify-content-center border-start p-3 overflow-hidden"
              @click="showAddModal = true"
            >
              <label class="pointer text-center text-primary text-break mb-1">{{ $t('ui.add.label') }}</label>
              <font-awesome-icon :icon="['fas', 'plus']" class="text-success" />
            </div>
          </div>
        </div>
      </template>

      <div v-if="!loaded || !canGrant" class="d-flex align-items-center justify-content-center h-100 pb-4">
        <div v-if="!loaded">
          <div class="spinner-border align-middle m-5" />
          <div>{{ $t('ui.loading') }}</div>
        </div>
        <div v-else-if="!canGrant" class="text-danger">{{ $t('ui.not-allowed') }}</div>
      </div>

      <div v-else class="overflow-auto p-0">
        <div v-for="(type, i) in sortedPermissions" :key="type">
          <div class="row g-0 bg-light border-bottom text-primary sticky-top align-items-stretch">
            <div class="col-4 align-self-center p-3 text-start">
              <span class="h6 mb-0">{{ getTranslation(type) }}</span>
            </div>
            <div v-for="role in roles" :key="role.ID" class="col d-flex align-items-center justify-content-center overflow-hidden p-3 text-center border-start not-allowed">
              <p v-if="i === 0" :title="$t(`ui.${role.mode === 'edit' ? 'edit' : 'evaluate'}.title`)" class="mb-0">{{ $t(`ui.${role.mode === 'edit' ? 'edit' : 'evaluate'}.title`) }}</p>
            </div>
            <div class="col p-3 border-start not-allowed" />
          </div>
          <div v-for="operation in permissions[type].ops" :key="operation" :data-test-id="`permission-${operation}`" class="row g-0">
            <div class="col-4 border-bottom text-start p-3">
              <span :title="getTranslation(type, operation)">{{ getTranslation(type, operation) }}</span>
            </div>
            <div
              v-for="role in roles" :key="role.ID"
              class="col d-flex align-items-center justify-content-center border-bottom border-start p-3 pointer active-cell h5 mb-0"
              :class="{
                'not-allowed bg-extra-light': role.mode === 'eval',
                'bg-warning': checkChange(role.ID, permissions[type].any, operation)
              }"
              :title="getRuleTooltip(checkRule(role.ID, permissions[type].any, operation, 'unknown-context'), !!role.userID)"
              @click="role.mode === 'edit' ? ruleChange($event, role.ID, permissions[type].any, operation) : undefined"
            >
              <font-awesome-icon v-if="checkRule(role.ID, permissions[type].any, operation, 'unknown-context')" data-test-id="permission-unknown" :icon="['fas', 'question']" class="text-secondary" />
              <font-awesome-icon v-else-if="checkRule(role.ID, permissions[type].any, operation, 'allow')" data-test-id="permission-allowed" :icon="['fas', 'check']" class="text-success" />
              <font-awesome-icon v-else data-test-id="permission-denied" :icon="['fas', 'times']" class="text-danger" />
            </div>
            <div v-if="roles.length < 8" class="col border-bottom border-start p-3 not-allowed bg-extra-light" />
          </div>
        </div>
      </div>

      <div v-if="loaded && canGrant" class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
        <c-button-submit
          :processing="processing"
          :success="success"
          :text="$t('ui.save')"
          class="ms-auto"
          @submit="onSubmit"
        />
      </div>
    </div>

    <div v-if="showAddModal" class="modal fade show d-block" tabindex="-1">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('ui.edit-or-eval') }}</h5>
            <button type="button" class="btn-close" @click="showAddModal = false" />
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <div class="btn-group w-100" role="group">
                <button
                  v-for="opt in modeOptions" :key="opt.value"
                  type="button"
                  :class="['btn', add.mode === opt.value ? 'btn-primary' : 'btn-outline-primary']"
                  @click="add.mode = opt.value"
                >
                  {{ opt.text }}
                </button>
              </div>
            </div>

            <p>{{ addModeDescription }}</p>

            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('ui.add.role.label') }}</label>
              <c-input-role
                v-model="add.roleID"
                :data-test-id="`select-${add.mode}-roles`"
                :placeholder="$t('ui.add.role.placeholder')"
                :visible="isRoleVisible"
                :multiple="add.mode === 'eval'"
                :disabled="add.mode === 'eval' && !!add.userID"
              />
            </div>

            <div v-if="add.mode === 'eval'" class="mb-0">
              <label class="form-label text-primary">{{ $t('ui.add.user.label') }}</label>
              <c-input-select
                v-model="add.userID"
                :data-test-id="`select-${add.mode}-users`"
                :disabled="!!add.roleID.length"
                :options="userOptions"
                :get-option-label="getUserLabel"
                :placeholder="$t('ui.add.user.placeholder')"
                :filterable="false"
                @search="searchUsers"
              />
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-outline-secondary" @click="showAddModal = false">{{ $t('label.cancel') }}</button>
            <button type="button" class="btn btn-primary" :disabled="!addEnabled" @click="onAdd">{{ $t('ui.add.save') }}</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showAddModal" class="modal-backdrop fade show" />
  </div>
</template>

<script setup>
import { ref, computed, reactive, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import _ from 'lodash'
import { components } from 'corteza-lib/vue/dist'
const { CInputRole } = components

const { t } = useI18n()

const props = defineProps({
  roles: { type: Array, required: true },
  permissions: { type: Object, required: true },
  rolePermissions: { type: Array, required: true },
  canGrant: { type: Boolean, value: false },
  loaded: { type: Boolean, value: false },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  component: { type: String, required: true },
})

const emit = defineEmits(['submit', 'add', 'hide'])

const showAddModal = ref(false)
const add = reactive({
  mode: 'edit',
  roleID: [],
  userID: undefined,
})
const modeOptions = [
  { text: 'Edit', value: 'edit' },
  { text: 'Evaluate', value: 'eval' },
]
const userOptions = ref([])
const permissionChanges = ref([])
const fetchedUsers = reactive({})

const sortedPermissions = computed(() => Object.keys(props.permissions).sort())

const addModeDescription = computed(() => add.mode === 'edit' ? t('ui.add.edit.description') : t('ui.add.evaluate.description'))

const addEnabled = computed(() => {
  const { mode, roleID = [], userID } = add
  if (mode === 'edit') {
    return Array.isArray(roleID) ? roleID.length : roleID
  } else if (mode === 'eval') {
    return (roleID && roleID.length) || userID
  }
  return false
})

watch(() => add.mode, (mode) => {
  add.roleID = mode === 'eval' ? [] : undefined
  add.userID = undefined
})

onMounted(() => {
  searchUsers('', () => {})
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function checkRule(ID, res, op, access) {
  const key = `${op}@${res}`
  return (props.rolePermissions.find(r => r.ID === ID) || { rules: {} }).rules[key] === access
}

function checkChange(ID, res, op) {
  const key = `${op}@${res}`
  const current = (props.rolePermissions.find(r => r.ID === ID) || { rules: {} }).rules[key]
  const initial = (permissionChanges.value.find(r => r.ID === ID) || { rules: {} }).rules[key]
  if (initial) {
    return current !== initial
  }
  return false
}

function ruleChange(event, ID, res, op) {
  const key = `${op}@${res}`
  let access = (props.rolePermissions.find(r => r.ID === ID) || { rules: {} }).rules[key]

  if (!(permissionChanges.value.find(r => r.ID === ID) || { rules: {} }).rules[key]) {
    permissionChanges.value.push({ ID, rules: {} })
    if (!access) {
      access = 'inherit'
    }
    permissionChanges.value.find(r => r.ID === ID).rules[key] = access
  }

  if (access === 'allow') {
    access = 'inherit'
  } else {
    access = 'allow'
  }

  props.rolePermissions.find(r => r.ID === ID).rules[key] = access
}

function searchUsers(query = '', loading) {
  loading(true)

  window.__systemAPI.userList({ query, limit: 15 })
    .then(({ set }) => {
      userOptions.value = set.reduce((acc, { userID, name, username, email }) => {
        if (!fetchedUsers[userID]) {
          fetchedUsers[userID] = name || username || email || `<@${userID}>`
        }
        acc.push(userID)
        return acc
      }, [])
    })
    .finally(() => {
      loading(false)
    })
}

function isRoleVisible({ isBypass }) {
  return add.mode === 'edit' || !isBypass
}

function getUserLabel(userID) {
  return fetchedUsers[userID]
}

function getTranslation(resource, operation = '') {
  resource = _.kebabCase(resource.split(':')[3]) || 'component'
  if (operation) {
    return t(`resources.${props.component}.${resource}.operations.${operation}.title`)
  }
  return t(`resources.${props.component}.${resource}.label`)
}

function getRuleTooltip(isUnknown = false, isUser) {
  if (!isUnknown) return ''
  return t(`ui.tooltip.unknown-context.${isUser ? 'user' : 'role'}`)
}

function onSubmit() {
  emit('submit', props.rolePermissions)
  permissionChanges.value = []
}

function onAdd() {
  let { userID } = add
  if (userID) {
    userID = { userID: add.userID, name: fetchedUsers[add.userID] }
  }
  emit('add', { ...add, userID })
  add.mode = 'edit'
  add.roleID = []
  add.userID = undefined
  showAddModal.value = false
}

function onHideRole(role) {
  emit('hide', role)
}

function setDefaultValues() {
  add.mode = 'edit'
  add.roleID = []
  add.userID = undefined
  userOptions.value = []
  permissionChanges.value = []
  Object.keys(fetchedUsers).forEach(k => delete fetchedUsers[k])
}
</script>

<style lang="scss" scoped>
.pointer {
  cursor: pointer;
}
.not-allowed {
  cursor: not-allowed;
}
.active-cell:hover {
  background-color: var(--gray-200);
}
.rotate {
  transform: rotate(45deg);
}
.hide-role:hover {
  .rotate {
    color: var(--dark) !important;
  }
}
</style>

<style lang="scss">
.mode {
  .btn {
    background-color: var(--light);
    border: none;
  }
  .btn:nth-child(2), .btn:nth-child(3) {
    margin-left: 0.2rem !important;
  }
}
</style>
