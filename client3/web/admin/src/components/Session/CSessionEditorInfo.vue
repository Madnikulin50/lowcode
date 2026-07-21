<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div v-if="session && session.sessionID" class="row g-3 p-3">
      <div v-if="session.workflowID" class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('workflowID') }}</label>
          <p class="form-control-plaintext">{{ session.workflowID }}</p>
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('status') }}</label>
          <p class="form-control-plaintext mb-0">{{ session.status }}</p>
        </div>
      </div>

      <div v-if="session.error" class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('error') }}</label>
          <p class="form-control-plaintext">{{ session.error }}</p>
        </div>
      </div>

      <div v-if="session.resourceType" class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('resourceType') }}</label>
          <p class="form-control-plaintext">{{ session.resourceType }}</p>
        </div>
      </div>

      <div v-if="session.eventType" class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('eventType') }}</label>
          <p class="form-control-plaintext">{{ session.eventType }}</p>
        </div>
      </div>

      <div v-if="!session.completedAt" class="col-12">
        <div class="mb-3">
          <button
            type="button"
            class="btn btn-danger"
            :disabled="processing"
            @click="$emit('cancel')"
          >
            {{ $t('cancel') }}
          </button>
        </div>
      </div>
    </div>

    <c-system-fields
      v-if="session && session.sessionID"
      :id="session.sessionID"
      :resource="session"
    >
      <template #custom-field>
        <div v-if="createdByUserName" class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('createdByUserName') }}</label>
            <p class="form-control-plaintext">{{ createdByUserName }}</p>
          </div>
        </div>
      </template>
    </c-system-fields>

    <div v-if="session && session.workflowID" class="card-footer border-top d-flex flex-wrap gap-1">
      <router-link
        class="btn btn-outline-secondary"
        :to="{ name: 'automation.workflow.edit', params: { workflowID: session.workflowID } }"
      >
        {{ $t('openWorkflow') }}
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  session: { type: Object, required: true },
  user: { type: Object, required: true },
  processing: { type: Boolean, value: false },
})

defineEmits(['cancel'])

const createdByUserName = computed(() => {
  const { userID, name, username, email } = props.user
  return name || username || email || `<@${userID}>`
})
</script>