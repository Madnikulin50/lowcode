<template>
  <div class="d-flex flex-column h-100">
    <div class="card shadow-sm mb-3">
      <div class="card-body">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('request.edit.export.data-type.label') }}</label>
          <div class="form-check ms-2 mb-1">
            <input
              id="checkbox-profile"
              v-model="payload.profile"
              type="checkbox"
              class="form-check-input-v3"
              data-test-id="checkbox-profile-information"
            />
            <label class="form-check-label" for="checkbox-profile">{{ t('request.edit.export.data-type.profile-information') }}</label>
          </div>
          <div class="form-check ms-2">
            <input
              id="checkbox-application"
              v-model="payload.application"
              type="checkbox"
              class="form-check-input-v3"
              data-test-id="checkbox-application-data"
            />
            <label class="form-check-label" for="checkbox-application">{{ t('request.edit.export.data-type.application-data') }}</label>
          </div>
        </div>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="text-primary form-label">{{ t('request.edit.export.date-range.label') }}</label>
              <select v-model="payload.range" class="form-select">
                <option value="all">{{ t('request.edit.export.date-range.all') }}</option>
              </select>
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="text-primary form-label">{{ t('request.edit.export.file-format.label') }}</label>
              <select v-model="payload.format" class="form-select" data-test-id="select-file-format">
                <option value="json">{{ t('request.edit.export.file-format.json') }}</option>
                <option value="csv">{{ t('request.edit.export.file-format.csv') }}</option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>

    <Teleport to="#editor-toolbar-target">
      <editor-toolbar
        :processing="processing"
        :back-link="{ name: 'data-overview.application' }"
        submit-show
        :submit-label="t('request.edit.export.submit')"
        :submit-disabled="!(payload.profile || payload.application)"
        @submit="emit('submit', { kind: 'export', payload })"
      />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import EditorToolbar from '../../Common/EditorToolbar.vue'

const { t } = useI18n()
const emit = defineEmits(['submit'])

const processing = ref(false)

const payload = reactive({
  profile: false,
  application: false,
  range: 'all',
  format: 'json',
})
</script>
