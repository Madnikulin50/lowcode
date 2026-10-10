<template>
  <div
    v-if="info"
    class="d-flex align-items-center flex-wrap gap-2 small"
    data-test-id="file-versions"
  >
    <span class="text-break">{{ info.name }}</span>
    <span
      v-if="label"
      class="badge bg-secondary"
    >{{ label }}</span>

    <button
      v-if="editable"
      type="button"
      class="btn btn-sm btn-outline-primary"
      :disabled="uploading"
      data-test-id="button-new-version"
      @click="picker.click()"
    >
      <span
        v-if="uploading"
        class="spinner-border spinner-border-sm me-1"
      />
      {{ $t('kind.file.versions.upload') }}
    </button>
    <input
      ref="picker"
      type="file"
      class="d-none"
      :accept="accept"
      @change="onPicked"
    >

    <button
      type="button"
      class="btn btn-sm btn-outline-secondary"
      data-test-id="button-file-history"
      @click="showHistory"
    >
      {{ $t('kind.file.versions.history') }}
    </button>

    <Teleport to="body">
      <div
        ref="modalEl"
        class="modal fade"
        tabindex="-1"
      >
        <div class="modal-dialog modal-lg modal-dialog-scrollable">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">
                {{ $t('kind.file.versions.title') }}: {{ info.name }}
              </h5>
              <button
                type="button"
                class="btn-close"
                @click="modal && modal.hide()"
              />
            </div>
            <div class="modal-body">
              <div
                v-if="loading"
                class="text-center py-3"
              >
                <span class="spinner-border" />
              </div>
              <table
                v-else
                class="table table-sm align-middle mb-0"
              >
                <tbody>
                  <tr
                    v-for="v in versions"
                    :key="v.attachmentID"
                  >
                    <td class="text-nowrap">
                      <strong>{{ versionLabel(v) || '—' }}</strong>
                      <span
                        v-if="isCurrent(v)"
                        class="badge bg-success ms-1"
                      >{{ $t('kind.file.versions.current') }}</span>
                    </td>
                    <td class="text-break">
                      {{ v.name }}
                      <div class="text-muted">
                        {{ size(v) }} · {{ date(v) }}
                      </div>
                    </td>
                    <td class="text-end text-nowrap">
                      <a
                        :href="v.download"
                        class="btn btn-sm btn-outline-secondary"
                      >{{ $t('kind.file.versions.download') }}</a>
                      <button
                        v-if="editable && canMakeCurrent([attachmentId], v.attachmentID)"
                        type="button"
                        class="btn btn-sm btn-outline-primary ms-1"
                        data-test-id="button-make-current"
                        @click="makeCurrent(v)"
                      >
                        {{ $t('kind.file.versions.makeCurrent') }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { Modal } from 'bootstrap'
import numeral from 'numeral'
import moment from 'moment'
import { shared } from 'corteza-lib/js/dist'
import { url, composables } from 'corteza-lib/vue/dist'
import { versionLabel, canMakeCurrent } from 'corteza-webapp-compose/src/lib/file-versions'

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  // ID of the file shown (any version of it)
  attachmentId: { type: String, required: true },
  // editors may add versions and make one current; viewers only read
  editable: { type: Boolean, default: false },
  accept: { type: String, default: '' },
})

// `replace`: the value of the field should point to `to` instead of `from`
const emit = defineEmits(['replace'])

const { t: $t } = useI18n({ useScope: 'global' })
const { toastErrorHandler, toastSuccess } = composables.useToast()
const $ComposeAPI = inject('$ComposeAPI', typeof window !== 'undefined' ? window.__composeAPI : undefined)
const baseURL = url.Make({ url: window.CortezaAPI + '/compose' })

const picker = ref(null)
const modalEl = ref(null)
const loading = ref(false)
const uploading = ref(false)
const versions = ref([])
let modal

const info = computed(() => versions.value.find(v => v.attachmentID === props.attachmentId) || versions.value[0])
const label = computed(() => versionLabel(info.value))

const ids = () => ({
  namespaceID: props.namespace.namespaceID,
  moduleID: props.record.moduleID,
  recordID: props.record.recordID,
  fieldName: props.field.name,
  attachmentID: props.attachmentId,
})

async function load () {
  loading.value = true
  try {
    const { set } = await $ComposeAPI.recordAttachmentVersions(ids())
    versions.value = (set || []).map(a => new shared.Attachment(a, baseURL))
  } catch (e) {
    toastErrorHandler($t('kind.file.versions.loadFailed'))(e)
  } finally {
    loading.value = false
  }
}

// "current" is what the form holds now, which may be unsaved
const isCurrent = v => String(v.attachmentID) === String(props.attachmentId)
const size = v => numeral(v.meta && v.meta.original ? v.meta.original.size : 0).format('0b')
const date = v => (v.createdAt ? moment(v.createdAt).format('lll') : '')

async function showHistory () {
  if (!modal) modal = new Modal(modalEl.value)
  modal.show()
  await load()
}

async function onPicked (ev) {
  const file = ev.target.files && ev.target.files[0]
  ev.target.value = ''
  if (!file) return

  uploading.value = true
  try {
    const att = await $ComposeAPI.recordAttachmentVersionUpload({ ...ids(), file })
    toastSuccess($t('kind.file.versions.uploaded', { version: att.version }))
    emit('replace', { from: props.attachmentId, to: att.attachmentID })
  } catch (e) {
    toastErrorHandler($t('kind.file.versions.uploadFailed'))(e)
  } finally {
    uploading.value = false
  }
}

function makeCurrent (v) {
  emit('replace', { from: props.attachmentId, to: v.attachmentID })
  if (modal) modal.hide()
}

// the file shown changed (new version, or another one made current): refresh
watch(() => props.attachmentId, load)
onMounted(load)
onBeforeUnmount(() => {
  if (modal) {
    modal.dispose()
    modal = undefined
  }
})
</script>
