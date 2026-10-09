<template>
  <tr class="w-100 d-inline-block">
    <td
      title=""
      class="handle align-middle w-100 d-inline-block"
      :class="{ 'cursor-grab': !fixed }"
    >
      <font-awesome-icon
        v-if="!fixed"
        :icon="['fas', 'bars']"
        class="text-secondary me-1"
      />
      <slot
        name="report-label"
        :report="report"
      />
      <button
        type="button"
        class="btn btn-link ms-1 float-end text-dark p-0"
        @click="$emit('edit', report)"
      >
        <font-awesome-icon :icon="['far', 'edit']" />
      </button>

      <c-input-confirm
        show-icon
        class="float-end"
        @confirmed="$emit('remove', report)"
      />
    </td>
  </tr>
</template>

<script setup>
defineProps({
  report: {
    type: Object,
    required: true,
    default: () => ({}),
  },
  fixed: {
    type: Boolean,
    required: false,
    default: false,
  },
})

defineEmits(['edit', 'remove'])
</script>

<style scoped lang="scss">
.cursor-grab {
  cursor: grab;
}
</style>
