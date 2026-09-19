<template>
  <Wrap v-bind="$props">
    <div class="risk-block h-100 p-3" :style="blockStyle">
      <div v-if="!bindingID" class="text-secondary small">
        Блок не настроен: выберите привязку модели риска в настройках блока.
      </div>

      <div v-else-if="loading" class="d-flex align-items-center justify-content-center h-100">
        <span class="spinner-border spinner-border-sm" />
      </div>

      <!-- record mode: current record's latest assessment -->
      <template v-else-if="mode === 'record'">
        <div v-if="!recordID" class="text-secondary small">
          Сохраните запись, чтобы увидеть оценку риска.
        </div>
        <div v-else-if="!assessment" class="text-secondary small">
          Для этой записи ещё нет оценки риска.
        </div>
        <div v-else>
          <div class="d-flex align-items-center gap-3 flex-wrap mb-2">
            <svg
              viewBox="0 0 200 105"
              class="risk-gauge"
              role="img"
              :aria-label="`Остаточный риск ${round1(assessment.residualScore)} из 100, уровень ${assessment.level}`"
            >
              <path d="M20,95 A80,80 0 0,1 43.43,38.43" class="risk-gauge-zone" stroke="#28a745" />
              <path d="M43.43,38.43 A80,80 0 0,1 100,15" class="risk-gauge-zone" stroke="#ffc107" />
              <path d="M100,15 A80,80 0 0,1 156.57,38.43" class="risk-gauge-zone" stroke="#fd7e14" />
              <path d="M156.57,38.43 A80,80 0 0,1 180,95" class="risk-gauge-zone" stroke="#dc3545" />
              <line :x1="100" :y1="95" :x2="needleTip.x" :y2="needleTip.y" class="risk-gauge-needle" />
              <circle cx="100" cy="95" r="6" class="risk-gauge-pivot" />
            </svg>
            <div class="flex-grow-1" style="min-width:8rem">
              <div class="d-flex align-items-baseline gap-2 flex-wrap">
                <span class="risk-score">{{ round1(assessment.residualScore) }}</span>
                <span class="badge" :class="levelBadgeClass(assessment.level)">{{ assessment.level }}</span>
              </div>
              <div class="text-secondary small">
                inherent {{ round1(assessment.inherentScore) }} · {{ formatDate(assessment.assessedAt) }}
              </div>
            </div>
          </div>
          <p v-if="showExplanation && assessment.explanation" class="small mb-2">
            {{ assessment.explanation }}
          </p>
          <div v-if="showAttribution && assessment.attribution">
            <div
              v-for="c in sortedContributions(assessment.attribution.contributions)"
              :key="c.nodeID"
              class="mb-1"
            >
              <div class="d-flex justify-content-between small">
                <span>{{ c.factorHandle }}</span>
                <span>{{ c.value >= 0 ? '+' : '' }}{{ round1(c.value) }}</span>
              </div>
              <div class="progress" style="height:5px">
                <div
                  class="progress-bar"
                  :class="c.value >= 0 ? 'bg-primary' : 'bg-success'"
                  :style="{ width: barWidth(c.value) + '%' }"
                />
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- portfolio mode: ranked list + level distribution across every subject -->
      <template v-else>
        <div v-if="!portfolio.length" class="text-secondary small">
          По этой привязке ещё нет ни одной оценки.
        </div>
        <div v-else>
          <div class="d-flex gap-2 mb-3 flex-wrap">
            <span
              v-for="lvl in ['critical', 'high', 'medium', 'low']"
              :key="lvl"
              class="badge"
              :class="levelBadgeClass(lvl)"
            >
              {{ lvl }}: {{ levelCounts[lvl] || 0 }}
            </span>
          </div>
          <table class="table table-sm mb-0">
            <thead>
              <tr>
                <th>Объект</th>
                <th class="text-end">Residual</th>
                <th>Уровень</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="a in topRows" :key="a.subjectRecordID">
                <td>#{{ a.subjectRecordID }}</td>
                <td class="text-end">{{ round1(a.residualScore) }}</td>
                <td>
                  <span class="badge" :class="levelBadgeClass(a.level)">{{ a.level }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </Wrap>
</template>

<script setup>
import { ref, computed, watch, onMounted, inject } from 'vue'
import { useRoute } from 'vue-router'
import { NoID } from 'corteza-lib/js/dist'
import Wrap from './Wrap/index.js'

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, default: undefined },
  record: { type: Object, default: undefined },
  mode: { type: String, default: '' },
  editable: { type: Boolean, default: false },
  resizing: { type: Boolean, default: false },
  magnified: { type: Boolean, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, default: false },
  errors: { type: Object, default: () => ({}) },
})

const $ComposeAPI = inject('$ComposeAPI', window.__composeAPI)
const route = useRoute()

const bindingID = computed(() => props.block.options?.bindingID || '')
const mode = computed(() => props.block.options?.mode || 'record')
const showAttribution = computed(() => props.block.options?.showAttribution !== false)
const showExplanation = computed(() => props.block.options?.showExplanation !== false)
const topN = computed(() => props.block.options?.topN || 10)

function asRecordID (id) {
  if (id == null || id === '' || id === NoID || id === '0' || id === 0) return ''
  return String(id)
}

// Vue Router reuses this component instance across /record/:recordID ->
// /record/:otherRecordID (same route, different params), so props.record
// isn't guaranteed to have been refreshed to the new record yet when this
// recomputes. route.params updates synchronously with every in-page
// navigation — including the prev/next "< >" record switcher, which itself
// calls router.push({ params: { ...route.params, recordID } }) — so it goes
// first; props.record is only the fallback, for contexts where the URL
// doesn't carry the id at all (e.g. a modal record viewer that swaps
// records via a plain prop instead of a route change).
const recordID = computed(() =>
  asRecordID(route.params?.recordID) || asRecordID(route.query?.recordID) || asRecordID(props.record?.recordID),
)

const loading = ref(false)
const assessment = ref(null)
const portfolio = ref([])

const levelCounts = computed(() => {
  const out = {}
  for (const a of portfolio.value) out[a.level] = (out[a.level] || 0) + 1
  return out
})
const topRows = computed(() => [...portfolio.value].sort((a, b) => b.residualScore - a.residualScore).slice(0, topN.value))

function levelBadgeClass (level) {
  return { low: 'bg-success', medium: 'bg-warning text-dark', high: 'bg-orange text-white', critical: 'bg-danger' }[level] || 'bg-secondary'
}
function severityColor (level) {
  return { low: '#28a745', medium: '#ffc107', high: '#fd7e14', critical: '#dc3545' }[level] || '#adb5bd'
}
// Left-edge stripe on the whole block — an ambient signal readable even when
// the block is scrolled past or collapsed, on top of the gauge/badge detail.
// No overflow here — Wrap's own card-body is already the single
// overflow-auto owner (see Wrap/Card.vue); adding a second one here nests
// two scrollbars for the same content.
const blockStyle = computed(() => {
  if (mode.value !== 'record' || !assessment.value) return {}
  return { borderLeft: `4px solid ${severityColor(assessment.value.level)}` }
})
// Gauge needle: a 0..100 score sweeps the semicircle from 180° (left, low)
// through 90° (top) to 0° (right, critical) — see the four static zone arcs
// in the template, which use the same four generic thresholds (25/50/75)
// this maps against. A model's own calibrated bands can differ slightly;
// the needle still lands in the right zone for any reasonable calibration.
const needleTip = computed(() => {
  const score = Math.min(100, Math.max(0, assessment.value?.residualScore || 0))
  const deg = 180 - (score / 100) * 180
  const rad = (deg * Math.PI) / 180
  const len = 68
  return { x: 100 + len * Math.cos(rad), y: 95 - len * Math.sin(rad) }
})
function round1 (v) {
  return Math.round((v || 0) * 10) / 10
}
function barWidth (v) {
  return Math.min(100, Math.abs(v) * 2)
}
function formatDate (v) {
  return v ? new Date(v).toLocaleString() : ''
}
function sortedContributions (list) {
  return [...(list || [])].sort((a, b) => Math.abs(b.value) - Math.abs(a.value))
}

async function load () {
  if (!bindingID.value || !$ComposeAPI) return
  loading.value = true
  try {
    if (mode.value === 'record') {
      assessment.value = null
      if (!recordID.value) return
      const { assessments } = await $ComposeAPI.riskAssessmentList({
        bindingID: bindingID.value,
        subjectRecordID: recordID.value,
        latestPerSubject: 1,
      })
      assessment.value = (assessments || [])[0] || null
    } else {
      const { assessments } = await $ComposeAPI.riskAssessmentList({
        bindingID: bindingID.value,
        latestPerSubject: 1,
      })
      portfolio.value = assessments || []
    }
  } catch {
    assessment.value = null
    portfolio.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([bindingID, mode, recordID], load)
// Belt-and-suspenders: route.fullPath is a plain string, so this fires on
// every navigation unambiguously — a direct backstop in case recordID's
// dependency on route.params (through optional chaining, computed) ever
// misses an update for a reason the [bindingID, mode, recordID] watch above
// doesn't catch (e.g. the prev/next "< >" switcher, which pushes only new
// params on the current route).
watch(() => route.fullPath, load)
</script>

<style scoped>
.risk-block .badge.bg-orange {
  background-color: #fd7e14;
}
.risk-gauge {
  width: 6.5rem;
  flex: 0 0 auto;
}
.risk-gauge-zone {
  fill: none;
  stroke-width: 14;
}
.risk-gauge-needle {
  stroke: #212529;
  stroke-width: 3;
  stroke-linecap: round;
}
.risk-gauge-pivot {
  fill: #212529;
}
.risk-score {
  font-size: 1.75rem;
  font-weight: 700;
  line-height: 1;
}
</style>
