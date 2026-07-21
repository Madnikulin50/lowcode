<template>
  <c-input-select
    :value="userGroup.value"
    data-test-id="select-user-group"
    :options="userGroup.options"
    :get-option-label="getOptionLabel"
    :get-option-key="getOptionKey"
    :placeholder="placeholder"
    :loading="processing"
    :filterable="false"
    v-bind="$attrs"
    @search="search"
    @input="onUserGroupUpdate"
  />
</template>

<script setup lang="ts">
import { ref, reactive, getCurrentInstance } from 'vue'
import { NoID, system } from '@cortezaproject/corteza-js'
import { debounce } from 'lodash'
import axios from 'axios'

defineOptions({ inheritAttrs: false })

const vm = getCurrentInstance()!
const $SystemAPI = (vm.appContext.config.globalProperties as any).$SystemAPI

const props = withDefaults(defineProps<{
  modelValue?: string | null
  placeholder?: string
}>(), {
  modelValue: null,
  placeholder: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | undefined]
}>()

const processing = ref(false)
const cancelRequest = ref<(() => void) | null>(null)

const userGroup = reactive({
  options: [] as any[],
  value: undefined as any,
  filter: {
    query: null as string | null,
    limit: 20,
  },
})

// Created equivalent
;(async () => {
  await fetchUserGroups()
  getUserGroupByID(props.modelValue)
})()

const search = debounce(function (query: string) {
  if (query !== userGroup.filter.query) {
    userGroup.filter.query = query
    userGroup.filter.page = 1
  }
  fetchUserGroups()
}, 300)

function fetchUserGroups() {
  processing.value = true

  if (cancelRequest.value) {
    cancelRequest.value()
    cancelRequest.value = null
  }

  const { response, cancel } = $SystemAPI.userGroupListCancellable(userGroup.filter)
  cancelRequest.value = cancel

  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(([{ set }]) => {
      userGroup.options = set.map((m: any) => new system.UserGroup(m))
      processing.value = false
    })
    .catch((e: any) => {
      if (axios.isCancel(e)) return
      processing.value = false
      throw e
    })
}

function getUserGroupByID(userGroupID: string | null | undefined) {
  if (!userGroupID || userGroupID === NoID) {
    userGroup.value = userGroup.options.find(({ isRoot }: any) => !!isRoot)
    emit('update:modelValue', userGroup.value?.userGroupID)
    return
  }

  const found = userGroup.options.find((o: any) => o.userGroupID === userGroupID)

  if (found) {
    userGroup.value = found
  }
}

function onUserGroupUpdate(val: any) {
  userGroup.value = val
  emit('update:modelValue', val.userGroupID)
}

function getOptionKey({ userGroupID }: any) {
  return userGroupID
}

function getOptionLabel({ handle, meta = {}, userGroupID }: any) {
  return meta.short || handle || userGroupID
}
</script>

<style scoped>
</style>
