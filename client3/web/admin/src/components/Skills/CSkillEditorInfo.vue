<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="ae-section-title">
        {{ $t('automation.skills.editor.info.title') }}
      </h4>
    </div>

    <div class="card-body">
      <div
        v-if="error"
        class="alert alert-danger py-2"
      >
        {{ error }}
      </div>

      <form @submit.prevent="$emit('submit')">
        <div class="row g-3">
          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.editor.info.handle') }}</label>
              <input
                v-model="form.handle"
                :class="['form-control', { 'is-invalid': handleState === false }]"
                :disabled="!isNew"
                spellcheck="false"
              >
              <div
                v-if="handleState === false"
                class="invalid-feedback"
              >
                {{ $t('automation.skills.editor.info.invalid-handle') }}
              </div>
              <div
                v-else-if="!isNew"
                class="form-text"
              >
                {{ $t('automation.skills.editor.info.handleHint') }}
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-8">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.editor.info.description') }}</label>
              <input
                v-model="form.description"
                class="form-control"
              >
              <div class="form-text">
                {{ $t('automation.skills.editor.info.descriptionHint') }}
              </div>
            </div>
          </div>

          <div
            v-if="!isNew && history.length"
            class="col-12"
          >
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.editor.info.versions') }}</label>
              <div class="d-flex flex-wrap gap-1">
                <button
                  v-for="v in history"
                  :key="v.version"
                  type="button"
                  class="btn btn-sm"
                  :class="v.version === viewing ? 'btn-primary' : 'btn-outline-secondary'"
                  :title="[v.note, v.createdAt].filter(Boolean).join(' · ')"
                  @click="$emit('view', v.version)"
                >
                  v{{ v.version }}
                  <span
                    v-if="v.active"
                    class="badge bg-success ms-1"
                  >{{ $t('automation.skills.active') }}</span>
                </button>
              </div>
              <div
                v-if="viewingBrief"
                class="form-text"
              >
                {{ viewingBrief.createdAt }}<template v-if="viewingBrief.note">
                  · {{ viewingBrief.note }}
                </template>
                <button
                  v-if="!viewingBrief.active"
                  type="button"
                  class="btn btn-sm btn-link py-0"
                  @click="$emit('activate', viewingBrief.version)"
                >
                  {{ $t('automation.skills.makeActive') }}
                </button>
              </div>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.text') }}</label>
              <textarea
                v-model="form.text"
                class="form-control font-monospace"
                rows="16"
                spellcheck="false"
              />
              <div class="form-text">
                {{ $t('automation.skills.textHint') }}
              </div>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.requires') }}</label>
              <div class="d-flex flex-wrap gap-3">
                <div
                  v-for="kit in allKits"
                  :key="kit"
                  class="form-check"
                >
                  <input
                    :id="'req-' + kit"
                    class="form-check-input"
                    type="checkbox"
                    :checked="form.requires.includes(kit)"
                    @change="toggleKit(kit)"
                  >
                  <label
                    class="form-check-label"
                    :for="'req-' + kit"
                  >{{ kit }}</label>
                </div>
                <span
                  v-if="!allKits.length"
                  class="text-muted"
                >{{ $t('automation.skills.noKits') }}</span>
              </div>
              <div class="form-text">
                {{ $t('automation.skills.requiresHint') }}
              </div>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.resources') }}</label>

              <div
                v-for="(file, i) in form.resources"
                :key="i"
                class="border rounded p-2 mb-2"
              >
                <div class="d-flex gap-2 mb-2">
                  <input
                    v-model="file.path"
                    class="form-control font-monospace"
                    :placeholder="$t('automation.skills.filePath')"
                    spellcheck="false"
                  >
                  <button
                    type="button"
                    class="btn btn-outline-danger"
                    @click="form.resources.splice(i, 1)"
                  >
                    {{ $t('automation.skills.removeFile') }}
                  </button>
                </div>
                <textarea
                  v-model="file.content"
                  class="form-control font-monospace"
                  rows="6"
                  spellcheck="false"
                />
                <div class="form-text">
                  {{ fileSize(file) }}
                </div>
              </div>

              <div class="d-flex gap-2">
                <button
                  type="button"
                  class="btn btn-outline-secondary btn-sm"
                  @click="form.resources.push({ path: '', content: '' })"
                >
                  {{ $t('automation.skills.addFile') }}
                </button>
                <button
                  type="button"
                  class="btn btn-outline-secondary btn-sm"
                  @click="fileInput.click()"
                >
                  {{ $t('automation.skills.uploadFile') }}
                </button>
                <input
                  ref="fileInput"
                  type="file"
                  multiple
                  class="d-none"
                  @change="onUpload"
                >
              </div>
              <div
                v-if="resourcesError"
                class="invalid-feedback d-block"
              >
                {{ resourcesError }}
              </div>
              <div
                v-else
                class="form-text"
              >
                {{ $t('automation.skills.resourcesHint') }}
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-8">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.note') }}</label>
              <input
                v-model="form.note"
                class="form-control"
                :placeholder="$t('automation.skills.noteHint')"
              >
            </div>
          </div>

          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.skills.activateOnSave') }}</label>
              <c-input-checkbox
                v-model="form.activate"
                switch
                :labels="{ on: $t('label.general.yes'), off: $t('label.general.no') }"
              />
            </div>
          </div>
        </div>

        <input
          type="submit"
          class="d-none"
          :disabled="saveDisabled"
        >
      </form>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="!isNew"
        :text="$t('automation.skills.editor.info.delete')"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <button
        v-if="!isNew"
        type="button"
        class="btn btn-outline-secondary"
        @click="$emit('export')"
      >
        {{ $t('automation.skills.export') }}
      </button>

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('automation.skills.save')"
        class="ms-auto"
        @submit="$emit('submit')"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'automation.skills' } })
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const MAX_FILES = 50
const MAX_FILE = 100 * 1024
const MAX_TOTAL = 512 * 1024

const props = defineProps({
  // {handle, description, text, note, activate, requires: [], resources: [{path, content}]}
  form: { type: Object, required: true },
  isNew: { type: Boolean, default: false },
  // [{version, active, note, createdAt}], oldest first
  history: { type: Array, default: () => [] },
  viewing: { type: Number, default: 0 },
  // toolkit names that exist right now
  kitOptions: { type: Array, default: () => [] },
  processing: { type: Boolean, default: false },
  success: { type: Boolean, default: false },
  error: { type: String, default: '' },
})

defineEmits(['submit', 'delete', 'view', 'activate', 'export'])

const fileInput = ref(null)

const viewingBrief = computed(() => props.history.find(v => v.version === props.viewing))

// the known toolkits, plus any the skill names that are not there now (a remote
// toolkit may be down): they stay visible and can be unticked
const allKits = computed(() => [...new Set([...props.kitOptions, ...props.form.requires])].sort())

function toggleKit (kit) {
  const i = props.form.requires.indexOf(kit)
  if (i >= 0) {
    props.form.requires.splice(i, 1)
  } else {
    props.form.requires.push(kit)
  }
}

const bytes = (s) => new Blob([s]).size

function fileSize (file) {
  const n = bytes(file.content)
  return n > 1024 ? `${(n / 1024).toFixed(1)} KB` : `${n} B`
}

async function onUpload (ev) {
  for (const file of Array.from(ev.target.files || [])) {
    try {
      props.form.resources.push({ path: file.name, content: await file.text() })
    } catch (e) {
      // not readable as text: skip it, the limits message says what is accepted
    }
  }
  ev.target.value = ''
}

// same limits the server enforces, so the problem shows before saving
const resourcesError = computed(() => {
  const files = props.form.resources
  if (files.length > MAX_FILES) return t('automation.skills.tooManyFiles', { max: MAX_FILES })
  const seen = new Set()
  let total = 0
  for (const f of files) {
    const path = f.path.trim()
    if (!path) continue
    if (path.startsWith('/') || path.split('/').includes('..') || path.toLowerCase() === 'skill.md') {
      return t('automation.skills.badPath', { path })
    }
    if (seen.has(path)) return t('automation.skills.duplicatePath', { path })
    seen.add(path)
    const n = bytes(f.content)
    if (n > MAX_FILE) return t('automation.skills.fileTooBig', { path, max: MAX_FILE / 1024 })
    total += n
  }
  return total > MAX_TOTAL ? t('automation.skills.filesTooBig', { max: MAX_TOTAL / 1024 }) : ''
})

// same shape the server accepts: lowercase letters, digits, _ - . ; a letter first
const handleState = computed(() => {
  if (!props.isNew || !props.form.handle) return null
  return /^[a-z][a-z0-9_.-]{0,63}$/.test(props.form.handle) ? null : false
})

const saveDisabled = computed(() => !props.form.text.trim() || !props.form.handle || handleState.value === false || !!resourcesError.value)
</script>
