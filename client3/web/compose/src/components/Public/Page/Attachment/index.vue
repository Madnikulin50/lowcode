<template>
  <div>
    <div v-if="mode === 'list'">
      <a :href="attachment.download">
        <font-awesome-icon :icon="['fas', 'download']" />
        {{ attachment.name }}
      </a>
      <i18next
        path="label.attachmentFileInfo"
        tag="label"
      >
        <span>{{ attachment.size }}</span>
        <span>{{ attachment.changedAt }}</span>
      </i18next>
    </div>

    <div v-if="mode === 'grid'">
      <a :href="attachment.download">
        <font-awesome-icon
          :icon="['far', 'file-'+ext(attachment)]"
        />
        {{ attachment.name }}
      </a>
      <i18next
        path="label.attachmentFileInfo"
        tag="label"
      >
        <span>{{ attachment.size }}</span>
        <span>{{ attachment.changedAt }}</span>
      </i18next>
    </div>

    <div
      v-else
      class="single"
    >
      <div v-if="isImage(attachment)">
        <img
          :src="attachment.previewUrl"
          @click="openLightbox(index)"
        >
      </div>
      <div v-else>
        <font-awesome-icon
          :icon="['far', 'file-'+ext(attachment)]"
        />
        <a :href="attachment.download">
          {{ $t('label.download') }}
        </a>
      </div>
      {{ attachment.name }}
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  kind: { type: String, required: true },
  mode: { type: String, required: true },
  value: { type: [Object, String], required: true },
})

const emit = defineEmits(['input'])

const attachment = ref({})

watch(() => props.value, (value) => {
  if (typeof value === 'string') {
    $ComposeAPI.attachmentRead({ kind: props.kind, attachmentID: value }).then(a => {
      attachment.value = a
    })
  } else {
    attachment.value = value
  }
}, { immediate: true })
</script>
