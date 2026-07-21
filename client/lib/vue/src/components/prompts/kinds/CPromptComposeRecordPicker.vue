<template>
  <div class="d-flex flex-column gap-1">
    <p
      v-if="!!message"
      class="text-break"
      v-html="message"
    />
    <label class="text-primary">{{ label }}</label>
    <c-input-select
      v-model="value"
      :options="options"
      :get-option-key="r => r.recordID"
      :loading="processing"
      append-to-body
      option-value="recordID"
      :placeholder="placeholder"
      :filterable="false"
      :reduce="r => r.recordID"
      class="w-100"
      @search="search"
    >
      <template #list-footer>
        <c-pagination
          v-if="showPagination"
          :has-prev-page="hasPrevPage"
          :has-next-page="hasNextPage"
          @prev="goToPage(false)"
          @next="goToPage(true)"
        />
      </template>
    </c-input-select>
    <button
      :disabled="loading"
      class="btn btn-primary ms-auto"
      @click="emit('submit', { value: encodeValue() })"
    >
      {{ pVal('buttonLabel', 'Submit') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onBeforeUnmount } from 'vue'
import { getCurrentInstance } from 'vue'
import { pVal as _pVal, pType as _pType } from '../utils'
import CPagination from '../common/CPagination.vue'
import CInputSelect from '../../input/CInputSelect.vue'
import { compose, NoID } from '@cortezaproject/corteza-js'
import { debounce } from 'lodash'
import axios from 'axios'

const { $ComposeAPI } = getCurrentInstance()!.appContext.config.globalProperties as any

const props = withDefaults(defineProps<{
  loading?: boolean
  payload?: Record<string, any>
}>(), {
  loading: false,
  payload: () => ({}),
})

const emit = defineEmits<{
  (e: 'submit', value: Record<string, any>): void
}>()

const processing = ref(false)
let cancelRequest: (() => void) | null = null
const query = ref('')
const filter = reactive({
  query: '',
  sort: '',
  limit: 10,
  pageCursor: '',
  prevPage: '',
  nextPage: '',
})
const namespaceID = ref(NoID)
const module = ref<any>(undefined)
const options = ref<any[]>([])
const value = ref<any>(undefined)

const message = computed(() => _pVal(props.payload, 'message', ''))
const label = computed(() => _pVal(props.payload, 'label', ''))

const labelField = computed(() => {
  return module.value?.fields.find((f: any) => f.name === pVal('labelField'))
})

const showPagination = computed(() => hasPrevPage.value || hasNextPage.value)

const hasPrevPage = computed(() => !!filter.prevPage)

const hasNextPage = computed(() => !!filter.nextPage)

const placeholder = computed(() => pVal('placeholder', 'Select a record'))

function pVal(k: string, def?: any) {
  return _pVal(props.payload, k, def)
}

function pType(k: string, def?: any) {
  return _pType(props.payload, k, def)
}

watch(() => filter.pageCursor, (pageCursor) => {
  if (pageCursor) {
    fetchPrefiltered(filter)
  }
})

async function init() {
  const mod = pVal('module')
  const moduleType = pType('module')
  const ns = pVal('namespace')
  const namespaceType = pType('namespace')

  if (namespaceType === 'ID') {
    namespaceID.value = ns
  } else if (namespaceType === 'ComposeNamespace') {
    namespaceID.value = ns.namespaceID
  } else {
    const { set: nn } = await $ComposeAPI.namespaceList({ slug: ns })
    if (!nn || nn.length !== 1) {
      throw new Error('namespace not resolved')
    }
    namespaceID.value = nn[0].namespaceID
  }

  if (moduleType === 'ID') {
    module.value = await $ComposeAPI.moduleRead({ namespaceID: namespaceID.value, moduleID: mod })
    if (!module.value) {
      throw new Error('module not resolved')
    }
  } else if (moduleType === 'ComposeModule') {
    module.value = mod
  } else {
    const { set: nn } = await $ComposeAPI.moduleList({ handle: mod, namespaceID: namespaceID.value })
    if (!nn || nn.length !== 1) {
      throw new Error('module not resolved')
    }
    module.value = nn[0]
  }

  loadLatest()
}

init()

onBeforeUnmount(() => {
  setDefaultValues()
})

function encodeValue() {
  if (!value.value) {
    return { '@type': 'Any', '@value': null }
  }
  const { record = {} } = options.value.find(({ recordID }: any) => recordID === value.value) || {}
  return { '@type': 'ComposeRecord', '@value': record }
}

function loadLatest() {
  const nsID = namespaceID.value
  const modID = module.value?.moduleID
  const { limit } = filter
  if (modID && modID !== NoID) {
    fetchPrefiltered({ namespaceID: nsID, moduleID: modID, limit })
  }
}

const search = debounce(function (queryInput = '') {
  if (queryInput !== query.value) {
    query.value = queryInput
    filter.pageCursor = ''
  }

  const { limit, pageCursor } = filter
  const nsID = namespaceID.value
  const modID = module.value?.moduleID
  const queryFields = pVal('queryFields') || []

  if (modID && modID !== NoID) {
    let qf = queryFields.map((f: any) => f['@value']).filter((f: any) => !!f)
    if ((!qf || qf.length === 0) && pVal('labelField')) {
      qf = [pVal('labelField')]
    }

    let qStr = queryInput
    if (qStr.length > 0) {
      qStr = qf.map((qf: string) => `${qf} LIKE '%${qStr}%'`).join(' OR ')
    }

    const sort = qf.filter((f: any) => !!f).join(', ')
    fetchPrefiltered({ namespaceID: nsID, moduleID: modID, query: qStr, sort, limit })
  }
}, 600)

function fetchPrefiltered(q: Record<string, any>) {
  processing.value = true

  let qStr = q.query || ''
  if (pVal('prefilter')) {
    const pf = pVal('prefilter')
    if (qStr) {
      qStr = `(${pf}) AND (${qStr})`
    } else {
      qStr = pf
    }
  }

  if (cancelRequest) {
    cancelRequest()
    cancelRequest = null
  }

  const { response, cancel } = $ComposeAPI.recordListCancellable({ ...q, query: qStr })
  cancelRequest = cancel

  Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(([{ filter: f, set }]: any) => {
      Object.assign(filter, {
        query: f.query || '',
        sort: f.sort || '',
        limit: f.limit || 10,
        pageCursor: f.pageCursor || '',
        prevPage: f.prevPage || '',
        nextPage: f.nextPage || '',
      })

      options.value = set.map((r: any) => {
        const record = new compose.Record(module.value, r)

        let lbl
        if (labelField.value) {
          lbl = labelField.value.isMulti
            ? record.values[pVal('labelField')].join(', ')
            : record.values[pVal('labelField')]
        }

        return {
          recordID: record.recordID,
          label: lbl || record.recordID,
          record,
        }
      })
      processing.value = false
      return { filter: f, set }
    })
    .catch((e: any) => {
      if (axios.isCancel(e)) return
      processing.value = false
      throw e
    })
}

function goToPage(next = true) {
  filter.pageCursor = next ? filter.nextPage : filter.prevPage
}

function setDefaultValues() {
  processing.value = false
  query.value = ''
  filter.query = ''
  filter.sort = ''
  filter.limit = 10
  filter.pageCursor = ''
  filter.prevPage = ''
  filter.nextPage = ''
  namespaceID.value = NoID
  module.value = undefined
  options.value = []
  value.value = undefined
}
</script>
