<script setup>
// Dynamics of the corrections (ТЗ 3.9.4): one stacked column per day of the registry's snapshots,
// split by working status; the line is the overdue remarks.
import { computed } from 'vue'
import { STATUS_LABEL, STATUS_ORDER, fmtDate } from './labels.js'

const props = defineProps({ dynamics: { type: Array, default: () => [] } })

const COLORS = {
  awaiting_answer: '#8b949e',
  being_fixed: '#d29922',
  answer_review: '#2e59d9',
  rework: '#cf222e',
  resolved: '#1a7f37',
}
const W = 640
const H = 170
const PAD = { l: 34, r: 8, t: 8, b: 22 }

const series = computed(() => props.dynamics.slice(-60))
const max = computed(() => Math.max(1, ...series.value.map(s => s.total)))
const step = computed(() => (W - PAD.l - PAD.r) / Math.max(series.value.length, 1))
const barW = computed(() => Math.min(40, step.value * 0.7))
const y = (n) => PAD.t + (H - PAD.t - PAD.b) * (1 - n / max.value)

const bars = computed(() => series.value.map((s, i) => {
  let acc = 0
  const x = PAD.l + step.value * i + (step.value - barW.value) / 2
  const parts = STATUS_ORDER.map((code) => {
    const n = s.byStatus?.[code] || 0
    const top = y(acc + n)
    const h = y(acc) - top
    acc += n
    return { code, n, x, y: top, h }
  }).filter(p => p.n > 0)
  return { s, x, parts, cx: x + barW.value / 2, overdueY: y(s.overdue) }
}))
const ticks = computed(() => [0, Math.ceil(max.value / 2), max.value].filter((v, i, a) => a.indexOf(v) === i))
const label = (i) => series.value.length <= 8 || i % Math.ceil(series.value.length / 8) === 0
</script>

<template>
  <div class="rm-dyn">
    <div v-if="series.length < 2" class="muted">
      Динамика копится по дням: срез реестра пишется при каждой загрузке выгрузки, проверке по документам и смене статуса.
      <template v-if="series.length === 1">Пока есть один срез — {{ fmtDate(series[0].date) }}: замечаний {{ series[0].total }}, просрочено {{ series[0].overdue }}.</template>
    </div>
    <template v-else>
      <svg :viewBox="`0 0 ${W} ${H}`" class="rm-dyn-svg" role="img" aria-label="Динамика устранения замечаний по дням">
        <g v-for="t in ticks" :key="t">
          <line :x1="PAD.l" :x2="W - PAD.r" :y1="y(t)" :y2="y(t)" class="rm-dyn-grid" />
          <text :x="PAD.l - 5" :y="y(t) + 3" text-anchor="end" class="rm-dyn-t">{{ t }}</text>
        </g>
        <g v-for="(b, i) in bars" :key="b.s.date">
          <rect v-for="p in b.parts" :key="p.code" :x="b.x" :y="p.y" :width="barW" :height="p.h" :fill="COLORS[p.code]">
            <title>{{ b.s.date }} · {{ STATUS_LABEL[p.code] }}: {{ p.n }}</title>
          </rect>
          <text v-if="label(i)" :x="b.cx" :y="H - 6" text-anchor="middle" class="rm-dyn-t">{{ b.s.date.slice(5) }}</text>
        </g>
        <polyline :points="bars.map(b => `${b.cx},${b.overdueY}`).join(' ')" class="rm-dyn-overdue" />
        <circle v-for="b in bars" :key="'o' + b.s.date" :cx="b.cx" :cy="b.overdueY" r="2.5" class="rm-dyn-overdue-dot"><title>{{ b.s.date }} · просрочено: {{ b.s.overdue }}</title></circle>
      </svg>
      <div class="rm-dyn-legend">
        <span v-for="code in STATUS_ORDER" :key="code"><i :style="{ background: COLORS[code] }"></i>{{ STATUS_LABEL[code] }}</span>
        <span><i class="line"></i>просрочено</span>
      </div>
    </template>
  </div>
</template>
