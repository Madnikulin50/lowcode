<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.apigw.title')" />

    <div class="d-flex flex-column h-100">
      <c-settings-editor
        :settings="apigwSettings"
        :processing="settings.processing"
        :success="settings.success"
        class="mb-3"
        @submit="onSettingsSubmit"
      />

      <c-route-list class="flex-fill" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { inject } from 'vue'
import { useToast } from 'corteza-lib/vue/dist'
import CSettingsEditor from '../../../components/Apigw/CSettingsEditor.vue'
import CRouteList from '../../../components/Apigw/CRouteList.vue'

const $SystemAPI = inject('$SystemAPI')
const $Settings = inject('$Settings')
const { toastSuccess, toastErrorHandler } = useToast()

const settings = ref({
  processing: false,
  success: false,
  items: [],
})

const apigwSettings = computed(() => {
  if (settings.value.items.length > 0) {
    return settings.value.items.reduce((map, obj) => {
      const { name, value } = obj
      const split = name.split('.')
      if (split[0] === 'apigw') {
        map[name] = value
      }
      return map
    }, {})
  }
  return {}
})

function onSettingsSubmit(s) {
  settings.value.processing = true
  const values = Object.entries(s).map(([name, value]) => ({ name, value }))

  $SystemAPI.settingsUpdate({ values })
    .then(() => {
      settings.value.success = true
      toastSuccess('apigw.settings.success')
      $Settings.fetch()
    })
    .catch(toastErrorHandler({ title: 'apigw.settings.error' }))
    .finally(() => {
      settings.value.processing = false
    })
}

function fetchSettings() {
  $SystemAPI.settingsList()
    .catch(toastErrorHandler({ title: 'apigw.settings.fetch.error' }))
    .then((items = []) => {
      settings.value.items = items
    })
}

onMounted(fetchSettings)
</script>
