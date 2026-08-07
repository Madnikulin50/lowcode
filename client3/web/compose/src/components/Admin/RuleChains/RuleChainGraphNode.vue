<template>
  <div
    class="rule-node"
    :class="{ selected, entry: data.entry }"
  >
    <Handle
      type="target"
      :position="Position.Left"
    />
    <button
      class="rule-node-delete"
      title="Delete node"
      @click.stop="$emit('delete', id)"
    >×</button>
    <div class="rule-node-header">
      <span
        v-if="data.entry"
        class="badge bg-success me-1"
        title="Entry node"
      >▶</span>
      <span class="rule-node-type">{{ nodeTypeLabel || data.nodeType || '?' }}</span>
    </div>
    <div class="rule-node-body">
      {{ data.label || id }}
    </div>
    <Handle
      type="source"
      :position="Position.Right"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'

const props = defineProps({
  id: { type: String, required: true },
  data: { type: Object, default: () => ({}) },
  selected: { type: Boolean, default: false },
  entryNode: { type: String, default: '' },
})

const nodeTypeLabel = computed(() => {
  if (!props.data.nodeType) return ''
  const caps = props.data.nodeType.replace(/([A-Z])/g, ' $1')
  return caps.charAt(0).toUpperCase() + caps.slice(1)
})
</script>

<style lang="scss" scoped>
.rule-node {
  background: #fff;
  border: 2px solid var(--bs-border-color, #dee2e6);
  border-radius: 8px;
  min-width: 160px;
  font-size: 0.8125rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  transition: box-shadow 0.15s, border-color 0.15s;

  &:hover {
    box-shadow: 0 2px 8px rgba(0,0,0,0.12);
    .rule-node-delete { opacity: 1; }
  }

  .rule-node-delete {
    position: absolute;
    top: -6px;
    right: -6px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 1px solid var(--bs-border-color, #dee2e6);
    background: #fff;
    color: var(--bs-danger, #dc3545);
    font-size: 12px;
    line-height: 1;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.15s;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10;
    &:hover { background: var(--bs-danger, #dc3545); color: #fff; }
  }

  &.selected {
    border-color: var(--bs-primary, #0d6efd);
    box-shadow: 0 0 0 2px rgba(13, 110, 253, 0.2);
  }

  &.entry {
    border-color: var(--bs-success, #198754);
    &.selected {
      box-shadow: 0 0 0 2px rgba(25, 135, 84, 0.2);
    }
  }

  .rule-node-header {
    background: #f8f9fa;
    padding: 4px 10px;
    border-bottom: 1px solid var(--bs-border-color-translucent, rgba(0,0,0,0.05));
    font-size: 0.6875rem;
    font-weight: 600;
    color: var(--bs-secondary-color, #6c757d);
    text-transform: uppercase;
    letter-spacing: 0.025em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .rule-node-body {
    padding: 6px 10px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
</style>
