<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('title') }}
      </h4>
    </div>

    <form
      @submit.prevent="emit('submit', settings)"
    >
      <div class="card-body">
        <div class="mb-3">
          <label class="form-label text-primary">
            {{ t('sidebar.title') }}
          </label>
          <div class="form-check">
            <input
              id="hide-namespace-list"
              v-model="sidebar.hideNamespaceList"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-namespace-list"
            >{{ t('sidebar.hide-namespace-list') }}</label>
          </div>
          <div class="form-check">
            <input
              id="hide-namespace-list-link"
              v-model="sidebar.hideNamespaceListLink"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-namespace-list-link"
            >{{ t('sidebar.hide-namespace-list-link') }}</label>
          </div>
          <div class="mt-2">
            <label class="form-label text-primary" for="sidebar-density">{{ t('sidebar.density') }}</label>
            <select
              id="sidebar-density"
              v-model="sidebar.density"
              class="form-select form-control"
            >
              <option value="comfortable">{{ t('sidebar.densityOptions.comfortable') }}</option>
              <option value="compact">{{ t('sidebar.densityOptions.compact') }}</option>
            </select>
          </div>
        </div>

        <div class="mb-0">
          <label class="form-label text-primary">
            {{ t('record-toolbar.title') }}
          </label>
          <div class="form-check">
            <input
              id="hide-submit"
              v-model="recordToolbar.hideSubmit"
              class="form-check-input-v3"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-submit"
            >{{ t('record-toolbar.hide-submit') }}</label>
          </div>
          <div class="form-check">
            <input
              id="hide-delete"
              v-model="recordToolbar.hideDelete"
              class="form-check-input-v3"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-delete"
            >{{ t('record-toolbar.hide-delete') }}</label>
          </div>
          <div class="form-check">
            <input
              id="hide-edit"
              v-model="recordToolbar.hideEdit"
              class="form-check-input-v3"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-edit"
            >{{ t('record-toolbar.hide-edit') }}</label>
          </div>
          <div class="form-check">
            <input
              id="hide-new"
              v-model="recordToolbar.hideNew"
              class="form-check-input-v3"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-new"
            >{{ t('record-toolbar.hide-new') }}</label>
          </div>
          <div class="form-check">
            <input
              id="hide-clone"
              v-model="recordToolbar.hideClone"
              class="form-check-input-v3"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-clone"
            >{{ t('record-toolbar.hide-clone') }}</label>
          </div>
          <div class="form-check">
            <input
              id="hide-back"
              v-model="recordToolbar.hideBack"
              class="form-check-input-v3"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="hide-back"
            >{{ t('record-toolbar.hide-back') }}</label>
          </div>
        </div>
      </div>
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        v-if="canManage"
        :processing="processing"
        :success="success"
        :text="t('admin.general.label.submit')"
        class="ms-auto"
        @submit="onSubmit"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'compose.settings', keyPrefix: 'editor.ui' } })
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: _t } = useI18n()

function t (key, ...args) {
  if (key.startsWith('label.') || key.startsWith('admin.')) return _t(key, ...args)
  return _t('editor.ui.' + key, ...args)
}

const props = defineProps({
  settings: {
    type: Object,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
  canManage: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['submit'])

const sidebar = ref({})
const recordToolbar = ref({})

watch(() => props.settings, (settings) => {
  sidebar.value = {
    density: 'comfortable',
    ...(settings['compose.ui.sidebar'] || {}),
  }
  recordToolbar.value = settings['compose.ui.record-toolbar'] || {}
}, { immediate: true })

function onSubmit () {
  emit('submit', {
    'compose.ui.sidebar': sidebar.value,
    'compose.ui.record-toolbar': recordToolbar.value,
  })
}
</script>
