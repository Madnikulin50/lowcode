<template>
  <div class="container-fluid d-flex flex-column p-3">
    <div v-if="processing.request" class="d-flex align-items-center justify-content-center h-100">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">{{ t('resourceList.loading') }}</span>
      </div>
    </div>

    <template v-else-if="request">
      <request-viewer
        :request="request"
        class="mb-3"
      />

      <request-comments
        :comments="comments"
        :processing="processing.comments"
        :sort="sort"
        @sort="sort = $event"
        @submit="submitComment"
      />
    </template>

    <Teleport to="#editor-toolbar-target">
      <editor-toolbar
        :processing="processing.toolbar"
        :processing-confirm="processingReject"
        :back-link="{ name: 'request.list' }"
        :delete-show="isDC"
        :delete-disabled="!request || !isPending"
        :delete-label="t('request.view.reject')"
        @delete="handleRequest('rejected')"
      >
        <c-input-confirm
          v-if="!isDC"
          :disabled="!request || !isPending"
          :processing="processingCancel"
          :text="t('request.view.cancel')"
          variant="outline-secondary"
          size="lg"
          size-confirm="lg"
          @confirmed="handleRequest('canceled')"
        />

        <c-input-confirm
          v-else
          :disabled="!request || !isPending"
          :processing="processingApprove"
          :text="t('request.view.approve')"
          variant="primary"
          variant-ok="primary"
          size="lg"
          size-confirm="lg"
          class="ms-2"
          @confirmed="handleRequest('approved')"
        />
      </editor-toolbar>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import EditorToolbar from '../../../components/Common/EditorToolbar.vue'
import RequestViewer from '../../../components/Requests/Viewer/index.vue'
import RequestComments from '../../../components/Requests/Comments.vue'

const props = defineProps({
  requestID: { type: String, required: false, default: '' },
})

const { t } = useI18n()
const router = useRouter()

const processing = reactive({
  comments: false,
  request: false,
  toolbar: false,
})
const processingApprove = ref(false)
const processingReject = ref(false)
const processingCancel = ref(false)
const isDC = ref(null)
const sort = ref('createdAt DESC')
const request = ref(undefined)
const comments = ref([])

const isPending = computed(() => request.value && request.value.status === 'pending')

watch(() => props.requestID, {
  immediate: true,
  handler () {
    if (props.requestID) {
      fetchRequest()
      fetchComments()
    }
  },
})

watch(sort, () => {
  if (props.requestID) {
    fetchComments()
  }
})

function checkIsDC () {
  window.__systemAPI.roleList({ query: 'data-privacy-officer', memberID: window.__auth.user.userID })
    .then(({ set = [] }) => {
      isDC.value = !!set.length
    })
}

function fetchRequest (requestID = props.requestID) {
  processing.request = true
  return window.__systemAPI.dataPrivacyRequestRead({ requestID })
    .then(r => { request.value = r })
    .catch((err) => { window.__toastErrorHandler(t('notification.list.load.error'))(err) })
    .finally(() => { processing.request = false })
}

function fetchComments (requestID = props.requestID) {
  processing.comments = true
  return window.__systemAPI.dataPrivacyRequestCommentList({ requestID, sort: sort.value })
    .then(({ set }) => { comments.value = set })
    .catch((err) => { window.__toastErrorHandler(t('notification.list.load.error'))(err) })
    .finally(() => { processing.comments = false })
}

function handleRequest (status) {
  processing.toolbar = true
  if (status === 'approved') {
    processingApprove.value = true
  } else if (status === 'rejected') {
    processingReject.value = true
  } else {
    processingCancel.value = true
  }

  window.__systemAPI.dataPrivacyRequestUpdateStatus({ requestID: props.requestID, status })
    .then(() => { router.push({ name: 'request.list' }) })
    .finally(() => {
      processing.toolbar = false
      if (status === 'approved') {
        processingApprove.value = false
      } else if (status === 'rejected') {
        processingReject.value = false
      } else {
        processingCancel.value = false
      }
    })
}

function submitComment (comment) {
  processing.comments = true
  window.__systemAPI.dataPrivacyRequestCommentCreate({ requestID: props.requestID, comment })
    .then(() => fetchComments())
    .finally(() => { processing.comments = false })
}

onMounted(() => checkIsDC())
</script>