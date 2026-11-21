<template>
  <li class="d-flex mt-1 mx-1">
    <button
      class="btn btn-outline-secondary btn-sm flex-grow-1"
      :disabled="!hasPrevPage"
      @click="$emit('prev')"
    >
      <font-awesome-icon
        :icon="['fas', 'chevron-left']"
      />
    </button>
    <button
      class="btn btn-outline-secondary btn-sm flex-grow-1 ms-1"
      :disabled="!hasNextPage"
      @click="$emit('next')"
    >
      <font-awesome-icon
        :icon="['fas', 'chevron-right']"
      />
    </button>
  </li>
</template>

<script setup>
defineProps({
  hasPrevPage: {
    type: Boolean,
    required: true,
  },
  hasNextPage: {
    type: Boolean,
    required: true,
  },
})

defineEmits(['prev', 'next'])
</script>
