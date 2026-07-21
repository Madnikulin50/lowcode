<template>
  <div class="card shadow-sm" data-test-id="card-user-group-info">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <form @submit.prevent="$emit('submit', userGroup)">
      <div class="row g-3 p-3">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('meta.short') }}</label>
            <input
              v-model="userGroup.meta.short"
              data-test-id="input-name"
              required
              class="form-control"
            >
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3" :class="{ 'mb-0': !userGroup.userGroupID }">
            <label class="form-label text-primary">{{ $t('handle') }}</label>
            <input
              v-model="userGroup.handle"
              data-test-id="input-handle"
              :placeholder="$t('placeholder-handle')"
              :class="['form-control', { 'is-invalid': handleState === false }]"
            >
            <div v-if="handleState === false" class="invalid-feedback">
              {{ $t('invalid-handle-characters') }}
            </div>
          </div>
        </div>
      </div>

      <div class="row g-3 px-3 pb-3">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('meta.description') }}</label>
            <textarea
              v-model="userGroup.meta.description"
              data-test-id="textarea-description"
              class="form-control"
            ></textarea>
          </div>
        </div>
      </div>

      <c-system-fields
        :id="userGroup.userGroupID"
        :resource="userGroup"
      />

      <template v-if="!isRoot">
        <hr class="mx-3">

        <div v-if="!isRoot" class="px-3 pb-3">
          <h5 class="mb-3">{{ $t('parents.title') }}</h5>

          <c-form-table-wrapper
            :labels="{
              addButton: $t('label.add')
            }"
            class="my-3"
            @add-item="addParent"
          >
            <table
              v-if="userGroup.config.path"
              class="table table-sm table-borderless"
            >
              <thead>
                <tr>
                  <th class="text-primary">{{ $t('parents.parent.label') }}</th>
                  <th class="text-primary">{{ $t('parents.name.label') }}</th>
                  <th v-if="userGroup.config.path.length > 1"></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(parent, i) in userGroup.config.path" :key="i">
                  <td class="align-middle" style="min-width: 250px;">
                    <c-input-user-group
                      v-model="parent.selfID"
                      :placeholder="$t('parents.parent.placeholder')"
                    />
                  </td>
                  <td class="align-middle" style="min-width: 200px;">
                    <input
                      v-model="parent.name"
                      :placeholder="$t('parents.name.placeholder')"
                      class="form-control"
                    >
                  </td>
                  <td v-if="userGroup.config.path.length > 1" class="text-end align-middle" style="width: 1%">
                    <c-input-confirm
                      show-icon
                      @confirmed="deleteParent(i)"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </c-form-table-wrapper>
        </div>
      </template>

      <input
        type="submit"
        class="d-none"
        :disabled="saveDisabled"
      >
    </form>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="!fresh && userGroup.canDeleteUserGroup"
        :data-test-id="deletedButtonStatusCypressId"
        :text="getDeleteStatus"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-corredor-manual-buttons
        ui-page="user-group/editor"
        ui-slot="infoFooter"
        resource-type="system:user-group"
        default-variant="outline-secondary"
        @click="dispatchCortezaSystemUserGroupEvent($event, { userGroup })"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', userGroup)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { handle, components } from 'corteza-lib/vue/dist'
const { CInputUserGroup } = components

const { t } = useI18n()

const props = defineProps({
  userGroup: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit', 'delete'])

const isRoot = computed(() => props.userGroup.isRoot)
const getDeleteStatus = computed(() => props.userGroup.deletedAt ? t('undelete') : t('delete'))
const userGroupID = computed(() => props.userGroup ? props.userGroup.userGroupID : undefined)
const fresh = computed(() => !userGroupID.value || userGroupID.value === NoID)
const editable = computed(() => fresh.value ? props.canCreate : props.userGroup.canUpdateUserGroup)
const nameState = computed(() => props.userGroup.meta.short ? null : false)
const handleState = computed(() => handle.handleState(props.userGroup.handle))
const parentState = computed(() => {
  if (isRoot.value) return null
  if (!props.userGroup.config.path || props.userGroup.config.path.length === 0) return false
  return props.userGroup.config.path.every(parent => parent.selfID) ? null : false
})
const saveDisabled = computed(() => !editable.value || [nameState.value, handleState.value, parentState.value].includes(false))
const deletedButtonStatusCypressId = computed(() => `button-${getDeleteStatus.value.toLowerCase()}`)

watch(userGroupID, () => {
  const { config = {} } = props.userGroup || {}
  const { path = [] } = config || {}
  if (!isRoot.value && path.length === 0) {
    addParent()
  }
}, { immediate: true })

function addParent() {
  props.userGroup.config.path.push({
    selfID: '',
    name: '',
  })
}

function deleteParent(i) {
  props.userGroup.config.path.splice(i, 1)
}

function dispatchCortezaSystemUserGroupEvent($event, { userGroup }) {}
</script>
