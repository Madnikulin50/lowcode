<template>
  <div class="container-fluid">
    <div class="row">
      <div class="col-3">
        <ul class="list-group list-group-flush">
          <li
            v-for="element in items"
            :key="element.kind"
            class="list-group-item text-truncate"
            style="cursor: pointer;"
            @mouseover="currentValue = element.value"
            @click="$emit('select', element.kind)"
          >
            {{ element.label }}
          </li>
        </ul>
      </div>
      <div
        v-if="currentValue"
        class="col-9"
        :class="{ 'my-auto': displayMode === 'image' }"
      >
        <img
          v-if="displayMode === 'image'"
          class="img-fluid img-thumbnail"
          :src="currentValue"
          alt=""
        />

        <p
          v-else-if="displayMode === 'text'"
        >
          {{ currentValue }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  items: {
    type: Array,
    required: true,
  },
  displayMode: {
    type: String,
    default: 'image',
  },
})

defineEmits(['select'])

const currentValue = ref(undefined)

watch(() => props.items, (items = []) => {
  const { value } = items[0]
  if (value) {
    currentValue.value = value
  }
}, { immediate: true })
</script>
