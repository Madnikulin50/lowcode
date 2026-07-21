<template>
  <div v-if="!processing" class="card shadow-sm">
    <div class="card-header border-bottom">
      <Teleport to="#topbar-title-target">{{ t(`request:kind.${request.kind}`) }}</Teleport>

      <h5 class="d-flex align-items-center justify-content-between">
        <span>{{ formattedDate }}</span>
        <span
          :data-test-id="`badge-${request.status}`"
          class="badge rounded-pill px-2 py-1"
          :class="`bg-${statusVariants[request.status]}`"
        >
          {{ t(`request:status.${request.status}`) }}
        </span>
      </h5>

      <p data-test-id="request-author" class="text-primary">
        {{ formattedUsers[request.requestedBy] || 'Unknown user' }}
      </p>
    </div>

    <div class="card-body p-2">
      <component :is="request.kind" :request="request" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fmt, NoID } from 'corteza-lib/js/dist'
import Correct from './Correct.vue'
import Delete from './Delete.vue'
import Export from './Export.vue'

const props = defineProps({
  request: { type: Object, required: true },
})

const { t } = useI18n()

const processing = ref(false)
const formattedUsers = ref({})

const statusVariants = {
  canceled: 'secondary',
  pending: 'warning',
  rejected: 'danger',
  approved: 'success',
}

const formattedDate = computed(() => {
  return props.request ? fmt.fullDateTime(props.request.requestedAt.toLocaleString()) : 'Unknown date'
})

const payload = computed(() => {
  const { payload = [] } = props.request || {}
  return payload[0]
})

watch(() => props.request.requestedBy, {
  immediate: true,
  handler (userID) {
    if (userID === window.__auth.user.userID) {
      const { name, username, email, handle } = window.__auth.user
      formattedUsers.value[userID] = name || username || email || handle || userID || ''
      return
    }

    if (userID !== NoID && !formattedUsers.value[userID]) {
      processing.value = true
      window.__systemAPI.userRead({ userID })
        .then(({ name, username, email, handle }) => {
          formattedUsers.value[userID] = name || username || email || handle || userID || ''
        })
        .finally(() => { processing.value = false })
    }
  },
})
</script>