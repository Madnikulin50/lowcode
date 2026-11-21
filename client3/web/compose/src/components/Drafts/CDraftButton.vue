<template>
  <button
    class="btn btn-outline-extra-light btn-lg nav-icon rounded-circle text-center border-0 d-flex align-items-center justify-content-center position-relative"
    title="Drafts"
    @click="toggleDrafts"
  >
    <font-awesome-icon
      :icon="['far', 'file']"
      class="text-dark"
    />
    <span
      v-if="draftCount > 0"
      class="badge rounded-pill bg-primary position-absolute draft-badge"
    >
      {{ draftCount > 9 ? '9+' : draftCount }}
    </span>
  </button>
</template>

<script setup>
import { computed } from 'vue'
import { useDraftsStore } from 'corteza-lib/vue/dist'

const draftsStore = useDraftsStore()

const drafts = computed(() => draftsStore.getAllDrafts)
const draftCount = computed(() => (drafts.value || []).length)

function toggleDrafts() {
  draftsStore.toggleVisibility()
}
</script>

<style lang="scss" scoped>
.draft-badge {
  top: 0;
  right: 0;
  transform: translate(25%, -25%);
  font-size: 0.7rem;
}
</style>
