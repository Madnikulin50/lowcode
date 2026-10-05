<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="ae-section-title">
        {{ $t('automation.prompts.editor.info.title') }}
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
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.prompts.editor.info.handle') }}</label>
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
                {{ $t('automation.prompts.editor.info.invalid-handle') }}
              </div>
              <div
                v-else-if="!isNew"
                class="form-text"
              >
                @prompt:{{ form.handle }}
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.prompts.editor.info.description') }}</label>
              <input
                v-model="form.description"
                class="form-control"
              >
            </div>
          </div>

          <div
            v-if="!isNew && history.length"
            class="col-12"
          >
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.prompts.editor.info.versions') }}</label>
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
                  >{{ $t('automation.prompts.active') }}</span>
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
                  {{ $t('automation.prompts.makeActive') }}
                </button>
              </div>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.prompts.text') }}</label>
              <textarea
                v-model="form.text"
                class="form-control font-monospace"
                rows="12"
                spellcheck="false"
              />
              <div class="form-text">
                {{ $t('automation.prompts.textHint') }}
              </div>
            </div>
          </div>

          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.prompts.cases') }}</label>
              <textarea
                v-model="form.casesText"
                :class="['form-control font-monospace', { 'is-invalid': !!casesError }]"
                rows="6"
                spellcheck="false"
                :placeholder="casesExample"
              />
              <div
                v-if="casesError"
                class="invalid-feedback d-block"
              >
                {{ casesError }}
              </div>
              <div
                v-else
                class="form-text"
              >
                {{ $t('automation.prompts.casesHint') }}
              </div>
            </div>
          </div>

          <div class="col-12 col-lg-8">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.prompts.note') }}</label>
              <input
                v-model="form.note"
                class="form-control"
                :placeholder="$t('automation.prompts.noteHint')"
              >
            </div>
          </div>

          <div class="col-12 col-lg-4">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('automation.prompts.activateOnSave') }}</label>
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
        :text="$t('automation.prompts.editor.info.delete')"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('automation.prompts.save')"
        class="ms-auto"
        @submit="$emit('submit')"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'automation.prompts' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  // {handle, description, text, note, casesText, activate}
  form: { type: Object, required: true },
  isNew: { type: Boolean, default: false },
  // [{version, active, note, createdAt}], oldest first
  history: { type: Array, default: () => [] },
  viewing: { type: Number, default: 0 },
  processing: { type: Boolean, default: false },
  success: { type: Boolean, default: false },
  error: { type: String, default: '' },
})

defineEmits(['submit', 'delete', 'view', 'activate'])

const casesExample = '[{"name": "outage", "inputs": {"ticket": "server down"}, "expect": {"priority": "high"}}]'

const viewingBrief = computed(() => props.history.find(v => v.version === props.viewing))

// same shape the server accepts: lowercase letters, digits, _ - . ; a letter first
const handleState = computed(() => {
  if (!props.isNew || !props.form.handle) return null
  return /^[a-z][a-z0-9_.-]{0,63}$/.test(props.form.handle) ? null : false
})

// the test cases must be a JSON list (or empty: none)
const casesError = computed(() => {
  const text = props.form.casesText.trim()
  if (!text) return ''
  try {
    return Array.isArray(JSON.parse(text)) ? '' : t('automation.prompts.casesNotArray')
  } catch (e) {
    return e.message
  }
})

const saveDisabled = computed(() => !props.form.text.trim() || !props.form.handle || handleState.value === false || !!casesError.value)
</script>
