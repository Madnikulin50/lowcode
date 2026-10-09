<template>
  <div
    class="comment-item"
    :class="{ 'comment-highlighted': highlighted, 'no-hover': disableHover }"
  >
    <div
      v-if="showHeader"
      class="comment-header d-flex align-items-center gap-2 px-2"
    >
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

    <div class="card comment-card rounded position-relative">
      <div class="card-body comment-card-body d-flex rounded">
        <div
          v-if="!isEditing"
          class="comment-toolbox d-flex align-items-center justify-content-end bg-light rounded gap"
        >
          <button
            v-if="reactionsField"
            :id="reactionPickerId"
            title="React"
            class="btn btn-outline-light btn-sm py-1"
          >
            <font-awesome-icon :icon="['far', 'face-smile']" />
          </button>
          <button
            title="Reply"
            class="btn btn-outline-light btn-sm py-1"
            @click.stop="$emit('reply')"
          >
            <font-awesome-icon :icon="['fas', 'reply']" />
          </button>
          <button
            v-if="canEdit"
            title="Edit"
            class="btn btn-outline-light btn-sm py-1"
            @click.stop="onEdit"
          >
            <font-awesome-icon :icon="['fas', 'pen']" />
          </button>
        </div>

        <div
          :title="commentFullDateTime"
          :class="['comment-time', 'text-nowrap', 'text-muted', 'ms-1', 'overflow-hidden', { 'always-visible': showTimeAlways }]"
        >
          <small>{{ commentTime }}</small>
        </div>

        <div class="d-flex flex-column w-100 overflow-hidden gap">
          <comment-reply
            v-if="comment.reply"
            :reply="comment.reply"
            :title-field="titleField"
            :content-field="contentField"
            :namespace="namespace"
            @click.native="$emit('reply-click', comment.reply.recordID)"
          />

          <div
            v-if="isEditing"
            class="d-flex flex-column"
          >
            <input
              v-if="titleField"
              v-model="editValue.title"
              class="form-control mb-1"
              :placeholder="$t('comment.title.placeholder')"
            />
            <c-rich-text-input
              v-if="contentField"
              v-model="editValue.content"
              hide-toolbar
              :placeholder="$t('comment.content.placeholder')"
              :labels="{
                urlPlaceholder: $t('content.urlPlaceholder'),
                ok: $t('content.ok'),
                emojiPicker: {
                  search: $t('content.emojiPicker.search'),
                  searchResults: $t('content.emojiPicker.searchResults'),
                  frequentlyUsed: $t('content.emojiPicker.frequentlyUsed'),
                  noResults: $t('content.emojiPicker.noResults'),
                  quickReactions: $t('content.emojiPicker.quickReactions'),
                },
              }"
              min-body-height="4rem"
              max-body-height="10rem"
              body-class="overflow-auto"
              style="border: none !important;"
            />
            <div class="d-flex justify-content-end gap-1 my-1">
              <button
                class="btn btn-outline-secondary btn-sm"
                @click="onCancel"
              >
                {{ $t('label.cancel') }}
              </button>
              <button
                class="btn btn-primary btn-sm"
                :disabled="isProcessing || !isValid"
                @click="onSave"
              >
                {{ $t('label.save') }}
              </button>
            </div>
          </div>

          <template v-else-if="!isEditing">
            <field-viewer
              v-if="shouldShowTitle"
              :field="titleField"
              :record="comment"
              :namespace="namespace"
              value-only
              class="fw-bold text-muted h5"
            />
            <small
              v-else-if="showTitle && titleField && !titleField.canReadRecordValue"
              class="text-secondary"
            >
              {{ $t('field.noPermission') }}
            </small>

            <field-viewer
              v-if="shouldShowContent"
              :field="contentField"
              :record="comment"
              :namespace="namespace"
              value-only
              class="multiline"
            />
            <small
              v-else-if="showContent && contentField && !contentField.canReadRecordValue"
              class="text-secondary"
            >
              {{ $t('field.noPermission') }}
            </small>

            <field-viewer
              v-if="showAttachments"
              :field="attachmentField"
              :record="comment"
              :namespace="namespace"
              value-only
            />

            <div
              v-if="hasReactions"
              class="comment-reactions d-flex flex-wrap align-items-center gap-1 mt-1"
            >
              <button
                v-for="(userIDs, emoji) in reactions"
                :key="emoji"
                :title="reactionTooltip(emoji, userIDs)"
                type="button"
                class="reaction-badge"
                :class="{ 'reaction-mine': userIDs.includes(currentUserID) }"
                @click.stop="$emit('react', emoji)"
              >
                <span class="reaction-emoji">{{ emoji }}</span>
                <span class="reaction-count">{{ userIDs.length }}</span>
              </button>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed } from 'vue'
import { fmt } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'
import CommentReply from './Reply.vue'
const { CRichTextInput, CEmojiPicker } = components

const emit = defineEmits(['reply', 'edit', 'react', 'reply-click'])

const props = defineProps({
  comment: { type: Object, required: true },
  titleField: { type: Object, default: undefined },
  contentField: { type: Object, default: undefined },
  attachmentField: { type: Object, default: undefined },
  reactionsField: { type: Object, default: undefined },
  emojiData: { type: Array, default: () => [] },
  namespace: { type: Object, required: true },
  showHeader: { type: Boolean, default: true },
  showTimeAlways: { type: Boolean, default: false },
  showTitle: { type: Boolean, default: true },
  showContent: { type: Boolean, default: true },
  highlighted: { type: Boolean, default: false },
  disableHover: { type: Boolean, default: false },
  isProcessing: { type: Boolean, default: false },
  currentUserID: { type: String, default: '' },
  findUserByID: { type: Function, default: () => () => undefined },
})

let reactionPickerCounter = 0
const isEditing = ref(false)
const editValue = ref({ title: '', content: '' })
const reactionPickerId = `reaction-picker-${++reactionPickerCounter}`

const commentTime = computed(() => fmt.time((props.comment || {}).updatedAt || (props.comment || {}).createdAt))
const commentFullDateTime = computed(() => fmt.fullDateTime((props.comment || {}).updatedAt || (props.comment || {}).createdAt))
const authorName = computed(() => ((props.comment || {}).author || {}).name || '')
const authorInitials = computed(() => ((props.comment || {}).author || {}).initials || '')
const authorIsCurrentUser = computed(() => Boolean(((props.comment || {}).author || {}).isCurrentUser))
const canEdit = computed(() => authorIsCurrentUser.value && !props.comment.deletedAt)

const shouldShowTitle = computed(() => {
  if (!props.showTitle || !props.titleField || !props.titleField.canReadRecordValue) return false
  const v = props.comment.values[props.titleField.name]
  return !!v && v.toString().trim().length > 0
})

const shouldShowContent = computed(() => {
  if (!props.showContent || !props.contentField || !props.contentField.canReadRecordValue) return false
  const v = props.comment.values[props.contentField.name]
  return !!v && v.toString().trim().length > 0
})

const showAttachments = computed(() => {
  if (!props.attachmentField || !props.attachmentField.canReadRecordValue) return false
  const v = props.comment.values[props.attachmentField.name]
  if (props.attachmentField.isMulti) return Array.isArray(v) && v.length > 0
  return !!v
})

const isValid = computed(() => !!editValue.value.title || !!editValue.value.content)

const reactions = computed(() => {
  if (!props.reactionsField) return {}
  try {
    const val = props.comment.values[props.reactionsField.name]
    return JSON.parse(val || '{}') || {}
  } catch { return {} }
})

const hasReactions = computed(() => Object.keys(reactions.value).length > 0)

function onEdit () {
  isEditing.value = true
  editValue.value.title = props.titleField ? props.comment.values[props.titleField.name] : ''
  editValue.value.content = props.contentField ? props.comment.values[props.contentField.name] : ''
}

function onCancel () { isEditing.value = false }

function onSave () {
  emit('edit', { title: editValue.value.title, content: editValue.value.content })
  isEditing.value = false
}

function reactionTooltip (emoji, userIDs) {
  const names = userIDs.map(id => {
    if (id === props.currentUserID) return 'You'
    const user = props.findUserByID(id)
    return (user || {}).name || (user || {}).handle || (user || {}).email || 'Unknown'
  })
  return names.join(', ')
}

function onReactionPickerShown () {
  nextTick(() => {
    if (reactionPicker) reactionPicker.reset()
  })
}

function onReactionSelect (emoji) {
  if (emoji && emoji.emoji) emit('react', emoji.emoji)
}
</script>

<style lang="scss" scoped>
.avatar { width: 2.25rem; height: 2.25rem; border-radius: 50%; user-select: none; }

.comment-item {
  .comment-time { display: block; min-width: 3.25rem; opacity: 0; transition: opacity 0.2s ease; }
  .comment-toolbox { position: sticky; top: 0; align-self: flex-start; margin-left: auto; opacity: 0; transition: opacity 0.2s ease; z-index: 1; flex-shrink: 0; order: 3; }
  .comment-card { overflow: visible; transition: background-color 0.2s ease;
    .comment-card-body { padding: 0.2rem 0.25rem; }
    &.comment-highlighted { background-color: var(--light); }
  }
  &.comment-highlighted .comment-card { background-color: var(--light); }
  &:hover .comment-toolbox { opacity: 1; }
  &:not(.no-hover):hover {
    .comment-card { background-color: var(--light); }
    .comment-time { opacity: 1; }
  }
}

.comment-reactions {
  .reaction-badge {
    display: inline-flex; align-items: center; gap: 0.2rem; padding: 0.1rem 0.4rem;
    border: 1px solid var(--extra-light, #e0e0e0); border-radius: 1rem;
    background: var(--white, #fff); cursor: pointer; font-size: 0.8rem; line-height: 1.4;
    transition: background-color 0.15s, border-color 0.15s;
    &:hover { background-color: var(--light, #f5f5f5); border-color: var(--secondary, #ccc); }
    &.reaction-mine { background-color: rgba(var(--primary-rgb, 64, 128, 255), 0.08); border-color: var(--primary, #4080ff); }
    .reaction-emoji { font-size: 0.9rem; }
    .reaction-count { font-size: 0.75rem; font-weight: 600; color: var(--secondary, #888); }
    &.reaction-mine .reaction-count { color: var(--primary, #4080ff); }
  }
}
</style>

<style lang="scss">
.reaction-emoji-popover { background: var(--white, #fff);
  .popover-body { padding: 0; }
  .arrow::after { border-bottom-color: var(--white, #fff); }
}
</style>
