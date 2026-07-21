<template>
  <div
    v-if="items.length"
    class="emoji-dropdown"
  >
    <button
      v-for="(item, index) in items"
      :key="item.name"
      type="button"
      :class="[
        'emoji-option',
        { 'emoji-option--highlighted': index === selectedIndex }
      ]"
      @click="handleClick(index)"
      @mouseenter="selectedIndex = index"
    >
      <span class="emoji-option-emoji">{{ item.emoji }}</span>
      <span class="emoji-option-name">:{{ item.name }}:</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  items: any[]
  command: Function
}>()

const selectedIndex = ref(0)

watch(() => props.items, () => {
  selectedIndex.value = 0
})

function onKeyDown({ event }: { event: KeyboardEvent }) {
  if (event.key === 'ArrowUp') {
    upHandler()
    return true
  }

  if (event.key === 'ArrowDown') {
    downHandler()
    return true
  }

  if (event.key === 'Enter') {
    enterHandler()
    return true
  }

  return false
}

function upHandler() {
  selectedIndex.value = ((selectedIndex.value + props.items.length) - 1) % props.items.length
}

function downHandler() {
  selectedIndex.value = (selectedIndex.value + 1) % props.items.length
}

function enterHandler() {
  selectItem(selectedIndex.value)
}

function handleClick(index: number) {
  selectItem(index)
}

function selectItem(index: number) {
  const item = props.items[index]

  if (item) {
    props.command({ name: item.name })
  }
}

defineExpose({ onKeyDown })
</script>

<style lang="scss" scoped>
.emoji-dropdown {
  background: var(--white);
  border: 1px solid var(--extra-light);
  border-radius: 0.25rem;
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
  max-height: 200px;
  overflow-y: auto;
  font-size: 0.9rem;
  font-family: var(--font-regular);
}

.emoji-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  background: var(--white);
  color: var(--black);
  padding: 0.35rem 0.75rem;
  border: none;
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;

  &:hover,
  &.emoji-option--highlighted {
    background: var(--light);
    color: var(--black);
  }

  &:active,
  &:focus {
    color: var(--white);
    background-color: var(--primary);
    outline: none;
  }
}

.emoji-option-emoji {
  font-size: 1.2em;
  line-height: 1;
}

.emoji-option-name {
  color: var(--secondary);
  font-size: 0.85em;
}
</style>
