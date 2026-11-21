<template>
  <Wrap v-bind="$props" @refreshBlock="refresh">
    <div v-if="profile" class="p-2 h-100">
      <Timeline v-if="isTwitter" :id="profile.twitterHandle" :key="key" class="h-100" :options="{ tweetLimit: 9 }" source-type="profile">
        <div class="d-flex align-items-center justify-content-center h-100">
          <div class="spinner-border" />
        </div>
      </Timeline>
    </div>
    <div v-else class="px-3">
      <p>{{ $t('socialFeed.noInput') }}</p>
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePageBlockBase } from './usePageBlockBase'
import { Timeline } from 'vue-tweet-embed'
import Wrap from './Wrap/index.js'

const { t: $t } = useI18n({ useScope: 'global' })

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
})

const { options, refreshBlock } = usePageBlockBase(props, {})

const key = ref(0)

const profile = computed(() => extractSocialUrl(options.value.profileSourceField, options.value.profileUrl, props.record))
const isTwitter = computed(() => profile.value.socialNetwork === 'Twitter')

function getTwitterHandle(url) {
  const parts = url.split('/')
  return parts.length === 4 ? parts[3] : ''
}

function extractSocialUrl(profileSourceField, profileUrl, record) {
  let url = ''
  let socialNetwork = ''
  let twitterHandle = ''
  if (profileSourceField && record) {
    const v = record.values[profileSourceField]
    url = (Array.isArray(v) && v.length > 0 ? v[0] : v) || ''
  } else if (profileUrl) {
    url = profileUrl
  }
  if (url && url.includes('twitter.com')) {
    url = typeof window.checkValidURL === 'function' ? window.checkValidURL(url) : url
    twitterHandle = getTwitterHandle(url)
    if (twitterHandle) socialNetwork = 'Twitter'
  }
  return { url, socialNetwork, twitterHandle }
}

function refresh() { key.value++ }

refreshBlock(refresh)
</script>
