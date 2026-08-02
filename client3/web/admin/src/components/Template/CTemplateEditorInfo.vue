<template>
  <div class="card shadow-sm" data-test-id="card-template-info">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit="$emit('submit', template)">
        <div class="row g-3 p-3">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('meta.short') }}</label>
              <input
                v-model="template.meta.short"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': shortState === false }"
                data-test-id="input-short-name"
                required
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" :class="{ 'mb-0': !template.templateID }">
              <label class="form-label text-primary">{{ $t('handle') }}</label>
              <input
                v-model="template.handle"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': handleState === false }"
                data-test-id="input-handle"
                :placeholder="$t('system.templates.editor.info.placeholder-handle')"
              >
              <div v-if="handleState === false" class="invalid-feedback">{{ $t('system.templates.editor.info.invalid-handle-characters') }}</div>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.templates.editor.info.type') }}</label>
              <select
                v-model="template.type"
                class="form-select"
                data-test-id="select-template-type"
              >
                <option v-for="opt in contentTypes" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
              </select>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.templates.editor.info.meta.description') }}</label>
              <textarea
                v-model="template.meta.description"
                class="form-control"
                data-test-id="textarea-description"
              />
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.templates.editor.info.partial') }}</label>
              <small class="form-text text-muted">{{ $t('system.templates.editor.info.partialDescription') }}</small>
              <c-input-checkbox
                v-model="template.partial"
                data-test-id="checkbox-is-partial-template"
                switch
                :labels="checkboxLabel"
                name="checkbox-1"
              />
            </div>
          </div>
        </div>
  
        <c-system-fields :resource="template" />
  
        <input
          type="submit"
          class="d-none"
          :disabled="saveDisabled"
        >
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="!fresh && template.canDeleteTemplate"
        :data-test-id="getDeletedButtonStatusCypressId"
        :text="getDeleteStatus"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', template)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { handle } from 'corteza-lib/vue/dist'

const { t } = useI18n()

const props = defineProps({
  template: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit', 'delete'])

const contentTypes = [
  { value: 'text/html', text: t('contentType.text_html') },
  { value: 'text/plain', text: t('contentType.text_plain') },
]

const checkboxLabel = { on: t('label.general.yes'), off: t('label.general.no') }

const fresh = computed(() => !props.template.templateID || props.template.templateID === NoID)
const editable = computed(() => fresh.value ? props.canCreate : true)
const shortState = computed(() => props.template.meta.short ? null : false)
const handleState = computed(() => handle.handleState(props.template.handle))
const saveDisabled = computed(() => !editable.value || [shortState.value, handleState.value].includes(false))
const getDeleteStatus = computed(() => props.template.deletedAt ? t('undelete') : t('delete'))
const getDeletedButtonStatusCypressId = computed(() => `button-${getDeleteStatus.value.toLowerCase()}`)
</script>