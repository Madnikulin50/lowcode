<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="title">
      <span class="text-nowrap">
        <button
          v-if="$Settings.get('apigw.profiler.enabled', false)"
          class="btn btn-info ms-2"
          @click="$router.push({ name: 'system.apigw.profiler' })"
        >
          {{ $t('system.apigw.profiler.label') }}
        </button>
      </span>
    </c-content-header>

    <c-profiler-route-hits
      :route="$route.params.routeID"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: ['system.apigw'], keyPrefix: 'profiler' } })
import { ref, watch, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import CProfilerRouteHits from 'corteza-webapp-admin/src/components/Apigw/Profiler/CProfilerRouteHits'

const { t } = useI18n()
const $route = useRoute()
const $router = useRouter()
const $Settings = inject('$Settings', {})

const title = ref('')

watch(() => $route.params.routeID, {
  immediate: true,
  handler() {
    title.value = `${t('system.apigw.profiler.title')} - ${decodeRouteID($route.params.routeID)}`
  },
})

function decodeRouteID(routeID) {
  return atob(routeID)
}
</script>
