<template>
  <div class="c-emoji-picker">
    <div class="c-emoji-picker-search-wrap">
      <input
        ref="searchInputRef"
        type="text"
        class="c-emoji-picker-search-input"
        :placeholder="labels.search || 'Search'"
        @input="onSearch"
        @click.stop
        @keydown.stop
      >
      <svg
        class="c-emoji-picker-search-icon"
        viewBox="0 0 16 16"
        fill="currentColor"
      >
        <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85zm-5.44 1.156a5 5 0 1 1 0-10 5 5 0 0 1 0 10z" />
      </svg>
    </div>

    <div
      ref="viewportRef"
      class="c-emoji-picker-viewport"
      :style="{ height: viewportHeight + 'px', width: viewportWidth + 'px' }"
      @scroll="onScroll"
    >
      <div
        ref="scrollContentRef"
        class="c-emoji-picker-scroll-content"
      />
    </div>

    <div
      v-if="showQuickReactions"
      class="c-emoji-picker-quick"
    >
      <div class="c-emoji-picker-quick-label">
        {{ labels.quickReactions || 'Quick Reactions' }}
      </div>
      <div class="c-emoji-picker-quick-row">
        <span
          v-for="qr in quickReactionsList"
          :key="qr.emoji"
          class="epi"
          :title="qr.name"
          @click.stop="onQuickReactionClick(qr)"
        >
          {{ qr.emoji }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'

const STORAGE_KEY = 'corteza:emoji:frequently-used'
const MAX_FREQUENT = 18

const ITEM_SIZE = 30
const LABEL_HEIGHT = 26
const ITEMS_PER_ROW = 8
const BUFFER_PX = 90

const EMOJI_BLACKLIST = new Set(['relaxed', 'frowning_face'])

const GROUP_CONFIG = [
  { name: 'smileys-people', icon: '😀', label: 'Smileys & People', sources: ['', 'people & body'] },
  { name: 'animals & nature', icon: '🐶', label: 'Animals & Nature', sources: ['animals & nature'] },
  { name: 'food & drink', icon: '🍎', label: 'Food & Drink', sources: ['food & drink'] },
  { name: 'travel & places', icon: '🚗', label: 'Travel & Places', sources: ['travel & places'] },
  { name: 'activities', icon: '⚽', label: 'Activities', sources: ['activities'] },
  { name: 'objects', icon: '💡', label: 'Objects', sources: ['objects'] },
  { name: 'symbols', icon: '❤️', label: 'Symbols', sources: ['symbols'] },
] as const

interface Emoji {
  emoji: string
  name: string
  shortcodes?: string[]
  tags?: string[]
  group?: string
}

interface VirtualRow {
  type: 'label' | 'emojis' | 'empty'
  text?: string
  items?: Emoji[]
  y: number
  height: number
}

const props = withDefaults(defineProps<{
  emojis?: Emoji[]
  viewportHeight?: number
  viewportWidth?: number
  showFrequent?: boolean
  showQuickReactions?: boolean
  labels?: Record<string, string>
}>(), {
  emojis: () => [],
  viewportHeight: 260,
  viewportWidth: 260,
  showFrequent: true,
  showQuickReactions: true,
  labels: () => ({}),
})

const emit = defineEmits<{
  (e: 'select', emoji: Emoji | { name: string; emoji: string }): void
}>()

const searchInputRef = ref<HTMLInputElement | null>(null)
const viewportRef = ref<HTMLDivElement | null>(null)
const scrollContentRef = ref<HTMLDivElement | null>(null)

const search = ref('')
const frequentlyUsed = ref<string[]>([])
const virtualRows = ref<VirtualRow[]>([])
const totalHeight = ref(0)
const renderedRange = ref({ start: -1, end: -1 })

const allEmojis = computed(() => props.emojis || [])

const emojiByGroup = computed(() => {
  const raw: Record<string, Emoji[]> = {}
  for (const emoji of allEmojis.value) {
    if (emoji.name.startsWith('regional_indicator')) continue
    if (EMOJI_BLACKLIST.has(emoji.name)) continue
    if (!raw[emoji.group!]) raw[emoji.group!] = []
    raw[emoji.group!].push(emoji)
  }

  const map: Record<string, Emoji[]> = {}
  for (const group of GROUP_CONFIG) {
    const merged: Emoji[] = []
    for (const src of group.sources) {
      if (raw[src]) merged.push(...raw[src])
    }
    if (merged.length) map[group.name] = merged
  }
  return map
})

const groups = computed(() => {
  const byGroup = emojiByGroup.value
  return GROUP_CONFIG.filter(g => byGroup[g.name]?.length)
})

const frequentEmojis = computed(() => {
  if (!props.showFrequent) return []
  return frequentlyUsed.value
    .map(name => allEmojis.value.find(e => e.name === name))
    .filter((e): e is Emoji => !!e)
})

const quickReactionsList = computed(() => {
  const names = ['+1', '-1', 'smile', 'tada', 'blush', 'rocket', 'eyes']
  return names
    .map(name => allEmojis.value.find(e => e.name === name))
    .filter((e): e is Emoji => !!e)
})

const filteredEmojis = computed(() => {
  if (!search.value) return []
  const q = search.value.toLowerCase()
  return allEmojis.value.filter((emoji) => {
    if (emoji.name.startsWith('regional_indicator')) return false
    if (EMOJI_BLACKLIST.has(emoji.name)) return false
    if (emoji.name.toLowerCase().includes(q)) return true
    if (emoji.shortcodes && emoji.shortcodes.some(s => s.toLowerCase().includes(q))) return true
    if (emoji.tags && emoji.tags.some(t => t.toLowerCase().includes(q))) return true
    return false
  }).slice(0, 60)
})

onMounted(() => {
  loadFrequentlyUsed()
})

function loadFrequentlyUsed(): void {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    frequentlyUsed.value = stored ? JSON.parse(stored) : []
  } catch {
    frequentlyUsed.value = []
  }
}

function saveFrequentlyUsed(emojiName: string): void {
  const list = [emojiName, ...frequentlyUsed.value.filter(n => n !== emojiName)].slice(0, MAX_FREQUENT)
  frequentlyUsed.value = list
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(list))
  } catch {
    // ignore
  }
}

function onSearch(e: Event): void {
  search.value = (e.target as HTMLInputElement).value
  rebuildVirtualRows()
  renderedRange.value = { start: -1, end: -1 }
  if (viewportRef.value) {
    viewportRef.value.scrollTop = 0
  }
  renderVisible()
}

function onEmojiClick(e: MouseEvent): void {
  const el = (e.target as HTMLElement).closest('[data-emoji]') as HTMLElement | null
  if (el) {
    e.stopPropagation()
    const name = el.dataset.emoji
    const emoji = allEmojis.value.find(em => em.name === name)
    saveFrequentlyUsed(name!)
    emit('select', emoji || { name: name!, emoji: el.textContent || '' })
  }
}

function rebuildVirtualRows(): void {
  const rows: VirtualRow[] = []
  let y = 0

  const addSection = (label: string, emojis: Emoji[]) => {
    rows.push({ type: 'label', text: label, y, height: LABEL_HEIGHT })
    y += LABEL_HEIGHT
    for (let i = 0; i < emojis.length; i += ITEMS_PER_ROW) {
      const chunk = emojis.slice(i, i + ITEMS_PER_ROW)
      rows.push({ type: 'emojis', items: chunk, y, height: ITEM_SIZE })
      y += ITEM_SIZE
    }
  }

  if (search.value) {
    const results = filteredEmojis.value
    if (results.length) {
      addSection(props.labels.searchResults || 'Search Results', results)
    } else {
      rows.push({ type: 'empty', y, height: 60 })
      y += 60
    }
  } else {
    const freq = frequentEmojis.value
    if (freq.length) {
      addSection(props.labels.frequentlyUsed || 'Frequently Used', freq)
    }
    const byGroup = emojiByGroup.value
    for (const group of groups.value) {
      const emojis = byGroup[group.name]
      if (emojis) {
        addSection(group.label, emojis)
      }
    }
  }

  virtualRows.value = rows
  totalHeight.value = y
}

function onScroll(): void {
  renderVisible()
}

function renderVisible(): void {
  const viewport = viewportRef.value
  const content = scrollContentRef.value
  if (!viewport || !content) return

  const rows = virtualRows.value
  content.style.height = totalHeight.value + 'px'

  if (!rows.length) {
    content.innerHTML = ''
    return
  }

  const scrollTop = viewport.scrollTop
  const rangeStart = scrollTop - BUFFER_PX
  const rangeEnd = scrollTop + props.viewportHeight + BUFFER_PX

  let first = findFirstRow(rangeStart)
  let last = findLastRow(rangeEnd)

  if (first === renderedRange.value.start && last === renderedRange.value.end) return
  renderedRange.value = { start: first, end: last }

  const parts: string[] = []
  for (let i = first; i <= last; i++) {
    const row = rows[i]
    if (row.type === 'label') {
      parts.push(`<div class="c-emoji-picker-section-label" style="position:absolute;top:${row.y}px;left:0;right:0;">${esc(row.text || '')}</div>`)
    } else if (row.type === 'emojis') {
      parts.push(`<div class="c-emoji-picker-vrow" style="position:absolute;top:${row.y}px;left:0;right:0;height:${ITEM_SIZE}px;">`)
      for (const e of row.items || []) {
        parts.push(`<span class="epi" data-emoji="${e.name}" title=":${e.name}:">${e.emoji}</span>`)
      }
      parts.push('</div>')
    } else if (row.type === 'empty') {
      parts.push(`<div class="c-emoji-picker-empty" style="position:absolute;top:${row.y}px;left:0;right:0;">${esc(props.labels.noResults || 'No emojis found')}</div>`)
    }
  }

  content.innerHTML = parts.join('')
}

function findFirstRow(y: number): number {
  const rows = virtualRows.value
  let lo = 0
  let hi = rows.length - 1
  while (lo < hi) {
    const mid = (lo + hi) >> 1
    if (rows[mid].y + rows[mid].height <= y) {
      lo = mid + 1
    } else {
      hi = mid
    }
  }
  return lo
}

function findLastRow(y: number): number {
  const rows = virtualRows.value
  let lo = 0
  let hi = rows.length - 1
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1
    if (rows[mid].y >= y) {
      hi = mid - 1
    } else {
      lo = mid
    }
  }
  return Math.min(lo, rows.length - 1)
}

function onQuickReactionClick(emoji: Emoji): void {
  saveFrequentlyUsed(emoji.name)
  emit('select', emoji)
}

function esc(str: string): string {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

async function reset(): Promise<void> {
  search.value = ''
  loadFrequentlyUsed()

  await nextTick()
  if (searchInputRef.value) {
    searchInputRef.value.value = ''
    searchInputRef.value.focus()
  }

  const content = scrollContentRef.value
  if (content && !(content as any)._bound) {
    content.addEventListener('click', onEmojiClick)
    ;(content as any)._bound = true
  }

  rebuildVirtualRows()
  renderedRange.value = { start: -1, end: -1 }
  if (viewportRef.value) {
    viewportRef.value.scrollTop = 0
  }
  renderVisible()
}

defineExpose({ reset })
</script>

<style lang="scss">
.c-emoji-picker-search-wrap {
  position: relative;
  padding: 0.4rem 0.5rem;
}

.c-emoji-picker-search-input {
  width: 100%;
  padding: 0.3rem 0.5rem 0.3rem 1.75rem;
  border: 1px solid var(--extra-light, #ddd);
  border-radius: 0.35rem;
  font-size: 0.8rem;
  outline: none;
  background: var(--white, #fff);
  color: var(--dark, #333);

  &:focus {
    border-color: var(--primary, #4080ff);
    box-shadow: 0 0 0 2px rgba(64, 128, 255, 0.15);
  }

  &::placeholder {
    color: var(--secondary, #999);
  }
}

.c-emoji-picker-search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  width: 0.8rem;
  height: 0.8rem;
  color: var(--secondary, #999);
  pointer-events: none;
}

.c-emoji-picker-viewport {
  overflow-y: auto;
  position: relative;
  padding: 0 0.4rem;
}

.c-emoji-picker-scroll-content {
  position: relative;
  width: 100%;
}

.c-emoji-picker-section-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--secondary, #888);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  padding: 0.3rem 0.15rem 0.15rem;
  background: var(--white, #fff);
  line-height: 1;
  height: 26px;
  display: flex;
  align-items: center;
}

.c-emoji-picker-vrow {
  display: flex;
}

.epi {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  padding: 0;
  margin: 0;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 1.15rem;
  line-height: 1;

  &:hover {
    background-color: var(--light, #f0f0f0);
    transform: scale(1.15);
  }

  &:active {
    transform: scale(0.95);
  }
}

.c-emoji-picker-empty {
  text-align: center;
  color: var(--secondary, #999);
  padding: 1.5rem 0;
  font-size: 0.8rem;
}

.c-emoji-picker-quick {
  border-top: 1px solid var(--extra-light, #e0e0e0);
  padding: 0.25rem 0.4rem 0.3rem;
}

.c-emoji-picker-quick-label {
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--secondary, #888);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  padding: 0.1rem 0.15rem;
}

.c-emoji-picker-quick-row {
  display: flex;
}
</style>
