<template>
  <Wrap
    v-bind="$props"
    :scrollable-body="false"
    @refreshBlock="refresh"
  >
    <div
      v-if="!isBlockConfigured"
      class="d-flex h-100 align-items-center justify-content-center"
    >
      <p class="mb-0 my-3">
        {{ $t('noConfiguration') }}
      </p>
    </div>

    <template v-else-if="roModule">
      <div class="d-flex flex-column h-100">
        <div
          v-if="isProcessing"
          class="d-flex align-items-center justify-content-center h-100"
        >
          <span class="spinner-border" />
        </div>

        <section
          v-else-if="comments.length"
          ref="chatContainer"
          class="flex-grow-1 py-2 px-1 overflow-auto"
        >
          <div
            v-if="showNewestFirst && hasNextPage"
            class="text-center"
          >
            <c-button-submit
              :text="$t('comment.load.older')"
              :processing="loadingMore"
              variant="extra-light"
              class="mb-1"
              @submit="loadMoreMessages"
            />
          </div>

          <div
            v-for="dateGroup in comments"
            :key="dateGroup.date"
            class="date-group d-flex flex-column gap-2 mt-2"
          >
            <div
              v-if="comments.length > 1"
              class="d-flex align-items-center justify-content-center gap-3 mx-2 text-muted"
            >
              <hr class="flex-grow-1 m-0" />
              <span>{{ dateGroup.date }}</span>
              <hr class="flex-grow-1 m-0" />
            </div>

            <div
              v-for="(messageGroup, index) in dateGroup.messages"
              :key="index"
              class="message-group"
            >
              <comment-item
                v-for="(comment, ci) in messageGroup.comments"
                :id="`comment-${comment.recordID}`"
                :key="comment.recordID"
                :comment="comment"
                :title-field="titleField"
                :content-field="contentField"
                :attachment-field="attachmentField"
                :reactions-field="reactionsField"
                :emoji-data="emojiData"
                :namespace="namespace"
                :show-header="ci === 0"
                :show-title="showTitle(comment)"
                :show-content="showContent(comment)"
                :highlighted="highlightedCommentId === comment.recordID"
                :current-user-i-d="currentUserID"
                :find-user-by-i-d="findUserByID"
                class="mb-1"
                @reply="replyToComment(comment)"
                @edit="onEditComment(comment, $event)"
                @react="onReact(comment, $event)"
                @reply-click="handleReplyClick"
                @mouseleave="resetHighlightedComment(comment.recordID)"
              />
            </div>
          </div>

          <div
            v-if="!showNewestFirst && hasNextPage"
            class="text-center"
          >
            <c-button-submit
              :text="$t('comment.load.newer')"
              :processing="loadingMore"
              variant="extra-light"
              class="mt-1"
              @submit="loadMoreMessages"
            />
          </div>
        </section>

        <div
          v-else
          class="d-flex align-items-center justify-content-center h-100"
        >
          <p class="mb-0 my-3">
            {{ $t('comment.noComments') }}
          </p>
        </div>

        <section
          v-if="canAddRecord"
          class="d-flex flex-column bg-white border-top"
        >
          <div
            v-if="newRecord.replyTo"
            class="reply-to-container p-3 position-relative"
          >
            <p class="text-muted">
              Replying to
            </p>

            <div class="position-relative">
              <div class="reply-to-toolbox position-absolute top-0 end-0">
                <button
                  class="btn btn-outline-light btn-sm py-1"
                  @click="newRecord.replyTo = null"
                >
                  <font-awesome-icon :icon="['fas', 'times']" />
                </button>
              </div>

              <comment-reply
                :reply="newRecord.replyTo"
                :title-field="titleField"
                :content-field="contentField"
                :namespace="namespace"
                @click.native="handleReplyClick(newRecord.replyTo.recordID)"
              />
            </div>
          </div>

          <input
            v-if="titleField"
            v-model="newRecord.title"
            class="form-control mb-2"
            :placeholder="$t('comment.title.placeholder')"
          />

          <c-rich-text-input
            ref="richTextInput"
            v-model="newRecord.content"
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
            @upload="handleFileUpload"
          />

          <c-uploader
            v-if="attachmentField"
            ref="uploader"
            :endpoint="fileUploadEndpoint"
            :form-data="uploaderFormData"
            :accepted-files="mimetypes"
            :max-filesize="maxSize"
            :max-files="attachmentField.isMulti ? undefined : 1"
            class="d-none"
            @upload="appendAttachment"
          />

          <list-loader
            v-if="attachmentField && newRecord.attachmentIDs.length"
            kind="record"
            v-model:set="newRecord.attachmentIDs"
            :namespace="namespace"
            :enable-order="attachmentField.isMulti"
            enable-delete
            mode="list"
            :hide-file-name="attachmentField.options.hideFileName"
            :preview-options="attachmentField.options"
            class="px-2"
          />

          <div class="d-flex align-items-center justify-content-end m-2 gap-1">
            <button
              v-if="attachmentField"
              :title="$t('comment.tooltip.attach')"
              class="btn btn-outline-light text-secondary border-0"
              @click="openFileUpload"
            >
              <font-awesome-icon :icon="['fas', 'paperclip']" />
            </button>

            <c-button-submit
              :text="$t('comment.submit')"
              :disabled="!isValid || isProcessing"
              :processing="submitting"
              @submit="submitComment()"
            />
          </div>
        </section>

        <div
          v-if="replyModal.show"
          class="modal fade show d-block"
          tabindex="-1"
        >
          <div class="modal-dialog modal-dialog-centered modal-xl">
            <div class="modal-content">
              <div class="modal-body p-2">
                <div
                  v-if="!replyModal.comment"
                  class="d-flex align-items-center justify-content-center p-3"
                >
                  <span class="spinner-border" />
                </div>

                <div v-else>
                  <div class="d-flex align-items-center justify-content-center gap-3 mx-2 text-muted">
                    <hr class="flex-grow-1 m-0" />
                    <span>{{ getFormattedDate((replyModal.comment || {}).createdAt) }}</span>
                    <hr class="flex-grow-1 m-0" />
                  </div>

                  <comment-item
                    :comment="replyModal.comment"
                    :title-field="titleField"
                    :content-field="contentField"
                    :attachment-field="attachmentField"
                    :namespace="namespace"
                    :show-time-always="true"
                    :show-title="showTitle(replyModal.comment)"
                    :show-content="showContent(replyModal.comment)"
                    :highlighted="false"
                    :disable-hover="true"
                    @reply="replyToComment(replyModal.comment)"
                    @reply-click="openReplyInModal"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
        <div
          v-if="replyModal.show"
          class="modal-backdrop fade show"
          @click="replyModal.show = false"
        />
      </div>
    </template>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, nextTick, inject } from 'vue'
import { useStore } from '../../../store'
import { usePageBlockBase } from '../usePageBlockBase'
import { NoID, compose, fmt } from 'corteza-lib/js/dist'
import { components, composables } from 'corteza-lib/vue/dist'
import { evaluatePrefilter, getFieldFilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import axios from 'axios'
import CommentItem from './Item.vue'
import CommentReply from './Reply.vue'
import ListLoader from 'corteza-webapp-compose/src/components/Public/Page/Attachment/ListLoader'
import Wrap from '../Wrap/index.js'

const { CRichTextInput, CUploader, CButtonSubmit, emojiData } = components
const { toastErrorHandler } = composables.useToast()

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
  previewMode: { type: Boolean, required: false, default: false },
})

const emit = defineEmits(['errors'])
const store = useStore()
const $auth = inject('$auth')
const $ComposeAPI = inject('$ComposeAPI')

const { options, isProcessing, processing, refreshBlock, setBaseDefaultValues } = usePageBlockBase(props, emit)

const chatContainer = ref(null)
const richTextInput = ref(null)
const uploader = ref(null)

const filter = reactive({
  sort: '',
  filter: '',
  limit: 50,
  pageCursor: '',
  prevPage: '',
  nextPage: '',
})
const comments = ref([])
const newRecord = reactive({
  title: '',
  content: '',
  replyTo: null,
  attachmentIDs: [],
})
const submitting = ref(false)
const loadingMore = ref(false)
const abortableRequests = ref([])
const showNewestFirst = ref(true)
const highlightedCommentId = ref(null)
const replyModal = reactive({
  show: false,
  comment: null,
})
const commentRefreshInterval = ref(null)
const autoFetching = ref(false)

const getModuleByID = computed(() => store.module.getByID)
const getPageByID = computed(() => store.page.getByID)
const findUserByID = computed(() => store.user.findByID)
const findRecordByID = computed(() => store.record.findByID)

const lastCommentTimestamp = computed(() => {
  if (comments.value.length === 0) return null
  const msgs = comments.value[comments.value.length - 1]?.messages || []
  if (msgs.length === 0) return null
  const cs = msgs[msgs.length - 1]?.comments || []
  if (cs.length === 0) return null
  return cs[cs.length - 1]?.createdAt || null
})

const roModule = computed(() => getModuleByID.value(props.module?.moduleID || ''))

const moduleID = computed(() => options.value.moduleID)

const titleField = computed(() => {
  const fn = options.value.titleField
  if (!fn || !roModule.value) return undefined
  return roModule.value.fields.find(f => f.name === fn)
})

const contentField = computed(() => {
  const fn = options.value.contentField
  if (!fn || !roModule.value) return undefined
  return roModule.value.fields.find(f => f.name === fn)
})

const referenceField = computed(() => {
  const fn = options.value.referenceField
  if (!fn || !roModule.value) return undefined
  return roModule.value.fields.find(f => f.name === fn) || {}
})

const attachmentField = computed(() => {
  const fn = options.value.attachmentField
  if (!fn || !roModule.value) return undefined
  const f = roModule.value.fields.find(f => f.name === fn)
  if (!f) return undefined
  const af = compose.ModuleFieldMaker(f)
  af.options.mode = 'list'
  return af
})

const reactionsField = computed(() => {
  const fn = options.value.reactionsField
  if (!fn || !roModule.value) return undefined
  return roModule.value.fields.find(f => f.name === fn)
})

const currentUserID = computed(() => ($auth?.user?.userID) || '')

const fileUploadEndpoint = computed(() => {
  if (!attachmentField.value) return undefined
  const mid = moduleID.value
  const rid = NoID
  const nid = props.namespace.namespaceID
  return $ComposeAPI.baseURL + $ComposeAPI.recordUploadEndpoint({ namespaceID: nid, moduleID: mid, recordID: rid, fieldName: attachmentField.value.name })
})

const uploaderFormData = computed(() => {
  if (!attachmentField.value) return {}
  return { fieldName: attachmentField.value.name }
})

const mimetypes = computed(() => {
  if (!attachmentField.value) return []
  const a = (attachmentField.value.options.mimetypes || '').trim()
  return a ? a.split(',').map(p => p.trim()) : []
})

const maxSize = computed(() => {
  if (!attachmentField.value) return 100
  return attachmentField.value.options.maxSize || 100
})

const replyField = computed(() => {
  const fn = options.value.replyField
  if (!fn || !roModule.value) return undefined
  return roModule.value.fields.find(f => f.name === fn) || {}
})

const canAddRecord = computed(() => roModule.value && roModule.value.canCreateRecord)

const isValid = computed(() => (!!newRecord.title || !!newRecord.content || newRecord.attachmentIDs.length) && !isNewRecord.value)

const isNewRecord = computed(() => {
  if (props.record) return props.record.recordID === NoID
  return false
})

const reference = computed(() => {
  if (props.record) return props.record.recordID !== NoID ? props.record.recordID : NoID
  return NoID
})

const isBlockConfigured = computed(() => !!contentField.value)

const hasPrevPage = computed(() => !!filter.prevPage)

const hasNextPage = computed(() => !!filter.nextPage)

watch(() => props.record?.recordID, () => {
  showNewestFirst.value = options.value.sortDirection === 'asc'
  refresh()
}, { immediate: true })

watch(() => options.value, () => {
  showNewestFirst.value = options.value.sortDirection === 'asc'
  refresh()
}, { deep: true })

onMounted(() => {
  refreshBlock(refresh)
  startAutoRefresh()
  createEvents()
})

onBeforeUnmount(() => {
  abortRequests()
  destroyEvents()
  setDefaultValues()
})

function createEvents () {
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('refetch-records', refresh)
}

function startAutoRefresh () {
  commentRefreshInterval.value = setInterval(() => {
    if (autoFetching.value || submitting.value || loadingMore.value) return
    if (!showNewestFirst.value && filter.nextPage) return
    autoFetching.value = true
    Promise.all([
      loadNewComments(),
      loadUpdatedComments(),
    ]).finally(() => { autoFetching.value = false })
  }, 5000)
}

function refetchOnPrefilterValueChange ({ fieldName }) {
  if (isFieldInFilter(fieldName, options.value.filter)) refresh()
}

function getFormattedDateTime (date) {
  return fmt.fullDateTime(date)
}

function getFormattedDate (timestamp) {
  const date = new Date(timestamp)
  const today = new Date()
  const yesterday = new Date()
  yesterday.setDate(yesterday.getDate() - 1)
  const cd = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const ct = new Date(today.getFullYear(), today.getMonth(), today.getDate())
  const cy = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate())
  if (cd.getTime() === ct.getTime()) return 'Today'
  if (cd.getTime() === cy.getTime()) return 'Yesterday'
  return fmt.date(timestamp, { dateStyle: 'long' })
}

function getFormattedTime (date) {
  return fmt.time(date)
}

async function loadNewComments () {
  const f = [
    expandFilter(),
    lastCommentTimestamp.value ? `${getFieldFilter('createdAt', 'DateTime', lastCommentTimestamp.value, '>')}` : '',
  ].filter(Boolean).join(' AND ')
  const wasAtBottom = isScrollAtBottom()
  const newGroups = await fetchCommentRecords(roModule.value, f, false)
  comments.value = mergeMessageGroups(comments.value, newGroups, true)
  if (wasAtBottom) {
    nextTick(() => scrollToLatest())
  }
}

async function loadUpdatedComments () {
  if (!reactionsField.value) return
  const recordIDs = []
  comments.value.forEach(dg => {
    (dg.messages || []).forEach(mg => {
      (mg.comments || []).forEach(c => recordIDs.push(c.recordID))
    })
  })
  if (!recordIDs.length) return
  const idFilter = recordIDs.map(id => `recordID = '${id}'`).join(' OR ')
  const f = `(${idFilter})`
  const updatedGroups = await fetchCommentRecords(roModule.value, f, false)
  updateExistingComments(updatedGroups)
}

function updateExistingComments (updatedGroups) {
  if (!updatedGroups?.length) return
  const fn = reactionsField.value.name
  const updatedMap = {}
  updatedGroups.forEach(dg => {
    (dg.messages || []).forEach(mg => {
      (mg.comments || []).forEach(c => { updatedMap[c.recordID] = c })
    })
  })
  comments.value.forEach(dg => {
    dg.messages.forEach(mg => {
      mg.comments.forEach((c, i) => {
        const u = updatedMap[c.recordID]
        if (!u) return
        if (c.values[fn] === u.values[fn]) return
        u.author = u.author || c.author
        u.reply = u.reply || c.reply
        mg.comments.splice(i, 1, u)
      })
    })
  })
}

function mergeMessageGroups (existing, newGroups, newestFirst = showNewestFirst.value) {
  if (!existing.length || !newGroups.length) {
    return newestFirst ? [...existing, ...newGroups] : [...newGroups, ...existing]
  }
  const [eg, ng] = newestFirst
    ? [existing[existing.length - 1], newGroups[0]]
    : [existing[0], newGroups[newGroups.length - 1]]
  if (eg.date === ng.date) {
    if (newestFirst) {
      ng.messages.forEach(nm => {
        const le = eg.messages[eg.messages.length - 1]
        if (le && le.authorId === nm.authorId) {
          le.comments = [...le.comments, ...nm.comments]
        } else {
          eg.messages.push(nm)
        }
      })
    } else {
      eg.messages = [...ng.messages, ...eg.messages]
    }
    newestFirst ? newGroups.shift() : newGroups.pop()
  }
  return newestFirst ? [...existing, ...newGroups] : [...newGroups, ...existing]
}

function getAuthor (userID) {
  const user = findUserByID.value(userID) || {}
  const name = user.name || user.handle || user.email || ''
  let initials = '?'
  if (name) {
    const words = name.trim().split(/\s+/)
    initials = words.length === 1
      ? words[0].substring(0, 2).toUpperCase()
      : words.slice(0, 2).map(w => w.charAt(0).toUpperCase()).join('')
  }
  return { name, initials, isCurrentUser: Boolean($auth?.user?.userID === userID) }
}

function loadMoreMessages () {
  loadingMore.value = true
  const container = chatContainer.value
  const currentScrollTop = container ? container.scrollTop : 0
  const currentScrollHeight = container ? container.scrollHeight : 0
  fetchCommentRecords(roModule.value, expandFilter()).then(newGroups => {
    comments.value = mergeMessageGroups(comments.value, newGroups, !showNewestFirst.value)
  }).finally(() => {
    nextTick(() => {
      if (container && showNewestFirst.value) {
        container.scrollTop = currentScrollTop + (container.scrollHeight - currentScrollHeight)
      }
    })
    loadingMore.value = false
  })
}

function refreshOnRelatedRecordsUpdate ({ moduleID: mid } = {}) {
  if (options.value.moduleID === mid) refresh()
}

async function refresh () {
  if (!options.value.moduleID) throw new Error('Module or page not set')
  if (roModule.value && contentField.value) {
    processing.value = true
    filter.nextPage = ''
    try {
      const grouped = await fetchCommentRecords(roModule.value, expandFilter())
      if (showNewestFirst.value) {
        comments.value = grouped.sort((a, b) => new Date(a.date) - new Date(b.date))
      } else {
        comments.value = grouped.sort((a, b) => new Date(b.date) - new Date(a.date))
      }
    } catch (e) { console.error(e) }
    setTimeout(() => {
      processing.value = false
      nextTick(() => scrollToPosition())
    }, 300)
  }
}

function scrollToPosition () {
  const container = chatContainer.value
  if (!container) return
  if (showNewestFirst.value) {
    container.scrollTop = container.scrollHeight
  } else {
    container.scrollTop = 0
  }
}

function scrollToLatest () {
  const container = chatContainer.value
  if (!container) return
  container.scrollTop = container.scrollHeight
}

function isScrollAtBottom () {
  const container = chatContainer.value
  if (!container) return false
  const threshold = 25
  return container.scrollTop + container.clientHeight >= container.scrollHeight - threshold
}

function handleFileUpload (files) {
  if (!attachmentField.value) return
  const u = uploader.value
  if (u?.$refs?.dropzone) {
    Array.from(files).forEach(file => { u.$refs.dropzone.addFile(file) })
  }
}

function openFileUpload () {
  const u = uploader.value
  if (u?.$refs?.dropzone) {
    u.$refs.dropzone.dropzone.hiddenFileInput.click()
  }
}

function appendAttachment ({ attachmentID } = {}) {
  if (!attachmentID) return
  if (attachmentField.value?.isMulti) {
    newRecord.attachmentIDs = [...newRecord.attachmentIDs, attachmentID]
  } else {
    newRecord.attachmentIDs = [attachmentID]
  }
}

function submitComment () {
  if (!isValid.value) return
  submitting.value = true
  const record = new compose.Record(roModule.value)
  if (titleField.value) record.values[titleField.value.name] = newRecord.title
  if (contentField.value) record.values[contentField.value.name] = newRecord.content
  if (referenceField.value) record.values[referenceField.value.name] = reference.value
  if (replyField.value && newRecord.replyTo) record.values[replyField.value.name] = newRecord.replyTo.recordID
  if (attachmentField.value && newRecord.attachmentIDs.length) {
    record.values[attachmentField.value.name] = attachmentField.value.isMulti ? newRecord.attachmentIDs : newRecord.attachmentIDs[0]
  }
  return $ComposeAPI.recordCreate(record).then(rec => {
    rec = new compose.Record(roModule.value, rec)
    newRecord.title = ''
    newRecord.content = ''
    newRecord.replyTo = null
    newRecord.attachmentIDs = []
    if (showNewestFirst.value) {
      return loadNewComments()
    } else {
      showNewestFirst.value = true
      filter.nextPage = ''
      return fetchCommentRecords(roModule.value, expandFilter()).then(groups => { comments.value = groups })
    }
  }).catch(toastErrorHandler('Failed to create comment')).finally(() => {
    submitting.value = false
    nextTick(() => scrollToLatest())
  })
}

function onEditComment (comment, { title, content }) {
  const record = new compose.Record(roModule.value, { ...comment })
  if (titleField.value) record.values[titleField.value.name] = title
  if (contentField.value) record.values[contentField.value.name] = content
  return $ComposeAPI.recordUpdate(record).then(rec => {
    const updated = new compose.Record(roModule.value, rec)
    updated.author = comment.author
    updated.reply = comment.reply
    comments.value.forEach(dg => {
      dg.messages.forEach(mg => {
        const idx = mg.comments.findIndex(c => c.recordID === updated.recordID)
        if (idx > -1) mg.comments.splice(idx, 1, updated)
      })
    })
  }).catch(toastErrorHandler('Failed to update comment'))
}

function onReact (comment, emoji) {
  if (!reactionsField.value) return
  const record = new compose.Record(roModule.value, { ...comment })
  const fn = reactionsField.value.name
  let reactions = {}
  try { reactions = JSON.parse(record.values[fn] || '{}') || {} } catch { reactions = {} }
  const uid = currentUserID.value
  if (!uid) return
  if (!reactions[emoji]) reactions[emoji] = []
  const idx = reactions[emoji].indexOf(uid)
  if (idx > -1) {
    reactions[emoji].splice(idx, 1)
    if (reactions[emoji].length === 0) delete reactions[emoji]
  } else {
    reactions[emoji].push(uid)
  }
  record.values[fn] = JSON.stringify(reactions)
  return $ComposeAPI.recordUpdate(record).then(rec => {
    const updated = new compose.Record(roModule.value, rec)
    updated.author = comment.author
    updated.reply = comment.reply
    resolveReactionUsers(updated)
    comments.value.forEach(dg => {
      dg.messages.forEach(mg => {
        const idx = mg.comments.findIndex(c => c.recordID === updated.recordID)
        if (idx > -1) mg.comments.splice(idx, 1, updated)
      })
    })
  }).catch(toastErrorHandler('Failed to update comment'))
}

function resolveReactionUsers (record) {
  if (!reactionsField.value) return
  try {
    const reactions = JSON.parse(record.values[reactionsField.value.name] || '{}') || {}
    const userIDs = [...new Set(Object.values(reactions).flat())].filter(Boolean)
    if (userIDs.length) store.dispatch('user/resolveUsers', userIDs)
  } catch {}
}

function expandFilter () {
  if (!props.record) {
    if ((options.value.filter || '').includes('${record')) throw new Error('Invalid record variable')
    if ((options.value.filter || '').includes('${ownerID}')) throw new Error('Invalid owner variable')
  }
  if (options.value.filter) {
    try {
      return evaluatePrefilter(options.value.filter, {
        record: props.record,
        user: $auth.user || {},
        recordID: (props.record || {}).recordID || NoID,
        ownerID: (props.record || {}).ownedBy || NoID,
        userID: ($auth.user || {}).userID || NoID,
      })
    } catch (e) { return e }
  }
  return ''
}

async function fetchCommentRecords (module, query, useNextPage = true) {
  if (module.moduleID !== options.value.moduleID) throw new Error('Module mismatch')
  let q = query || ''
  if (referenceField.value) {
    if (q.length) q += ' AND '
    q += `${referenceField.value.name} = '${reference.value}' `
  }
  let sort = showNewestFirst.value ? 'createdAt DESC' : 'createdAt ASC'
  if (useNextPage && filter.nextPage) sort = ''
  const { moduleID: mid, namespaceID: nid } = module
  const params = {
    namespaceID: nid,
    moduleID: mid,
    query: q,
    sort,
    limit: useNextPage ? filter.limit : 500,
    pageCursor: useNextPage ? filter.nextPage : '',
  }
  const { response, cancel } = $ComposeAPI.recordListCancellable(params)
  abortableRequests.value.push(cancel)
  try {
    const { set = [], filter: f = {} } = await response()
    if (useNextPage) filter.nextPage = f.nextPage || ''
    const comments = set.map(r => new compose.Record(module, r))
    await Promise.all([
      fetchUsersModule([{ name: 'createdBy', kind: 'User', isSystem: true, isMulti: false }], comments),
      fetchReplyRecords(comments),
    ])
    comments.forEach(c => resolveReactionUsers(c))
    if (showNewestFirst.value) comments.reverse()
    const groups = {}
    comments.forEach(comment => {
      const date = getFormattedDate(comment.createdAt)
      const authorId = comment.createdBy
      comment.reply = getReplyComment(comment)
      comment.author = getAuthor(authorId)
      if (!groups[date]) groups[date] = { date, messages: [] }
      const last = groups[date].messages[groups[date].messages.length - 1]
      if (last && last.authorId === authorId) {
        last.comments.push(comment)
      } else {
        groups[date].messages.push({ authorId, comments: [comment] })
      }
    })
    return Object.values(groups)
  } catch (e) {
    if (!axios.isCancel(e)) console.error(e)
    return []
  }
}

function fetchUsersModule (fields = [], records = []) {
  if (!records.length || !fields.length) return
  const list = [...new Set(records.map(r => fields.filter(c => c.kind === 'User').map(f => f.isSystem ? [r[f.name]] : f.isMulti ? r.values[f.name] : [r.values[f.name]])).flat(Infinity))].filter(uID => uID !== NoID)
  if (list.length) return store.dispatch('user/resolveUsers', list)
}

async function fetchReplyRecords (records) {
  if (!replyField.value || !records.length) return
  const fields = [replyField.value]
  return fetchRecordsModule(props.namespace.namespaceID, fields, records)
}

async function fetchRecordsModule (namespaceID, fields = [], records = []) {
  if (!records.length || !fields.length) return
  const moduleRecords = {}
  fields.filter(c => c.kind === 'Record').forEach(f => {
    const { moduleID: fmid } = f.options || {}
    if (!fmid) return
    if (!moduleRecords[fmid]) moduleRecords[fmid] = new Set()
    records.forEach(r => {
      const ids = f.isSystem ? [r[f.name]] : f.isMulti ? r.values[f.name] : [r.values[f.name]]
      ids.forEach(id => { if (id) moduleRecords[fmid].add(id) })
    })
  })
  return Promise.all(Object.entries(moduleRecords).map(([fmid, ids]) => {
    ids = [...ids]
    return ids.length ? store.dispatch('record/resolveRecords', { namespaceID, moduleID: fmid, recordIDs: ids }) : Promise.resolve([])
  }))
}

function replyToComment (comment) {
  newRecord.replyTo = comment
  replyModal.show = false
  nextTick(() => {
    const rti = richTextInput.value
    if (rti) {
      if (typeof rti.focus === 'function') rti.focus()
      else if (rti.editor && typeof rti.editor.focus === 'function') rti.editor.focus()
    }
  })
}

function handleReplyClick (recordID) {
  const el = document.getElementById(`comment-${recordID}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    highlightedCommentId.value = recordID
  } else {
    openReplyInModal(recordID)
  }
}

function openReplyInModal (recordID) {
  const { namespaceID: nid, moduleID: mid } = roModule.value || {}
  if (!nid || !mid) return
  replyModal.show = true
  replyModal.comment = null
  let comment = findRecordByID.value(recordID)
  if (!comment) return
  comment = new compose.Record(roModule.value, comment)
  fetchReplyRecords([comment]).then(() => {
    comment.reply = getReplyComment(comment)
    comment.author = getAuthor(comment.createdBy)
    replyModal.comment = comment
  }).catch(e => {
    replyModal.show = false
    toastErrorHandler('Failed to load comment')(e)
  })
}

function resetHighlightedComment (recordID) {
  if (highlightedCommentId.value === recordID) highlightedCommentId.value = null
}

function showTitle (comment) {
  return Boolean(titleField.value && titleField.value.canReadRecordValue && comment.values[titleField.value.name])
}

function showContent (comment) {
  return Boolean(contentField.value && contentField.value.canReadRecordValue && comment.values[contentField.value.name])
}

function showReply (comment) {
  return Boolean(replyField.value && replyField.value.canReadRecordValue && comment.values[replyField.value.name])
}

function getReplyComment (comment) {
  if (!showReply(comment)) return null
  let replyRecord = findRecordByID.value(comment.values[replyField.value.name])
  if (!replyRecord) return null
  replyRecord = new compose.Record(roModule.value, replyRecord)
  replyRecord.author = getAuthor(replyRecord.createdBy)
  return replyRecord
}

function setDefaultValues () {
  filter.sort = ''
  filter.filter = ''
  filter.limit = 50
  filter.pageCursor = ''
  filter.prevPage = ''
  filter.nextPage = ''
  comments.value = []
  newRecord.title = ''
  newRecord.content = ''
  newRecord.replyTo = null
  newRecord.attachmentIDs = []
  abortableRequests.value = []
  submitting.value = false
  loadingMore.value = false
  showNewestFirst.value = true
  highlightedCommentId.value = null
  replyModal.show = false
  replyModal.comment = null
  autoFetching.value = false
  if (commentRefreshInterval.value) {
    clearInterval(commentRefreshInterval.value)
    commentRefreshInterval.value = null
  }
}

function abortRequests () {
  abortableRequests.value.forEach(cancel => cancel())
}

function destroyEvents () {
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('refetch-records', refresh)
}
</script>

<style lang="scss" scoped>
.reply-to-container {
  .reply-to-toolbox {
    opacity: 0;
    transition: opacity 0.2s ease;
    z-index: 1;
  }

  &:hover {
    .reply-to-toolbox {
      opacity: 1;
    }
  }
}
</style>
