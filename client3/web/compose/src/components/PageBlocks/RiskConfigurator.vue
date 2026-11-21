<template>
  <div>
    <div class="mb-3">
      <label class="form-label">Привязка модели риска</label>
      <select v-model="options.bindingID" class="form-select">
        <option value="">— выберите —</option>
        <option v-for="b in bindings" :key="b.bindingID" :value="b.bindingID">
          Модуль #{{ b.moduleID }} — модель #{{ b.modelID }}
        </option>
      </select>
      <small class="form-text text-muted">
        Заводится в разделе «Реестр рисков» (Admin → Risk → Registry) для нужного namespace.
      </small>
    </div>

    <div class="mb-3">
      <label class="form-label">Режим</label>
      <select v-model="options.mode" class="form-select">
        <option value="record">Текущая запись (карточка записи)</option>
        <option value="portfolio">Портфель — все объекты привязки (отчёт/дашборд)</option>
      </select>
    </div>

    <div v-if="options.mode === 'portfolio'" class="mb-3">
      <label class="form-label">Показывать топ N по остаточному риску</label>
      <input v-model.number="options.topN" type="number" class="form-control" style="max-width:120px" />
    </div>

    <div class="form-check form-switch mb-2">
      <input id="risk-show-attribution" v-model="options.showAttribution" class="form-check-input" type="checkbox" role="switch" />
      <label class="form-check-label" for="risk-show-attribution">Показывать разложение по факторам (SHAP)</label>
    </div>
    <div v-if="options.mode === 'record'" class="form-check form-switch mb-3">
      <input id="risk-show-explanation" v-model="options.showExplanation" class="form-check-input" type="checkbox" role="switch" />
      <label class="form-check-label" for="risk-show-explanation">Показывать текстовое объяснение</label>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, inject } from 'vue'

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  namespace: { type: Object, required: true },
})

const options = computed(() => {
  if (!props.block.options || typeof props.block.options !== 'object') {
    props.block.options = {}
  }
  return props.block.options
})

const $ComposeAPI = inject('$ComposeAPI', window.__composeAPI)
const bindings = ref([])

async function loadBindings () {
  if (!$ComposeAPI) return
  try {
    const { bindings: set } = await $ComposeAPI.riskBindingList({ namespaceID: props.namespace?.namespaceID })
    bindings.value = set || []
  } catch {
    bindings.value = []
  }
}

onMounted(loadBindings)
watch(() => props.namespace?.namespaceID, loadBindings)
</script>
