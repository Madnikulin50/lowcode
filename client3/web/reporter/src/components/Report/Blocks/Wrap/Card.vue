<template>
  <div class="h-100 p-2 overflow-hidden">
    <div class="h-100 border-0 shadow-sm overflow-hidden d-flex flex-column">
      <div v-if="$slots.header" class="sticky-top border-0 p-0">
        <slot name="header" />
      </div>
      <div v-if="$slots.default" class="p-0 h-100">
        <slot />
      </div>
    </div>
  </div>
</template>