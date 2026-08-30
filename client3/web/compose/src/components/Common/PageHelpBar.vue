<template>
  <div
    v-if="hasAny"
    class="page-help-bar d-flex align-items-start gap-2 px-3 py-2 mb-2 rounded border bg-white"
  >
    <div class="flex-grow-1 min-w-0">
      <div
        v-if="description"
        class="text-body"
      >
        {{ description }}
      </div>
      <div
        v-else
        class="text-muted small"
      >
        {{ emptyLabel }}
      </div>
    </div>
    <button
      type="button"
      class="btn btn-sm btn-outline-primary text-nowrap d-flex align-items-center gap-1 flex-shrink-0"
      @click="open = true"
    >
      <font-awesome-icon :icon="['far', 'question-circle']" />
      {{ openLabel }}
    </button>
    <c-help-panel
      v-model:open="open"
      v-bind="triggerProps"
    />
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useHelp } from '../../composables/useHelp'
import { pageHelpDocs } from '../../help/appDocs'

const props = defineProps({
  page: { type: Object, required: true },
  namespace: { type: Object, default: null },
})

const { t } = useI18n({ useScope: 'global' })
const open = ref(false)

const docs = computed(() => pageHelpDocs(props.namespace, props.page))
const { hasAny, triggerProps, app } = useHelp('compose.page.view', docs)

const description = computed(() => app.value.description)

function label (key, fallback) {
  const value = t(key)
  return !value || value === key ? fallback : value
}

const openLabel = computed(() => label('help.bar.open', 'Справка'))
const emptyLabel = computed(() => label('help.bar.empty', 'Откройте справку, чтобы узнать, как пользоваться этой страницей'))
</script>
