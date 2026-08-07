<template>
  <div class="c-toolbar d-flex flex-column flex-sm-row bg-white gap-1 flex-nowrap overflow-auto">
    <div
      class="d-flex align-items-center gap-1 flex-nowrap flex-shrink-0"
    >
      <slot name="start" />
    </div>

    <div
      class="d-flex align-items-center gap-1 flex-nowrap flex-shrink-0"
    >
      <slot name="center" />
    </div>

    <div
      class="d-flex align-items-center gap-1 flex-nowrap flex-shrink-0"
    >
      <slot name="end" />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.c-toolbar {
  z-index: 1036;
  position: relative;
  min-height: 42px;
  padding: 0.75rem;
  flex-shrink: 0;
}

.fill-width > * {
  flex: 1;
}

@media (min-width: 576px) {
  .c-toolbar > div:last-child {
    margin-left: auto !important;
  }
}
</style>
