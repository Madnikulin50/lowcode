<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h5 class="d-flex align-items-center justify-content-between mb-0">
        {{ t('request.comments.label') }}

        <div class="dropdown">
          <button
            class="btn btn-link text-muted text-decoration-none dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
          >
            {{ sort.includes('DESC') ? t('request.comments.sort.first.newest') : t('request.comments.sort.first.oldest') }}
          </button>
          <ul class="dropdown-menu">
            <li>
              <button
                class="dropdown-item"
                :disabled="sort.includes('DESC')"
                @click="emit('sort', 'createdAt DESC')"
              >
                {{ t('request.comments.sort.first.newest') }}
              </button>
            </li>
            <li>
              <button
                class="dropdown-item"
                :disabled="!sort.includes('DESC')"
                @click="emit('sort', 'createdAt')"
              >
                {{ t('request.comments.sort.first.oldest') }}
              </button>
            </li>
          </ul>
        </div>
      </h5>
    </div>

    <div class="card-body">
      <div class="d-flex flex-column">
        <textarea
          id="textarea"
          v-model="comment"
          class="form-control mb-2"
          :placeholder="t('request.comments.enter')"
          rows="2"
        />
        <button
          class="btn btn-primary ms-auto"
          :disabled="!comment"
          @click="submitComment()"
        >
          {{ t('request.comments.submit') }}
        </button>
      </div>

      <hr v-if="comments.length || processing">

      <div v-if="processing" class="d-flex align-items-center justify-content-center py-3">
        <div class="spinner-border" role="status">
          <span class="visually-hidden">{{ t('resourceList.loading') }}</span>
        </div>
      </div>

      <template v-else>
        <div
          v-for="(c, index) in comments"
          :key="c.commentID"
          :class="{ 'mt-3': index }"
          class="overflow-auto"
        >
          <div class="d-flex align-items-center flex-wrap border p-2">
            <h6 class="text-primary mb-0">
              <span v-if="formatting" class="spinner-border spinner-border-sm" role="status">
                <span class="visually-hidden">{{ t('resourceList.loading') }}</span>
              </span>
              <span v-else>
                {{ formattedUsers[c.createdBy] || t('request.comments.unknown.user') }}
              </span>
            </h6>
            <span class="ms-auto text-muted">{{ formatDate(c.createdAt) }}</span>
          </div>
          <div class="border p-3">
            <p class="mb-0 multiline">{{ c.comment }}</p>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'request', keyPrefix: 'comments' } })
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { fmt, NoID } from 'corteza-lib/js/dist'

const props = defineProps({
  comments: { type: Array, required: true },
  processing: { type: Boolean, required: true },
  sort: { type: String, required: true },
})

const emit = defineEmits(['sort', 'submit'])

const { t } = useI18n()

const comment = ref('')
const formatting = ref(false)
const formattedUsers = ref({})

watch(() => props.comments, {
  immediate: true,
  handler (comments) {
    formatUsers(comments)
  },
})

function submitComment () {
  emit('submit', comment.value)
  comment.value = ''
}

function formatDate (date) {
  return date ? fmt.fullDateTime(date.toLocaleString()) : t('request.comments.unknown.date')
}

function formatUsers (comments = []) {
  const userID = []

  comments.forEach(({ createdBy }) => {
    if (createdBy !== NoID && !formattedUsers.value[createdBy]) {
      userID.push(createdBy)
    }
  })

  if (userID.length) {
    formatting.value = true
    window.__systemAPI.userList({ userID, suspended: 1, deleted: 1 })
      .then(({ set }) => {
        set.forEach(({ userID, name, username, email, handle }) => {
          formattedUsers.value[userID] = name || username || email || handle || userID || ''
        })
      })
      .finally(() => { formatting.value = false })
  }
}
</script>

<style lang="scss" scoped>
.multiline {
  white-space: pre-line;
}
</style>