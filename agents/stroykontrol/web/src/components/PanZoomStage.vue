<script setup>
// Просмотр страницы документа с панорамированием и масштабом для больших чертежей.
// Колесо — масштаб к курсору, перетаскивание — перемещение, щипок на touch-экранах,
// двойной клик — приблизить/вернуть, миникарта при увеличении, полноэкранный режим,
// клавиатура (когда область в фокусе): + − 0 F стрелки PgUp/PgDn Home/End.
// Масштаб 100% = страница по ширине области просмотра.
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  src: { type: String, default: '' },
  page: { type: Number, default: 1 },
  count: { type: Number, default: 1 },
  focusBox: { type: Array, default: null }, // [x0,y0,x1,y1] в долях страницы: показать этот фрагмент после загрузки
  showPageControls: { type: Boolean, default: true },
})
const emit = defineEmits(['update:page', 'failed'])

const MIN_SCALE = 0.1
const MAX_SCALE = 16
const EDGE_MARGIN = 80 // сколько пикселей страницы всегда остаётся в области просмотра
const PAN_STEP = 80

const stage = ref(null)
const canvas = ref(null)
const minimap = ref(null)
const v = reactive({ s: 1, x: 0, y: 0, nw: 0, nh: 0, sw: 0, sh: 0 })
const fitMode = ref('width') // 'width' | 'page' | '' (пользователь изменил вид вручную)
const dragging = ref(false)
const fullscreen = ref(false)

const contentH = computed(() => (v.nw ? (v.sw * v.nh) / v.nw : 0))
const contentStyle = computed(() => ({
  width: v.sw + 'px',
  transform: `translate(${v.x}px, ${v.y}px) scale(${v.s})`,
}))
const percent = computed(() => Math.round(v.s * 100))
const showMap = computed(() => Boolean(props.src) && v.nw > 0 && v.s > 1.05)
const mapRect = computed(() => {
  if (!v.nw || !v.sw || !contentH.value) return {}
  const left = Math.max(0, -v.x / v.s / v.sw)
  const top = Math.max(0, -v.y / v.s / contentH.value)
  const right = Math.min(1, (-v.x + v.sw) / v.s / v.sw)
  const bottom = Math.min(1, (-v.y + v.sh) / v.s / contentH.value)
  return { left: left * 100 + '%', top: top * 100 + '%', width: Math.max(0, right - left) * 100 + '%', height: Math.max(0, bottom - top) * 100 + '%' }
})

function clampPan () {
  const w = v.sw * v.s
  const h = contentH.value * v.s
  v.x = w <= v.sw ? (v.sw - w) / 2 : Math.min(EDGE_MARGIN, Math.max(v.sw - w - EDGE_MARGIN, v.x))
  v.y = h <= v.sh ? (v.sh - h) / 2 : Math.min(EDGE_MARGIN, Math.max(v.sh - h - EDGE_MARGIN, v.y))
}

function measure () {
  if (!stage.value) return
  v.sw = stage.value.clientWidth
  v.sh = stage.value.clientHeight
  if (!props.src) {
    v.nw = v.sw
    v.nh = canvas.value?.scrollHeight || canvas.value?.offsetHeight || 0
  }
}

function zoomAt (factor, cx = v.sw / 2, cy = v.sh / 2) {
  const next = Math.min(MAX_SCALE, Math.max(MIN_SCALE, v.s * factor))
  const ratio = next / v.s
  v.x = cx - (cx - v.x) * ratio
  v.y = cy - (cy - v.y) * ratio
  v.s = next
  fitMode.value = ''
  clampPan()
}

function fitWidth () {
  v.s = 1
  v.x = 0
  v.y = 0
  fitMode.value = 'width'
  clampPan()
}

function fitPage () {
  v.s = contentH.value ? Math.min(1, v.sh / contentH.value) : 1
  fitMode.value = 'page'
  clampPan()
}

function showRegion ([x0, y0, x1, y1]) {
  const bw = Math.max(0.01, x1 - x0) * v.sw
  const bh = Math.max(0.01, y1 - y0) * contentH.value
  v.s = Math.min(8, Math.max(1, Math.min((v.sw * 0.6) / bw, (v.sh * 0.6) / bh)))
  v.x = v.sw / 2 - ((x0 + x1) / 2) * v.sw * v.s
  v.y = v.sh / 2 - ((y0 + y1) / 2) * contentH.value * v.s
  fitMode.value = ''
  clampPan()
}

function toggleZoom (e) {
  const r = stage.value.getBoundingClientRect()
  if (v.s > 1.05) fitWidth()
  else zoomAt(3, e.clientX - r.left, e.clientY - r.top)
}

function onLoad (e) {
  v.nw = e.target.naturalWidth
  v.nh = e.target.naturalHeight
  measure()
  if (props.focusBox) showRegion(props.focusBox)
  else if (fitMode.value === 'width') fitWidth()
  else if (fitMode.value === 'page') fitPage()
  else clampPan()
}

watch(() => props.src, () => { v.nw = 0; v.nh = 0 })

function onWheel (e) {
  e.preventDefault()
  const r = stage.value.getBoundingClientRect()
  if (!e.ctrlKey && (e.shiftKey || Math.abs(e.deltaX) > Math.abs(e.deltaY))) {
    v.x -= e.shiftKey && !e.deltaX ? e.deltaY : e.deltaX
    fitMode.value = ''
    clampPan()
    return
  }
  const speed = e.ctrlKey ? 0.01 : 0.0015
  zoomAt(Math.exp(-e.deltaY * speed), e.clientX - r.left, e.clientY - r.top)
}

const pointers = new Map()
let pinchDist = 0

function onPointerDown (e) {
  if (e.button > 0) return
  stage.value.focus({ preventScroll: true })
  stage.value.setPointerCapture(e.pointerId)
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY })
  dragging.value = true
  pinchDist = 0
}

function onPointerMove (e) {
  const p = pointers.get(e.pointerId)
  if (!p) return
  const dx = e.clientX - p.x
  const dy = e.clientY - p.y
  p.x = e.clientX
  p.y = e.clientY
  if (pointers.size === 1) {
    v.x += dx
    v.y += dy
    fitMode.value = ''
    clampPan()
    return
  }
  if (pointers.size === 2) {
    const [a, b] = [...pointers.values()]
    const dist = Math.hypot(a.x - b.x, a.y - b.y)
    if (pinchDist) {
      const r = stage.value.getBoundingClientRect()
      zoomAt(dist / pinchDist, (a.x + b.x) / 2 - r.left, (a.y + b.y) / 2 - r.top)
    }
    pinchDist = dist
  }
}

function onPointerUp (e) {
  pointers.delete(e.pointerId)
  pinchDist = 0
  if (!pointers.size) dragging.value = false
}

function goPage (n) {
  const next = Math.min(Math.max(1, n), Math.max(1, props.count))
  if (next !== props.page) emit('update:page', next)
}

function onKey (e) {
  if (e.target.tagName === 'INPUT') return
  const k = e.key
  const pan = (dx, dy) => { v.x += dx; v.y += dy; fitMode.value = ''; clampPan() }
  if (k === '+' || k === '=') zoomAt(1.25)
  else if (k === '-' || k === '_') zoomAt(0.8)
  else if (k === '0') fitWidth()
  else if (k === 'f' || k === 'F' || k === 'а' || k === 'А') fitPage()
  else if (k === 'ArrowLeft') pan(PAN_STEP, 0)
  else if (k === 'ArrowRight') pan(-PAN_STEP, 0)
  else if (k === 'ArrowUp') pan(0, PAN_STEP)
  else if (k === 'ArrowDown') pan(0, -PAN_STEP)
  else if (k === 'PageUp') goPage(props.page - 1)
  else if (k === 'PageDown') goPage(props.page + 1)
  else if (k === 'Home') goPage(1)
  else if (k === 'End') goPage(props.count)
  else return
  e.preventDefault()
}

function mapMove (e) {
  const r = minimap.value.getBoundingClientRect()
  const fx = Math.min(1, Math.max(0, (e.clientX - r.left) / r.width))
  const fy = Math.min(1, Math.max(0, (e.clientY - r.top) / r.height))
  v.x = v.sw / 2 - fx * v.sw * v.s
  v.y = v.sh / 2 - fy * contentH.value * v.s
  fitMode.value = ''
  clampPan()
}
function mapDown (e) { minimap.value.setPointerCapture(e.pointerId); mapMove(e) }
function mapDrag (e) { if (e.buttons & 1) mapMove(e) }

function toggleFullscreen () {
  if (document.fullscreenElement) document.exitFullscreen()
  else stage.value.requestFullscreen?.()
}
function onFullscreenChange () {
  fullscreen.value = document.fullscreenElement === stage.value
}

let observer = null
onMounted(() => {
  measure()
  stage.value.addEventListener('wheel', onWheel, { passive: false })
  document.addEventListener('fullscreenchange', onFullscreenChange)
  observer = new ResizeObserver(() => {
    measure()
    if (fitMode.value === 'page') fitPage()
    else if (fitMode.value === 'width') fitWidth()
    else clampPan()
  })
  observer.observe(stage.value)
  if (canvas.value) observer.observe(canvas.value)
})
onBeforeUnmount(() => {
  stage.value?.removeEventListener('wheel', onWheel)
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  observer?.disconnect()
})
</script>

<template>
  <div
    ref="stage" class="pz-stage" :class="{ dragging, fullscreen }" tabindex="0"
    @pointerdown="onPointerDown" @pointermove="onPointerMove" @pointerup="onPointerUp" @pointercancel="onPointerUp"
    @dblclick="toggleZoom" @keydown="onKey"
  >
    <div ref="canvas" class="pz-content" :style="contentStyle">
      <img v-if="src" :src="src" alt="страница" draggable="false" @load="onLoad" @error="emit('failed')">
      <slot v-else />
    </div>

    <div class="pz-toolbar" @pointerdown.stop @dblclick.stop @wheel.stop>
      <template v-if="showPageControls">
        <button :disabled="page <= 1" title="Предыдущая страница (PgUp)" @click="goPage(page - 1)">‹</button>
        <span class="pz-page">
          <input type="number" min="1" :max="count" :value="page" title="Номер страницы" @change="e => goPage(Number(e.target.value))"> / {{ count }}
        </span>
        <button :disabled="page >= count" title="Следующая страница (PgDn)" @click="goPage(page + 1)">›</button>
      </template>
      <span class="pz-sep"></span>
      <button title="Уменьшить (−)" @click="zoomAt(0.8)">−</button>
      <span class="pz-zoom">{{ percent }}%</span>
      <button title="Увеличить (+)" @click="zoomAt(1.25)">+</button>
      <button :class="{ on: fitMode === 'width' }" title="По ширине (0)" @click="fitWidth">По ширине</button>
      <button :class="{ on: fitMode === 'page' }" title="Страница целиком (F)" @click="fitPage">Целиком</button>
      <button :title="fullscreen ? 'Выйти из полноэкранного режима' : 'На весь экран'" @click="toggleFullscreen">{{ fullscreen ? '✕' : '⛶' }}</button>
    </div>

    <div v-if="showMap" ref="minimap" class="pz-map" :style="{ aspectRatio: v.nw + ' / ' + v.nh }" title="Обзор: клик или перетаскивание — перейти к месту" @pointerdown.stop="mapDown" @pointermove.stop="mapDrag" @dblclick.stop>
      <img :src="src" alt="" draggable="false">
      <div class="pz-map-view" :style="mapRect"></div>
    </div>

    <div class="pz-hint">Колесо — масштаб · перетаскивание — перемещение · двойной клик — приблизить</div>
  </div>
</template>

<style scoped>
.pz-stage { position: relative; overflow: hidden; height: 85vh; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--panel); cursor: grab; touch-action: none; user-select: none; outline: none; }
.pz-stage:focus-visible { border-color: var(--accent); }
.pz-stage.dragging { cursor: grabbing; }
.pz-stage.fullscreen { height: 100%; border-radius: 0; }
.pz-content { position: absolute; left: 0; top: 0; transform-origin: 0 0; will-change: transform; }
.pz-content > img { display: block; width: 100%; max-width: none; pointer-events: none; }
.pz-toolbar { position: absolute; top: 8px; left: 50%; transform: translateX(-50%); display: flex; align-items: center; gap: 4px; padding: 4px 6px; background: var(--panel); border: 1px solid var(--border); border-radius: var(--radius-lg); box-shadow: 0 2px 8px rgba(0, 0, 0, .18); font-size: 12px; cursor: default; white-space: nowrap; max-width: calc(100% - 16px); overflow-x: auto; }
.pz-toolbar button { border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: var(--radius); padding: 2px 9px; font-size: 12px; cursor: pointer; }
.pz-toolbar button.on { border-color: var(--accent); color: var(--accent); }
.pz-toolbar button:disabled { opacity: .5; cursor: default; }
.pz-toolbar input { width: 52px; border: 1px solid var(--border); background: var(--bg); color: var(--fg); border-radius: var(--radius); padding: 1px 4px; }
.pz-sep { width: 1px; align-self: stretch; background: var(--border); margin: 0 4px; }
.pz-zoom { min-width: 42px; text-align: center; }
.pz-map { position: absolute; right: 10px; bottom: 10px; width: 160px; max-height: 40%; border: 1px solid var(--border); background: var(--bg); box-shadow: 0 2px 8px rgba(0, 0, 0, .25); cursor: pointer; overflow: hidden; }
.pz-map img { display: block; width: 100%; height: 100%; object-fit: fill; opacity: .75; pointer-events: none; }
.pz-map-view { position: absolute; border: 2px solid #e8590c; background: rgba(232, 89, 12, .15); pointer-events: none; box-sizing: border-box; }
.pz-hint { position: absolute; left: 10px; bottom: 8px; font-size: 11px; color: var(--muted); background: var(--panel); padding: 2px 6px; border-radius: var(--radius); opacity: .8; pointer-events: none; }
@media (max-width: 700px) { .pz-hint { display: none; } }
</style>
