<template>
  <button
    v-if="list.length > 0"
    data-test-id="button-export"
    class="btn btn-outline-secondary"
    @click="jsonExport(list, type)"
  >
    {{ $t('label.export') }}
  </button>
</template>

<script setup lang="js">
import { saveAs } from 'file-saver'
import { useStore } from '../../store'

defineOptions({
  i18nOptions: {
    namespaces: 'general',
  },
})

const props = defineProps({
  list: {
    type: Array,
    required: true,
  },
  type: {
    type: String,
    required: true,
  },
})

const store = useStore()

function findModuleByID (...args) {
  return store.dispatch('module/findByID', ...args)
}

function jsonExport (list, type) {
  Promise.all(list.map(i => i.export(findModuleByID))).then(list => {
    const blob = new Blob([JSON.stringify({ type, list }, null, 2)], { type: 'application/json' })
    saveAs(blob, `${type}-export.json`)
  })
}
</script>
