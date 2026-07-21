  <template>
    <Wrap
      v-bind="$props"
      @refreshBlock="refresh"
    >
      <div class="d-flex flex-column calendar-container p-2 h-100">
        <div v-if="!header.hide">
          <div
            v-if="!header.hidePrevNext || !header.hideTitle"
            class="d-flex align-items-baseline justify-content-center mb-2"
          >
            <button
              v-if="!header.hidePrevNext"
              class="btn btn-link text-dark"
              @click="api().prev()"
            >
              <font-awesome-icon :icon="['fas', 'angle-left']" />
            </button>
            <span
              v-if="!header.hideTitle"
              class="h5"
            >
            {{ title }}
          </span>
            <button
              v-if="!header.hidePrevNext"
              class="btn btn-link text-dark"
              @click="api().next()"
            >
              <font-awesome-icon :icon="['fas', 'angle-right']" />
            </button>
          </div>
          <div class="row g-0">
            <div
              class="col-12 col-sm-10 col-md-9 col-lg-8 col-xl-9 d-flex justify-content-sm-start justify-content-center flex-wrap"
            >
              <button
                v-for="view in views"
                :key="view"
                class="btn btn-outline-secondary me-1 mb-1"
                @click="api().changeView(view)"
              >
                {{ $t(`calendar.view.${view}`) }}
              </button>
            </div>
            <div
              v-if="!header.hideToday && !header.hide"
              class="col-12 col-sm-2 col-md-3 col-lg-4 col-xl-3 d-flex justify-content-end"
            >
              <button
                class="btn btn-outline-secondary mb-1 w-100"
                @click="api().today()"
              >
                {{ $t('calendar.today') }}
              </button>
            </div>
          </div>
        </div>

        <div
          :ref="`cc-${blockIndex}`"
          class="d-flex flex-column flex-fill"
        >
          <div
            v-if="isProcessing"
            class="d-flex align-items-center justify-content-center h-100"
          >
            <span class="spinner-border spinner-border-sm" />
          </div>

          <full-calendar
            v-show="!isProcessing"
            :ref="`fc-${blockIndex}`"
            :key="key"
            :height="getHeight()"
            :events="events"
            v-bind="config"
            class="flex-fill"
            @eventClick="handleEventClick"
          />
        </div>
      </div>
    </Wrap>
  </template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick, getCurrentInstance } from 'vue'
import moment from 'moment'
import axios from 'axios'
import { useStore } from '../../store'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import timeGridPlugin from '@fullcalendar/timegrid'
import listPlugin from '@fullcalendar/list'
import '@fullcalendar/core/main.css'
import '@fullcalendar/daygrid/main.css'
import { BootstrapTheme } from '@fullcalendar/bootstrap'
import { createPlugin } from '@fullcalendar/core'
import { evaluatePrefilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { useRouter } from 'vue-router'

const { getWeekStartDay } = shared

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

const emit = defineEmits(['errors'])
const store = useStore()
const router = useRouter()
const gp = getCurrentInstance()?.appContext?.app?.config?.globalProperties || {}
const $auth = gp.$auth || window.__auth
const $ComposeAPI = gp.$ComposeAPI || window.__composeAPI
const $SystemAPI = gp.$SystemAPI || window.__systemAPI
const $root = typeof window !== 'undefined' ? window : null

const { key, options, isProcessing, inModal, refreshBlock, setBaseDefaultValues } = usePageBlockBase(props, emit)

class CortezaTheme extends BootstrapTheme {}
CortezaTheme.prototype.classes.widget = 'corteza-unthemed'
CortezaTheme.prototype.classes.button = 'btn btn-outline-primary'
CortezaTheme.prototype.baseIconClass = 'fc-icon'
CortezaTheme.prototype.iconClasses = {
  close: 'fc-icon-x',
  prev: 'fc-icon-chevron-left',
  next: 'fc-icon-chevron-right',
  prevYear: 'fc-icon-chevrons-left',
  nextYear: 'fc-icon-chevrons-right',
}
CortezaTheme.prototype.iconOverrideOption = 'buttonIcons'
CortezaTheme.prototype.iconOverrideCustomButtonOption = 'icon'
CortezaTheme.prototype.iconOverridePrefix = 'fc-icon-'

const show = ref(false)
const events = ref([])
const locale = ref(undefined)
const title = ref('')
const loaded = ref({ start: null, end: null })
const refreshing = ref(false)
const cancelTokenSource = axios.CancelToken.source()
const pages = computed(() => store.page.set)

const config = computed(() => ({
  header: false,
  themeSystem: 'corteza',
  defaultView: options.value.defaultView,
  editable: false,
  eventLimit: true,
  locale: locale.value,
  firstDay: weekStartDay.value,
  plugins: [
    dayGridPlugin,
    timeGridPlugin,
    listPlugin,
    createPlugin({ themeClasses: { corteza: CortezaTheme } }),
  ],
  datesRender: renderDate,
  eventRender: ({ event, el }) => {
    if (!event.title) return
    const titleEl = el && el.querySelector('.fc-title, .fc-list-item-title')
    if (!titleEl) return
    titleEl.classList.add('rt-content')
    titleEl.innerHTML = event.title
  },
}))

const header = computed(() => props.block.options.header)
const views = computed(() => {
  if (header.value.hide) return []
  return props.block.reorderViews(header.value.views)
})

const weekStartDay = computed(() => getWeekStartDay(browserLocale()))

watch(() => options.value, () => { updateSize(); refresh() }, { deep: true })
watch(() => props.block.xywh, () => { updateSize() }, { deep: true })
watch(() => props.record?.recordID, () => { refresh() })

onMounted(() => {
  changeLocale(browserLocale()).then(() => {
    refreshBlock(refresh)
    createEvents()
  })
})

onBeforeUnmount(() => {
  setBaseDefaultValues()
  abortRequests()
  destroyEvents()
})

function browserLocale () {
  return navigator.language || navigator.languages?.[0] || 'en-US'
}

function createEvents () {
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('refetch-records', refresh)
}

function refetchOnPrefilterValueChange ({ fieldName }) {
  const { feeds } = options.value
  if (feeds.some(({ options: o }) => isFieldInFilter(fieldName, o.prefilter))) refresh()
}

function updateSize () {
  nextTick(() => { api() && api().updateSize() })
}

function refreshOnRelatedRecordsUpdate ({ moduleID } = {}) {
  options.value.feeds.forEach((feed) => {
    const { moduleID: feedModuleID } = feed.options
    if (feedModuleID === moduleID) refresh()
  })
}

async function changeLocale (lng = 'en-gb') {
  try {
    if (lng === 'en') lng = 'en-gb'
    else {
      const langParts = lng.split('-')
      if (langParts.length > 1) lng = langParts[0].toLowerCase() + '-' + langParts[1].toLowerCase()
    }
    const mod = await import(/* @vite-ignore */ `@fullcalendar/core/locales/${lng}`)
    locale.value = mod.default || mod
  } catch (e) {
    const mod = await import('@fullcalendar/core/locales/en-gb')
    locale.value = mod.default || mod
  }
}

function api () {
  const refKey = `fc-${props.blockIndex}`
  return null
}

function loadEvents (start, end) {
  if (!start || !end) return
  if (start.isSame(loaded.value.start) && end.isSame(loaded.value.end) && !refreshing.value) return
  loaded.value.start = start
  loaded.value.end = end
  events.value = []
  processing.value = true

  Promise.all(options.value.feeds.map(feed => {
    switch (feed.resource) {
      case compose.PageBlockCalendar.feedResources.record:
        return store.module.findByID({ namespace: props.namespace, moduleID: feed.options.moduleID })
          .then(module => {
            const ff = compose.PageBlockCalendar.makeFeed(feed)
            if (ff.options.prefilter) {
              ff.options.prefilter = evaluatePrefilter(ff.options.prefilter, {
                record: props.record,
                user: $auth.user || {},
                recordID: (props.record || {}).recordID || NoID,
                ownerID: (props.record || {}).ownedBy || NoID,
                userID: ($auth.user || {}).userID || NoID,
              })
            }
            return compose.PageBlockCalendar.RecordFeed($ComposeAPI, module, props.namespace, ff, loaded.value, { cancelToken: cancelTokenSource.token })
              .then(evts => { events.value.push(...evts) })
          })
      case compose.PageBlockCalendar.feedResources.reminder:
        return compose.PageBlockCalendar.ReminderFeed($SystemAPI, $auth.user, feed, loaded.value)
          .then(evts => { events.value.push(...evts) })
      default:
        return Promise.resolve([])
    }
  })).finally(() => {
    setTimeout(() => {
      processing.value = false
      refreshing.value = false
      updateSize()
    }, 300)
  })
}

function renderDate ({ view: { activeStart, activeEnd, title: t } } = {}) {
  loadEvents(moment(activeStart), moment(activeEnd))
  title.value = t
}

function handleEventClick ({ event: { extendedProps: { recordID, moduleID } } }) {
  if (!moduleID || !recordID) return
  const page = pages.value.find(p => p.moduleID === moduleID)
  if (!page) return
  const route = { name: 'page.record', params: { recordID, pageID: page.pageID } }
  if (options.value.eventDisplayOption === 'modal' || inModal.value) {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID, recordPageID: page.pageID } }))
  } else if (options.value.eventDisplayOption === 'newTab') {
    window.open(router.resolve(route).href)
  } else {
    router.push(route)
  }
}

function getHeight () {
  return 'auto'
}

function refresh () {
  refreshing.value = true
  new Promise(resolve => resolve(api().refetchEvents()))
    .then(() => key.value++)
}

function abortRequests () {
  cancelTokenSource.cancel(`cancel-record-list-request-${props.block.blockID}`)
}

function destroyEvents () {
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('refetch-records', refresh)
}
</script>

<style lang="scss" scoped>
</style>

<style lang="scss">
.calendar-container {
  .fc-content,
  .event-record { cursor: pointer; }
  .fc-day-header { white-space: pre-wrap; }
}

.fc-popover {
  .fc-header { padding: 0.5rem; }
  .fc-body {
    padding: 0;
    .fc-event-container {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }
  }
}
</style>
