<template>
  <button
    class="btn"
    :class="`btn-${variant}`"
    :style="size === 'lg' ? 'font-size: 1.25rem;' : size === 'sm' ? 'font-size: 0.875rem;' : ''"
    @click="jsonExport(workflows)"
  >
    <slot />
    {{ $t('export') }}
  </button>
</template>

<script setup>
import { inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import { saveAs } from 'file-saver'

const { t } = useI18n()
const toast = useToast()
const $AutomationAPI = inject('$AutomationAPI', {})

const props = defineProps({
  workflows: { type: Array, default: () => ([]) },
  fileName: { type: String, default: 'workflows-export' },
  size: { type: String, default: 'md' },
  variant: { type: String, default: 'light' },
})

async function jsonExport(workflowID = []) {
  const triggers = {}
  let workflows = []

  await $AutomationAPI.triggerList({ workflowID, disabled: 1 })
    .then(({ set = [] }) => {
      set.forEach(({ workflowID, resourceType, eventType, constraints, enabled, stepID, meta }) => {
        if (!triggers[workflowID]) {
          triggers[workflowID] = []
        }
        triggers[workflowID].push({ resourceType, eventType, constraints, enabled, stepID, meta })
      })
    })
    .catch((e) => toast.error(t('notification.failed-fetch-triggers'), e))

  await $AutomationAPI.workflowList({ workflowID, disabled: 1, subWorkflow: 1 })
    .then(({ set = [] }) => {
      workflows = set.map(({ workflowID, handle, enabled, keepSessions, steps, paths, meta }) => {
        return { handle, enabled, meta, keepSessions, steps, paths, triggers: triggers[workflowID] }
      })
    })
    .catch((e) => toast.error(t('notification.failed-fetch-workflows'), e))

  const blob = new Blob([JSON.stringify({ workflows }, null, 2)], { type: 'application/json' })
  const filename = props.fileName.replace(/[/\\?%*:|"<>]/g, '')
  saveAs(blob, `${filename}.json`)
}
</script>
