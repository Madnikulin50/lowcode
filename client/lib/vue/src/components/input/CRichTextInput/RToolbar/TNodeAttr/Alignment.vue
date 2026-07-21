<template>
  <div class="dropdown">
    <button
      class="btn btn-link dropdown-toggle text-dark fw-bold"
      data-bs-toggle="dropdown"
      aria-expanded="false"
    >
      <span class="text-dark fw-bold">
        <span :class="{ 'text-primary': !!activeType && activeType !== 'left' }">
          <font-awesome-icon
            v-if="activeIcon"
            :icon="['fas', activeIcon]"
          />
          <span v-else>
            {{ format.label }}
          </span>
        </span>
      </span>
    </button>

    <ul class="dropdown-menu text-center bg-white">
      <li
        v-for="v of format.variants"
        :key="v.variant"
      >
        <button
          class="dropdown-item"
          @click="$emit('click', { type: 'alignment', attrs: v.attrs })"
        >
          <font-awesome-icon
            v-if="format.icon"
            :icon="v.icon"
          />
          <span v-else>
            {{ v.label }}
          </span>
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  editor: any
  format: any
  isActive?: Record<string, any>
}>()

defineEmits<{
  (e: 'click', payload: { type: string; attrs: Record<string, any> }): void
}>()

const activeType = computed(() => {
  const alignments = ['left', 'center', 'right', 'justify']
  return alignments.find(alignment =>
    props.editor.isActive({ textAlign: alignment }),
  )
})

const activeIcon = computed(() => {
  const alignmentMap: Record<string, string> = {
    left: 'align-left',
    center: 'align-center',
    right: 'align-right',
    justify: 'align-justify',
  }
  return alignmentMap[activeType.value] || 'align-left'
})
</script>
