<template>
  <div>
    <div class="mb-3">
      <div class="form-check">
        <input
          id="incl-role-membership"
          v-model="inclRoleMembership"
          class="form-check-input-v3"
          data-test-id="checkbox-include-role-membership"
          type="checkbox"
          :true-value="true"
          :false-value="false"
        >
        <label
          class="form-check-label"
          for="incl-role-membership"
        >{{ t('export.inclRoleMembership') }}</label>
      </div>
    </div>

    <div class="mb-0">
      <div
        class="form-text mb-2"
        v-text="!inclRoleMembership ? t('export.membershipRequiredLabel') : ''"
      />
      <div class="form-check">
        <input
          id="incl-roles"
          v-model="inclRoles"
          class="form-check-input-v3"
          data-test-id="checkbox-include-roles"
          type="checkbox"
          :true-value="true"
          :false-value="false"
          :disabled="!inclRoleMembership"
        >
        <label
          class="form-check-label"
          for="incl-roles"
        >{{ t('export.inclRoles') }}</label>
      </div>
    </div>

    <div class="d-flex justify-content-end mt-3">
      <button
        data-test-id="button-export"
        class="btn btn-primary"
        @click="nextStep"
      >
        {{ t('export.export') }}
      </button>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.users' } })
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const emit = defineEmits(['configured'])

const inclRoleMembership = ref(false)
const inclRoles = ref(false)

function nextStep () {
  const rtr = {
    inclRoleMembership: inclRoleMembership.value,
    inclRoles: inclRoles.value,
  }

  emit('configured', rtr)
}
</script>
