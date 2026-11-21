<template>
  <span>
    <button
      data-test-id="button-version-history"
      class="btn btn-outline-secondary text-nowrap"
      :class="{ 'btn-lg': !compact }"
      type="button"
      :disabled="disabled"
      @click="show"
    >
      <font-awesome-icon :icon="['fas', 'clock']" />
      {{ $t('open') }}
    </button>

    <Teleport to="body">
      <div
        ref="modalEl"
        class="modal fade"
        tabindex="-1"
        data-test-id="modal-version-history"
      >
        <div class="modal-dialog modal-xl modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">
                {{ $t('title') }}
              </h5>
              <button
                type="button"
                class="btn-close"
                :aria-label="$t('close')"
                @click="hide"
              />
            </div>

            <div class="modal-body">
              <div
                v-if="loading"
                class="text-center py-4"
              >
                <span class="spinner-border" />
              </div>

              <p
                v-else-if="!versions.length"
                class="text-center text-muted my-4"
              >
                {{ $t('empty') }}
              </p>

              <div
                v-else
                class="row"
              >
                <div class="col-md-4 mb-3">
                  <div class="list-group">
                    <button
                      v-for="(v, idx) in versions"
                      :key="v.versionID"
                      type="button"
                      class="list-group-item list-group-item-action"
                      :class="{ active: selected && selected.version === v.version }"
                      @click="select(v)"
                    >
                      <div class="d-flex justify-content-between">
                        <strong>{{ $t('version') }} {{ v.version }}</strong>
                        <small v-if="idx === 0">{{ $t('current') }}</small>
                      </div>
                      <small>{{ formatDate(v.createdAt) }}</small>
                      <div
                        v-if="v.note"
                        class="small text-truncate"
                      >
                        {{ v.note }}
                      </div>
                    </button>
                  </div>
                </div>

                <div class="col-md-8">
                  <div
                    v-if="loadingDiff"
                    class="text-center py-4"
                  >
                    <span class="spinner-border" />
                  </div>

                  <template v-else-if="selected">
                    <div class="d-flex justify-content-between align-items-center mb-2">
                      <h6 class="mb-0">
                        {{ $t('compare') }}
                      </h6>
                      <button
                        v-if="canRestore"
                        type="button"
                        class="btn btn-primary"
                        :disabled="restoring"
                        data-test-id="button-restore-version"
                        @click="restore"
                      >
                        <span
                          v-if="restoring"
                          class="spinner-border spinner-border-sm me-1"
                        />
                        {{ $t('restore') }}
                      </button>
                    </div>

                    <p
                      v-if="!previous"
                      class="text-muted"
                    >
                      {{ $t('firstVersion') }}
                    </p>
                    <p
                      v-else-if="!changes.length"
                      class="text-muted"
                    >
                      {{ $t('noChanges') }}
                    </p>

                    <table
                      v-else
                      class="table table-sm align-middle"
                    >
                      <tbody>
                        <tr
                          v-for="c in changes"
                          :key="c.path"
                        >
                          <td
                            class="text-nowrap"
                            style="width: 1%;"
                          >
                            <span
                              class="badge"
                              :class="badgeClass[c.kind]"
                            >{{ $t(c.kind) }}</span>
                          </td>
                          <td class="text-break">
                            <code>{{ c.path }}</code>
                            <div
                              v-if="c.kind !== 'added'"
                              class="text-danger small text-break"
                            >
                              − {{ short(c.before) }}
                            </div>
                            <div
                              v-if="c.kind !== 'removed'"
                              class="text-success small text-break"
                            >
                              + {{ short(c.after) }}
                            </div>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </span>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'versions' } })
import { ref, computed, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { Modal } from 'bootstrap'
import { composables } from 'corteza-lib/vue/dist'
import { diff, parseSnapshot } from 'corteza-webapp-compose/src/lib/json-diff'

const props = defineProps({
  // page, module, chart, rulechain or workflow (picks the confirmation text)
  kind: { type: String, required: true },
  namespaceId: { type: String, default: '' },
  resourceId: { type: String, default: '' },
  disabled: { type: Boolean, default: false },

  // for resources that are not served by the compose version API:
  // { list(), get(version), restore(version) } returning what the API returns
  // ({ set }, a version with its snapshot, anything)
  adapter: { type: Object, default: undefined },

  // a regular size button (for headers) instead of the large toolbar one
  compact: { type: Boolean, default: false },
})

const emit = defineEmits(['restored'])

const { t: $t } = useI18n({ useScope: 'global' })
const { toastErrorHandler, toastSuccess } = composables.useToast()

const badgeClass = { added: 'bg-success', removed: 'bg-danger', changed: 'bg-warning text-dark' }

const modalEl = ref(null)
const loading = ref(false)
const loadingDiff = ref(false)
const restoring = ref(false)
const versions = ref([])
const selected = ref(null)
const previous = ref(null)
const changes = ref([])

let modal

const ids = () => ({ namespaceID: props.namespaceId, kind: props.kind, resourceID: props.resourceId })

const api = () => props.adapter || {
  list: () => window.__composeAPI.resourceVersions(ids()),
  get: version => window.__composeAPI.resourceVersion({ ...ids(), version }),
  restore: version => window.__composeAPI.resourceVersionRestore({ ...ids(), version }),
}

// the newest version is what is saved now: nothing to restore
const canRestore = computed(() => versions.value.length > 0 && selected.value && selected.value.version !== versions.value[0].version)

function formatDate (ts) {
  return ts ? new Date(ts).toLocaleString() : ''
}

function short (v) {
  const s = v === null || v === undefined ? 'null' : String(v)
  return s.length > 300 ? s.slice(0, 300) + '…' : s
}

async function show () {
  if (!modal) modal = new Modal(modalEl.value)
  modal.show()

  loading.value = true
  selected.value = null
  try {
    const { set } = await api().list()
    versions.value = set || []
    if (versions.value.length) await select(versions.value[0])
  } catch (e) {
    toastErrorHandler($t('loadFailed'))(e)
  } finally {
    loading.value = false
  }
}

function hide () {
  if (modal) modal.hide()
}

async function select (v) {
  selected.value = v
  previous.value = null
  changes.value = []

  loadingDiff.value = true
  try {
    const before = versions.value.find(o => o.version === v.version - 1)
    const [cur, prev] = await Promise.all([
      api().get(v.version),
      before ? api().get(before.version) : Promise.resolve(null),
    ])

    previous.value = prev
    if (prev) changes.value = diff(parseSnapshot(prev.snapshot), parseSnapshot(cur.snapshot))
  } catch (e) {
    toastErrorHandler($t('loadFailed'))(e)
  } finally {
    loadingDiff.value = false
  }
}

async function restore () {
  const version = selected.value.version
  if (!window.confirm($t(`restoreConfirm.${props.kind}`, { version }))) return

  restoring.value = true
  try {
    await api().restore(version)
    toastSuccess($t('restored', { version }))
    hide()
    emit('restored', version)
  } catch (e) {
    toastErrorHandler($t('restoreFailed'))(e)
  } finally {
    restoring.value = false
  }
}

onBeforeUnmount(() => {
  if (modal) {
    modal.dispose()
    modal = undefined
  }
})
</script>
