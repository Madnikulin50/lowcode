<template>
  <div class="card">
    <div class="card-body">
      <c-progress
        :value="localProgress.completed"
        :max="localProgress.entryCount"
        labeled
        progress
        :animated="!localProgress.finishedAt"
        :relative="false"
        variant="success"
        text-style="font-size: 1.5rem;"
        style="height: 4rem;"
        class="mb-4"
      />

      <div v-if="!localProgress.finishedAt" class="d-flex">
        <span class="text-secondary">
          <span class="spinner-border spinner-border-sm text-secondary" />
          {{ $t('recordList.import.importing') }}
        </span>
        <button
          class="btn btn-outline-secondary ms-auto"
          @click="$emit('close')"
        >
          {{ $t('label.cancel') }}
        </button>
      </div>

      <div v-if="localProgress.finishedAt && !localProgress.failed" class="d-flex">
        <span class="text-success">
          {{ $t('recordList.import.success') }}
        </span>
        <button
          class="btn btn-outline-secondary ms-auto"
          @click="$emit('close')"
        >
          {{ $t('label.close') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CProgress } = components
const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  session: { type: Object, required: true, default: () => ({}) },
  noPool: { type: Boolean, default: false },
})

const emit = defineEmits(['importSuccessful', 'importFailed', 'close'])

let toHandle = null

const localProgress = ref(props.session.progress || {})

watch(localProgress, ({ finishedAt, failed }) => {
  if (finishedAt && failed) {
    clearTimer()
    emit('importFailed', localProgress.value)
  } else if (finishedAt) {
    clearTimer()
    window.dispatchEvent(new CustomEvent('recordList.refresh', { detail: props.session }))
    emit('importSuccessful')
  }
})

function clearTimer() {
  if (toHandle !== null) {
    window.clearTimeout(toHandle)
    toHandle = null
  }
}

function pool() {
  $ComposeAPI.recordImportProgress(props.session)
    .then(({ progress }) => {
      localProgress.value = progress
      toHandle = window.setTimeout(pool, 2000)
    })
}

onMounted(() => {
  if (!props.noPool) {
    pool()
  }
})

onBeforeUnmount(() => {
  clearTimer()
})
</script>

<style lang="scss" scoped>
.progress-label {
  font-size: 15px;
}
</style>
