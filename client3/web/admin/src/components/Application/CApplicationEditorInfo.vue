<template>
  <div class="card shadow-sm" data-test-id="card-application-info">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ $t('title') }}
      </h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit', application)">
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('name') }}</label>
              <input
                v-model="application.name"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': nameState === false }"
                data-test-id="input-name"
                required
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" :class="{ 'mb-0': !application.applicationID }">
              <label class="form-label text-primary">{{ $t('enabled') }}</label>
              <c-input-checkbox
                v-model="application.enabled"
                data-test-id="checkbox-enabled"
                :labels="checkboxLabel"
                switch
              />
            </div>
          </div>
        </div>
  
        <c-system-fields
          :id="application.applicationID"
          :resource="application"
        />
  
        <input
          type="submit"
          class="d-none"
          :disabled="saveDisabled"
        >
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="application && application.applicationID && application.canDeleteApplication"
        :data-test-id="deleteButtonStatusCypressId"
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
        @submit="$emit('submit', application)"
      />
    </div>
  </div>
</template>

<script setup>
import { useNsI18n } from 'corteza-lib/vue/dist'
defineOptions({ i18nOptions: { namespaces: 'system.applications', keyPrefix: 'editor.info' } })
import { computed } from 'vue'

import { NoID } from 'corteza-lib/js/dist'

const t = useNsI18n()

const props = defineProps({
  application: {
    type: Object,
    required: true,
  },
  processing: {
    type: Boolean,
    value: false,
  },
  success: {
    type: Boolean,
    value: false,
  },
  canCreate: {
    type: Boolean,
    required: true,
  },
})

defineEmits(['submit', 'delete'])

const checkboxLabel = {
  on: t('label.general.yes'),
  off: t('label.general.no'),
}

const fresh = computed(() => {
  return !props.application.applicationID || props.application.applicationID === NoID
})

const editable = computed(() => {
  return fresh.value ? props.canCreate : props.application.canUpdateApplication
})

const saveDisabled = computed(() => {
  return !editable.value || nameState.value === false
})

const nameState = computed(() => {
  const { name } = props.application
  return name ? null : false
})

const getDeleteStatus = computed(() => {
  return props.application.deletedAt ? t('undelete') : t('delete')
})

const deleteButtonStatusCypressId = computed(() => {
  return `button-${getDeleteStatus.value.toLowerCase()}`
})
</script>