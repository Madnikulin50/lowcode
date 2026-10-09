<script setup>
// Unified, git-diff-style page text: PD is "a", RD is "b", removed lines get
// a red "−" row, added lines a green "+" row, matching lines are dimmed
// context — same layout `git diff`/GitHub give two versions of a source
// file. Long unchanged stretches collapse behind a click-to-expand fold
// (like `git diff -U3`), and a clean 1:1 line replacement gets word-level
// highlighting inside the row instead of a flat color. Ported from the
// original vanilla-JS buildTextDiff().
import { computed } from 'vue'
import { pageText } from '../composables/useComparison.js'
import { textLines, diffSeq, wordDiffSpans, buildBlocks } from '../composables/useTextDiff.js'
import { DIFF_CONTEXT } from '../constants.js'
import DiffLine from './DiffLine.vue'

const props = defineProps({
  pdPages: Object,
  rdPages: Object,
  page: Number,
  foldState: Object,
  expandFold: Function,
})

const pdText = computed(() => pageText(props.pdPages, props.page))
const rdText = computed(() => pageText(props.rdPages, props.page))
const isEmpty = computed(() => !pdText.value && !rdText.value)

const blocks = computed(() => {
  if (isEmpty.value) return []
  const ops = diffSeq(textLines(pdText.value), textLines(rdText.value))
  return buildBlocks(ops)
})
const isFullMatch = computed(() => blocks.value.length === 1 && blocks.value[0].kind === 'ctx')

// Pre-computes everything the template needs per block (paired word-diffs,
// or which context lines are hidden behind a fold) so the template itself
// stays a plain data walk with no logic in it.
const displayBlocks = computed(() => blocks.value.map((block, blockIdx) => {
  if (block.kind === 'change') {
    const paired = block.dels.length === block.adds.length
    return {
      kind: 'change',
      paired,
      pairs: paired ? block.dels.map((d, k) => wordDiffSpans(d, block.adds[k])) : null,
      dels: block.dels,
      adds: block.adds,
    }
  }
  const isFirst = blockIdx === 0
  const isLast = blockIdx === blocks.value.length - 1
  const headN = isFirst ? 0 : DIFF_CONTEXT
  const tailN = isLast ? 0 : DIFF_CONTEXT
  const key = props.page + ':' + blockIdx
  const expanded = !!props.foldState[key]
  const hiddenCount = block.lines.length - headN - tailN
  const folded = !expanded && hiddenCount > 0
  return {
    kind: 'ctx',
    key,
    folded,
    hiddenCount,
    head: folded ? block.lines.slice(0, headN) : [],
    tail: folded ? block.lines.slice(block.lines.length - tailN) : [],
    all: folded ? [] : block.lines,
  }
}))
</script>

<template>
  <div>
    <div class="text-legend">
      <span><span class="swatch del"></span>только в ПД</span>
      <span><span class="swatch add"></span>только в РД</span>
    </div>

    <div v-if="isEmpty" class="empty">На этой странице нет извлечённого текста ни в одном из файлов.</div>

    <div v-else-if="isFullMatch" class="text-diff">
      <div class="tdiff-match">✓ Текст страницы совпадает построчно.</div>
    </div>

    <div v-else class="text-diff">
      <template v-for="(block, blockIdx) in displayBlocks" :key="blockIdx">
        <template v-if="block.kind === 'change'">
          <template v-if="block.paired">
            <template v-for="(pair, k) in block.pairs" :key="k">
              <DiffLine kind="del" marker="−" :spans="pair.delSpans" />
              <DiffLine kind="add" marker="+" :spans="pair.addSpans" />
            </template>
          </template>
          <template v-else>
            <DiffLine v-for="(l, k) in block.dels" :key="'d' + k" kind="del" marker="−" :text="l" />
            <DiffLine v-for="(l, k) in block.adds" :key="'a' + k" kind="add" marker="+" :text="l" />
          </template>
        </template>
        <template v-else>
          <template v-if="block.folded">
            <DiffLine v-for="(l, k) in block.head" :key="'h' + k" kind="ctx" marker=" " :text="l" />
            <div class="tdiff-fold" @click="expandFold(block.key)">⋯ {{ block.hiddenCount }} совпадающих строк без изменений — показать ⋯</div>
            <DiffLine v-for="(l, k) in block.tail" :key="'t' + k" kind="ctx" marker=" " :text="l" />
          </template>
          <template v-else>
            <DiffLine v-for="(l, k) in block.all" :key="k" kind="ctx" marker=" " :text="l" />
          </template>
        </template>
      </template>
    </div>
  </div>
</template>
