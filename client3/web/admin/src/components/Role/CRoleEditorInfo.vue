<template>
  <div class="card shadow-sm" data-test-id="card-role-info">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <form @submit.prevent="submit()">
      <div class="row g-3 p-3">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('name') }}</label>
            <input
              v-model="role.name"
              type="text"
              class="form-control"
              :class="{ 'is-invalid': nameState === false }"
              data-test-id="input-name"
              :disabled="!editable"
            >
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('handle') }}</label>
            <input
              v-model="role.handle"
              type="text"
              class="form-control"
              :class="{ 'is-invalid': handleState === false }"
              data-test-id="input-handle"
              :disabled="!editable"
              :placeholder="$t('placeholder-handle')"
            >
            <div v-if="handleState === false" class="invalid-feedback">{{ $t('invalid-handle-characters') }}</div>
          </div>
        </div>

        <div class="col-12">
          <div v-if="role.meta" class="mb-3">
            <label class="form-label text-primary">{{ $t('description') }}</label>
            <textarea
              v-model="role.meta.description"
              class="form-control"
              data-test-id="textarea-description"
              :disabled="!editable"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('context.label') }}</label>
            <c-input-checkbox
              v-model="isContextual"
              data-test-id="checkbox-is-contextual"
              switch
              :labels="checkboxLabel"
              :disabled="!editable"
            />
          </div>
        </div>
      </div>

      <div v-if="isContextual" class="row g-3 my-3 px-3">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('context.expression-label') }}</label>
            <input
              v-model="role.meta.context.expr"
              type="text"
              class="form-control"
              data-test-id="input-expression"
              :disabled="!editable"
            >
          </div>
        </div>

        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('context.resource-types-label') }}</label>
            <div v-for="(resourceType, i) in resourceTypeOptions" :key="i" class="form-check">
              <input
                :id="'resource-type-' + i"
                v-model="role.meta.context.resourceTypes"
                type="checkbox"
                class="form-check-input"
                :data-test-id="`checkbox-resource-type-${resourceType.text}`"
                :value="resourceType.value"
                :disabled="!editable"
              >
              <label :for="'resource-type-' + i" class="form-check-label">{{ resourceType.text }}</label>
            </div>
          </div>
        </div>
      </div>

      <c-system-fields :resource="role" />

      <input
        type="submit"
        class="d-none"
        :disabled="saveDisabled"
      >
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="!fresh && editable && role.canDeleteRole && !isDataPrivacyOfficer"
        :data-test-id="deletedButtonStatusCypressId"
        :text="getDeleteStatus"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-input-confirm
        v-if="!fresh && editable && !isDataPrivacyOfficer"
        :data-test-id="archivedButtonStatusCypressId"
        :text="getArchiveStatus"
        variant="secondary"
        size="md"
        @confirmed="$emit('status')"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="submit()"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { system, NoID } from 'corteza-lib/js/dist'
import { handle } from 'corteza-lib/vue/dist'

const { t } = useI18n()
const props = defineProps({
  role: { type: system.Role, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  isContext: { type: Boolean, required: true },
  canCreate: { type: Boolean, required: true },
})

const emit = defineEmits(['submit', 'delete', 'status', 'update:is-context'])

const checkboxLabel = { on: t('label.general.yes'), off: t('label.general.no') }

const isContextual = computed({
  get() { return props.isContext },
  set(isContext) {
    props.role.meta.context.resourceTypes = []
    props.role.meta.context.expr = ''
    emit('update.is-context', isContext)
  },
})

const fresh = computed(() => !props.role.roleID || props.role.roleID === NoID)
const editable = computed(() => fresh.value ? props.canCreate : !props.role.isSystem && props.role.canUpdateRole)
const saveDisabled = computed(() => !editable.value || [nameState.value, handleState.value].includes(false))
const nameState = computed(() => props.role.name ? null : false)
const handleState = computed(() => handle.handleState(props.role.handle))
const getDeleteStatus = computed(() => props.role.deletedAt ? t('undelete') : t('delete'))
const getArchiveStatus = computed(() => props.role.archivedAt ? t('unarchive') : t('archive'))

const resourceTypes = [
  'corteza::system:auth-client',
  'corteza::system:role',
  'corteza::system:user',
  'corteza::compose:module',
  'corteza::compose:namespace',
  'corteza::compose:page',
  'corteza::compose:record',
  'corteza::automation:workflow',
]

const resourceTypeOptions = computed(() => resourceTypes.map(value => ({
  text: value.replace('corteza::', ''),
  value,
})))

const deletedButtonStatusCypressId = computed(() => `button-${getDeleteStatus.value.toLowerCase()}`)
const archivedButtonStatusCypressId = computed(() => `button-${getArchiveStatus.value.toLowerCase()}`)
const isDataPrivacyOfficer = computed(() => props.role.handle === 'data-privacy-officer')

function submit() {
  if (!isContextual.value && props.role.isContext) {
    props.role.meta.context.resourceTypes = []
    props.role.meta.context.expr = ''
  }
  emit('submit', props.role)
}
</script>