<script setup>
import { computed } from 'vue'
import { STATUS_LABEL, STATUS_CLASS, STATUS_ICON } from '../constants.js'
import Icon from './Icon.vue'

const props = defineProps({ comparison: Object, recordId: String })

const statusCls = computed(() => STATUS_CLASS[props.comparison.status] || '')
const spinning = computed(() => props.comparison.status === 'processing')
const title = computed(() => props.comparison.title || ('Сравнение #' + props.recordId))
const pct = computed(() => Math.max(0, Math.min(100, Number(props.comparison.similarityPercent) || 0)))
const barColor = computed(() => pct.value >= 80 ? 'var(--ok)' : pct.value >= 50 ? 'var(--warn)' : 'var(--bad)')
</script>

<template>
  <div class="head">
    <h1>{{ title }}</h1>
    <span class="badge" :class="statusCls">
      <Icon :svg="STATUS_ICON[comparison.status] || '•'" :cls="spinning ? 'spin' : ''" />
      {{ STATUS_LABEL[comparison.status] || comparison.status || '—' }}
    </span>
  </div>

  <div v-if="comparison.pdFileName || comparison.rdFileName" class="files">
    <span v-if="comparison.pdFileName">ПД: <b>{{ comparison.pdFileName }}</b></span>
    <span v-if="comparison.rdFileName">РД: <b>{{ comparison.rdFileName }}</b></span>
  </div>

  <div v-if="comparison.similarityPercent" class="simbar-row">
    <span class="simlabel">Схожесть</span>
    <div class="simbar"><div :style="{ width: pct + '%', background: barColor }"></div></div>
    <span class="simlabel">{{ pct }}%</span>
  </div>

  <div v-if="comparison.comment" class="comment">{{ comparison.comment }}</div>
</template>
