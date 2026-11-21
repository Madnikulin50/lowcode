<template>
  <div class="dropdown">
    <button
      class="btn btn-link dropdown-toggle text-dark fw-bold"
      data-bs-toggle="dropdown"
      aria-expanded="false"
    >
      <span class="text-dark fw-bold">
        <font-awesome-icon
          v-if="format.icon"
          :icon="format.icon"
        />
        <span v-else>
          {{ format.label }}
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
          @click="emitClick(v)"
        >
          {{ v.label }}
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { nodeTypes } from '../../lib/formats'

const props = defineProps<{
  editor: any
  format: any
  isActive?: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'click', payload: { type: string; attrs: Record<string, any> }): void
}>()

function activeNodes(types: string[], attrs?: Record<string, any>) {
  const ed = props.editor
  const rtr: any[] = []
  ed.state.doc.nodesBetween(
    ed.state.selection.from,
    ed.state.selection.to,
    (n: any, pos: number) => {
      if (types.includes(n.type.name)) {
        if (attrs) {
          if (!Object.entries(attrs || {}).find(([k, v]) => n.attrs[k] !== v)) {
            rtr.push({ node: n, position: pos })
          }
        } else {
          rtr.push({ node: n, position: pos })
        }
      }
    },
  )

  return rtr
}

function activeNode(types: string[], attrs?: Record<string, any>) {
  const ann = activeNodes(types, attrs)
  if (!ann) {
    return undefined
  }
  return ann[0]
}

function activeClasses(attrs?: Record<string, any>) {
  const an = activeNode(nodeTypes, attrs)
  if (!an || !an.node) {
    return undefined
  }

  const ac = (type: string, attrs: Record<string, any>) => {
    return props.editor.isActive(type, attrs)
  }

  if (ac(an.node.type.name, { ...an.node.attrs, ...attrs })) {
    return ['text-primary']
  }

  return undefined
}

function dispatchTransaction(v: any) {
  const ann = activeNodes(nodeTypes)
  const tr = props.editor.state.tr
  for (const an of ann) {
    tr.setNodeMarkup(an.position, an.node.type, { ...an.node.attrs, ...v.attrs })
  }
  props.editor.dispatchTransaction(tr)
}

function emitClick(v: any) {
  emit('click', { type: v.type, attrs: { ...v.attrs } })
}

function rootActiveClasses(v: any) {
  if (props.format.variants.find(({ type, attrs }: any) => activeClasses(attrs))) {
    return ['text-primary']
  }
}
</script>
