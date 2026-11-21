<template>
  <div class="h-100 d-flex flex-column">
    <div class="overflow-auto flex-grow-1 h-100">
      <div
        v-if="loading"
        class="d-flex justify-content-center p-5"
      >
        <span class="spinner-border text-primary" />
      </div>

      <div
        v-if="drafts.length > 0 && !loading"
        class="d-flex align-items-center justify-content-between p-2 ps-3 border-bottom bg-light"
      >
        <div class="text-secondary small fw-bold">
          {{ $t('count', { count: drafts.length }) }}
        </div>

        <div class="d-flex align-items-center gap-1">
          <c-input-select
            v-model="sortOrder"
            :options="sortOptions"
            :reduce="option => option.value"
            :clearable="false"
            :searchable="false"
            size="sm"
            style="width: 10rem;"
            class="border-0"
          >
            <template #option="option">
              <div class="d-flex align-items-center gap-1">
                <font-awesome-icon
                  v-if="option && option.icon"
                  :icon="['fas', option.icon]"
                  class="text-secondary"
                />
                {{ option.text }}
              </div>
            </template>
            <template #selected-option="option">
              <div
                v-if="option"
                class="d-flex align-items-center gap-1 text-secondary"
              >
                <font-awesome-icon
                  v-if="option.icon"
                  :icon="['fas', option.icon]"
                />
                {{ option.text }}
              </div>
            </template>
          </c-input-select>

          <c-input-confirm
            title="Delete all drafts"
            show-icon
            @confirmed="onClearAll"
          />
        </div>
      </div>

      <div v-if="sortedDrafts.length > 0 && !loading" class="list-group">
        <DraftItem
          v-for="(draft, idx) in sortedDrafts"
          :key="draft.revision.changeID || idx"
          :draft="draft"
          :namespace="namespaces[draft.revision.resource.split('/')[1]]"
          :module="modules[draft.revision.resource.split('/')[2]]"
          :active="$route.query.draftID === String(draft.revision.changeID)"
          @click="onDraftClick(draft)"
          @view="onDraftClick(draft, true)"
          @delete="onDeleteDraft"
        />
      </div>

      <div
        v-else-if="!loading"
        class="text-center p-5"
      >
        <font-awesome-icon
          :icon="['far', 'edit']"
          class="text-secondary mb-3"
          size="3x"
        />
        <p class="text-secondary">
          {{ $t('empty') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: ['drafts', 'notifications', 'general'] } })
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { composables, useDraftsStore } from 'corteza-lib/vue/dist'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import DraftItem from 'corteza-webapp-compose/src/components/Drafts/DraftItem.vue'

const { CInputConfirm, CInputSelect } = components
const { toastSuccess, toastDanger } = composables.useToast()
const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI
const draftsStore = useDraftsStore()
const route = useRoute()
const router = useRouter()

const modules = ref({})
const namespaces = ref({})
const sortOrder = ref('desc')

const drafts = computed(() => draftsStore.getAllDrafts)
const loading = computed(() => draftsStore.isLoading)

const sortedDrafts = computed(() => {
  return [...drafts.value].sort((a, b) => {
    const aTime = new Date(a.revision.timestamp)
    const bTime = new Date(b.revision.timestamp)
    return sortOrder.value === 'desc' ? bTime - aTime : aTime - bTime
  })
})

const sortOptions = computed(() => [
  { value: 'desc', text: $t('newestFirst'), icon: 'sort-amount-down', label: $t('newestFirst') },
  { value: 'asc', text: $t('oldestFirst'), icon: 'sort-amount-up', label: $t('oldestFirst') },
])

watch(drafts, (val) => {
  fetchMetadata(val)
}, { immediate: true })

async function fetchMetadata(draftsList) {
  try {
    for (const draft of draftsList) {
      const parts = draft.revision.resource.split('/')
      const namespaceID = parts[1]
      const moduleID = parts[2]

      if (!namespaces.value[namespaceID]) {
        const ns = await $ComposeAPI.namespaceRead({ namespaceID })
        if (ns) {
          namespaces.value[namespaceID] = new compose.Namespace(ns)
        }
      }

      if (!modules.value[moduleID]) {
        const mod = await $ComposeAPI.moduleRead({ namespaceID, moduleID })
        if (mod) {
          modules.value[moduleID] = new compose.Module(mod)
        }
      }
    }
  } catch (e) {
    console.error('Failed to fetch metadata:', e)
  }
}

async function onDraftClick(draft, view = false) {
  const { revision } = draft
  const parts = revision.resource.split('/')
  if (parts.length < 4) return

  const namespaceID = parts[1]
  const moduleID = parts[2]
  const recordID = parts[3] === '0' ? undefined : parts[3]

  let pageID
  try {
    const pages = await $ComposeAPI.pageList({ namespaceID, moduleID })
    const recordPage = (pages.set || []).find(p => p.moduleID === moduleID)
    if (recordPage) pageID = recordPage.pageID
  } catch (e) {
    console.error('Failed to fetch page metadata:', e)
  }

  if (!pageID) {
    toastDanger($t('notifications.recordRedirectError'))
    return
  }

  const ns = namespaces.value[namespaceID]
  const slug = ns ? ns.slug || ns.namespaceID : namespaceID

  const routeObj = {
    name: view ? 'page.record' : (recordID ? 'page.record.edit' : 'page.record.create'),
    params: { slug, pageID, recordID },
    query: view ? {} : { draftID: revision.changeID },
  }

  router.push(routeObj).catch(err => {
    console.error('Draft navigation failed:', err)
  })
}

function onDeleteDraft({ revision }) {
  draftsStore.removeDraft({ changeID: revision.changeID })
    .then(() => { toastSuccess($t('deleted')) })
    .catch(() => { toastDanger($t('deleteError')) })
}

function onClearAll() {
  draftsStore.clearDrafts()
    .then(() => { toastSuccess($t('allDeleted')) })
    .catch(() => { toastDanger($t('deleteError')) })
}
</script>
