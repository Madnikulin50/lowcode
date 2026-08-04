<template>
  <div class="d-flex align-items-center flex-wrap gap-1 ml-8">
    <div class="btn-group" role="group">
      <input
        :id="uid + '0'"
        v-model="selectedValue"
        type="radio"
        class="btn-check"
        :name="uid"
        :value="0"
        autocomplete="off"
      >
      <label
        :for="uid + '0'"
        class="btn btn-outline-primary btn-sm"
      >{{ excludedLabel }}</label>

      <input
        :id="uid + '1'"
        v-model="selectedValue"
        type="radio"
        class="btn-check"
        :name="uid"
        :value="1"
        autocomplete="off"
      >
      <label
        :for="uid + '1'"
        class="btn btn-outline-primary btn-sm"
      >{{ inclusiveLabel }}</label>

      <input
        :id="uid + '2'"
        v-model="selectedValue"
        type="radio"
        class="btn-check"
        :name="uid"
        :value="2"
        autocomplete="off"
      >
      <label
        :for="uid + '2'"
        class="btn btn-outline-primary btn-sm"
      >{{ exclusiveLabel }}</label>
    </div>
    <span class="text-nowrap ml-2">{{ label }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: { type: Number, default: 0 },
  label: { type: String, default: '' },
  excludedLabel: { type: String, default: 'excluded' },
  inclusiveLabel: { type: String, default: 'inclusive' },
  exclusiveLabel: { type: String, default: 'exclusive' },
})

const emit = defineEmits(['update:modelValue', 'change'])

const uid = 'rbs-' + Math.random().toString(36).slice(2, 8)

const selectedValue = computed({
  get() { return props.modelValue },
  set(v) {
    emit('update:modelValue', Number(v))
    emit('change')
  },
})
</script>
