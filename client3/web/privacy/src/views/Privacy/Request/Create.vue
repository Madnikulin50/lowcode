<template>
  <div class="container-fluid d-flex flex-column p-3">
    <request-editor
      v-if="kind"
      :kind="kind"
      @submit="onSubmit"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import RequestEditor from '../../../components/Requests/Editor/index.vue'

const props = defineProps({
  kind: { type: String, required: true },
})

const { t } = useI18n()
const router = useRouter()

const processing = ref(false)

function onSubmit ({ kind, payload }) {
  processing.value = true
  payload = [payload]

  return window.__systemAPI.dataPrivacyRequestCreate({ kind, payload })
    .then(({ requestID, kind } = {}) => {
      router.push({ name: 'request.view', params: { requestID, kind } })
    })
    .catch((err) => {
      window.__toastErrorHandler(t('notification.list.load.error'))(err)
    })
    .finally(() => {
      processing.value = false
    })
}
</script>