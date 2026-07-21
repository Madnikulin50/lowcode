<template>
  <c-input-select
    ref="userSelect"
    :value="user.value"
    data-test-id="select-user"
    :options="user.options"
    :get-option-label="getOptionLabel"
    :get-option-key="getOptionKey"
    :placeholder="placeholder"
    :loading="processing"
    :filterable="false"
    :clearable="clearable"
    v-bind="$attrs"
    @search="search"
    @input="onUserUpdate"
  />
</template>

<script setup lang="ts">
import { ref, reactive, getCurrentInstance } from 'vue'
import { NoID } from '@cortezaproject/corteza-js'
import { debounce } from 'lodash'
import axios from 'axios'

defineOptions({ inheritAttrs: false })

const vm = getCurrentInstance()!
const $SystemAPI = (vm.appContext.config.globalProperties as any).$SystemAPI

const props = withDefaults(defineProps<{
  modelValue?: string | null
  placeholder?: string
  clearable?: boolean
  clearOnSelect?: boolean
}>(), {
  modelValue: null,
  placeholder: '',
  clearable: false,
  clearOnSelect: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string | undefined]
  'input-object': [user: any]
}>()

const userSelect = ref<any>(null)
const processing = ref(false)
const cancelRequest = ref<(() => void) | null>(null)

const user = reactive({
  options: [] as any[],
  value: undefined as any,
  filter: {
    query: null as string | null,
    limit: 20,
  },
})

// Created equivalent
;(async () => {
  await fetchUsers()
  getUserByID(props.modelValue)
})()

const search = debounce(function (query: string) {
  if (query !== user.filter.query) {
    user.filter.query = query
  }
  fetchUsers()
}, 300)

function fetchUsers() {
  processing.value = true

  if (cancelRequest.value) {
    cancelRequest.value()
    cancelRequest.value = null
  }

  const { response, cancel } = $SystemAPI.userListCancellable(user.filter)
  cancelRequest.value = cancel

  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(([{ set }]) => {
      user.options = set.map((m: any) => Object.freeze(m))
      processing.value = false
    })
    .catch((e: any) => {
      if (axios.isCancel(e)) return
      processing.value = false
      throw e
    })
}

function getUserByID(userID: string | null | undefined) {
  if (!userID || userID === NoID) {
    user.value = undefined
    return
  }

  const found = user.options.find((o: any) => o.userID === userID)

  if (found) {
    user.value = found
  } else {
    return $SystemAPI.userRead({ userID }).then((userData: any) => {
      user.value = userData
      user.options.push(Object.freeze(userData))
    })
  }
}

function onUserUpdate(userData: any) {
  if (props.clearOnSelect && userSelect.value) {
    userSelect.value._data._value = undefined
  } else {
    user.value = userData
  }

  emit('update:modelValue', userData.userID)
  emit('input-object', userData)
}

function getOptionKey({ userID }: any) {
  return userID
}

function getOptionLabel({ userID, email, name, username }: any) {
  return name || username || email || `<@${userID}>`
}
</script>

<style scoped>
</style>
