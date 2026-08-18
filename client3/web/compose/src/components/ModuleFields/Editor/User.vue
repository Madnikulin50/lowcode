<template>
  <div class="mb-3" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" :class="labelColClass">
      <div class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-inline-block mw-100">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>
    </div>
    <div :class="contentColClass">
    <multi
      v-if="field.isMulti"
      v-model:value="value"
      :errors="errors"
      :single-input="field.options.selectType !== 'each'"
      :show-list="field.options.selectType !== 'multiple'"
    >
      <template #single>
        <c-input-select
          v-if="field.options.selectType === 'default'"
          ref="singleSelect"
          :placeholder="t('kind.user.suggestionPlaceholder')"
          :options="options"
          :get-option-label="getOptionLabel"
          :get-option-key="getOptionKey"
          :filterable="false"
          :selectable="isSelectable"
          :loading="processing"
          @search="search"
          @input="updateValue($event)"
        >
          <template #list-footer>
            <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
          </template>
        </c-input-select>
        <c-input-select
          v-else-if="field.options.selectType === 'multiple'"
          v-model="multipleSelected"
          :placeholder="t('kind.user.suggestionPlaceholder')"
          :options="options"
          :get-option-label="getOptionLabel"
          :get-option-key="getOptionKey"
          :filterable="false"
          :selectable="isSelectable"
          :loading="processing"
          multiple
          @search="search"
        >
          <template #list-footer>
            <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
          </template>
        </c-input-select>
      </template>
      <template #default="ctx">
        <c-input-select
          v-if="field.options.selectType === 'each'"
          :placeholder="t('kind.user.suggestionPlaceholder')"
          :options="options"
          :get-option-label="getOptionLabel"
          :get-option-key="getOptionKey"
          :value="getUserIDByIndex(ctx.index)"
          :filterable="false"
          :selectable="isSelectable"
          :loading="processing"
          :clearable="false"
          @search="search"
          @input="updateValue($event, ctx.index)"
        >
          <template #list-footer>
            <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
          </template>
        </c-input-select>
        <span v-else>{{ getOptionLabel(getUserIDByIndex(ctx.index)) }}</span>
      </template>
    </multi>

    <template v-else>
      <c-input-select
        :placeholder="t('kind.user.suggestionPlaceholder')"
        :options="options"
        :get-option-label="getOptionLabel"
        :get-option-key="getOptionKey"
        :value="getUserIDByIndex()"
        :clearable="field.name !== 'ownedBy'"
        :filterable="false"
        :selectable="isSelectable"
        :loading="processing"
        @input="updateValue($event)"
        @search="search"
      >
        <template #list-footer>
          <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
        </template>
      </c-input-select>
      <FieldErrors :errors="errors" />
    </template>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, computed, inject, watch, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { debounce } from 'lodash'
import axios from 'axios'
import { NoID } from 'corteza-lib/js/dist'
import { useEditorBase } from './base'
import { useUserStore } from 'corteza-webapp-compose/src/stores/user'
import Pagination from '../Common/Pagination.vue'
import FieldErrors from '../errors'
import multi from './multi'

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description } = useEditorBase(props, emit)

const $SystemAPI = inject('$SystemAPI')
const $auth = inject('$auth')

const userStore = useUserStore()

const processing = ref(false)
const cancelRequest = ref(null)
const users = ref([])
const filter = ref({ query: null, limit: 10, pageCursor: '', prevPage: '', nextPage: '', roles: [] })

const options = computed(() => users.value)

const multipleSelected = computed({
  get () {
    const map = userID => {
      return userID && userID !== NoID ? userStore.findByID(userID) || { userID } : undefined
    }
    return props.field.isMulti ? (value.value || []).map(map) : map(value.value)
  },
  set (userList) {
    if (userList && Array.isArray(userList)) {
      userList.forEach(u => userStore.push(u))
      value.value = userList.map(({ userID }) => userID)
    }
  },
})

const showPagination = computed(() => hasPrevPage.value || hasNextPage.value)
const hasPrevPage = computed(() => !!filter.value.prevPage)
const hasNextPage = computed(() => !!filter.value.nextPage)

watch(() => value.value, async (val) => {
    const ids = props.field.isMulti ? [...(val || [])] : [val]
    if (ids.length) {
      await userStore.resolveUsers(ids)
    }
  })

watch(() => filter.value.pageCursor, (pageCursor) => {
  if (pageCursor) fetchUsers()
})

const isNewRecord = !props.record || props.record.recordID === NoID
if (isNewRecord && (!value.value || value.value.length === 0) && (props.field.options.presetWithAuthenticated || props.field.name === 'ownedBy')) {
  updateValue(($auth || {}).user)
}
fetchUsers()

onBeforeUnmount(() => {
  setDefaultValues()
})

function getOptionKey (user) {
  if (typeof user === 'string') return user
  return user.userID
}

function getOptionLabel (user) {
  if (typeof user === 'string') user = userStore.findByID(user)
  const { name, username, email, userID: uid } = user || {}
  return name || username || email || `<@${uid}>`
}

function isSelectable ({ userID } = {}) {
  if (!userID) return false
  if (props.field.isMulti) return !props.field.options.isUniqueMultiValue || !((value.value || [])).includes(userID)
  return value.value !== userID
}

function updateValue (user, index = -1) {
  if (user) {
    userStore.push({ ...user })
    const { userID } = user
    if (props.field.isMulti) {
      if (index >= 0) {
        const arr = [...(value.value || [])]
        arr.splice(index, 1, userID)
        value.value = arr
      } else {
        value.value = [...(value.value || []), userID]
      }
    } else {
      value.value = userID
    }
  } else {
    if (index >= 0) {
      const arr = [...(value.value || [])]
      arr.splice(index, 1)
      value.value = arr
    } else {
      value.value = undefined
    }
  }
  emit('change', value.value)
}

function getUserIDByIndex (index = 0) {
  const v = props.field.isMulti ? (value.value || [])[index] : value.value
  return v && v !== NoID ? v : undefined
}

const search = debounce(function (query = '') {
  if (query !== filter.value.query) {
    filter.value.query = query
    filter.value.pageCursor = undefined
  }
  fetchUsers()
}, 300)

function fetchUsers () {
  processing.value = true
  const roleID = props.field.options.roles || []

  if (cancelRequest.value) {
    cancelRequest.value()
    cancelRequest.value = null
  }

  const api = ($SystemAPI.value || {})
  if (!api.userListCancellable) {
    processing.value = false
    return
  }

  const { response, cancel } = api.userListCancellable({ ...filter.value, roleID })
  cancelRequest.value = cancel

  Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(([{ filter: f, set }]) => {
      filter.value = { ...filter.value, ...f }
      filter.value.nextPage = f.nextPage
      filter.value.prevPage = f.prevPage
      users.value = set.map(m => Object.freeze(m))
      processing.value = false
    })
    .catch((e) => {
      if (axios.isCancel(e)) return
      processing.value = false
    })
}

function goToPage (next = true) {
  filter.value.pageCursor = next ? filter.value.nextPage : filter.value.prevPage
}

function setDefaultValues () {
  processing.value = false
  users.value = []
  filter.value = {}
}
</script>
