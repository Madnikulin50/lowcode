<template>
  <div>
    <div
      id="permissions-modal"
      ref="modalRef"
      class="modal fade h-100 overflow-hidden"
      tabindex="-1"
      data-bs-backdrop="static"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ translatedTitle }}
            </h5>
          </div>
          <div class="modal-body d-flex flex-column p-0">
            <div class="row g-0 bg-light border-bottom">
              <div class="col-lg-4 p-3">
                {{ labels.edit?.description }}
              </div>
              <div class="col d-none d-lg-block border-start p-3">
                {{ labels.evaluate?.description }}
              </div>
            </div>

            <div class="row g-0 bg-light align-items-stretch">
              <div class="col-lg-4 d-flex align-items-center p-3 border-bottom">
                <div class="mb-0 w-100">
                  <label class="text-primary mb-1">{{ labels.edit?.label }}</label>
                  <c-input-role
                    v-model="currentRoleID"
                    data-test-id="select-user-list-roles"
                    :visible="isRoleVisible"
                    :clearable="false"
                    preselect
                    @input="onRoleChange"
                  />
                </div>
              </div>

              <div
                v-for="(e, i) in evaluate"
                :key="i"
                data-test-id="icon-remove"
                class="col-lg-2 pointer hide-eval border-bottom d-none d-lg-flex flex-column align-items-center justify-content-center overflow-hidden border-start p-3"
                @click="onHideEval(i)"
              >
                <label
                  v-for="(n, index) in getEvalName(e)"
                  :key="index"
                  :title="n"
                  class="pointer text-center text-primary mb-1"
                >
                  {{ n }}
                </label>
                <font-awesome-icon
                  :icon="['fas', 'plus']"
                  class="text-secondary rotate mt-1"
                />
              </div>

              <div
                v-if="evaluate.length < 4"
                data-test-id="icon-add"
                class="d-none d-lg-flex pointer border-bottom flex-column align-items-center justify-content-center overflow-hidden border-start p-3"
                data-bs-toggle="modal"
                data-bs-target="#permissions-modal-eval"
              >
                <label class="pointer text-center text-primary mb-1">
                  {{ labels.add?.label }}
                </label>
                <font-awesome-icon
                  :icon="['fas', 'plus']"
                  class="text-success d-block mx-auto mt-1"
                />
              </div>
            </div>

            <div
              v-if="processing"
              class="d-flex flex-column align-items-center justify-content-center h-100 py-4"
              style="min-height: 50vh;"
            >
              <span class="spinner-border spinner-border-sm" />
              <div>
                {{ labels.loading }}
              </div>
            </div>

            <div
              v-else
              class="row g-0"
            >
              <div class="col-lg-4 p-3">
                <rules
                  v-model:rules="rules"
                />
              </div>

              <div
                v-for="(e, i) in evaluate"
                :key="i"
                class="col-lg-2 d-none d-lg-flex border-start p-3 bg-light not-allowed"
              >
                <div class="d-flex flex-column align-items-center justify-content-between mt-4 w-100">
                  <h5
                    v-for="r in e.rules"
                    :key="r.operation"
                    :title="getRuleTooltip(r.access === 'unknown-context', !!e.userID)"
                    class="text-center mb-1 mt-2 w-100"
                  >
                    <font-awesome-icon
                      v-if="r.access === 'unknown-context'"
                      :icon="['fas', 'question']"
                      class="text-secondary"
                    />
                    <font-awesome-icon
                      v-else-if="r.access === 'allow'"
                      :icon="['fas', 'check']"
                      class="text-success"
                    />
                    <font-awesome-icon
                      v-else
                      :icon="['fas', 'times']"
                      class="text-danger"
                    />
                  </h5>
                </div>
              </div>

              <div class="col d-none d-lg-block pt-4 border-start" />
            </div>
          </div>

          <div class="modal-footer">
            <button
              data-test-id="button-cancel"
              type="button"
              class="btn btn-light"
              @click="onHide"
            >
              {{ labels.cancel }}
            </button>
            <c-button-submit
              data-test-id="button-save"
              :disabled="submitDisabled"
              :processing="submitting"
              :text="labels.save"
              @submit="onSubmit"
            />
          </div>
        </div>
      </div>
    </div>

    <div
      id="permissions-modal-eval"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ labels.add?.title }}
            </h5>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="text-primary mb-1">{{ labels.add?.role?.label }}</label>
              <c-input-role
                v-model="add.roleID"
                data-test-id="select-role"
                :placeholder="labels.add?.role?.placeholder"
                multiple
                :disabled="!!add.userID"
              />
            </div>
            <div class="mb-0">
              <label class="text-primary mb-1">{{ labels.add?.user?.label }}</label>
              <c-input-select
                v-model="add.userID"
                data-test-id="select-user"
                :disabled="!!add.roleID.length"
                :options="userOptions"
                :get-option-label="getUserLabel"
                :placeholder="labels.add?.user?.placeholder"
                :filterable="false"
                @search="searchUsers"
              />
            </div>
          </div>
          <div class="modal-footer">
            <c-button-submit
              data-test-id="button-save"
              :disabled="!addEnabled"
              :processing="processing"
              :text="labels.add?.save"
              @submit="onAddEval"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, getCurrentInstance, nextTick } from 'vue'
import { Modal } from 'bootstrap'
import { modalOpenEventName, split } from './def.ts'
import CInputSelect from '../input/CInputSelect.vue'
import CInputRole from '../input/CInputRole.vue'
import Rules from './form/Rules.vue'
import { useToast } from '../../composables/useToast'

interface Rule {
  operation: string
  access: string
  current?: string
  resource?: string
}

interface EvaluateEntry {
  roleID?: Array<{ roleID: string }>
  userID?: string
  rules: Rule[]
}

interface AddEntry {
  roleID: Array<{ roleID: string }>
  userID?: string
}

const props = withDefaults(defineProps<{
  labels?: Record<string, any>
}>(), {
  labels: () => ({}),
})

const instance = getCurrentInstance()
const $t = instance!.appContext.config.globalProperties.$t
const { toastSuccess, toastErrorHandler } = useToast()

function getApi(name: string) {
  const key = '$' + name.charAt(0).toUpperCase() + name.slice(1) + 'API'
  return (instance!.appContext.config.globalProperties as Record<string, any>)[key]
}

const modalRef = ref<HTMLDivElement | null>(null)
let modalInstance: Modal | null = null

const processing = ref(false)
const submitting = ref(false)

const backendComponentName = ref<string | undefined>()
const resource = ref<string | undefined>()
const title = ref<string | undefined>()
const target = ref<string | undefined>()
const allSpecific = ref(false)

const userOptions = ref<string[]>([])
const permissions = ref<Rule[]>([])
const rules = ref<Rule[]>([])
const currentRoleID = ref<{ roleID: string; name: string; isBypass: boolean } | undefined>()
const evaluate = ref<EvaluateEntry[]>([])
const fetchedUsers = ref<Record<string, string>>({})
const add = ref<AddEntry>({ roleID: [], userID: undefined })

const api = computed(() => {
  const s = backendComponentName.value
  return s ? getApi(s) : undefined
})

const showModal = computed(() => !!(resource.value && api.value))

const dirty = computed(() => collectChangedRules().length > 0)

const submitDisabled = computed(() => !dirty.value || processing.value || submitting.value)

const addEnabled = computed(() => {
  const { roleID = [], userID } = add.value
  return (roleID && roleID.length > 0) || !!userID
})

const translatedTitle = computed(() => {
  if (resource.value) {
    const { i18nPrefix } = split(resource.value)
    let targetText: string
    if (allSpecific.value) {
      targetText = $t(`${i18nPrefix}.all-specific`, { target: title.value, interpolation: { escapeValue: false } })
    } else if (title.value) {
      targetText = $t(`${i18nPrefix}.specific`, { target: title.value, interpolation: { escapeValue: false } })
    } else {
      targetText = $t(`${i18nPrefix}.all`)
    }
    return $t('ui.set-for', { target: targetText, interpolation: { escapeValue: false } })
  }
  return undefined
})

function loadModal(e: Event) {
  const { resource: r, title: t, target: tg, allSpecific: a } = (e as CustomEvent).detail
  processing.value = true

  resource.value = r
  title.value = t
  target.value = tg
  allSpecific.value = a
  backendComponentName.value = r.split(':')[2]

  fetchPermissions().then(() => {
    if (currentRoleID.value) {
      const { roleID: rid } = currentRoleID.value
      return reEvaluatePermissions(rid)
    }
  }).finally(() => {
    processing.value = false
  })

  nextTick(() => {
    if (modalRef.value && !modalInstance) {
      modalInstance = new Modal(modalRef.value, { backdrop: 'static' })
    }
    modalInstance?.show()
  })
}

function onHide() {
  resource.value = undefined
  title.value = undefined
  target.value = undefined
  modalInstance?.hide()
}

function onRoleChange({ roleID: rid }: { roleID: string }) {
  processing.value = true
  fetchRules(rid).finally(() => {
    processing.value = false
  })
}

async function onSubmit() {
  submitting.value = true
  const changedRules = collectChangedRules()
  const { roleID: rid } = currentRoleID.value!
  try {
    await api.value.permissionsUpdate({ roleID: rid, rules: changedRules })
    await reEvaluatePermissions(rid)
    toastSuccess($t('ui.notification.save.success'))
  } catch (err) {
    toastErrorHandler($t('ui.notification.save.failed'))(err as Error)
  } finally {
    setTimeout(() => {
      submitting.value = false
    }, 300)
  }
}

async function fetchPermissions() {
  rules.value = []
  permissions.value = []
  const pp = await api.value.permissionsList()
  permissions.value = filterPermissions(pp)
}

async function fetchRules(roleID: string) {
  const rr = await api.value.permissionsRead({ roleID, resource: resource.value })
  rules.value = normalizeRules(rr)
}

function isRoleVisible({ isBypass }: { isBypass: boolean }) {
  return !isBypass
}

async function evaluatePermissions({ resource: res = resource.value, roleID, userID }: { resource?: string; roleID: string[]; userID?: string }) {
  processing.value = true
  try {
    return await api.value.permissionsTrace({ resource: res, roleID, userID })
  } finally {
    processing.value = false
  }
}

async function reEvaluatePermissions(roleID: string) {
  await fetchRules(roleID)
  const results = await Promise.all(
    evaluate.value.map(async (e) => {
      let rids: string[] = []
      if (e.roleID) {
        rids = e.roleID.map(({ roleID: rid }) => rid)
      }
      const rules = await evaluatePermissions({ roleID: rids, userID: e.userID })
      return { ...e, rules: normalizeRules(rules, true) }
    }),
  )
  evaluate.value = results
}

function searchUsers(query = '', loading: (v: boolean) => void) {
  loading(true)
  const sysApi = (instance!.appContext.config.globalProperties as Record<string, any>)['$SystemAPI']
  sysApi.userList({ query, limit: 15 })
    .then(({ set }: { set: Array<Record<string, any>> }) => {
      userOptions.value = set.reduce((acc: string[], { userID, name, username, email, handle }: Record<string, any>) => {
        if (!fetchedUsers.value[userID]) {
          fetchedUsers.value[userID] = name || username || email || `<@${userID}>`
        }
        acc.push(userID)
        return acc
      }, [])
    })
    .finally(() => {
      loading(false)
    })
}

function getUserLabel(userID: string) {
  return fetchedUsers.value[userID]
}

function onAddEval() {
  const userID = add.value.userID || undefined
  let roleIDs: string[] = []
  if (add.value.roleID && add.value.roleID.length > 0) {
    roleIDs = add.value.roleID.map(({ roleID: rid }) => rid)
  }
  evaluatePermissions({ roleID: roleIDs, userID }).then((rr) => {
    evaluate.value.push({
      ...add.value as unknown as EvaluateEntry,
      rules: normalizeRules(rr, true),
    })
    add.value = { roleID: [], userID: undefined }
    const evalModal = document.getElementById('permissions-modal-eval')
    if (evalModal) {
      const inst = Modal.getInstance(evalModal)
      inst?.hide()
    }
  })
}

function onHideEval(i: number) {
  evaluate.value.splice(i, 1)
}

function getEvalName(e: EvaluateEntry) {
  if (e.userID) {
    return [fetchedUsers.value[e.userID]]
  }
  return (e.roleID || []).map(({ name }) => name)
}

function normalizeRules(rr: Rule[] | null, fallback = false): Rule[] {
  const inherit = 'inherit'

  function findCurrent({ operation }: { operation: string }) {
    if (!rr) return inherit
    const found = rr.find(r => r.operation === operation) || {} as Rule
    let { resolution, access = inherit } = found
    if (resolution === 'unknown-context') {
      access = 'unknown-context'
    } else if (fallback && access === inherit) {
      access = 'deny'
    }
    return access
  }

  return permissions.value.map(p => {
    const current = findCurrent(p)
    return { ...p, access: current, current }
  })
}

function filterPermissions(pp: Array<Record<string, any>>): Rule[] {
  const [resourceType] = resource.value!.split('/', 2)
  return pp
    .filter(({ type }) => resourceType === type)
    .map(({ type, op: operation }) => {
      return {
        ...describePermission({ resource: type, operation }),
        operation,
        resource: resource.value,
      }
    })
}

function collectChangedRules(): Array<{ resource: string; operation: string; access: string }> {
  return rules.value
    .filter(r => r.access !== r.current)
    .map(({ resource, operation, access }) => ({ resource, operation, access }))
}

function describePermission({ resource: res, operation }: { resource: string; operation: string }) {
  const i18nPrefix = split(res).i18nPrefix + `.operations.${operation}`
  let translatedTitle = ''
  if (allSpecific.value) {
    translatedTitle = $t(`${i18nPrefix}.all-specific`, { target: target.value, interpolation: { escapeValue: false } })
  } else if (target.value) {
    translatedTitle = $t(`${i18nPrefix}.specific`, { target: target.value, interpolation: { escapeValue: false } })
  } else {
    translatedTitle = $t(`${i18nPrefix}.title`)
  }
  return {
    title: translatedTitle,
    description: $t(`${i18nPrefix}.description`),
  }
}

function getRuleTooltip(isUnknown = false, isUser?: boolean) {
  if (!isUnknown) return ''
  return $t(`ui.tooltip.unknown-context.${isUser ? 'user' : 'role'}`)
}

onMounted(() => {
  searchUsers('', () => {})
  window.addEventListener(modalOpenEventName, loadModal as any)

  if (modalRef.value) {
    modalRef.value.addEventListener('hidden.bs.modal', () => {
      resource.value = undefined
      title.value = undefined
      target.value = undefined
    })
  }
})

onBeforeUnmount(() => {
  window.removeEventListener(modalOpenEventName, loadModal as any)
})
</script>

<style scoped lang="scss">
.not-allowed {
  cursor: not-allowed;
}

.pointer {
  cursor: pointer;
}

.rotate {
  transform: rotate(45deg);
}

.hide-eval:hover {
  .rotate {
    color: var(--primary) !important;
  }
}
</style>
