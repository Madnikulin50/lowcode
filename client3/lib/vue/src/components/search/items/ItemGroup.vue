<template>
  <div
    v-if="items.length > 0"
  >
    <div
      :data-bs-toggle="'collapse'"
      :data-bs-target="`#${collapseId}`"
      class="group-header d-flex align-items-center justify-content-between text-muted shadow-sm"
      :class="{ 'bg-white border-left border-bottom p-2 pe-3': subgroup, 'bg-light p-2 pe-3': !subgroup }"
      :style="{ top: subgroup ? '36px' : '0', zIndex: subgroup ? 10 : 20 }"
      @click="toggleCollapse"
    >
      <div class="d-flex align-items-center gap-1">
        <span
          class="badge"
          :class="subgroup ? 'bg-light text-dark' : 'bg-primary'"
          style="font-size: 90%;"
        >
          {{ title }}
        </span>
        <span
          v-if="labels.numberOfResults"
          class="small text-muted"
        >
          {{ labels.numberOfResults(items.length) }}
        </span>
      </div>
      <font-awesome-icon
        :icon="['fas', 'chevron-right']"
        class="chevron-icon ms-2 small"
        :class="{ 'chevron-collapsed': !internalExpanded }"
      />
    </div>
    <div
      :id="collapseId"
      ref="collapseEl"
      class="collapse border-bottom"
      :class="{ show: internalExpanded }"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faChevronRight } from '@fortawesome/free-solid-svg-icons'
import { Collapse } from 'bootstrap'

library.add(faChevronRight)

const props = defineProps<{
  title: string
  items: any[]
  collapseId: string
  expanded?: boolean
  subgroup?: boolean
  labels?: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'update:expanded', value: boolean): void
}>()

const internalExpanded = ref(props.expanded ?? true)
const collapseEl = ref<HTMLElement | null>(null)
let bsCollapse: Collapse | null = null

watch(() => props.expanded, (val) => {
  internalExpanded.value = val ?? true
})

watch(internalExpanded, (val) => {
  emit('update:expanded', val)
})

function toggleCollapse() {
  internalExpanded.value = !internalExpanded.value
}

onMounted(() => {
  if (collapseEl.value) {
    bsCollapse = new Collapse(collapseEl.value, { toggle: false })
    if (!internalExpanded.value) {
      bsCollapse.hide()
    }
  }
})
</script>

<style lang="scss" scoped>
.group-header {
  position: sticky;
  cursor: pointer;
  transition: all 0.2s ease;

  .chevron-icon {
    transition: transform 0.2s ease;
    transform: rotate(90deg);
  }

  .chevron-collapsed {
    transform: rotate(0deg);
  }

  &:hover {
    .chevron-icon {
      color: var(--primary) !important;
    }
  }
}
</style>
