<template>
  <div
    v-if="buttons.length"
    class="btn-group"
    :class="sizeClass"
  >
    <button
      v-for="(b, i) in buttons"
      :key="i"
      class="btn"
      :class="[b.variant || defaultVariant, buttonClass]"
      @click="$emit('click', b)"
    >
      {{ b.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance } from 'vue'

const { proxy }: any = getCurrentInstance()!

const props = withDefaults(defineProps<{
  resourceType: string
  uiSlot: string
  uiPage: string
  defaultVariant?: string
  buttonClass?: string
  size?: string
}>(), {
  defaultVariant: 'link',
  buttonClass: 'me-1',
  size: 'md',
})

defineEmits<{
  click: [value: any]
}>()

const buttons = computed(() => proxy.$UIHooks.Find(props.resourceType, props.uiPage, props.uiSlot))

const sizeClass = computed(() => {
  if (props.size === 'sm') return 'btn-group-sm'
  if (props.size === 'lg') return 'btn-group-lg'
  return ''
})
</script>
