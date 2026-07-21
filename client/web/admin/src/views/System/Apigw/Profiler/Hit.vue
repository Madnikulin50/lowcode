<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('hit.title')">
      <span class="text-nowrap" />
    </c-content-header>

    <c-profiler-hit-info
      :processing="info.processing"
      :success="info.success"
      :can-create="canCreate"
      :hit="hit"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useStore } from 'corteza-webapp-admin/src/store'
import { useEditorHelpers } from 'corteza-webapp-admin/src/mixins/editorHelpers'
import CProfilerHitInfo from 'corteza-webapp-admin/src/components/Apigw/Profiler/CProfilerHitInfo'

const { t } = useI18n()
const $route = useRoute()
const store = useStore()
const { incLoader, decLoader } = useEditorHelpers()

const hit = ref({})
const info = reactive({ processing: false, success: false })

const canCreate = computed(() => store.rbac.can('system/', 'apigw-route.create'))

watch(() => $route.params.hitID, {
  immediate: true,
  handler() {
    if ($route.params.hitID) {
      fetchHit()
    } else {
      hit.value = {}
    }
  },
})

onMounted(() => {
  fetchHit()
})

function fetchHit() {
  incLoader()

  window.__systemAPI.apigwProfilerHit({ hitID: $route.params.hitID })
    .then(h => { hit.value = h })
    .catch(window.__toastError(t('notification.queue.fetch.error')))
    .finally(() => {
      decLoader()
    })
}
</script>
