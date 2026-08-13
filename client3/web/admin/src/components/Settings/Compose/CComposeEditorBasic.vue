<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('title') }}
      </h4>
    </div>

    <form
      @submit.prevent="emit('submit', basic)"
    >
      <div class="card-body">
        <div class="pb-3">
          <h5>{{ t('attachments.namespace') }}</h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('attachments.max-size') }}</label>
                <input
                  v-model.number="basic['compose.namespace.attachments.max-size']"
                  class="form-control"
                  type="number"
                  placeholder="0"
                >
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-0">
                <label class="form-label text-primary">{{ t('attachments.type.whitelist') }}</label>
                <div class="form-text mb-2">{{ t('attachments.type.description') }}</div>
                <input
                  v-model="namespaceAttachmentWhitelist"
                  class="form-control"
                >
              </div>
            </div>
          </div>
        </div>

        <div class="pb-3">
          <h5>{{ t('attachments.page') }}</h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('attachments.max-size') }}</label>
                <input
                  v-model.number="basic['compose.page.attachments.max-size']"
                  class="form-control"
                  type="number"
                >
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-0">
                <label class="form-label text-primary">{{ t('attachments.type.whitelist') }}</label>
                <div class="form-text mb-2">{{ t('attachments.type.description') }}</div>
                <input
                  v-model="pageAttachmentWhitelist"
                  class="form-control"
                >
              </div>
            </div>
          </div>
        </div>

        <div class="pb-3">
          <h5>{{ t('attachments.record') }}</h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('attachments.max-size') }}</label>
                <input
                  v-model.number="basic['compose.record.attachments.max-size']"
                  class="form-control"
                  type="number"
                >
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-0">
                <label class="form-label text-primary">{{ t('attachments.type.whitelist') }}</label>
                <div class="form-text mb-2">{{ t('attachments.type.description') }}</div>
                <input
                  v-model="recordAttachmentWhitelist"
                  class="form-control"
                >
              </div>
            </div>
          </div>
        </div>

        <div>
          <h5>{{ t('attachments.icon') }}</h5>

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ t('attachments.max-size') }}</label>
                <input
                  v-model.number="basic['compose.icon.attachments.max-size']"
                  class="form-control"
                  type="number"
                >
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-0">
                <label class="form-label text-primary">{{ t('attachments.type.whitelist') }}</label>
                <div class="form-text mb-2">{{ t('attachments.type.description') }}</div>
                <input
                  v-model="iconAttachmentWhitelist"
                  class="form-control"
                >
              </div>
            </div>
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
        @submit="emit('submit', basic)"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'compose.settings', keyPrefix: 'editor.basic' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: _t } = useI18n()

function t (key, ...args) {
  if (key.startsWith('label.') || key.startsWith('admin.')) return _t(key, ...args)
  return _t('editor.basic.' + key, ...args)
}

const props = defineProps({
  basic: {
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

const pageAttachmentWhitelist = computed({
  get () {
    return (props.basic['compose.page.attachments.mimetypes'] || []).join(',')
  },
  set (value) {
    props.basic['compose.page.attachments.mimetypes'] = convertToExternal(value)
  },
})

const recordAttachmentWhitelist = computed({
  get () {
    return (props.basic['compose.record.attachments.mimetypes'] || []).join(',')
  },
  set (value) {
    props.basic['compose.record.attachments.mimetypes'] = convertToExternal(value)
  },
})

const iconAttachmentWhitelist = computed({
  get () {
    return (props.basic['compose.icon.attachments.mimetypes'] || []).join(',')
  },
  set (value) {
    props.basic['compose.icon.attachments.mimetypes'] = convertToExternal(value)
  },
})

const namespaceAttachmentWhitelist = computed({
  get () {
    return (props.basic['compose.namespace.attachments.mimetypes'] || []).join(',')
  },
  set (value) {
    props.basic['compose.namespace.attachments.mimetypes'] = convertToExternal(value)
  },
})

function convertToExternal (value) {
  return (value || '').split(',').map(v => {
    return v.replace(/ /g, '')
  }).filter(v => {
    return v.match(/^[-\w.]+\/[-\w/+.]+$/g) !== null
  })
}
</script>
