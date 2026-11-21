<template>
  <div class="container-fluid">
    <div class="row">
      <div class="col-3">
        <div class="list-group">
          <button
            v-for="element in items"
            :key="element.kind"
            class="list-group-item list-group-item-action text-truncate"
            @mouseover="currentValue = element.value"
            @click="$emit('select', element.kind)"
          >
            {{ element.label }}
          </button>
        </div>
      </div>
      <div v-if="currentValue" class="col-9">
        <img v-if="displayMode === 'image'" :src="currentValue" class="img-fluid img-thumbnail" />
        <p v-else-if="displayMode === 'text'">{{ currentValue }}</p>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, watch } from 'vue'

defineProps({
  items: { type: Array, required: true },
  displayMode: { type: String, default: 'image' },
})
defineEmits(['select'])

const currentValue = ref(undefined)

watch(() => props.items, (items = []) => {
  const { value } = items[0]
  if (value) currentValue.value = value
}, { immediate: true })
</script>