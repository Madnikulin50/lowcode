<template>
  <button
    :title="(labels.search || 'Search')"
    type="button"
    class="btn btn-outline-extra-light btn-lg nav-icon rounded-circle text-center border-0 d-flex align-items-center justify-content-center"
    @click="openSearch"
  >
    <font-awesome-icon
      :icon="['fas', 'search']"
      class="text-dark"
    />
  </button>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  labels?: Record<string, string>
}>(), {
  labels: () => ({}),
})

function openSearch() {
  window.dispatchEvent(new CustomEvent('search:open'))
}
</script>

<style lang="scss" scoped>
$nav-icon-size: calc(var(--topbar-height) - 24px);

.nav-icon {
  width: $nav-icon-size !important;
  height: $nav-icon-size !important;
  padding: 0 !important;
}
</style>
