<script setup>
import { computed } from 'vue'

// Either a flat `text` row, or word-level highlight `spans` (from
// wordDiffSpans) for a 1:1 changed-line pair — same idiom as the original
// lineRow()/lineRowHi().
const props = defineProps({
  kind: String, // 'del' | 'add' | 'ctx'
  marker: String,
  text: { type: String, default: null },
  spans: { type: Array, default: null },
})

// Each span after the first needs a plain (unhighlighted) separating space —
// computed here as real text content so it survives Vue's template
// whitespace handling untouched.
const items = computed(() => props.spans ? props.spans.map((s, idx) => ({ ...s, sep: idx > 0 ? ' ' : '' })) : null)
const hiClass = props.kind === 'del' ? 'tdiff-del' : 'tdiff-add'
</script>

<template>
  <div class="tdiff-row" :class="kind">
    <span class="marker">{{ marker }}</span>
    <span class="content">
      <template v-if="items">
        <template v-for="(s, idx) in items" :key="idx">{{ s.sep }}<span v-if="s.hi" :class="hiClass">{{ s.t }}</span><template v-else>{{ s.t }}</template></template>
      </template>
      <template v-else>{{ text || ' ' }}</template>
    </span>
  </div>
</template>
