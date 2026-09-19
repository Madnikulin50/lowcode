<template>
  <div>
    <div class="mb-3">
      <label class="form-label">Поля записи для запроса (коды через запятую)</label>
      <input
        :value="(options.fields || []).join(', ')"
        class="form-control"
        placeholder="store_name, address"
        @change="setFields($event.target.value)"
      >
      <small class="form-text text-muted">
        Значения этих полей текущей записи склеиваются в поисковый запрос — например,
        название и адрес магазина.
      </small>
    </div>

    <div class="mb-3">
      <label class="form-label">Дополнительный текст запроса (необязательно)</label>
      <input
        v-model="options.extraQuery"
        class="form-control"
        placeholder="фасад магазина"
      >
    </div>

    <div class="mb-3">
      <label class="form-label">Число результатов</label>
      <input
        v-model.number="options.limit"
        type="number"
        min="1"
        max="20"
        class="form-control"
        style="max-width:120px"
      >
    </div>

    <p v-if="moduleFields.length" class="text-secondary small mb-0">
      Доступные поля модуля «{{ module?.name }}»: {{ moduleFields.map(f => f.name).join(', ') }}
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

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

const moduleFields = computed(() => props.module?.fields || [])

function setFields (raw) {
  options.value.fields = raw.split(',').map((s) => s.trim()).filter(Boolean)
}
</script>
