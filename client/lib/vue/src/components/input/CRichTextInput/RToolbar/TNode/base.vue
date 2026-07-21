<script setup lang="ts">
import { nodeTypes } from '../../lib/formats'

const props = defineProps<{
  editor: Record<string, any>
  format: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'click', value: { type: string; attrs: Record<string, any> }): void
}>()

function onClick(type: string, attrs?: Record<string, any>) {
  if (!attrs) {
    attrs = {}
  }

  const act = activeNode([type], attrs)
  if (act) {
    type = 'paragraph'
  }

  const nn = activeNode(nodeTypes)
  if (!nn) {
    throw new Error('no node selected')
  }
  const n = nn.node

  const cAttr = n.attrs
  const target = props.editor.schema.nodes[type]
  const nAttrs: Record<string, any> = {}

  const targetAttrs = (target && target.spec && target.spec.attrs) ? Object.keys(target.spec.attrs) : []
  for (const a of targetAttrs) {
    if (attrs[a] === undefined) {
      nAttrs[a] = cAttr[a]
    } else {
      nAttrs[a] = attrs[a]
    }
  }

  emit('click', { type, attrs: nAttrs })
}

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

function isActiveCheck(types: string[], attrs?: Record<string, any>) {
  return !!activeNode(types, attrs)
}

function activeClasses(attrs?: Record<string, any>) {
  if (isActiveCheck([props.format.type], attrs)) {
    return ['text-primary']
  }

  return undefined
}
</script>
