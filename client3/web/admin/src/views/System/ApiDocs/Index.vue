<template>
  <div class="api-docs">
    <c-content-header :title="title">
      <a
        class="btn btn-sm btn-outline-primary"
        :href="docsUrl"
        target="_blank"
        rel="noopener noreferrer"
      >
        {{ openLabel }}
      </a>
    </c-content-header>
    <iframe
      class="api-docs__frame"
      :src="docsUrl"
      :title="title"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useNsI18n } from 'corteza-lib/vue/dist'

defineOptions({ i18nOptions: { namespaces: 'navigation' } })

const t = useNsI18n()

function tr (key, fallback) {
  const v = t(key)
  if (!v || v === key || v.endsWith(`.${key}`)) return fallback
  return v
}

const title = computed(() => tr('system.items.apidocs', 'API docs'))
const openLabel = computed(() => tr('openApiDocs', 'Open'))

const docsUrl = computed(() => {
  const base = String(window.CortezaAPI || '/api').replace(/\/$/, '')
  return `${base}/docs/`
})
</script>

<style scoped>
.api-docs {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 64px);
  height: 100%;
}

.api-docs__frame {
  flex: 1 1 auto;
  width: 100%;
  min-height: calc(100vh - 64px);
  border: 0;
  background: #fafafa;
}
</style>
