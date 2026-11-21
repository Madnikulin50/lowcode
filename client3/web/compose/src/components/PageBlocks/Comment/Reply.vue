<template>
  <div class="comment-reply d-flex flex-column gap-1 overflow-hidden border rounded p-2 bg-white pointer">
    <div class="d-flex align-items-center gap-1">
      <div
        :title="authorName"
        class="avatar d-flex align-items-center justify-content-center bg-light fw-bold"
      >
        {{ authorInitials }}
      </div>
      <span
        :class="authorIsCurrentUser ? 'text-primary' : 'text-muted'"
        class="text-nowrap fw-bold text-truncate"
      >
        {{ authorName }}
      </span>
    </div>
    <div class="d-flex flex-column gap-2 overflow-hidden">
      <field-viewer
        v-if="titleField"
        :field="titleField"
        :record="reply"
        :namespace="namespace"
        value-only
        class="fw-bold text-muted h5"
      />
      <field-viewer
        v-if="contentField"
        :field="contentField"
        :record="reply"
        :namespace="namespace"
        value-only
        class="reply-content text-muted text-truncate w-100"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'

const props = defineProps({
  reply: { type: Object, required: true },
  titleField: { type: Object, default: undefined },
  contentField: { type: Object, default: undefined },
  namespace: { type: Object, required: true },
})

const authorName = computed(() => ((props.reply || {}).author || {}).name || '')
const authorInitials = computed(() => ((props.reply || {}).author || {}).initials || '')
const authorIsCurrentUser = computed(() => Boolean(((props.reply || {}).author || {}).isCurrentUser))
</script>

<style lang="scss" scoped>
.avatar {
  width: 2rem;
  height: 2rem;
  font-size: 0.8rem;
  border-radius: 50%;
  user-select: none;
}

.comment-reply {
  max-height: 6.5rem;
  border-left-color: var(--primary) !important;
  border-left-width: 3px !important;

  .reply-content { transition: color 0.2s ease; }

  &:hover {
    .reply-content { color: var(--dark) !important; }
  }
}
</style>
