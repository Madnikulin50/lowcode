<template>
  <div class="card shadow-sm" data-test-id="card-route-edit">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit', route)">
        <div class="row g-3 p-3">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary d-flex align-items-center" for="endpoint">
                {{ $t('system.apigw.editor.info.endpoint') }}
                <font-awesome-icon id="endpoint_info" class="ms-1" :icon="['far', 'question-circle']" />
              </label>
              <small v-if="routeEndpointDescription" class="text-danger d-block">{{ routeEndpointDescription }}</small>
              <input
                id="endpoint"
                v-model="route.endpoint"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': isValidEndpoint === false }"
                data-test-id="input-endpoint"
                required
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.apigw.editor.info.method') }}</label>
              <select
                v-model="route.method"
                class="form-select"
                data-test-id="select-method"
                required
              >
                <option v-for="m in methods" :key="m" :value="m">{{ m }}</option>
              </select>
            </div>
          </div>
  
          <div v-if="route.meta" class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('description') }}</label>
              <textarea v-model="route.meta.description" class="form-control" rows="3" />
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3" :class="{ 'mb-0': !route.routeID }">
              <label class="form-label text-primary">{{ $t('enabled') }}</label>
              <c-input-checkbox
                v-model="route.enabled"
                switch
                :labels="checkboxLabel"
                data-test-id="checkbox-enabled"
              />
            </div>
          </div>
        </div>
  
        <c-system-fields :id="route.routeID" :resource="route" />
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="route && route.routeID && route.canDeleteApigwRoute"
        :data-test-id="deletedButtonStatusCypressId"
        variant="danger"
        :text="getDeleteStatus"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', route)"
      />
    </div>
  </div>
</template>

<script setup>
import { useNsI18n } from 'corteza-lib/vue/dist'
defineOptions({ i18nOptions: { namespaces: ['system.apigw'], keyPrefix: 'editor.info' } })
import { computed } from 'vue'

import { NoID } from 'corteza-lib/js/dist'

const t = useNsI18n()
const props = defineProps({
  route: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit', 'delete'])
const methods = ['GET', 'POST', 'PUT', 'DELETE']
const checkboxLabel = { on: t('label.general.yes'), off: t('label.general.no') }

const fresh = computed(() => !props.route.routeID || props.route.routeID === NoID)
const editable = computed(() => fresh.value ? props.canCreate : true)
const saveDisabled = computed(() => !editable.value || isValidEndpoint.value === false)
const getDeleteStatus = computed(() => props.route.deletedAt ? t('undelete') : t('delete'))
const isValidEndpoint = computed(() => {
  const { endpoint } = props.route
  return (!!endpoint && /^(\/[\w-]+)+$/.test(endpoint)) ? null : false
})
const startsWithSlash = computed(() => props.route.endpoint ? /^\//.test(props.route.endpoint) : null)
const routeEndpointDescription = computed(() => {
  if (isValidEndpoint.value === false) {
    if (!startsWithSlash.value) return t('validation.slash')
    if (props.route.endpoint.length < 2) return t('validation.minLength')
    if (!/^([\w-]+)+$/.test(props.route.endpoint)) return t('validation.invalidCharacters')
  }
  return ''
})
const deletedButtonStatusCypressId = computed(() => `button-${getDeleteStatus.value.toLowerCase()}`)
</script>