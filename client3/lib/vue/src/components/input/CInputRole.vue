<template>
  <c-input-select
    ref="roleSelect"
    :value="modelValue"
    :options="roles"
    :placeholder="placeholder"
    :get-option-key="r => r.roleID"
    :get-option-label="r => getRoleLabel(r)"
    :filterable="false"
    :selectable="selectable"
    :multiple="multiple"
    :clearable="clearable"
    :loading="loading"
    @search="search"
    @input="updateValue"
  />
</template>

<script setup lang="ts">
import { ref, onMounted, getCurrentInstance } from 'vue'
import { debounce } from 'lodash'
import axios from 'axios'

const vm = getCurrentInstance()!
const $SystemAPI = (vm.appContext.config.globalProperties as any).$SystemAPI

const props = withDefaults(defineProps<{
  modelValue?: any
  visible?: (role: any) => boolean
  placeholder?: string
  multiple?: boolean
  clearOnSelect?: boolean
  selectable?: (role: any) => boolean
  clearable?: boolean
  preselect?: boolean
}>(), {
  modelValue: '',
  visible: () => true,
  placeholder: 'Start typing to search for roles',
  multiple: false,
  clearOnSelect: false,
  selectable: () => true,
  clearable: true,
  preselect: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: any]
}>()

const roleSelect = ref<any>(null)
const loading = ref(false)
const cancelRequest = ref<(() => void) | null>(null)
const roles = ref<any[]>([])
const filter = ref('')

onMounted(() => {
  fetchRoles(props.preselect)
})

function fetchRoles(preselect = false) {
  loading.value = true

  if (cancelRequest.value) {
    cancelRequest.value()
    cancelRequest.value = null
  }

  const { response, cancel } = $SystemAPI.roleListCancellable({ query: filter.value, limit: 20 })
  cancelRequest.value = cancel

  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(([{ set }]) => {
      roles.value = set.filter(props.visible)

      if (preselect && (!props.modelValue || !props.modelValue.length)) {
        updateValue(roles.value[0])
      }
      loading.value = false
    })
    .catch((e: any) => {
      if (axios.isCancel(e)) return
      loading.value = false
      throw e
    })
}

const search = debounce(function (query = '') {
  if (query !== filter.value) {
    filter.value = query
  }
  fetchRoles()
}, 400)

function updateValue(role: any) {
  if (props.clearOnSelect && roleSelect.value) {
    roleSelect.value._data._value = undefined
  }
  emit('update:modelValue', role)
}

function getRoleLabel({ name, handle, roleID }: any) {
  return name || handle || roleID
}
</script>
